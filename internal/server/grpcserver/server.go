// Пакет grpcserver содержит объект Server и его методы
package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	pb "github.com/pavlegich/metrics-alerting/internal/proto"
	ctrl "github.com/pavlegich/metrics-alerting/internal/server/grpcserver/handlers"
	"github.com/pavlegich/metrics-alerting/internal/server/grpcserver/interceptors"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	config *config.ServerConfig
}

func NewServer(ctx context.Context, memStorage *storage.MemStorage,
	database *storage.Database, file *storage.File, cfg *config.ServerConfig) interfaces.Server {
	controller := ctrl.NewController(ctx, memStorage, database, file)
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor(
		interceptors.WithUnaryLogging,
		interceptors.WithUnaryNetworking(cfg.Network),
	))
	opts = append(opts, grpc.ChainStreamInterceptor(
		interceptors.WithStreamLogging,
		interceptors.WithStreamNetworking(cfg.Network),
	))

	srv := grpc.NewServer(opts...)
	pb.RegisterMetricsServer(srv, controller)

	return &Server{
		server: srv,
		config: cfg,
	}
}

func (s *Server) GetAddress(ctx context.Context) string {
	return s.config.Grpc
}

func (s *Server) Serve(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.config.Grpc)
	if err != nil {
		return fmt.Errorf("Serve: announce listen failed %w", err)
	}
	return s.server.Serve(listen)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.server.Stop()
	// Будет ожидать отсутствия передачи данных.
	// Нужно агента тоже отключать, чтобы закрылся сервер.
	// s.server.GracefulStop()
	return nil
}
