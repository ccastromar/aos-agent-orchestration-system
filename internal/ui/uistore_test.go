package ui

import (
    "net/http"
    "net/http/httptest"
    "net/url"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
    "time"
)

// chdirToRepoRoot changes the working directory to the repository root
// so that relative template paths (templates/ui/...) resolve during tests.
func chdirToRepoRoot(t *testing.T) {
    t.Helper()
    // Determine this test file directory
    _, file, _, _ := runtime.Caller(0)
    // internal/ui/uistore_test.go -> repo root is two dirs up
    root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
    if err := os.Chdir(root); err != nil {
        t.Fatalf("chdir to repo root: %v", err)
    }
}

func TestUIStore_AddEventAndSnapshotIsolation(t *testing.T) {
    s := NewUIStore()
    s.AddEvent("task1", "agentA", "info", "hello", "10ms")
    s.AddEvent("task1", "agentB", "warn", "world", "5ms")

    snap := s.snapshot()
    if len(snap["task1"]) != 2 {
        t.Fatalf("expected 2 events, got %d", len(snap["task1"]))
    }

    // mutate snapshot and verify original store is not affected
    snap["task1"][0].Message = "hacked"
    again := s.snapshot()
    if again["task1"][0].Message == "hacked" {
        t.Fatalf("store should not reflect mutations to snapshot copy")
    }
}

func TestHandleIndex_OK_RendersAndOrdersByLastEvent(t *testing.T) {
    chdirToRepoRoot(t)

    s := NewUIStore()
    // Ensure ordering by making taskB have the most recent event
    s.AddEvent("taskA", "agent", "info", "msgA1", "1ms")
    time.Sleep(5 * time.Millisecond)
    s.AddEvent("taskB", "agent", "info", "msgB1", "1ms")

    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/ui", nil)
    s.HandleIndex(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    body := rr.Body.String()
    if !strings.Contains(body, "taskA") || !strings.Contains(body, "taskB") {
        t.Fatalf("expected both task IDs in output")
    }
    // Newest (taskB) should appear before taskA in the HTML
    if strings.Index(body, "taskB") > strings.Index(body, "taskA") {
        t.Fatalf("expected taskB to appear before taskA: body=\n%s", body)
    }
}

func TestHandleTask_MissingID_Redirects(t *testing.T) {
    chdirToRepoRoot(t)

    s := NewUIStore()
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/ui/task", nil)
    s.HandleTask(rr, req)

    if rr.Code != http.StatusFound {
        t.Fatalf("expected 302 redirect, got %d", rr.Code)
    }
}

func TestHandleTask_NotFound(t *testing.T) {
    chdirToRepoRoot(t)

    s := NewUIStore()
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/ui/task?id=unknown", nil)
    s.HandleTask(rr, req)

    if rr.Code != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", rr.Code)
    }
}

func TestHandleTask_OK(t *testing.T) {
    chdirToRepoRoot(t)

    s := NewUIStore()
    s.AddEvent("taskX", "agent1", "info", "first", "1ms")
    time.Sleep(2 * time.Millisecond)
    s.AddEvent("taskX", "agent2", "info", "second", "2ms")

    rr := httptest.NewRecorder()
    q := url.Values{"id": {"taskX"}}
    req := httptest.NewRequest(http.MethodGet, "/ui/task?"+q.Encode(), nil)
    s.HandleTask(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
    body := rr.Body.String()
    if !strings.Contains(body, "first") || !strings.Contains(body, "second") {
        t.Fatalf("expected event messages in body, got: %s", body)
    }
}
