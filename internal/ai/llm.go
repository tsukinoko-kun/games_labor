package ai

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/genai"
)

const (
	mainModel     = "gemini-2.0-flash"
	thinkingModel = "gemini-2.5-pro-preview-05-06"
)

type (
	EntityData struct {
		EntityName string `json:"entity"`
		Data       string `json:"data"`
	}

	// RollDice corresponds to the nested object within the roll_dice result.
	RollDice struct {
		Difficulty int `json:"difficulty"`
	}

	// ResponseSchema corresponds to the top-level object schema.
	ResponseSchema struct {
		NarratorText      string       `json:"narrator_text"`
		EventPlan         []string     `json:"event_plan"`
		EventLongHistory  []string     `json:"event_long_history"`
		EventShortHistory []string     `json:"event_short_history"`
		EntityData        []EntityData `json:"entity_data"`
		RollDice          *RollDice    `json:"roll_dice"`
	}

	PromptDataSchema struct {
		EventPlan         []string            `json:"event_plan"`
		EventLongHistory  []string            `json:"event_long_history"`
		EventShortHistory []string            `json:"event_short_history"`
		EntityData        map[string][]string `json:"entity_data"`
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
	systemInstructionTxt     string
	systemInstructionContent = genai.Text(systemInstructionTxt)
	//go:embed start.txt
	startPromptTxt string
)

var (
	topP             float32                      = 0.5
	topK             float32                      = 5
	temperature      float32                      = 0.7
	frequencyPenalty float32                      = 0.5
	presencePenalty  float32                      = 0.5
	mainConfig       *genai.GenerateContentConfig = &genai.GenerateContentConfig{
		TopP:             &topP,
		TopK:             &topK,
		Temperature:      &temperature,
		FrequencyPenalty: &frequencyPenalty,
		PresencePenalty:  &presencePenalty,
		ResponseMIMEType: "application/json",
		ResponseSchema:   llmResponseGenaiSchema,
	}
	thinkingConfig *genai.GenerateContentConfig = &genai.GenerateContentConfig{
		TopP:             &topP,
		TopK:             &topK,
		Temperature:      &temperature,
		ResponseMIMEType: "application/json",
		ResponseSchema:   llmResponseGenaiSchema,
	}
)

const maxRecentChatHistory = 10

func (llm *AI) Data() []*genai.Content {
	sb := strings.Builder{}
	sb.WriteString("Aktuelle Spieldaten: ")

	data := PromptDataSchema{
		llm.EventPlan,
		llm.EventLongHistory,
		llm.EventShortHistory,
		llm.EntityData,
		nil,
	}
	if len(llm.ChatHistory) > maxRecentChatHistory {
		data.RecentChatHistory = llm.ChatHistory[len(llm.ChatHistory)-maxRecentChatHistory:]
	} else {
		data.RecentChatHistory = llm.ChatHistory
	}
	je := json.NewEncoder(&sb)
	je.Encode(data)

	return genai.Text(sb.String())
}

func (llm *AI) Start(scenario string) ResponseSchema {
	fmt.Println("Starting scenario:", scenario)
	return llm.Text(false, systemInstructionContent, llm.Data(), genai.Text(fmt.Sprintf(startPromptTxt, scenario)))
}

func (llm *AI) Continue(text string) ResponseSchema {
	fmt.Println("Continuing:", text)
	return llm.Text(false, systemInstructionContent, llm.Data(), genai.Text(text))
}

func flatten[T any](slice [][]T) []T {
	l := 0
	for _, item := range slice {
		l += len(item)
	}
	flattened := make([]T, 0, l)
	for _, item := range slice {
		flattened = append(flattened, item...)
	}
	return flattened
}

func (ai *AI) Text(thinking bool, parts ...[]*genai.Content) ResponseSchema {
	var model string
	var config *genai.GenerateContentConfig
	if thinking {
		model = thinkingModel
		config = thinkingConfig
	} else {
		model = mainModel
		config = mainConfig
	}
	resp, err := ai.llmClient.Models.GenerateContent(ai.ctx, model, flatten(parts), config)
	if err != nil {
		fmt.Println("Error generating content:", err)
		return ResponseSchema{}
	}

	sb := strings.Builder{}
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			sb.WriteString(part.Text)
		}
	}

	jd := json.NewDecoder(strings.NewReader(sb.String()))
	respData := ResponseSchema{}
	jd.Decode(&respData)

	ai.applyResponse(respData)

	return respData
}

func appendTime(s string) string {
	return fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), s)
}

func (llm *AI) applyResponse(resp ResponseSchema) {
	if resp.EntityData != nil {
		for i, entityData := range resp.EntityData {
			if llm.EntityData == nil {
				fmt.Sprintln("ai.EntityData should not be nil at this point")
				os.Exit(1)
				llm.EntityData = make(map[string][]string)
			}
			if i == 0 {
				llm.EntityData[entityData.EntityName] = append(llm.EntityData[entityData.EntityName], appendTime(entityData.Data))
			} else {
				llm.EntityData[entityData.EntityName] = append(llm.EntityData[entityData.EntityName], entityData.Data)
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
			AudioEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
			Pitch:         -2.,
			SpeakingRate:  .9,
		},
	}

	ctx := context.Background()
	resp, err := ai.ttsClient.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", errors.Join(errors.New("failed to synthesize speech"), err)
	}
	return saveOgg(resp.GetAudioContent())
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
