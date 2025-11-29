package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/ui"
	"github.com/stretchr/testify/require"
)

func TestAPIAgent_HandleAsk(t *testing.T) {
	// Setup dependencies
	messageBus := bus.New()
	uiStore := ui.NewUIStore()
	apiAgent := NewAPIAgent(messageBus, uiStore)

	// Subscribe to inspector channel to intercept the message
	inspectorChan := make(chan bus.Message, 1)
	messageBus.Subscribe("inspector", inspectorChan)

	// Setup HTTP server
	mux := http.NewServeMux()
	apiAgent.RegisterHTTP(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Prepare request
	reqBody := map[string]string{
		"message": "test message",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Start a goroutine to simulate the backend processing
	go func() {
		select {
		case msg := <-inspectorChan:
			// Verify message content
			payload := msg.Payload
			if payload == nil {
				return
			}
			id, ok := payload["id"].(string)
			if !ok {
				return
			}

			// Simulate processing time
			time.Sleep(50 * time.Millisecond)

			// Store result
			storeResult(id, Result{
				Status: "completed",
				Data:   map[string]string{"reply": "processed"},
			})
		case <-time.After(2 * time.Second):
			// Timeout in test helper
		}
	}()

	// Execute request
	resp, err := http.Post(ts.URL+"/ask", "application/json", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	require.Equal(t, "completed", result["status"])

	data, ok := result["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "processed", data["reply"])
}
