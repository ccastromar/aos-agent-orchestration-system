package llm

import "context"

type LLMClient interface {
    Ping(ctx context.Context) error
    Chat(ctx context.Context, prompt string) (string, error)
    //DetectIntent(text string) (string, map[string]any, error)
    //Summarize(ctx map[string]any) (string, error)
}
