package openai

import (
	"context"
	"io"

	"github.com/kevinxvu/goone/pkg/logging"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"go.uber.org/zap"
)

// AudioRequest represents an audio transcription request
type AudioRequest struct {
	File     io.Reader `json:"-"`                         // Audio file reader
	FileName string    `json:"file_name"`                 // File name (required for API)
	Language *string   `json:"language,omitempty"`        // Optional language code (ISO-639-1, e.g., "en")
	Prompt   *string   `json:"prompt,omitempty"`          // Optional context to guide transcription
	Format   *string   `json:"response_format,omitempty"` // Optional format (json, text, srt, vtt)
}

// AudioResponse represents an audio transcription response
type AudioResponse struct {
	Text     string   `json:"text"`               // Transcribed text
	Language string   `json:"language,omitempty"` // Detected language (if available)
	Duration *float64 `json:"duration,omitempty"` // Duration in seconds (if available)
}

// TranscribeAudio transcribes audio using the Whisper model
func (s *Service) TranscribeAudio(ctx context.Context, req AudioRequest) (*AudioResponse, error) {
	logger := logging.Type("openai")

	// Validate request
	if req.File == nil {
		return nil, ErrNoFile
	}

	if req.FileName == "" {
		return nil, ErrNoFile.SetInternal(nil)
	}

	// Build parameters
	params := openai.AudioTranscriptionNewParams{
		File:  req.File,
		Model: openai.AudioModel(s.getAudioModel(nil)),
	}

	// Apply optional parameters
	if req.Language != nil && *req.Language != "" {
		params.Language = param.NewOpt(*req.Language)
	}

	if req.Prompt != nil && *req.Prompt != "" {
		params.Prompt = param.NewOpt(*req.Prompt)
	}

	if req.Format != nil && *req.Format != "" {
		params.ResponseFormat = openai.AudioResponseFormat(*req.Format)
	}

	// Log request
	logger.Info("transcribing audio",
		zap.String("file_name", req.FileName),
		zap.Any("language", req.Language),
	)

	// Call OpenAI API
	transcription, err := s.client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		logger.Error("audio transcription failed", zap.Error(err))
		return nil, ErrAPIError.SetInternal(err)
	}

	// Build response
	audioResp := &AudioResponse{
		Text: transcription.Text,
	}

	// Log response
	logger.Info("audio transcription succeeded",
		zap.Int("text_length", len(audioResp.Text)),
	)

	return audioResp, nil
}

// TranslateAudio translates audio to English using the Whisper model
func (s *Service) TranslateAudio(ctx context.Context, req AudioRequest) (*AudioResponse, error) {
	logger := logging.Type("openai")

	// Validate request
	if req.File == nil {
		return nil, ErrNoFile
	}

	if req.FileName == "" {
		return nil, ErrNoFile.SetInternal(nil)
	}

	// Build parameters
	params := openai.AudioTranslationNewParams{
		File:  req.File,
		Model: openai.AudioModel(s.getAudioModel(nil)),
	}

	// Apply optional parameters
	if req.Prompt != nil && *req.Prompt != "" {
		params.Prompt = param.NewOpt(*req.Prompt)
	}

	if req.Format != nil && *req.Format != "" {
		params.ResponseFormat = openai.AudioTranslationNewParamsResponseFormat(*req.Format)
	}

	// Log request
	logger.Info("translating audio",
		zap.String("file_name", req.FileName),
	)

	// Call OpenAI API
	translation, err := s.client.Audio.Translations.New(ctx, params)
	if err != nil {
		logger.Error("audio translation failed", zap.Error(err))
		return nil, ErrAPIError.SetInternal(err)
	}

	// Build response
	audioResp := &AudioResponse{
		Text:     translation.Text,
		Language: "en", // Translations are always to English
	}

	// Log response
	logger.Info("audio translation succeeded",
		zap.Int("text_length", len(audioResp.Text)),
	)

	return audioResp, nil
}
