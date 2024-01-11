package interceptors

import (
	"context"
	"net"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// WithStreamNetworking проверяет соответствие IP адреса клиента указанному доверенному диапозону.
func WithStreamNetworking(network *net.IPNet) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if network == nil {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			logger.Log.Info("WithStreamNetworking: get metadata error")
		}

		ipStr := md.Get("X-Real-IP")
		if len(ipStr) == 0 {
			logger.Log.Error("WithStreamNetworking: couldn't obtain ip value")
			return status.Errorf(codes.InvalidArgument, "couldn't obtain ip value")
		}
		ip := net.ParseIP(ipStr[0])
		if ip == nil || !network.Contains(ip) {
			logger.Log.Error("WithStreamNetworking: IP not in trusted subnet",
				zap.String("ip", ip.String()))
			return status.Errorf(codes.Unavailable, "IP %s not in trusted subnet", ip.String())
		}

		return handler(srv, ss)
	}
}

// WithUnaryNetworking проверяет соответствие IP адреса клиента указанному доверенному диапозону.
func WithUnaryNetworking(network *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if network == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Log.Info("WithUnaryNetworking: get metadata error")
		}

		ipStr := md.Get("X-Real-IP")
		if len(ipStr) == 0 {
			logger.Log.Error("WithUnaryNetworking: couldn't obtain ip value")
			return nil, status.Errorf(codes.InvalidArgument, "couldn't obtain ip value")
		}
		ip := net.ParseIP(ipStr[0])
		if ip == nil || !network.Contains(ip) {
			logger.Log.Error("WithUnaryNetworking: IP not in trusted subnet",
				zap.String("ip", ip.String()))
			return nil, status.Errorf(codes.Unavailable, "IP %s not in trusted subnet", ip.String())
		}

		return handler(ctx, req)
	}
}
