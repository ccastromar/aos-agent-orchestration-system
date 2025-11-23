package agent

import (
	"context"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/bus"
)

type Agent interface {
	Start(ctx context.Context) error
	Inbox() chan bus.Message
}
