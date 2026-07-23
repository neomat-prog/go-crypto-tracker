package server

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type ServerOpts struct {
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Logger       *log.Logger
}

type Server struct {
	opts ServerOpts
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		opts: opts,
	}
}

func Hello(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Hello\n"))
}

func (s *Server) Start() error {
	router := chi.NewRouter()

	s.registerRoutes(router)

	server := &http.Server{
		Addr:         s.opts.ListenAddr,
		Handler:      router,
		ReadTimeout:  s.opts.ReadTimeout,
		WriteTimeout: s.opts.WriteTimeout,
	}

	s.opts.Logger.Printf("Listening on %s", s.opts.ListenAddr)

	return server.ListenAndServe()

}
