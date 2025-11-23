package agent

import (
	"testing"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/llm"
	"github.com/stretchr/testify/require"
)

type dummyLLM struct {
	output string
}

// Ping implements llm.LLMClient.
func (d dummyLLM) Ping() error {
	panic("unimplemented")
}

func (d dummyLLM) Chat(prompt string) (string, error) {
	return d.output, nil
}

func TestDetectIntent(t *testing.T) {
	mock := dummyLLM{
		output: `{"intent":"banking.get_balance","required_params":["accountId"]}`,
	}

	// 🟢 IMPORTANTE: solo las KEYS importan para la validación actual
	schemas := map[string]any{
		"banking.get_balance": struct{}{},
	}

	di, err := llm.DetectIntent(mock, "saldo de mi cuenta", schemas)
	require.NoError(t, err)
	require.Equal(t, "banking.get_balance", di.Type)
	require.Equal(t, []string{"accountId"}, di.Params)
}
