package llm

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Client interface {
	Chat(prompt string) (string, error)
}

type OllamaClient struct {
	BaseURL string
	Model   string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
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

func (c *OllamaClient) Chat(prompt string) (string, error) {
	payload := map[string]any{
		"model": c.Model,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
		"stream": true,
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.BaseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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
