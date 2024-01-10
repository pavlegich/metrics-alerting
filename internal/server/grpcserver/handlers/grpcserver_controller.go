// Пакет grpcserver содержит объект и методы
// для работы с gRPC-сервером
package grpcserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	pb "github.com/pavlegich/metrics-alerting/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Controller struct {
	pb.UnimplementedWebhookServer

	MemStorage interfaces.MetricStorage
	Database   interfaces.Storage
	File       interfaces.Storage
}

func NewController(ctx context.Context, ms interfaces.MetricStorage, db interfaces.Storage, file interfaces.Storage) *Controller {
	return &Controller{
		MemStorage: ms,
		Database:   db,
		File:       file,
	}
}

func (c *Controller) Ping(ctx context.Context, _ *emptypb.Empty) (*pb.PingResponse, error) {
	err := c.Database.Ping(ctx)
	if err != nil {
		return &pb.PingResponse{Ok: false}, status.Errorf(codes.Internal, "Ping: connection with database is died %s", err)
	}

	return &pb.PingResponse{Ok: true}, nil
}

func (c *Controller) Updates(stream pb.Webhook_UpdatesServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return err
		}

		if in.Metric.Type != "gauge" && in.Metric.Type != "counter" {
			return fmt.Errorf("Updates: invalid metric type %s", in.Metric.Type)
		}

		c.MemStorage.Put(context.Background(), in.Metric.Type, in.Metric.Id, fmt.Sprint(in.Metric.Value))
	}
}

func (c *Controller) Value(ctx context.Context, in *pb.ValueRequest) (*pb.ValueResponse, error) {
	metric, statusCode := c.MemStorage.Get(ctx, in.Metric.Type, in.Metric.Id)
	if statusCode != http.StatusOK {
		switch statusCode {
		case http.StatusNotFound:
			return nil, status.Errorf(codes.NotFound, "Value: metric not found")
		case http.StatusNotImplemented:
			return nil, status.Errorf(codes.Unknown, "Value: unknown metric type")
		default:
			return nil, status.Errorf(codes.Internal, "Value: couldn't find metric")
		}
	}

	respMetric := &pb.Metric{
		Id:   in.Metric.Id,
		Type: in.Metric.Type,
	}

	switch in.Metric.Type {
	case "gauge":
		value, err := strconv.ParseFloat(metric, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Value: couldn't parse float")
		}
		respMetric.Value = value
	case "counter":
		value, err := strconv.ParseInt(metric, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Value: couldn't parse int")
		}
		respMetric.Delta = value
	default:
		logger.Log.Error("main: invalid metric type", zap.String("type", in.Metric.Type))
	}

	return &pb.ValueResponse{
		Metric: respMetric,
	}, nil
}
