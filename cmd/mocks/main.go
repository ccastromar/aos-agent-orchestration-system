package main

import (
	"log"
	"net/http"

	mockBanking "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/banking"
	mockCRM "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/crm"
	mockDevops "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/devops"
	mockHelpdesk "github.com/ccastromar/aos-agent-orchestration-system/internal/mocks/helpdesk"
)

func main() {
	mux := http.NewServeMux()

	// REGISTRAR ENDPOINTS DE CADA DOMINIO
	mockBanking.RegisterHandlers(mux)
	mockDevops.RegisterHandlers(mux)
	mockCRM.RegisterHandlers(mux)
	mockHelpdesk.RegisterHandlers(mux)

	log.Println("[MOCK SERVER] listening on :9000")
	http.ListenAndServe(":9000", mux)
}
