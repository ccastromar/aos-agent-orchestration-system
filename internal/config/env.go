package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type EnvVars struct {
    AppEnv       string        `env:"APP_ENV" default:"dev"`
    Port         int           `env:"PORT" default:"8080"`
    ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"5s"`
    WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"5s"`

    BusWorkers int `env:"BUS_WORKERS" default:"4"`
    BusBuffer  int `env:"BUS_BUFFER"  default:"100"`

    LLMApiKey  string        `env:"LLM_API_KEY" required:"true"`
    LLMBaseURL string        `env:"LLM_BASE_URL" default:"https://api.openai.com/v1"`
    LLMModel   string        `env:"LLM_MODEL" default:"gpt-4.1"`
    LLMTimeout time.Duration `env:"LLM_TIMEOUT" default:"10s"`

    // Ollama (local LLM) configuration
    OllamaBaseURL string `env:"OLLAMA_BASE_URL" default:"http://localhost:11434"`
    OllamaModel   string `env:"OLLAMA_MODEL" default:"qwen3:0.6b"`

    LogLevel string `env:"LOG_LEVEL" default:"info"`
}

func LoadEnv() (*EnvVars, error) {
	var v EnvVars
	if err := envconfig.Process("", &v); err != nil {
		return nil, err
	}
	return &v, nil
}
