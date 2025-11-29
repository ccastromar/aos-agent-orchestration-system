package agent

import (
	"context"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
	"github.com/ccastromar/aos-agent-orchestration-system/internal/logx"
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

func (i *Inspector) Start(ctx context.Context) error {
    defer func() {
        if r := recover(); r != nil {
            logx.Error("Inspector", "panic recovered in Start: %v", r)
        }
    }()
    for {
        select {
        case msg := <-i.inbox:
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        logx.Error("Inspector", "panic recovered in dispatch: %v", r)
                    }
                }()
                i.dispatch(msg)
            }()

		case <-ctx.Done():
			return nil
		}
	}
}

func (i *Inspector) dispatch(msg bus.Message) {
	switch msg.Type {
	case "new_task":
		id := msg.Payload["id"].(string)
		mode, _ := msg.Payload["mode"].(string)
		logx.Info("Inspector", "new task id=%s mode=%s", id, mode)

		i.bus.Send("planner", bus.Message{
			Type: "detect_intent",
			Payload: map[string]any{
				"id":      id,
				"message": msg.Payload["message"],
				"mode":    mode,
			},
		})

	default:
		logx.Warn("Inspector", "unknown message: %#v", msg)
	}

}
