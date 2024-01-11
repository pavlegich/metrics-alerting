// Пакет grpcagent содержит методы для работы
// с grpc-агентом.
package grpcagent

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	pb "github.com/pavlegich/metrics-alerting/internal/proto"
	utils "github.com/pavlegich/metrics-alerting/internal/utils/grpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

type Agent struct {
}

// SendStats создаёт worker-ы и отправляет данные из хранилища в работу worker-ам
// через канал с указанным интервалом.
func (a *Agent) SendStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
	interval := time.Duration(time.Duration(cfg.ReportInterval) * time.Second)

	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))

	conn, err := grpc.Dial(cfg.Grpc, opts...)

	if err != nil {
		logger.Log.Error("SendStats: create client connection failed", zap.Error(err))
	}
	client := pb.NewMetricsClient(conn)
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			jobs := make(chan interfaces.StatsStorage)
			g := new(errgroup.Group)

			// метадата, наподобие headers в http запросе
			md := metadata.Pairs(
				"X-Real-IP", cfg.IP,
			)
			mdCtx := metadata.NewOutgoingContext(ctx, md)

			stream, err := client.Updates(mdCtx)
			if err != nil {
				logger.Log.Error("SendStats: make client stream failed", zap.Error(err))
				return
			}
			for w := 1; w <= cfg.RateLimit; w++ {
				g.Go(func() error {
					return sendWorker(ctx, cfg, jobs, stream)
				})
			}
			jobs <- st
			close(jobs)
			if err := g.Wait(); err != nil {
				logger.Log.Error("SendStats: sendWorker run failed",
					zap.Error(err))
			}
			stream.CloseSend()
		}
		time.Sleep(interval)
	}
}

// sendWorker принимает метрики из канала и отправляет их по указанному адресу.
// Если соединение с сервером получить не удаётся, прерывает отправку метрик.
func sendWorker(ctx context.Context, cfg *config.AgentConfig, jobs <-chan interfaces.StatsStorage, stream pb.Metrics_UpdatesClient) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case job, ok := <-jobs:
			if !ok {
				return nil
			}
			// берутся все метрики и по очереди обрабатываются
			metrics := job.GetAll(ctx)
			for _, m := range metrics {
				// конвертация в pb.Metric формат
				pbMetric, err := utils.ConvertFromMetricsToGRPC(m)
				if err != nil {
					return fmt.Errorf("sendWorker: convert to grpc metric failed %w", err)
				}

				// попытка отправки метрики
				intervals := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}
				for _, interval := range intervals {
					time.Sleep(interval)
					err = stream.Send(&pb.UpdatesRequest{Metric: pbMetric})
					if !errors.Is(err, syscall.ECONNREFUSED) {
						break
					}
				}
				if err != nil {
					return fmt.Errorf("sendWorker: send metric in stream failed %w", err)
				}
			}
		}
	}
}
