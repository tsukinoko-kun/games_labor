package llm

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"gameslabor/internal/env"
	"strings"

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
}

type (
	// CharacterData corresponds to the object type within the character_data array.
	CharacterData struct {
		CharacterName string `json:"character_name"`
		Data          string `json:"data"`
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
		RollDice          []RollDice      `json:"roll_dice"`
	}

	PromptDataSchema struct {
		EventPlan         []string            `json:"event_plan"`
		EventLongHistory  []string            `json:"event_long_history"`
		EventShortHistory []string            `json:"event_short_history"`
		CharacterData     map[string][]string `json:"character_data"`
	}
)

//go:embed system.txt
var systemInstructionTxt string

var characterData = make(map[string]string)

func New(ctx context.Context) (*LLM, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(env.GOOGLE_AI_API_KEY))
	if err != nil {
		return nil, errors.Join(errors.New("failed to create gemini client"), err)
	}
	model := client.GenerativeModel("gemini-2.0-flash")
	model.ResponseMIMEType = "application/json"
	model.SystemInstruction = genai.NewUserContent(genai.Text(fmt.Sprintf(systemInstructionTxt, "Elbjorn")))
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"narrator_text": {
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
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"character_name": {
							Type: genai.TypeString,
						},
						"data": {
							Type: genai.TypeString,
						},
					},
				},
			},
			"roll_dice": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"result": {
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"difficulty": {
									Type: genai.TypeInteger,
								},
							},
						},
					},
				},
			},
		},
	}
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
	}
	je := json.NewEncoder(&sb)
	je.Encode(data)

	return genai.Text(sb.String())
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

func (llm *LLM) applyResponse(resp ResponseSchema) {
	if resp.CharacterData != nil {
		for _, characterData := range resp.CharacterData {
			llm.CharacterData[characterData.CharacterName] = append(llm.CharacterData[characterData.CharacterName], characterData.Data)
		}
	}
	if resp.EventLongHistory != nil {
		llm.EventLongHistory = append(llm.EventLongHistory, resp.EventLongHistory...)
	}
	if resp.EventShortHistory != nil {
		llm.EventShortHistory = append(llm.EventShortHistory, resp.EventShortHistory...)
	}
	if resp.EventPlan != nil {
		llm.EventPlan = append(llm.EventPlan, resp.EventPlan...)
	}
}
