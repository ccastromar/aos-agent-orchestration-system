package logx

import (
	"log"
	"time"
)

type Timer struct {
	start time.Time
	id    string
	comp  string
	op    string
}

func Start(id, comp, op string) *Timer {
	return &Timer{
		start: time.Now(),
		id:    id,
		comp:  comp,
		op:    op,
	}
}

func (t *Timer) End() {
	elapsed := time.Since(t.start)
	log.Printf("[%s][%s][TIMING] %s = %v", t.comp, t.id, t.op, elapsed)
}
