package handlers

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/status"
)

// writeJSON sends a JSON response with the given status code.
// Every handler uses this — it sets the Content-Type header automatically.
func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// writeError sends a JSON error response.
// Example: { "error": "user not found" }
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// grpcErrMsg extracts the human-readable message from a gRPC error.
// Without this, the error looks like: "rpc error: code = NotFound desc = user not found"
// With this, it just returns: "user not found"
func grpcErrMsg(err error) string {
	if st, ok := status.FromError(err); ok {
		return st.Message()
	}
	return err.Error()
}
