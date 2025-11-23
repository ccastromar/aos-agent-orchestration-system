package logx

import (
	"fmt"
	"log"
	"time"
)

func L(id, agent, msg string, args ...any) {
	prefix := fmt.Sprintf("[%s][%s][%s] ",
		time.Now().Format(time.RFC3339),
		agent,
		id,
	)
	log.Printf(prefix+msg, args...)
}

// Versión sin ID (para logs globales de arranque)
func G(agent, msg string, args ...any) {
	prefix := fmt.Sprintf("[%s][%s] ",
		time.Now().Format(time.RFC3339),
		agent,
	)
	log.Printf(prefix+msg, args...)
}
