package llm

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestOpenAI_Ping_OK(t *testing.T) {
    var gotAuth string
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/models" {
            t.Fatalf("unexpected path: %s", r.URL.Path)
        }
        gotAuth = r.Header.Get("Authorization")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"data":[]}`))
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "test-key", "gpt-4.1")
    c.Timeout = 500 * time.Millisecond

    if err := c.Ping(context.Background()); err != nil {
        t.Fatalf("Ping() unexpected error: %v", err)
    }
    if gotAuth != "Bearer test-key" {
        t.Fatalf("expected Authorization header to be set, got %q", gotAuth)
    }
}

func TestOpenAI_Ping_Non200(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "nope", http.StatusUnauthorized)
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "test-key", "gpt-4.1")
    c.Timeout = 200 * time.Millisecond

    err := c.Ping(context.Background())
    if err == nil {
        t.Fatalf("expected error for non-200 status")
    }
    if !errors.Is(err, err) { // keep compiler from complaining; we check message contents below
    }
    if have := err.Error(); !(contains(have, "bad status") && contains(have, "401") && contains(have, "nope")) {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestOpenAI_Chat_Success(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            t.Fatalf("expected POST, got %s", r.Method)
        }
        if r.URL.Path != "/chat/completions" {
            t.Fatalf("unexpected path: %s", r.URL.Path)
        }
        if ct := r.Header.Get("Content-Type"); ct != "application/json" {
            t.Fatalf("expected Content-Type application/json, got %s", ct)
        }
        if auth := r.Header.Get("Authorization"); auth != "Bearer key" {
            t.Fatalf("expected Authorization 'Bearer key', got %q", auth)
        }
        // return a minimal valid chat completion response
        resp := map[string]any{
            "choices": []any{
                map[string]any{
                    "message": map[string]any{
                        "content": "hello world",
                    },
                },
            },
        }
        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(resp)
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "key", "gpt-4.1")
    c.Timeout = 500 * time.Millisecond

    out, err := c.Chat(context.Background(), "hi")
    if err != nil {
        t.Fatalf("Chat() unexpected error: %v", err)
    }
    if out != "hello world" {
        t.Fatalf("unexpected chat output: %q", out)
    }
}

func TestOpenAI_Chat_Non200(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "boom", http.StatusInternalServerError)
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "key", "gpt-4.1")
    c.Timeout = 200 * time.Millisecond

    _, err := c.Chat(context.Background(), "hi")
    if err == nil {
        t.Fatalf("expected error for non-200 status")
    }
    // should include status code and body contents
    if have := err.Error(); !(contains(have, "status 500") && contains(have, "boom")) {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestOpenAI_Chat_EmptyChoices(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"choices":[]}`))
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "key", "gpt-4.1")
    c.Timeout = 200 * time.Millisecond

    _, err := c.Chat(context.Background(), "hi")
    if err == nil || !contains(err.Error(), "empty response") {
        t.Fatalf("expected empty response error, got %v", err)
    }
}

func TestOpenAI_Chat_BadJSON(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{malformed`))
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "key", "gpt-4.1")
    c.Timeout = 200 * time.Millisecond

    _, err := c.Chat(context.Background(), "hi")
    if err == nil {
        t.Fatalf("expected JSON decode error")
    }
}

func TestOpenAI_APIKey_Required(t *testing.T) {
    c := NewOpenAIClient("http://example", "", "gpt-4.1")
    if err := c.Ping(context.Background()); err == nil {
        t.Fatalf("expected error when API key is empty for Ping")
    }
    if _, err := c.Chat(context.Background(), "hello"); err == nil {
        t.Fatalf("expected error when API key is empty for Chat")
    }
}

func TestOpenAI_Chat_ContextTimeout(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(300 * time.Millisecond)
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"choices":[{"message":{"content":"late"}}]}`))
    }))
    defer ts.Close()

    c := NewOpenAIClient(ts.URL, "key", "gpt-4.1")
    c.Timeout = 100 * time.Millisecond // request should time out

    if _, err := c.Chat(context.Background(), "hi"); err == nil {
        t.Fatalf("expected timeout error from context")
    }
}

// contains is a small helper to avoid importing strings in every test
func contains(s, substr string) bool { return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && (indexOf(s, substr) >= 0))) }

func indexOf(s, substr string) int {
    // simple substring search to avoid importing strings; fine for tests
    n := len(s)
    m := len(substr)
    if m == 0 {
        return 0
    }
    if m > n {
        return -1
    }
    for i := 0; i <= n-m; i++ {
        if s[i:i+m] == substr {
            return i
        }
    }
    return -1
}
