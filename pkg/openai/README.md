# OpenAI Service Package

A reusable OpenAI service package for Go that provides easy access to OpenAI's API, including:

- **Chat Completions** — Text generation with GPT models (including vision with image URLs)
- **Streaming** — Real-time streaming responses via a callback handler
- **Audio Transcription** — Speech-to-text using Whisper (`TranscribeAudio`)
- **Audio Translation** — Translate audio to English using Whisper (`TranslateAudio`)

## Configuration

Environment variables (OS env > `.env.local` > `.env`):

```env
# Required
OPENAI_API_KEY=sk-...

# Optional (defaults shown)
OPENAI_BASE_URL=https://api.openai.com/v1  # Custom endpoint (Azure OpenAI, proxy, etc.)
OPENAI_TIMEOUT=60                          # Request timeout in seconds
OPENAI_MAX_RETRIES=2                       # Number of retry attempts

# Model defaults per capability
OPENAI_TEXT_MODEL=gpt-4o      # Default model for chat/text completions
OPENAI_AUDIO_MODEL=whisper-1  # Default model for audio transcription/translation
```

### Initializing the service

```go
svc := openai.New(openai.Config{
    APIKey:     cfg.OpenAIAPIKey,
    BaseURL:    cfg.OpenAIBaseURL,
    Timeout:    cfg.OpenAITimeout,
    MaxRetries: cfg.OpenAIMaxRetries,
    TextModel:  cfg.OpenAITextModel,
    AudioModel: cfg.OpenAIAudioModel,
})
```

## Usage Examples

### 1. Basic Chat Completion

```go
resp, err := svc.ChatCompletion(ctx, openai.ChatRequest{
    Messages: []openai.Message{
        {Role: "user", Content: "What is the capital of France?"},
    },
})
if err != nil {
    return err
}
fmt.Printf("Response: %s\n", resp.Content)
fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
```

### 2. Chat with System Prompt

```go
systemPrompt := "You are a helpful assistant that speaks like a pirate."

resp, err := svc.ChatCompletion(ctx, openai.ChatRequest{
    SystemPrompt: &systemPrompt,
    Messages: []openai.Message{
        {Role: "user", Content: "Tell me about treasure hunting."},
    },
})
```

### 3. Vision — Analyze Images

Pass image URLs alongside text in a user message (requires a vision-capable model such as `gpt-4o`):

```go
resp, err := svc.ChatCompletion(ctx, openai.ChatRequest{
    Messages: []openai.Message{
        {
            Role:      "user",
            Content:   "What is in this image?",
            ImageURLs: []string{"https://example.com/photo.jpg"},
        },
    },
    MaxTokens: openai.Int64Ptr(300),
})
```

### 4. Per-Request Overrides

```go
model := "gpt-4o-mini"
temperature := 0.9
maxTokens := int64(500)

resp, err := svc.ChatCompletion(ctx, openai.ChatRequest{
    Model:       &model,
    Temperature: &temperature,
    MaxTokens:   &maxTokens,
    Messages: []openai.Message{
        {Role: "user", Content: "Write a creative poem."},
    },
})
```

### 5. Multi-Turn Conversation

```go
resp, err := svc.ChatCompletion(ctx, openai.ChatRequest{
    Messages: []openai.Message{
        {Role: "user",      Content: "What is 2+2?"},
        {Role: "assistant", Content: "2+2 equals 4."},
        {Role: "user",      Content: "What about 2+3?"},
    },
})
```

### 6. Streaming Responses

```go
handler := func(chunk string, done bool) error {
    if done {
        fmt.Println() // newline at end
        return nil
    }
    fmt.Print(chunk) // print token as it arrives
    return nil
}

resp, err := svc.ChatCompletionStream(ctx, openai.ChatRequest{
    Messages: []openai.Message{
        {Role: "user", Content: "Write a short story about a robot."},
    },
}, handler)
if err != nil {
    return err
}
// resp.Content holds the full assembled response
fmt.Printf("\n%d chunks received\n", ...)
```

