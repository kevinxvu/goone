# OpenAI Service Package

A reusable OpenAI service package for Go that provides easy access to OpenAI's API, including:

- **Chat Completions** - Text generation with GPT models (including GPT-4 Vision for images)
- **Audio Transcription** - Speech-to-text using Whisper
- **Audio Translation** - Audio translation to English using Whisper
- **Streaming Support** - Real-time streaming responses with token tracking

This package follows the project's AWS service patterns and integrates seamlessly with the existing architecture.

## Installation

The OpenAI SDK is already included in the project dependencies:

```bash
go get github.com/openai/openai-go@latest
```

## Configuration

Configuration is loaded via environment variables with the following priority:
1. OS environment variables (highest priority)
2. `.env.local` (gitignored, for local overrides)
3. `.env` (committed defaults)

### Environment Variables

```env
# Required
OPENAI_API_KEY=sk-...

# Optional (with defaults shown)
OPENAI_BASE_URL=https://api.openai.com/v1  # Custom endpoint for proxies/Azure OpenAI
OPENAI_TIMEOUT=60                          # Request timeout in seconds
OPENAI_MAX_RETRIES=2                       # Number of retry attempts
OPENAI_DEFAULT_MODEL=gpt-4                 # Default model for chat completions
```

### Configuration in Code

The configuration is automatically loaded from `config.Configuration`:

```go
import (
    "github.com/kevinxvu/goone/config"
    "github.com/kevinxvu/goone/pkg/openai"
)

cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Create service
svc := openai.New(openai.Config{
    APIKey:       cfg.OpenAIAPIKey,
    BaseURL:      cfg.OpenAIBaseURL,
    Timeout:      cfg.OpenAITimeout,
    MaxRetries:   cfg.OpenAIMaxRetries,
    DefaultModel: cfg.OpenAIDefaultModel,
})
```

## Usage Examples

### 1. Basic Chat Completion

Simple text-based chat completion:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kevinxvu/goone/pkg/openai"
)

func main() {
    // Initialize service
    svc := openai.New(openai.Config{
        APIKey:       "sk-...",
        DefaultModel: "gpt-4",
    })

    // Create request
    req := openai.ChatRequest{
        Messages: []openai.Message{
            {
                Role:    "user",
                Content: "What is the capital of France?",
            },
        },
    }

    // Get completion
    resp, err := svc.ChatCompletion(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
        resp.Usage.TotalTokens,
        resp.Usage.PromptTokens,
        resp.Usage.CompletionTokens,
    )
}
```

### 2. Chat with System Prompt

Using a system prompt to set behavior:

```go
systemPrompt := "You are a helpful assistant that speaks like a pirate."

req := openai.ChatRequest{
    SystemPrompt: &systemPrompt,
    Messages: []openai.Message{
        {
            Role:    "user",
            Content: "Tell me about treasure hunting.",
        },
    },
}

resp, err := svc.ChatCompletion(context.Background(), req)
```

### 3. Vision - Chat with Images (GPT-4 Vision)

Analyze images using GPT-4 Vision:

```go
req := openai.ChatRequest{
    Model: openai.StringPtr("gpt-4-vision-preview"),
    Messages: []openai.Message{
        {
            Role:    "user",
            Content: "What's in this image?",
            ImageURLs: []string{
                "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg",
            },
        },
    },
    MaxTokens: openai.Int64Ptr(300),
}

resp, err := svc.ChatCompletion(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Image description: %s\n", resp.Content)
```

### 4. Audio Transcription (Whisper)

Transcribe audio files to text:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/kevinxvu/goone/pkg/openai"
)

func main() {
    svc := openai.New(openai.Config{
        APIKey: "sk-...",
    })

    // Open audio file
    file, err := os.Open("audio.mp3")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Create request
    language := "en"
    req := openai.AudioRequest{
        File:     file,
        FileName: "audio.mp3",
        Language: &language, // Optional: helps accuracy
    }

    // Transcribe
    resp, err := svc.TranscribeAudio(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Transcription: %s\n", resp.Text)
    fmt.Printf("Language: %s\n", resp.Language)
}
```

### 5. Audio Translation to English

Translate audio in any language to English:

```go
file, err := os.Open("german_audio.mp3")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

req := openai.AudioRequest{
    File:     file,
    FileName: "german_audio.mp3",
}

// Translate to English
resp, err := svc.TranslateAudio(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("English translation: %s\n", resp.Text)
```

### 6. Streaming Responses

Real-time streaming for long responses:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kevinxvu/goone/pkg/openai"
)

func main() {
    svc := openai.New(openai.Config{
        APIKey: "sk-...",
    })

    req := openai.ChatRequest{
        Messages: []openai.Message{
            {
                Role:    "user",
                Content: "Write a short story about a robot.",
            },
        },
    }

    // Define streaming handler
    handler := func(chunk string, done bool) error {
        if done {
            fmt.Println("\n[Stream complete]")
            return nil
        }
        // Print chunks as they arrive
        fmt.Print(chunk)
        return nil
    }

    // Stream response
    resp, err := svc.ChatCompletionStream(context.Background(), req, handler)
    if err != nil {
        log.Fatal(err)
    }

    // Final response with full content and token usage
    fmt.Printf("\nFull response received (%d tokens)\n", resp.Usage.TotalTokens)
}
```

### 7. Per-Request Configuration Override

Override default settings per request:

```go
temperature := 0.9
maxTokens := int64(500)
topP := 0.95

