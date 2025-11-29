package app

import (
	"context"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/logx"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/runtime"
	"golang.org/x/sync/errgroup"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/agent"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/config"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/llm"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/ui"
)

type App struct {
	cfg    *config.Config
	bus    *bus.Bus
	ui     *ui.UIStore
	agents []agent.Agent
	llm    llm.LLMClient
	http   *HTTPServer
}

func New() (*App, error) {
	cfg, err := config.LoadFromDir("definitions")
	if err != nil {
		return nil, err
	}

	uiStore := ui.NewUIStore()
	messageBus := bus.New()

	llmClient := llm.NewOllamaClient("http://localhost:11434", "qwen3:0.6b")

	r := &runtime.Runtime{
		SpecsLoaded: true,
		LLMClient:   llmClient,
	}

	// Crear todos los agentes
	apiAgent := agent.NewAPIAgent(messageBus, uiStore)
	inspector := agent.NewInspector(messageBus)
	planner := agent.NewPlanner(messageBus, cfg, llmClient, uiStore)
	verifier := agent.NewVerifier(messageBus, cfg, uiStore)
	analyst := agent.NewAnalyst(messageBus, llmClient, uiStore)

	// Registrar subscripciones
	//messageBus.Subscribe("api", apiAgent.Inbox())
	messageBus.Subscribe("inspector", inspector.Inbox())
	messageBus.Subscribe("planner", planner.Inbox())
	messageBus.Subscribe("verifier", verifier.Inbox())
	messageBus.Subscribe("analyst", analyst.Inbox())

	httpServer := NewHTTPServer(apiAgent, uiStore, r)

	return &App{
		cfg:    cfg,
		bus:    messageBus,
		ui:     uiStore,
		agents: []agent.Agent{inspector, planner, verifier, analyst},
		llm:    llmClient,
		http:   httpServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	// Lanzar agentes
	for _, ag := range a.agents {
		agent := ag
		g.Go(func() error {
			return agent.Start(gctx)
		})
	}

	// Lanzar HTTP server
	g.Go(func() error {
		return a.http.Start(gctx)
	})

	logx.Info("App", "AOS v0.2.0 started")

	return g.Wait()
}
