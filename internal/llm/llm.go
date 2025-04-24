package llm

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"gameslabor/internal/env"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type LLM struct {
	ctx               context.Context
	client            *genai.Client
	model             *genai.GenerativeModel
	EventPlan         []string
	EventLongHistory  []string
	EventShortHistory []string
	CharacterData     map[string][]string
	PlaceData         map[string][]string
	GroupData         map[string][]string
}

type (
	// CharacterData corresponds to the object type within the character_data array.
	CharacterData struct {
		CharacterName string `json:"character"`
		Data          string `json:"data"`
	}

	PlaceData struct {
		PlaceName string `json:"place"`
		Data      string `json:"data"`
	}

	GroupData struct {
		GroupName string `json:"group"`
		Data      string `json:"data"`
	}

	// RollDice corresponds to the nested object within the roll_dice result.
	RollDice struct {
		Difficulty int `json:"difficulty"`
	}

	// ResponseSchema corresponds to the top-level object schema.
	ResponseSchema struct {
		NarratorText      string          `json:"narrator_text"`
		EventPlan         []string        `json:"event_plan"`
		EventLongHistory  []string        `json:"event_long_history"`
		EventShortHistory []string        `json:"event_short_history"`
		CharacterData     []CharacterData `json:"character_data"`
		PlaceData         []PlaceData     `json:"place_data"`
		GroupData         []GroupData     `json:"group_data"`
		RollDice          *RollDice       `json:"roll_dice"`
	}

	PromptDataSchema struct {
		EventPlan         []string            `json:"event_plan"`
		EventLongHistory  []string            `json:"event_long_history"`
		EventShortHistory []string            `json:"event_short_history"`
		CharacterData     map[string][]string `json:"character_data"`
		PlaceData         map[string][]string `json:"place_data"`
		GroupData         map[string][]string `json:"group_data"`
	}
)

var (
	//go:embed system.txt
	systemInstructionTxt string
	//go:embed start.txt
	startPromptTxt string
)

func New(ctx context.Context) (*LLM, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(env.GOOGLE_AI_API_KEY))
	if err != nil {
		return nil, errors.Join(errors.New("failed to create gemini client"), err)
	}
	model := client.GenerativeModel("gemini-2.0-flash")
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type:     genai.TypeObject,
		Required: []string{"narrator_text", "event_plan", "event_long_history", "event_short_history", "character_data", "place_data", "group_data", "roll_dice"},
		Properties: map[string]*genai.Schema{
			"narrator_text": {
				Type: genai.TypeString,
			},
			"place": {
				Type: genai.TypeString,
			},
			"event_plan": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"event_long_history": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"event_short_history": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"character_data": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"character", "data"},
					Properties: map[string]*genai.Schema{
						"character": {
							Type: genai.TypeString,
						},
						"data": {
							Type: genai.TypeString,
						},
					},
				},
			},
			"place_data": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"place", "data"},
					Properties: map[string]*genai.Schema{
						"place": {
							Type: genai.TypeString,
						},
						"data": {
							Type: genai.TypeString,
						},
					},
				},
			},
			"group_data": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type:     genai.TypeObject,
					Required: []string{"group", "data"},
					Properties: map[string]*genai.Schema{
						"group": {
							Type: genai.TypeString,
						},
						"data": {
							Type: genai.TypeString,
						},
					},
				},
			},
			"roll_dice": {
				Type:     genai.TypeObject,
				Nullable: true,
				Required: []string{"difficulty"},
				Properties: map[string]*genai.Schema{
					"difficulty": {
						Type: genai.TypeInteger,
					},
				},
			},
		},
	}
	model.SystemInstruction = genai.NewUserContent(
		genai.Text(fmt.Sprintf(
			systemInstructionTxt,
			TranslateGenaiSchemaToJSONSchema(model.ResponseSchema),
			"Elbjorn",
		)))

	return &LLM{
		ctx:    ctx,
		client: client,
		model:  model,
	}, nil
}