req := openai.ChatRequest{
    Model:       openai.StringPtr("gpt-3.5-turbo"), // Override default model
    Temperature: &temperature,                       // Higher creativity
    MaxTokens:   &maxTokens,                        // Limit response length
    TopP:        &topP,                             // Nucleus sampling
    Messages: []openai.Message{
        {
            Role:    "user",
            Content: "Write a creative poem.",
        },
    },
}

resp, err := svc.ChatCompletion(context.Background(), req)
```

### 8. Multi-turn Conversation

Maintain conversation context:

```go
messages := []openai.Message{
    {
        Role:    "user",
        Content: "What is 2+2?",
    },
    {
        Role:    "assistant",
        Content: "2+2 equals 4.",
    },
    {
        Role:    "user",
        Content: "What about 2+3?",
    },
}

req := openai.ChatRequest{
    Messages: messages,
}

resp, err := svc.ChatCompletion(context.Background(), req)
```

### 9. Using Custom Base URL (Azure OpenAI / Proxies)

Configure custom endpoints:

```go
// For Azure OpenAI
svc := openai.New(openai.Config{
    APIKey:  "your-azure-key",
    BaseURL: "https://your-resource.openai.azure.com/openai/deployments/your-deployment",
})

// For custom proxy
svc := openai.New(openai.Config{
    APIKey:  "sk-...",
    BaseURL: "https://your-proxy.example.com/v1",
})
```

## Error Handling

The package provides custom error types for common scenarios:

```go
resp, err := svc.ChatCompletion(ctx, req)
if err != nil {
    switch err {
    case openai.ErrInvalidAPIKey:
        // Handle invalid API key
        log.Fatal("Invalid OpenAI API key")
    case openai.ErrNoMessages:
        // Handle empty message list
        log.Fatal("No messages provided")
    case openai.ErrAPIError:
        // Handle OpenAI API errors
        // Check err.Internal for underlying error
        log.Printf("OpenAI API error: %v", err)
    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

### Available Error Types

- `ErrInvalidAPIKey` - Invalid or missing API key
- `ErrEmptyPrompt` - Empty prompt provided
- `ErrUnsupportedFormat` - Unsupported file format
- `ErrStreamingFailed` - Streaming response failed
- `ErrNoMessages` - No messages provided
- `ErrNoFile` - No file provided for audio
- `ErrAPIError` - General OpenAI API error

## Token Usage Tracking

All completion methods return comprehensive token usage:

```go
resp, err := svc.ChatCompletion(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Prompt tokens: %d\n", resp.Usage.PromptTokens)
fmt.Printf("Completion tokens: %d\n", resp.Usage.CompletionTokens)
fmt.Printf("Total tokens: %d\n", resp.Usage.TotalTokens)

// Estimate cost (example rates)
promptCost := float64(resp.Usage.PromptTokens) * 0.00003    // $0.03/1K tokens
completionCost := float64(resp.Usage.CompletionTokens) * 0.00006  // $0.06/1K tokens
totalCost := promptCost + completionCost
fmt.Printf("Estimated cost: $%.6f\n", totalCost)
```

## Best Practices

### 1. Context Management

Always pass context for proper cancellation and timeout:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := svc.ChatCompletion(ctx, req)
```

### 2. Rate Limiting

Implement rate limiting for production use:

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 requests per second

if err := limiter.Wait(ctx); err != nil {
    return err
}
resp, err := svc.ChatCompletion(ctx, req)
```

### 3. Error Retry Logic

The SDK has built-in retry logic via `MaxRetries` config, but you can add custom logic:

```go
var resp *openai.ChatResponse
var err error

for i := 0; i < 3; i++ {
    resp, err = svc.ChatCompletion(ctx, req)
    if err == nil {
        break
    }
    if err == openai.ErrInvalidAPIKey {
        return err // Don't retry auth errors
    }
    time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
}
```

### 4. Logging

The package uses structured logging via `pkg/logging`:

```go
import "github.com/kevinxvu/goone/pkg/logging"

// Logs are automatically created with "type":"openai" field
// Enable debug logging to see all API calls:
logging.SetConfig(&logging.Config{
    Level: zapcore.DebugLevel,
})
```

## Supported Models

### Chat Models
- `gpt-4` - Most capable model (default)
- `gpt-4-turbo`
- `gpt-4-vision-preview` - For image analysis
- `gpt-3.5-turbo` - Faster and cheaper
- `gpt-3.5-turbo-16k` - Extended context

### Audio Models
- `whisper-1` - Transcription and translation

For the latest model list, see [OpenAI Models](https://platform.openai.com/docs/models).

## Dependency Injection

To use the service via DI (optional):

```go
// In internal/di/wire.go
func ProvideOpenAIService(cfg *config.Configuration) *openai.Service {
    return openai.New(openai.Config{
        APIKey:       cfg.OpenAIAPIKey,
        BaseURL:      cfg.OpenAIBaseURL,
        Timeout:      cfg.OpenAITimeout,
        MaxRetries:   cfg.OpenAIMaxRetries,
        DefaultModel: cfg.OpenAIDefaultModel,
    })
}

// Add to Application struct:
type Application struct {
    // ... existing fields
    OpenAI *openai.Service
}

// Add to wire.Build():
wire.Build(
    // ... existing providers
    ProvideOpenAIService,
    wire.Struct(new(Application), "*"),
)
```

Then run `make wire` to regenerate DI code.

## Testing

Mock the service for unit tests:

```go
type MockOpenAIService struct{}

func (m *MockOpenAIService) ChatCompletion(ctx context.Context, req openai.ChatRequest) (*openai.ChatResponse, error) {
    return &openai.ChatResponse{
        Content: "Mocked response",
        Usage: openai.TokenUsage{
            TotalTokens: 10,
        },
    }, nil
}
```

## License

This package follows the project's license.
