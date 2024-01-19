package grpc

import (
	"fmt"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/entities"
	pb "github.com/pavlegich/metrics-alerting/internal/proto"
	"google.golang.org/grpc/codes"
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

func ConvertCodeHTTPtoGRPC(code int) codes.Code {
	switch code {
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusNotImplemented:
		return codes.Unimplemented
	default:
		return codes.Unknown
	}
}

// func ConvertFromGRPCToMetrics(pbMetric *pb.Metric) (entities.Metrics, error) {
// 	metric := entities.Metrics{
// 		ID: pbMetric.Id,
// 		MType: pbMetric.Type,
// 	}

// 	switch metric.MType {
// 	case "gauge":
// 		metric.Value = &pbMetric.Value
// 	case "counter":
// 		pbMetric.Delta = *metric.Delta
// 	default:
// 		return nil, fmt.Errorf("ConvertFromMetricsToGRPC: invalid metric type %s", metric.MType)
// 	}

// 	return pbMetric, nil
// }
