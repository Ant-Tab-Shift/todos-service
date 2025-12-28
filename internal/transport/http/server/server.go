package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Ant-Tab-Shift/todos-service/internal/transport/http/models"
)

type Handler interface {
	Handlers() []models.Endpoint
}

type Server struct {
	srv http.Server
}

func New(baseContext context.Context, addr string) *Server {
	return &Server{srv: http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return baseContext
		},
	}}
}

func (s *Server) RegisterHandlers(handler Handler) {
	mux := http.NewServeMux()
	for _, endpoint := range handler.Handlers() {
		mux.HandleFunc(endpoint.Pattern, endpoint.Func)
	}

	s.srv.Handler = loggingMiddleware(mux)
}

func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
