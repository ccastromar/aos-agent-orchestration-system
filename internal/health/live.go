package health

import "net/http"

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(`{"status": "ok"}`))
}
