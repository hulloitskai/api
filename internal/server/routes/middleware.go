package routes

import (
	"net/http"
)

// corsMiddleware sets the "Access-Control-Allow-Origin" to "*" for all GET
// requests.
//
// If the header already set by handle, corsMiddleware does nothing.
func corsMiddleware(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle(w, r)
		if (r.Method != "GET") ||
			(w.Header().Get("Access-Control-Allow-Origin") != "") {
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
}
