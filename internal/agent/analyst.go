package agent

import (
	"log"

	"github.com/ccastromar/aos-banking-v2/internal/bus"
	"github.com/ccastromar/aos-banking-v2/internal/llm"
	"github.com/ccastromar/aos-banking-v2/internal/logx"
	"github.com/ccastromar/aos-banking-v2/internal/ui"
)

type Analyst struct {
	bus       *bus.Bus
	inbox     chan bus.Message
	llmClient llm.Client
	uiStore   *ui.UIStore
}

func NewAnalyst(b *bus.Bus, llmClient llm.Client, ui *ui.UIStore) *Analyst {
	return &Analyst{
		bus:       b,
		inbox:     make(chan bus.Message, 16),
		llmClient: llmClient,
		uiStore:   ui,
	}
}

func (a *Analyst) Inbox() chan bus.Message {
	return a.inbox
}

func (a *Analyst) Start() {
	for msg := range a.inbox {
		switch msg.Type {
		case "summarize":
			a.handleSummarize(msg)
		default:
			log.Printf("[Analyst] mensaje desconocido: %#v", msg)
		}
	}
}

func (a *Analyst) handleSummarize(msg bus.Message) {
	id := msg.Payload["id"].(string)
	intentType, _ := msg.Payload["intent"].(string)
	rawAny := msg.Payload["rawResult"]

	raw, ok := rawAny.(map[string]any)
	if !ok {
		log.Printf("[Analyst] rawResult inválido para id=%s", id)
		storeResult(id, Result{
			Status: "error",
			Err:    "resultado bruto inválido",
		})
		return
	}

	log.Printf("[Analyst] pidiendo summary al LLM...")
	log.Printf("[Analyst] rawResult recibido: %#v", raw)

	timer := logx.Start(id, "Analyst", "SummarizeLLM")
	summary, err := llm.SummarizeBankingResult(a.llmClient, intentType, raw)
	timer.End()

	if err != nil {
		log.Printf("[Analyst] error llamando al LLM: %v", err)
		// Degradamos de forma elegante: devolvemos solo el raw.
		storeResult(id, Result{
			Status: "ok",
			Data: map[string]any{
				"raw": raw,
			},
		})
		return
	}
	log.Printf("[Analyst] summary generado: %s", summary)
	a.uiStore.AddEvent(id, "Analyst", "summary", "summary LLM generado", "")

	storeResult(id, Result{
		Status: "ok",
		Data: map[string]any{
			"raw":     raw,
			"summary": summary,
		},
	})
}
