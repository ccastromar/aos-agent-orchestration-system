package llm

type LLMClient interface {
	Ping() error
	Chat(prompt string) (string, error)
	//DetectIntent(text string) (string, map[string]any, error)
	//Summarize(ctx map[string]any) (string, error)
}
