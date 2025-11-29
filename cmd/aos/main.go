package main

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/ccastromar/aos-agent-orchestration-system/internal/app"
)

// runner is the minimal interface our app must satisfy for running.
type runner interface{ Run(context.Context) error }

// appCtor is a constructor indirection to enable testing without launching the real app.
var appCtor = func() (runner, error) { return app.New() }

// fatalf indirection allows testing fatal paths without exiting the test process.
var fatalf = log.Fatalf

func run(ctx context.Context) {
    a, err := appCtor()
    if err != nil {
        fatalf("error initializing app: %v", err)
        return
    }
    if err := a.Run(ctx); err != nil {
        fatalf("error running app: %v", err)
        return
    }
}

func main() {
    // CLI flags
    port := flag.String("port", "9090", "HTTP port to listen on")
    flag.Parse()

    // apply runtime options
    app.SetHTTPPort(*port)

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    run(ctx)
}
