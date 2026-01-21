// В пакете server создаем "оболочку" (wrapper)
// вокруг встроенного метода http.Server
package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Проверяем: вдруг Run и Shutdown вызываются почти одновременно (или Run не успел создать httpServer).
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
