package routes

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	l *zap.SugaredLogger
}

func (rw *responseWriter) WriteJSON(v interface{}) error {
	rw.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")

	err := enc.Encode(v)
	if err != nil {
		rw.l.Errorf("Error writing JSON response: %v", err)
	}
	return err
}

type jsonError struct {
	Error string `json:"error"`
	Desc  string `json:"description,omitempty"`
	Code  int    `json:"code,omitempty"`
}

func jsonErrorFrom(err error, code int) jsonError {
	return jsonError{Error: err.Error(), Code: code}
}
