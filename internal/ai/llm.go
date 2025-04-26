package ai

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"gameslabor/internal/env"
	"strings"
	"time"

	tts "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AI struct {
	ctx               context.Context        `json:"-"`
	llmClient         *genai.Client          `json:"-"`
	llmModel          *genai.GenerativeModel `json:"-"`
	ttsClient         *tts.Client            `json:"-"`
	EventPlan         []string               `json:"event_plan"`
	EventLongHistory  []string               `json:"event_long_history"`
	EventShortHistory []string               `json:"event_short_history"`
	ChatHistory       []ChatMessage          `json:"chat_history"`
	CharacterData     map[string][]string    `json:"character_data"`
	PlaceData         map[string][]string    `json:"place_data"`
	GroupData         map[string][]string    `json:"group_data"`
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
		RecentChatHistory []ChatMessage       `json:"recent_chat_history"`
	}

	ChatMessage struct {
		Role     string `json:"role"`
		PlayerID string `json:"player,omitempty"`
		Message  string `json:"message"`
	}
)

var (
	//go:embed system.txt
	systemInstructionTxt string
	//go:embed start.txt
	startPromptTxt string
)

func New(ctx context.Context) (*AI, error) {
	llmClient, err := genai.NewClient(ctx, option.WithAPIKey(env.GOOGLE_API_KEY))
	if err != nil {
		return nil, errors.Join(errors.New("failed to create gemini client"), err)
	}
	llmModel := llmClient.GenerativeModel("gemini-2.0-flash")
	llmModel.ResponseMIMEType = "application/json"
	llmModel.ResponseSchema = llmResponseGenaiSchema
	llmModel.SystemInstruction = genai.NewUserContent(
		genai.Text(fmt.Sprintf(
			systemInstructionTxt,
			TranslateGenaiSchemaToJSONSchema(llmModel.ResponseSchema),
		)))

	ttsClient, err := tts.NewClient(ctx, option.WithAPIKey(env.GOOGLE_API_KEY))
	if err != nil {
		return nil, errors.Join(errors.New("failed to create tts client"), err)
	}

	return &AI{
		ctx:               ctx,
		llmClient:         llmClient,
		llmModel:          llmModel,
		ttsClient:         ttsClient,
		EventPlan:         make([]string, 0),
		EventLongHistory:  make([]string, 0),
		EventShortHistory: make([]string, 0),
		CharacterData:     make(map[string][]string),
		PlaceData:         make(map[string][]string),
		GroupData:         make(map[string][]string),
	}, nil
}

func (llm *AI) Close() {
	_ = llm.llmClient.Close()
	_ = llm.ttsClient.Close()
}

const maxRecentChatHistory = 10

func (llm *AI) Data() genai.Text {
	sb := strings.Builder{}
	sb.WriteString("Aktuelle Daten: ")

	data := PromptDataSchema{
		EventPlan:         llm.EventPlan,
		EventLongHistory:  llm.EventLongHistory,
		EventShortHistory: llm.EventShortHistory,
		CharacterData:     llm.CharacterData,
		PlaceData:         llm.PlaceData,
		GroupData:         llm.GroupData,
	}
	if len(llm.ChatHistory) > maxRecentChatHistory {
		data.RecentChatHistory = llm.ChatHistory[len(llm.ChatHistory)-maxRecentChatHistory:]
	} else {
		data.RecentChatHistory = llm.ChatHistory
	}
	je := json.NewEncoder(&sb)
	je.Encode(data)

	sb.WriteString("\n\nFühre die Geschichte fort.")

	return genai.Text(sb.String())
}

func (llm *AI) Start(scenario string) ResponseSchema {
	fmt.Println("Starting scenario:", scenario)
	return llm.Text(genai.Text(fmt.Sprintf(startPromptTxt, scenario)))
}

func (llm *AI) Continue(text string) ResponseSchema {
	fmt.Println("Continuing:", text)
	return llm.Text(llm.Data(), genai.Text(text))
}

func (llm *AI) Text(parts ...genai.Part) ResponseSchema {
	resp, err := llm.llmModel.GenerateContent(llm.ctx, parts...)
	if err != nil {
		fmt.Println("Error generating content:", err)
		return ResponseSchema{}
	}

	sb := strings.Builder{}
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				sb.WriteString(string(textPart))
			}
		}
	}

	jd := json.NewDecoder(strings.NewReader(sb.String()))
	respData := ResponseSchema{}
	jd.Decode(&respData)

	llm.applyResponse(respData)

	return respData
}

func appendTime(s string) string {
	return fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), s)
}

func (llm *AI) applyResponse(resp ResponseSchema) {
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

func (ai *AI) TTS(text string) (string, error) {
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "de-DE",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			Pitch:         -2.,
			SpeakingRate:  .9,
		},
	}

	ctx := context.Background()
	resp, err := ai.ttsClient.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", errors.Join(errors.New("failed to synthesize speech"), err)
	}
	return saveMp3(resp.GetAudioContent())
}

func (rs *ResponseSchema) JSON() string {
	sb := strings.Builder{}
	je := json.NewEncoder(&sb)
	je.SetIndent("", "  ")
	err := je.Encode(rs)
	if err != nil {
		return err.Error()
	}
	return sb.String()
}
