package tools_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ccastromar/aos-banking-v2/internal/config"
	"github.com/ccastromar/aos-banking-v2/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestExecuteTool_URLRendering(t *testing.T) {
	var receivedURL string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer ts.Close()

	tool := config.Tool{
		Name:      "core_get_balance",
		Type:      "http",
		Method:    "GET",
		URL:       ts.URL + "/mock/balance?accountId={{ .accountId }}",
		TimeoutMs: 2000,
	}

	params := map[string]string{"accountId": "555"}

	out, err := tools.ExecuteTool(tool, params)
	require.NoError(t, err)
	require.Equal(t, true, out["ok"])

	require.Equal(t, "/mock/balance?accountId=555", receivedURL)
}
