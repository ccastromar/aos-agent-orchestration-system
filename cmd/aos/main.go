package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ccastromar/aos-agent-orchestration-system/internal/app"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New()
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	if err := a.Run(ctx); err != nil {
		log.Fatalf("error running app: %v", err)
	}

}
