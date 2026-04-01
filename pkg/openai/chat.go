package openai

import (
	"context"

	"github.com/kevinxvu/goone/pkg/logging"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"go.uber.org/zap"
)

// Message represents a chat message with optional image URLs
type Message struct {
	Role      string   `json:"role"`       // system, user, or assistant
	Content   string   `json:"content"`    // Text content
	ImageURLs []string `json:"image_urls"` // Optional image URLs for vision models
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages     []Message `json:"messages"`                // List of messages
	Model        *string   `json:"model,omitempty"`         // Optional model override
	MaxTokens    *int64    `json:"max_tokens,omitempty"`    // Optional max tokens override
	Temperature  *float64  `json:"temperature,omitempty"`   // Optional temperature override (0.0 to 2.0)
	SystemPrompt *string   `json:"system_prompt,omitempty"` // Optional system prompt
	TopP         *float64  `json:"top_p,omitempty"`         // Optional nucleus sampling
	Stop         []string  `json:"stop,omitempty"`          // Optional stop sequences
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	Model        string     `json:"model"`
	FinishReason string     `json:"finish_reason"`
	Usage        TokenUsage `json:"usage"`
	CreatedAt    int64      `json:"created_at"`
}

// ChatCompletion performs a chat completion request
func (s *Service) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	logger := logging.Type("openai")

	// Validate request
	if len(req.Messages) == 0 && req.SystemPrompt == nil {
		return nil, ErrNoMessages
	}

	// Build messages
	messages, err := s.buildMessages(req.Messages, req.SystemPrompt)
	if err != nil {
		return nil, err
	}

	// Determine model to use
	model := s.getTextModel(req.Model)

	// Build parameters
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: messages,
	}

	// Apply optional parameters
	if req.MaxTokens != nil {
		params.MaxTokens = param.NewOpt(*req.MaxTokens)
	}

	if req.Temperature != nil {
		params.Temperature = param.NewOpt(*req.Temperature)
	}

	if req.TopP != nil {
		params.TopP = param.NewOpt(*req.TopP)
	}

	// Log request
	logger.Info("creating chat completion",
		zap.String("model", model),
		zap.Int("message_count", len(req.Messages)),
	)

	// Call OpenAI API
	response, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		logger.Error("chat completion failed", zap.Error(err))
		return nil, ErrAPIError.SetInternal(err)
	}

	// Extract response
	if len(response.Choices) == 0 {
		logger.Error("no choices in response")
		return nil, ErrAPIError.SetInternal(nil)
	}

	choice := response.Choices[0]
	content := choice.Message.Content

	// Build response
	chatResp := &ChatResponse{
		ID:           response.ID,
		Content:      content,
		Model:        string(response.Model),
		FinishReason: string(choice.FinishReason),
		CreatedAt:    response.Created,
		Usage: TokenUsage{
			PromptTokens:     int(response.Usage.PromptTokens),
			CompletionTokens: int(response.Usage.CompletionTokens),
			TotalTokens:      int(response.Usage.TotalTokens),
		},
	}

	// Log response
	logger.Info("chat completion succeeded",
		zap.String("id", chatResp.ID),
		zap.Int("prompt_tokens", chatResp.Usage.PromptTokens),
		zap.Int("completion_tokens", chatResp.Usage.CompletionTokens),
		zap.Int("total_tokens", chatResp.Usage.TotalTokens),
	)

	return chatResp, nil
}
