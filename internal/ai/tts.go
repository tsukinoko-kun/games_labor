package ai

import (
	"context"
	"errors"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

func (llm *AI) Close() {
	_ = llm.ttsClient.Close()
}

var (
	ttsVoice = &texttospeechpb.VoiceSelectionParams{
		LanguageCode: "de-DE",
		Name:         "de-DE-Chirp3-HD-Algenib",
	}
	ttsAudioConfig = &texttospeechpb.AudioConfig{
		AudioEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
	}
)

func (ai *AI) TTS(text string) (string, error) {
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice:       ttsVoice,
		AudioConfig: ttsAudioConfig,
	}

	ctx := context.Background()
	resp, err := ai.ttsClient.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", errors.Join(errors.New("failed to synthesize speech"), err)
	}
	return saveOgg(resp.GetAudioContent())
}
