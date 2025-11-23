package agent

import (
	"testing"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/llm"
	"github.com/stretchr/testify/require"
)

type dummyParams struct {
	data string
}

// Ping implements llm.LLMClient.
func (d dummyParams) Ping() error {
	panic("unimplemented")
}

func (d dummyParams) Chat(prompt string) (string, error) {
	return d.data, nil
}

func TestExtractParams(t *testing.T) {
	mock := dummyParams{
		data: `{"accountId":"999"}`,
	}

	params, err := llm.ExtractParams(mock, "quiero mi saldo", []string{"accountId"})
	require.NoError(t, err)
	require.Equal(t, "999", params["accountId"])
}
