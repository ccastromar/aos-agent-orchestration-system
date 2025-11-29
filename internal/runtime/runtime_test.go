package runtime

import "testing"

// The Runtime type is a simple data holder; this test ensures
// its fields can be set and read as expected.
type fakeLLM struct{}

func (f *fakeLLM) Ping() error                       { return nil }
func (f *fakeLLM) Chat(prompt string) (string, error) { return "", nil }

func TestRuntimeFields(t *testing.T) {
    rt := &Runtime{SpecsLoaded: true, LLMClient: &fakeLLM{}}

    if !rt.SpecsLoaded {
        t.Fatalf("SpecsLoaded should be true")
    }
    if rt.LLMClient == nil {
        t.Fatalf("LLMClient should not be nil")
    }
    if err := rt.LLMClient.Ping(); err != nil {
        t.Fatalf("Ping should succeed: %v", err)
    }
}
