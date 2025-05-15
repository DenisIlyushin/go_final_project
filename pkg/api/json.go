package api

import (
	"encoding/json"
	"net/http"
)

func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string) {
	writeJson(w, map[string]string{"error": msg})
}
