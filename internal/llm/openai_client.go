package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

type OpenAIClient struct {
    BaseURL string
    APIKey  string
    Model   string
    HTTP    *http.Client
    Timeout time.Duration
}

// Compile-time interface conformance
var _ LLMClient = (*OpenAIClient)(nil)

// NewOpenAIClient crea un nuevo proveedor OpenAI.
func NewOpenAIClient(baseURL, apiKey, model string) *OpenAIClient {
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }

	return &OpenAIClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

}

// Ping
func (c *OpenAIClient) Ping(ctx context.Context) error {
    if c.APIKey == "" {
        return fmt.Errorf("openai api key is empty")
    }

    // timeout configurable con valor por defecto
    to := c.Timeout
    if to <= 0 {
        to = 2 * time.Second
    }
    var cancel context.CancelFunc
    if ctx == nil {
        ctx = context.Background()
    }
    ctx, cancel = context.WithTimeout(ctx, to)
    defer cancel()

    url := strings.TrimRight(c.BaseURL, "/") + "/models"
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+c.APIKey)

    httpClient := c.HTTP
    if httpClient == nil {
        httpClient = &http.Client{Timeout: to}
    }

    resp, err := httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("openai ping failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("openai ping bad status: %d, body: %s", resp.StatusCode, string(b))
    }

    return nil

}

// Chat llama al modelo de OpenAI en modo no-stream
func (c *OpenAIClient) Chat(ctx context.Context, prompt string) (string, error) {
    if c.APIKey == "" {
        return "", fmt.Errorf("openai api key is empty")
    }

    payload := map[string]any{
        "model": c.Model,
        "messages": []map[string]string{
            {"role": "user", "content": prompt},
        },
        "temperature": 0,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("marshal payload: %w", err)
    }

    // timeout configurable con valor por defecto
    to := c.Timeout
    if to <= 0 {
        to = 30 * time.Second
    }
    var cancel context.CancelFunc
    if ctx == nil {
        ctx = context.Background()
    }
    ctx, cancel = context.WithTimeout(ctx, to)
    defer cancel()

    url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
    if err != nil {
        return "", err
    }

    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")

    httpClient := c.HTTP
    if httpClient == nil {
        httpClient = &http.Client{Timeout: to}
    }

    resp, err := httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("openai chat failed: status %d, body: %s", resp.StatusCode, string(b))
    }

    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}

	return result.Choices[0].Message.Content, nil

}
