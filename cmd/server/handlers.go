package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ApiError struct {
	Error string `json:"error"`
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (s *Server) handleHello(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, http.StatusOK, map[string]string{"message": "Hello"})
}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	return WriteJSON(w, http.StatusOK, map[string]string{
		"id": id,
	})
}