### 7. Audio Transcription (Whisper)

```go
file, err := os.Open("audio.mp3")
if err != nil {
    return err
}
defer file.Close()

language := "en"
resp, err := svc.TranscribeAudio(ctx, openai.AudioRequest{
    File:     file,
    FileName: "audio.mp3",
    Language: &language, // optional — improves accuracy
})
if err != nil {
    return err
}
fmt.Printf("Transcription: %s\n", resp.Text)
```

### 8. Audio Translation to English

```go
file, err := os.Open("german_audio.mp3")
if err != nil {
    return err
}
defer file.Close()

resp, err := svc.TranslateAudio(ctx, openai.AudioRequest{
    File:     file,
    FileName: "german_audio.mp3",
})
if err != nil {
    return err
}
fmt.Printf("English translation: %s\n", resp.Text)
```

### 9. Custom Base URL (Azure OpenAI / Proxy)

```go
svc := openai.New(openai.Config{
    APIKey:  "your-azure-key",
    BaseURL: "https://your-resource.openai.azure.com/openai/deployments/your-deployment",
})
```

## Error Handling

```go
resp, err := svc.ChatCompletion(ctx, req)
if err != nil {
    var he *apperr.HTTPError
    if errors.As(err, &he) {
        switch he.Type {
        case "NO_MESSAGES":
            // no messages provided
        case "EMPTY_PROMPT":
            // message content is empty
        case "NO_FILE":
            // audio file missing
        case "OPENAI_API_ERROR":
            // upstream API error — he.Internal has details
        }
    }
}
```

### Error reference

| Error var          | Type               | HTTP status |
|--------------------|--------------------|-------------|
| `ErrInvalidAPIKey` | `INVALID_API_KEY`  | 401         |
| `ErrEmptyPrompt`   | `EMPTY_PROMPT`     | 400         |
| `ErrNoMessages`    | `NO_MESSAGES`      | 400         |
| `ErrNoFile`        | `NO_FILE`          | 400         |
| `ErrStreamingFailed` | `STREAMING_FAILED` | 500       |
| `ErrAPIError`      | `OPENAI_API_ERROR` | 502         |

## Supported Models

### Text / Chat
- `gpt-4o` *(default)* — multimodal, latest
- `gpt-4o-mini` — faster and cheaper
- `gpt-4-turbo`

### Audio
- `whisper-1` *(default)* — transcription and translation

For the full list see [OpenAI Models](https://platform.openai.com/docs/models).

## Helper Utilities

```go
openai.StringPtr("value")    // *string
openai.Int64Ptr(500)         // *int64
openai.Float64Ptr(0.9)       // *float64
```

## Running Tests

### Unit tests (no API key required)

```bash
go test ./pkg/openai/... -run "^Test[^I]"
```

### Integration tests (require `OPENAI_API_KEY`)

```bash
OPENAI_API_KEY=sk-... go test ./pkg/openai/... -v -run "^TestIntegration"
```

### Audio integration tests (also require an audio file)

```bash
OPENAI_API_KEY=sk-... OPENAI_TEST_AUDIO_FILE=./testdata/sample.mp3 \
  go test ./pkg/openai/... -v -run "^TestIntegration_(Transcribe|Translate)"
```

## Dependency Injection

```go
// internal/di/wire.go
func ProvideOpenAIService(cfg *config.Configuration) *openai.Service {
    return openai.New(openai.Config{
        APIKey:     cfg.OpenAIAPIKey,
        BaseURL:    cfg.OpenAIBaseURL,
        Timeout:    cfg.OpenAITimeout,
        MaxRetries: cfg.OpenAIMaxRetries,
        TextModel:  cfg.OpenAITextModel,
        AudioModel: cfg.OpenAIAudioModel,
    })
}

// Application struct field:
OpenAI *openai.Service

// wire.Build():
wire.Build(..., ProvideOpenAIService, wire.Struct(new(Application), "*"))
```

Run `make wire` after modifying `wire.go`.

