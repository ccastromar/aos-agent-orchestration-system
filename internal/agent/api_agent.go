package agent

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ccastromar/aos-banking-v2/internal/bus"
	"github.com/ccastromar/aos-banking-v2/internal/ui"
)

type APIAgent struct {
	bus     *bus.Bus
	inbox   chan bus.Message
	uiStore *ui.UIStore // <-- nuevo

}

func NewAPIAgent(b *bus.Bus, ui *ui.UIStore) *APIAgent {
	return &APIAgent{
		bus:     b,
		inbox:   make(chan bus.Message, 16),
		uiStore: ui,
	}
}

func (a *APIAgent) Inbox() chan bus.Message {
	return a.inbox
}

func (a *APIAgent) Start() {
	for msg := range a.inbox {
		log.Printf("[API] mensaje interno ignorado: %#v", msg)
	}
}

type askRequest struct {
	Operation string         `json:"operation,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
	Message   string         `json:"message"`
}

type askNLPRequest struct {
	Message string `json:"message"`
}

type askResponse struct {
	ID     string      `json:"id"`
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// RegisterHTTP registra endpoints HTTP
func (a *APIAgent) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/ask", a.handleAsk) // modo estructurado
	//mux.HandleFunc("/ask_nlp", a.handleAskNLP) // modo lenguaje natural
}

func (a *APIAgent) handleAsk(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Message string `json:"message"`
	}

	var req Req
	json.NewDecoder(r.Body).Decode(&req)

	if req.Message == "" {
		http.Error(w, "message requerido", http.StatusBadRequest)
		return
	}

	id := randomID()

	log.Printf("[Api] nueva petición id=%s message='%s'", id, req.Message)
	a.uiStore.AddEvent(id, "Api", "request", req.Message, "")

	// Enviar al inspector con el message correcto
	a.bus.Send("inspector", bus.Message{
		Type: "new_task",
		Payload: map[string]any{
			"id":      id,
			"mode":    "structured",
			"message": req.Message, // ← ¡IMPORTANTE!
		},
	})

	// Esperar resultado final como en v2
	result := waitForResult(id, 15*time.Second)

	json.NewEncoder(w).Encode(result)
}

// /ask → operation + params (como v1)
func (a *APIAgent) handleAsk2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[API] error parseando request:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"json inválido"}`))
		return
	}

	if req.Operation == "" && req.Message == "" {
		http.Error(w, "operation o message requerido", 400)
		return
	}

	id := randomID()

	a.bus.Send("inspector", bus.Message{
		Type: "new_task",
		Payload: map[string]any{
			"id":        id,
			"mode":      "structured",
			"operation": req.Operation,
			"params":    req.Params,
		},
	})

	res := waitForResult(id, 30*time.Second)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if res.Err != "" {
		w.WriteHeader(http.StatusInternalServerError)
	}
	enc.Encode(askResponse{
		ID:     id,
		Status: res.Status,
		Result: res.Data,
		Error:  res.Err,
	})
}

// /ask_nlp → message (texto libre)
func (a *APIAgent) handleAskNLP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req askNLPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[API] error parseando request NLP:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"json inválido"}`))
		return
	}

	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"message requerido"}`))
		return
	}

	id := randomID()

	a.bus.Send("inspector", bus.Message{
		Type: "new_task",
		Payload: map[string]any{
			"id":      id,
			"mode":    "nlp",
			"message": req.Message,
		},
	})

	res := waitForResult(id, 30*time.Second)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if res.Err != "" {
		w.WriteHeader(http.StatusInternalServerError)
	}
	enc.Encode(askResponse{
		ID:     id,
		Status: res.Status,
		Result: res.Data,
		Error:  res.Err,
	})
}
