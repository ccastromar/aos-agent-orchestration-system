package main

import (
    "log"
    "net/http"

    mockBanking "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/banking"
    mockCRM "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/crm"
    mockDevops "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/devops"
    mockHelpdesk "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/helpdesk"
)

var listenAndServe = http.ListenAndServe

func buildMux() *http.ServeMux {
    mux := http.NewServeMux()
    // REGISTRAR ENDPOINTS DE CADA DOMINIO
    mockBanking.RegisterHandlers(mux)
    mockDevops.RegisterHandlers(mux)
    mockCRM.RegisterHandlers(mux)
    mockHelpdesk.RegisterHandlers(mux)
    return mux
}

func main() {
    mux := buildMux()
    log.Println("[MOCK SERVER] listening on :9000")
    listenAndServe(":9000", mux)
}
