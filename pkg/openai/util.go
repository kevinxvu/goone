package openai

import (
	"github.com/openai/openai-go"
)

// buildMessages converts custom Message DTOs to OpenAI SDK message types
// Handles text messages, system prompts, and vision messages with images
func (s *Service) buildMessages(messages []Message, systemPrompt *string) ([]openai.ChatCompletionMessageParamUnion, error) {
	var result []openai.ChatCompletionMessageParamUnion

	// Add system prompt if provided
	if systemPrompt != nil && *systemPrompt != "" {
		result = append(result, openai.SystemMessage(*systemPrompt))
	}

	// Convert each message
	for _, msg := range messages {
		// Validate message has content
		if msg.Content == "" && len(msg.ImageURLs) == 0 {
			return nil, ErrEmptyPrompt
		}

		switch msg.Role {
		case "system":
			result = append(result, openai.SystemMessage(msg.Content))

		case "assistant":
			result = append(result, openai.ChatCompletionMessageParamOfAssistant(msg.Content))

		case "user":
			// Check if this is a vision message (has images)
			if len(msg.ImageURLs) > 0 {
				// Build multi-part content with text and images
				parts := make([]openai.ChatCompletionContentPartUnionParam, 0)

				// Add text part if content is not empty
				if msg.Content != "" {
					parts = append(parts, openai.TextContentPart(msg.Content))
				}

				// Add image parts
				for _, imageURL := range msg.ImageURLs {
					imgParam := openai.ChatCompletionContentPartImageImageURLParam{
						URL: imageURL,
					}
					parts = append(parts, openai.ImageContentPart(imgParam))
				}

				result = append(result, openai.UserMessage(parts))
			} else {
				// Simple text message
				result = append(result, openai.UserMessage(msg.Content))
			}

		default:
			// Default to user message for unknown roles
			result = append(result, openai.UserMessage(msg.Content))
		}
	}

	return result, nil
}

// getTextModel returns the model from the request or the configured text/chat default
func (s *Service) getTextModel(model *string) string {
	if model != nil && *model != "" {
		return *model
	}
	if s.cfg.TextModel != "" {
		return s.cfg.TextModel
	}
	return "gpt-4o"
}

// getAudioModel returns the configured audio transcription/translation model default
func (s *Service) getAudioModel(model *string) string {
	if model != nil && *model != "" {
		return *model
	}
	if s.cfg.AudioModel != "" {
		return s.cfg.AudioModel
	}
	return "whisper-1"
}

// Helper functions for creating pointers (useful in DTOs)

// StringPtr returns a pointer to the string value
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr returns a pointer to the int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr returns a pointer to the float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}
