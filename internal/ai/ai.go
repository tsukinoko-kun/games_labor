package ai

import (
	"context"
	"errors"
	"gameslabor/internal/env"

	tts "cloud.google.com/go/texttospeech/apiv1"
	"google.golang.org/api/option"
	"google.golang.org/genai"
)

type AI struct {
	ctx               context.Context     `json:"-"`
	llmClient         *genai.Client       `json:"-"`
	ttsClient         *tts.Client         `json:"-"`
	EventPlan         []string            `json:"event_plan"`
	EventLongHistory  []string            `json:"event_long_history"`
	EventShortHistory []string            `json:"event_short_history"`
	ChatHistory       []ChatMessage       `json:"chat_history"`
	EntityData        map[string][]string `json:"entity_data"`
}

var (
	emptyStrSlice    = make([]string, 0)
	emptyChatHistory = make([]ChatMessage, 0)
	emptyEntityData  = make(map[string][]string)
)

func Empty() *AI {
	return &AI{
		EventPlan:         emptyStrSlice,
		EventLongHistory:  emptyStrSlice,
		EventShortHistory: emptyStrSlice,
		ChatHistory:       emptyChatHistory,
		EntityData:        emptyEntityData,
	}
}

func New(ctx context.Context) (*AI, error) {
	llmClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  env.GOOGLE_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, errors.Join(errors.New("failed to create gemini client"), err)
	}
	ttsClient, err := tts.NewClient(ctx, option.WithAPIKey(env.GOOGLE_API_KEY))
	if err != nil {
		return nil, errors.Join(errors.New("failed to create tts client"), err)
	}

	return &AI{
		ctx:               ctx,
		llmClient:         llmClient,
		ttsClient:         ttsClient,
		EventPlan:         make([]string, 0),
		EventLongHistory:  make([]string, 0),
		EventShortHistory: make([]string, 0),
		ChatHistory:       make([]ChatMessage, 0),
		EntityData:        make(map[string][]string),
	}, nil
}
