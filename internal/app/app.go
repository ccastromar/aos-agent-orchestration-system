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
	env    *config.EnvVars
	bus    *bus.Bus
	ui     *ui.UIStore
	agents []agent.Agent
	llm    llm.LLMClient
	http   *HTTPServer
}

// New loads environment variables if available and delegates to NewWithEnv.
// It is tolerant to missing env during tests (e.g., required vars not set).
func New() (*App, error) {
	env, err := config.LoadEnv()
	if err != nil {
		// Proceed without env to keep backward compatibility in tests
		return NewWithEnv(nil)
	}
	return NewWithEnv(env)
}

// NewWithEnv builds the App wiring using the provided environment variables.
func NewWithEnv(env *config.EnvVars) (*App, error) {
	cfg, err := config.LoadFromDir("definitions")
	if err != nil {
		return nil, err
	}

	uiStore := ui.NewUIStore()
	messageBus := bus.New()

	// Select LLM client parameters from env when available (Ollama)
	ollamaURL := "http://localhost:11434"
	ollamaModel := "qwen3:0.6b"
	if env != nil {
		if env.OllamaBaseURL != "" {
			ollamaURL = env.OllamaBaseURL
		}
		if env.OllamaModel != "" {
			ollamaModel = env.OllamaModel
		}
	}
	llmClient := llm.NewOllamaClient(ollamaURL, ollamaModel)

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
		env:    env,
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
		agentAI := ag
		g.Go(func() error {
			return agentAI.Start(gctx)
		})
	}

	// Lanzar HTTP server
	g.Go(func() error {
		return a.http.Start(gctx)
	})

	if a.env != nil {
		logx.Info("App", "AOS v0.2.0 started (env=%s)", a.env.AppEnv)
	} else {
		logx.Info("App", "AOS v0.2.0 started")
	}

	return g.Wait()
}
