package server

import (
	"encoding/json"
	"net/http"
)

type httpError struct {
	Error error `json:"error"`
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func handleFunc(f handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusInternalServerError, httpError{Error: err})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func withLogin(f handlerFunc) handlerFunc
