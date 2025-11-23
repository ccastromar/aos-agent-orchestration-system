package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/mock/crm/customer", handleCustomerProfile)
	mux.HandleFunc("/mock/crm/interactions", handleCustomerInteractions)
	mux.HandleFunc("/mock/crm/ticket", handleCreateTicket)
	mux.HandleFunc("/mock/crm/lead/status", handleUpdateLeadStatus)

	addr := ":9002"
	log.Printf("[MOCK-CRM] escuchando en %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("error arrancando mock CRM: %v", err)
	}
}

func handleCustomerProfile(w http.ResponseWriter, r *http.Request) {
	customerId := r.URL.Query().Get("customerId")
	if customerId == "" {
		http.Error(w, "customerId requerido", http.StatusBadRequest)
		return
	}

	resp := map[string]any{
		"customerId":    customerId,
		"name":          "Laura Fernández",
		"segment":       "Gold",
		"email":         "laura.fernandez@example.com",
		"lastPurchase":  "2025-10-10",
		"lifetimeValue": 12450.75,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCustomerInteractions(w http.ResponseWriter, r *http.Request) {
	customerId := r.URL.Query().Get("customerId")
	if customerId == "" {
		http.Error(w, "customerId requerido", http.StatusBadRequest)
		return
	}

	days := r.URL.Query().Get("days")
	if days == "" {
		days = "30"
	}

	resp := map[string]any{
		"customerId": customerId,
		"windowDays": days,
		"items": []map[string]any{
			{
				"date":    "2025-11-15",
				"type":    "email",
				"agent":   "Sofia",
				"summary": "Consulta sobre el estado de su último pedido.",
			},
			{
				"date":    "2025-11-08",
				"type":    "call",
				"agent":   "Miguel",
				"summary": "Duda sobre la próxima factura y métodos de pago.",
			},
			{
				"date":    "2025-10-30",
				"type":    "ticket",
				"agent":   "Equipo Soporte",
				"summary": "Incidencia con una devolución, resuelta satisfactoriamente.",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "solo POST permitido", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	resp := map[string]any{
		"ticketId":   "TCK-12345",
		"status":     "OPEN",
		"customerId": payload["customerId"],
		"subject":    payload["subject"],
		"priority":   payload["priority"],
		"channel":    payload["channel"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleUpdateLeadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "solo POST permitido", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	resp := map[string]any{
		"leadId":    payload["leadId"],
		"newStatus": payload["newStatus"],
		"updated":   true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
