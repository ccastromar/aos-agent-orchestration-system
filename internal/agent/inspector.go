package agent

import (
	"log"

	"github.com/ccastromar/aos-banking-v2/internal/bus"
)

type Inspector struct {
	bus   *bus.Bus
	inbox chan bus.Message
}

func NewInspector(b *bus.Bus) *Inspector {
	return &Inspector{
		bus:   b,
		inbox: make(chan bus.Message, 16),
	}
}

func (i *Inspector) Inbox() chan bus.Message {
	return i.inbox
}

func (i *Inspector) Start() {
	for msg := range i.inbox {
		switch msg.Type {
		case "new_task":
			id := msg.Payload["id"].(string)
			mode, _ := msg.Payload["mode"].(string)
			log.Printf("[Inspector] nueva tarea id=%s mode=%s", id, mode)

			i.bus.Send("planner", bus.Message{
				Type: "detect_intent",
				Payload: map[string]any{
					"id":      id,
					"message": msg.Payload["message"],
					"mode":    mode,
				},
			})

		default:
			log.Printf("[Inspector] mensaje desconocido: %#v", msg)
		}
	}
}
