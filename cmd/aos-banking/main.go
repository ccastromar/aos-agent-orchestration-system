package main

import (
	"log"
	"net/http"

	"github.com/ccastromar/aos-banking-v2/internal/agent"
	"github.com/ccastromar/aos-banking-v2/internal/bus"
	"github.com/ccastromar/aos-banking-v2/internal/config"
	"github.com/ccastromar/aos-banking-v2/internal/llm"
	"github.com/ccastromar/aos-banking-v2/internal/ui"
)

func main() {
	// Cargar configuración
	cfg, err := config.LoadFromDir("config")
	if err != nil {
		log.Fatalf("error cargando configuración: %v", err)
	}
	uiStore := ui.NewUIStore()

	// Crear bus
	b := bus.New()

	// Cliente LLM (Ollama + qwen3:8b)
	llmClient := llm.NewOllamaClient("http://localhost:11434", "qwen3:0.6b")

	// Crear agentes
	apiAgent := agent.NewAPIAgent(b, uiStore)
	inspector := agent.NewInspector(b)
	planner := agent.NewPlanner(b, cfg, llmClient, uiStore)
	verifier := agent.NewVerifier(b, cfg, uiStore)
	analyst := agent.NewAnalyst(b, llmClient, uiStore)

	// Registrar en bus
	b.Subscribe("api", apiAgent.Inbox())
	b.Subscribe("inspector", inspector.Inbox())
	b.Subscribe("planner", planner.Inbox())
	b.Subscribe("verifier", verifier.Inbox())
	b.Subscribe("analyst", analyst.Inbox())

	// Lanzar agentes
	go apiAgent.Start()
	go inspector.Start()
	go planner.Start()
	go verifier.Start()
	go analyst.Start()

	// HTTP API
	mux := http.NewServeMux()
	apiAgent.RegisterHTTP(mux)
	mux.HandleFunc("/ui", uiStore.HandleIndex)
	mux.HandleFunc("/ui/task", uiStore.HandleTask)

	log.Println("[AOS-BANKING v2] escuchando en :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("error en servidor HTTP: %v", err)
	}
}
