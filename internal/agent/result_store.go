package agent

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Result struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Err    string      `json:"error,omitempty"`
}

var (
	resultsMu sync.Mutex
	results   = make(map[string]Result)
)

func storeResult(id string, res Result) {
	resultsMu.Lock()
	defer resultsMu.Unlock()
	results[id] = res
}

func waitForResult(id string, timeout time.Duration) Result {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		resultsMu.Lock()
		r, ok := results[id]
		resultsMu.Unlock()
		if ok {
			return r
		}
	}
	return Result{
		Status: "timeout",
		Err:    "timeout esperando resultado",
	}
}

func randomID() string {
	return uuid.NewString()
}
