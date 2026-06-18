package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler responds to GET /health
// This is the first thing you test — if this works, the gateway is alive.
// Tools like Kubernetes also call /health to know if the app is running.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
