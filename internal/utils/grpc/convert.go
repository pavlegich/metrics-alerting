package grpc

import (
	"fmt"

	"github.com/pavlegich/metrics-alerting/internal/entities"
	pb "github.com/pavlegich/metrics-alerting/internal/proto"
)

func ConvertFromMetricsToGRPC(metric entities.Metrics) (*pb.Metric, error) {
	pbMetric := &pb.Metric{
		Id:   metric.ID,
		Type: metric.MType,
	}
	switch metric.MType {
	case "gauge":
		pbMetric.Value = *metric.Value
	case "counter":
		pbMetric.Delta = *metric.Delta
	default:
		return nil, fmt.Errorf("ConvertFromMetricsToGRPC: invalid metric type %s", metric.MType)
	}

	return pbMetric, nil
}
