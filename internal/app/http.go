package app

import (
	"context"
	"net/http"

	"time"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/agent"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/health"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/logx"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/runtime"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/ui"
)

type HTTPServer struct {
	srv *http.Server
}

func NewHTTPServer(apiAgent *agent.APIAgent, uiStore *ui.UIStore, rt *runtime.Runtime) *HTTPServer {
	mux := http.NewServeMux()

	apiAgent.RegisterHTTP(mux)
	mux.HandleFunc("/ui", uiStore.HandleIndex)
	mux.HandleFunc("/ui/task", uiStore.HandleTask)
	mux.HandleFunc("/health/live", health.LiveHandler)
	mux.HandleFunc("/health/ready", health.ReadyHandler(rt))

	return &HTTPServer{
		srv: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
	}
}

func (h *HTTPServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		logx.Info("HTTP", "listening on port :8080")
		errCh <- h.srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logx.Info("HTTP", "shutting down server...")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return h.srv.Shutdown(shutCtx)
	}
}