func (llm *LLM) Close() error {
	return llm.client.Close()
}

func (llm *LLM) Data() genai.Text {
	sb := strings.Builder{}
	data := PromptDataSchema{
		EventPlan:         llm.EventPlan,
		EventLongHistory:  llm.EventLongHistory,
		EventShortHistory: llm.EventShortHistory,
		CharacterData:     llm.CharacterData,
		PlaceData:         llm.PlaceData,
		GroupData:         llm.GroupData,
	}
	je := json.NewEncoder(&sb)
	je.Encode(data)

	return genai.Text(sb.String())
}

func (llm *LLM) Start(scenario string) ResponseSchema {
	llm.EventPlan = append(llm.EventPlan, scenario)
	return llm.Text(fmt.Sprintf(startPromptTxt, scenario))
}

func (llm *LLM) Text(text string) ResponseSchema {
	respIter := llm.model.GenerateContentStream(llm.ctx, genai.Text(text))
	restReader := NewGenAIResponseReader(respIter)
	jd := json.NewDecoder(restReader)
	resp := ResponseSchema{}
	jd.Decode(&resp)

	llm.applyResponse(resp)

	return resp
}

func appendTime(s string) string {
	return fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), s)
}

func (llm *LLM) applyResponse(resp ResponseSchema) {
	if resp.CharacterData != nil {
		for i, characterData := range resp.CharacterData {
			if llm.CharacterData == nil {
				llm.CharacterData = make(map[string][]string)
			}
			if i == 0 {
				llm.CharacterData[characterData.CharacterName] = append(llm.CharacterData[characterData.CharacterName], appendTime(characterData.Data))
			} else {
				llm.CharacterData[characterData.CharacterName] = append(llm.CharacterData[characterData.CharacterName], characterData.Data)
			}
		}
	}
	if resp.PlaceData != nil {
		for i, placeData := range resp.PlaceData {
			if llm.PlaceData == nil {
				llm.PlaceData = make(map[string][]string)
			}
			if i == 0 {
				llm.PlaceData[placeData.PlaceName] = append(llm.PlaceData[placeData.PlaceName], appendTime(placeData.Data))
			} else {
				llm.PlaceData[placeData.PlaceName] = append(llm.PlaceData[placeData.PlaceName], placeData.Data)
			}
		}
	}
	if resp.GroupData != nil {
		for i, groupData := range resp.GroupData {
			if llm.GroupData == nil {
				llm.GroupData = make(map[string][]string)
			}
			if i == 0 {
				llm.GroupData[groupData.GroupName] = append(llm.GroupData[groupData.GroupName], appendTime(groupData.Data))
			} else {
				llm.GroupData[groupData.GroupName] = append(llm.GroupData[groupData.GroupName], groupData.Data)
			}
		}
	}
	if resp.EventLongHistory != nil {
		for i, eventLongHistory := range resp.EventLongHistory {
			if i == 0 {
				llm.EventLongHistory = append(llm.EventLongHistory, appendTime(eventLongHistory))
			} else {
				llm.EventLongHistory = append(llm.EventLongHistory, eventLongHistory)
			}
		}
	}
	if resp.EventShortHistory != nil {
		for i, eventShortHistory := range resp.EventShortHistory {
			if i == 0 {
				llm.EventShortHistory = append(llm.EventShortHistory, appendTime(eventShortHistory))
			} else {
				llm.EventShortHistory = append(llm.EventShortHistory, eventShortHistory)
			}
		}
	}
	if resp.EventPlan != nil {
		for i, eventPlan := range resp.EventPlan {
			if i == 0 {
				llm.EventPlan = append(llm.EventPlan, appendTime(eventPlan))
			} else {
				llm.EventPlan = append(llm.EventPlan, eventPlan)
			}
		}
	}
}

func (resp ResponseSchema) JSON() string {
	sb := strings.Builder{}
	jd := json.NewEncoder(&sb)
	jd.SetIndent("", "  ")
	jd.Encode(resp)

	return sb.String()
}
