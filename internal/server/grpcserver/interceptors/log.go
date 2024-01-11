package interceptors

import (
	"context"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// WithStreamLogging логирует события из обработчиков.
func WithStreamLogging(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logger.Log.Info("stream method called",
		zap.String("method", info.FullMethod))

	err := handler(srv, ss)
	if err != nil {
		logger.Log.Error("stream error", zap.Error(err))
	}

	return err

}

// WithUnaryLogging логирует события из обработчиков.
func WithUnaryLogging(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger.Log.Info("unary method called",
		zap.String("method", info.FullMethod))

	resp, err := handler(ctx, req)

	if err != nil {
		status, ok := status.FromError(err)
		if ok {
			logger.Log.Error("unary method error",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.String("status", status.Code().String()))
		} else {
			logger.Log.Error("unary method error",
				zap.String("method", info.FullMethod),
				zap.Error(err))
		}
	} else {
		logger.Log.Info("unary method call success",
			zap.String("method", info.FullMethod))
	}

	return resp, err
}
