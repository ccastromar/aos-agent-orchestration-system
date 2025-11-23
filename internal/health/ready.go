package health

import (
	"net/http"

	"github.com/ccastromar/aos-banking-v2/internal/runtime"
)

func NewReadyHandler(rt *runtime.Runtime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !rt.SpecsLoaded {
			http.Error(w, "specs not loaded", 503)
			return
		}

		if err := rt.LLMClient.Ping(); err != nil {
			http.Error(w, "llm unreachable", 503)
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ready"}`))
	}
}
