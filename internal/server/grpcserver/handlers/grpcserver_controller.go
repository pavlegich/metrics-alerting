// Пакет grpcserver содержит объект и методы
// для работы с gRPC-сервером
package grpcserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	pb "github.com/pavlegich/metrics-alerting/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Controller содержит данные для работы с grpc-сервером
type Controller struct {
	pb.UnimplementedMetricsServer

	MemStorage interfaces.MetricStorage
	Database   interfaces.Storage
	File       interfaces.Storage
}

// NewController создаёт новый контроллер для grpc-сервера
func NewController(ctx context.Context, ms interfaces.MetricStorage, db interfaces.Storage, file interfaces.Storage) *Controller {
	return &Controller{
		MemStorage: ms,
		Database:   db,
		File:       file,
	}
}

// Updates обрабатывает и сохраняет полученные метрики.
func (c *Controller) Updates(stream pb.Metrics_UpdatesServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return fmt.Errorf("Updates: recieve stream failed %w", err)
		}

		var mValue string
		switch in.Metric.Type {
		case "gauge":
			mValue = fmt.Sprint(in.Metric.Value)
		case "counter":
			mValue = fmt.Sprint(in.Metric.Delta)
		default:
			return fmt.Errorf("Updates: invalid metric type %s", in.Metric.Type)
		}

		code := c.MemStorage.Put(stream.Context(), in.Metric.Type, in.Metric.Id, mValue)
		if code != http.StatusOK {
			return fmt.Errorf("Updates: put metric code %v", code)
		}
	}
}

// Update обрабатывает и сохраняет полученную в proto-формате метрику.
// В случае успешного сохранения обработчик получает новое значение метрики
// из хранилища и отправляет в ответ метрику в proto-формате.
func (c *Controller) Update(ctx context.Context, in *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var mValue string
	switch in.Metric.Type {
	case "gauge":
		mValue = fmt.Sprint(in.Metric.Value)
	case "counter":
		mValue = fmt.Sprint(in.Metric.Delta)
	default:
		return nil, fmt.Errorf("Update: invalid metric type %s", in.Metric.Type)
	}

	code := c.MemStorage.Put(ctx, in.Metric.Type, in.Metric.Id, mValue)
	if code != http.StatusOK {
		return nil, fmt.Errorf("Update: put metric code %v", code)
	}

	pbMetric := &pb.Metric{
		Id:   in.Metric.Id,
		Type: in.Metric.Type,
	}

	mValue, code = c.MemStorage.Get(ctx, in.Metric.Type, in.Metric.Id)
	if code != http.StatusOK {
		return nil, fmt.Errorf("Value: get metric code %v", code)
	}

	switch pbMetric.Type {
	case "gauge":
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Value: couldn't parse float")
		}
		pbMetric.Value = value
	case "counter":
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Value: couldn't parse int")
		}
		pbMetric.Delta = value
	}

	return &pb.UpdateResponse{
		Metric: pbMetric,
	}, nil
}

// Value обрабатывает запрос на получение значения метрики.
// Обработчик принимает в proto-формате название и тип метрики,
// в случае успешного получения значения метрики из хранилища,
// формирует и отправляет ответ с метрикой в proto-формате.
func (c *Controller) Value(ctx context.Context, in *pb.ValueRequest) (*pb.ValueResponse, error) {
	metric, code := c.MemStorage.Get(ctx, in.Metric.Type, in.Metric.Id)
	if code != http.StatusOK {
		return nil, fmt.Errorf("Value: get metric code %v", code)
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
	}

	return &pb.ValueResponse{
		Metric: respMetric,
	}, nil
}

func (c *Controller) Ping(ctx context.Context, _ *emptypb.Empty) (*pb.PingResponse, error) {
	err := c.Database.Ping(ctx)
	if err != nil {
		return &pb.PingResponse{Ok: false}, status.Errorf(codes.Internal, "Ping: connection with database is died %s", err)
	}

	return &pb.PingResponse{Ok: true}, nil
}
