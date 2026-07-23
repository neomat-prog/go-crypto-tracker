package server

import (
	"github.com/go-chi/chi/v5"
)

func (s *Server) registerRoutes(r chi.Router) {
	r.Get("/", makeHandler(s.handleHello))
	r.Get("/accounts/{id}", makeHandler(s.handleGetAccount))
}
