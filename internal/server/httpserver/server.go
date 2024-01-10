// Пакет httpserver содержит объект Server и его методы
package httpserver

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	ctrl "github.com/pavlegich/metrics-alerting/internal/server/httpserver/handlers"
	"github.com/pavlegich/metrics-alerting/internal/server/httpserver/middlewares"
)

type Server struct {
	server http.Server
	config *config.ServerConfig
}

func NewServer(ctx context.Context, memStorage interfaces.MetricStorage, database interfaces.Storage,
	file interfaces.Storage, cfg *config.ServerConfig) interfaces.Server {
	controller := ctrl.NewWebhook(ctx, memStorage, database, file, cfg)

	// Роутер
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", controller.Route(ctx))

	// Сервер
	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	return &Server{
		server: srv,
		config: cfg,
	}
}

func (s *Server) GetAddress(ctx context.Context) string {
	return s.config.Address
}

func (s *Server) Serve(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
