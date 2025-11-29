package agent

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/logx"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/ui"
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

// Max request size for POST /ask to protect the server (1MB)
const maxAskBodyBytes int64 = 1 << 20

func (a *APIAgent) Inbox() chan bus.Message {
	return a.inbox
}

func (a *APIAgent) Start(ctx context.Context) error {
	for {
		select {
		case msg := <-a.inbox:
			a.dispatch(msg)

		case <-ctx.Done():
			return nil
		}
	}
}

func (a *APIAgent) dispatch(msg bus.Message) {
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
    mux.HandleFunc("/ask", a.handleAsk)   // async structured mode
    mux.HandleFunc("/task", a.handleTask) // fetch task status/result
    //mux.HandleFunc("/ask_nlp", a.handleAskNLP) // modo lenguaje natural
}

func (a *APIAgent) handleAsk(w http.ResponseWriter, r *http.Request) {
    type Req struct {
        Message string `json:"message"`
    }

 // Limit request body size
 r.Body = http.MaxBytesReader(w, r.Body, maxAskBodyBytes)
 var req Req
 if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
     // If body too large, return 413; otherwise 400
     httpErr := http.StatusBadRequest
     if err != nil && err.Error() == "http: request body too large" {
         httpErr = http.StatusRequestEntityTooLarge
     }
     http.Error(w, "invalid request body", httpErr)
     return
 }

	if req.Message == "" {
		http.Error(w, "message requerido", http.StatusBadRequest)
		return
	}

 id := randomID()

 logx.Info("Api", "new request id=%s message='%s'", id, req.Message)
 a.uiStore.AddEvent(id, "Api", "request", req.Message, "")

 // Create and register a task context with a default TTL
 // Future: source from env
 _ = NewTaskContext(r.Context(), id, 60*time.Second)

    // Enviar al inspector con el message correcto
    a.bus.Send("inspector", bus.Message{
        Type: "new_task",
        Payload: map[string]any{
            "id":      id,
            "mode":    "structured",
            "message": req.Message, // ← ¡IMPORTANTE!
        },
    })

    // Respuesta asíncrona inmediata
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    _ = json.NewEncoder(w).Encode(map[string]any{
        "id":     id,
        "status": "accepted",
    })
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

// handleTask devuelve el estado/resultados de una tarea.
// GET /task?id=...
func (a *APIAgent) handleTask(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    id := r.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "id requerido", http.StatusBadRequest)
        return
    }

    // Consultar si ya hay resultado
    if res, ok := getResult(id); ok {
        // Limpiar almacenamiento para evitar fugas
        deleteResult(id)
        w.Header().Set("Content-Type", "application/json")
        // Mapear al formato de respuesta anterior
        _ = json.NewEncoder(w).Encode(map[string]any{
            "id":     id,
            "status": res.Status,
            "data":   res.Data,
            "error":  res.Err,
        })
        return
    }

    // Aún pendiente
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{
        "id":     id,
        "status": "pending",
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
