package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestBuildMux_RegistersBankingHandlers(t *testing.T) {
    mux := buildMux()
    server := httptest.NewServer(mux)
    defer server.Close()

    resp, err := http.Get(server.URL + "/mock/core/balance?accountId=abc")
    if err != nil { t.Fatalf("GET failed: %v", err) }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status: %d", resp.StatusCode)
    }

    var out map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        t.Fatalf("decode: %v", err)
    }

    if out["accountId"] != "abc" {
        t.Fatalf("expected accountId=abc, got %v", out["accountId"])
    }
}
