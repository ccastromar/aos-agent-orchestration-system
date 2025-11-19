package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/mock/core/balance", func(w http.ResponseWriter, r *http.Request) {
		accountId := r.URL.Query().Get("accountId")
		resp := map[string]any{
			"accountId": accountId,
			"currency":  "EUR",
			"balance":   15.56,
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/mock/core/movements", func(w http.ResponseWriter, r *http.Request) {
		accountId := r.URL.Query().Get("accountId")
		resp := map[string]any{
			"accountId": accountId,
			"currency":  "EUR",
			"movements": []map[string]any{
				{
					"date":   "2025-01-10",
					"amount": -25.00,
					"desc":   "Bizum Laura",
				},
				{
					"date":   "2025-01-05",
					"amount": 1500.00,
					"desc":   "Nómina",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/mock/payments/bizum", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		resp := map[string]any{
			"status": "ok",
			"detail": body,
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/mock/aml/check", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		resp := map[string]any{
			"riskScore":  12,
			"riskLevel":  "LOW",
			"sanctioned": false,
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/mock/notifications/send", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		log.Println("[MOCK NOTIF]", body)
		resp := map[string]any{
			"sent": true,
		}
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("[MOCK BACKEND] escuchando en :9000")
	http.ListenAndServe(":9000", mux)
}
