package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Defino la interfaz mas abajo
// type Client interface {
// 	Chat(prompt string) (string, error)
// }

type OllamaClient struct {
    BaseURL    string
    Model      string
    HTTPClient *http.Client
}

// Asegura que implementa la interfaz
var _ LLMClient = (*OllamaClient)(nil)

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func (c *OllamaClient) Chat(ctx context.Context, prompt string) (string, error) {
    payload := map[string]any{
        "model": c.Model,
        "messages": []map[string]any{
            {"role": "user", "content": prompt},
        },
        "stream": true,
    }

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

 // Context with timeout prevents hangs; derive from provided ctx
 if ctx == nil {
     ctx = context.Background()
 }
 ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
 defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama chat failed: status %d, body: %s", resp.StatusCode, string(b))
	}

	dec := json.NewDecoder(resp.Body)
	var out bytes.Buffer

	for {
		var chunk struct {
			Message *struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}

		if err := dec.Decode(&chunk); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}

		if chunk.Message != nil {
			out.WriteString(chunk.Message.Content)
		}

		if chunk.Done {
			break
		}
	}

	return out.String(), nil
}

// Ping checks if Ollama is reachable and responding.
func (c *OllamaClient) Ping(ctx context.Context) error {
    // Ollama health: GET /api/tags
    req, err := http.NewRequest("GET", c.BaseURL+"/api/tags", nil)
    if err != nil {
        return err
    }

    if ctx == nil {
        ctx = context.Background()
    }
    ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
    defer cancel()
    req = req.WithContext(ctx)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("llm ping failed: status %d", resp.StatusCode)
	}

	return nil
}
