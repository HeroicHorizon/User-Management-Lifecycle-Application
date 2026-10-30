// Package handler exposes the HTTP API.
package handler

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// envelope is the consistent response structure for every endpoint.
type envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *errorBody `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, envelope{Success: true, Data: data})
}
