// Пакет agent содержит методы для обновления метрик
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/crypto"
	"github.com/pavlegich/metrics-alerting/internal/infra/hash"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

// StatStorage хранит метрики агента.
type StatStorage struct {
	stats map[string]entities.Metrics
	mu    sync.Mutex
}

// NewStatStorage создаёт новый объект хранилища агента.
func NewStatStorage(ctx context.Context) *StatStorage {
	return &StatStorage{
		stats: make(map[string]entities.Metrics),
	}
}

// Put обрабатывает типы метрик gauge и counter, сохраняет их в хранилище.
func (st *StatStorage) Put(ctx context.Context, sType string, name string, value string) error {

	switch sType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("Put: parse float64 gauge %w", err)
		}
		st.mu.Lock()
		st.stats[name] = entities.Metrics{
			ID:    name,
			MType: sType,
			Value: &v,
		}
		st.mu.Unlock()
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("Put: parse int64 counter %w", err)
		}
		st.mu.Lock()
		st.stats[name] = entities.Metrics{
			ID:    name,
			MType: sType,
			Delta: &v,
		}
		st.mu.Unlock()
	}

	return nil
}

// Update сохраняет необходимые метрики runtime, счётчик и случайное число в хранилище агента.
func (st *StatStorage) Update(ctx context.Context, memStats runtime.MemStats, count int, rand float64) error {

	st.Put(ctx, "gauge", "Alloc", fmt.Sprintf("%v", memStats.Alloc))
	st.Put(ctx, "gauge", "BuckHashSys", fmt.Sprintf("%v", memStats.BuckHashSys))
	st.Put(ctx, "gauge", "Frees", fmt.Sprintf("%v", memStats.Frees))
	st.Put(ctx, "gauge", "GCCPUFraction", fmt.Sprintf("%v", memStats.GCCPUFraction))
	st.Put(ctx, "gauge", "GCSys", fmt.Sprintf("%v", memStats.GCSys))
	st.Put(ctx, "gauge", "HeapAlloc", fmt.Sprintf("%v", memStats.HeapAlloc))
	st.Put(ctx, "gauge", "HeapIdle", fmt.Sprintf("%v", memStats.HeapIdle))
	st.Put(ctx, "gauge", "HeapInuse", fmt.Sprintf("%v", memStats.HeapInuse))
	st.Put(ctx, "gauge", "HeapObjects", fmt.Sprintf("%v", memStats.HeapObjects))
	st.Put(ctx, "gauge", "HeapReleased", fmt.Sprintf("%v", memStats.HeapReleased))
	st.Put(ctx, "gauge", "HeapSys", fmt.Sprintf("%v", memStats.HeapSys))
	st.Put(ctx, "gauge", "LastGC", fmt.Sprintf("%v", memStats.LastGC))
	st.Put(ctx, "gauge", "Lookups", fmt.Sprintf("%v", memStats.Lookups))
	st.Put(ctx, "gauge", "MCacheInuse", fmt.Sprintf("%v", memStats.MCacheInuse))
	st.Put(ctx, "gauge", "MCacheSys", fmt.Sprintf("%v", memStats.MCacheSys))
	st.Put(ctx, "gauge", "MSpanInuse", fmt.Sprintf("%v", memStats.MSpanInuse))
	st.Put(ctx, "gauge", "MSpanSys", fmt.Sprintf("%v", memStats.MSpanSys))
	st.Put(ctx, "gauge", "Mallocs", fmt.Sprintf("%v", memStats.Mallocs))
	st.Put(ctx, "gauge", "NextGC", fmt.Sprintf("%v", memStats.NextGC))
	st.Put(ctx, "gauge", "NumForcedGC", fmt.Sprintf("%v", memStats.NumForcedGC))
	st.Put(ctx, "gauge", "NumGC", fmt.Sprintf("%v", memStats.NumGC))
	st.Put(ctx, "gauge", "OtherSys", fmt.Sprintf("%v", memStats.OtherSys))
	st.Put(ctx, "gauge", "PauseTotalNs", fmt.Sprintf("%v", memStats.PauseTotalNs))
	st.Put(ctx, "gauge", "StackInuse", fmt.Sprintf("%v", memStats.StackInuse))
	st.Put(ctx, "gauge", "StackSys", fmt.Sprintf("%v", memStats.StackSys))
	st.Put(ctx, "gauge", "Sys", fmt.Sprintf("%v", memStats.Sys))
	st.Put(ctx, "gauge", "TotalAlloc", fmt.Sprintf("%v", memStats.TotalAlloc))
	st.Put(ctx, "gauge", "RandomValue", fmt.Sprintf("%v", rand))
	st.Put(ctx, "counter", "PollCount", fmt.Sprintf("%v", count))

	return nil
}

// Send конвертирует метрики в JSON формат, сжимает и подписыает данные,
// формирует запрос POST на указанный адрес и отправляет данные.
func Send(ctx context.Context, target string, cfg *config.AgentConfig, stats ...entities.Metrics) error {
	req, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("Send: request marshal %w", err)
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	if _, err := zb.Write(req); err != nil {
		return fmt.Errorf("Send: write request into buffer %w", err)
	}
	if err = zb.Close(); err != nil {
		return err
	}

	// Шифрование
	if cfg.CryptoKey != "" {
		encryptedReq, err := crypto.Encrypt(buf.Bytes(), cfg.CryptoKey)
		if err != nil {
			return fmt.Errorf("Send: encrypt failed %w", err)
		}
		buf = bytes.NewBuffer(encryptedReq)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, target, buf)
	if err != nil {
		return fmt.Errorf("Send: new post request %w", err)
	}

	if cfg.Key != "" {
		hash, err := hash.Sign(buf.Bytes(), []byte(cfg.Key))
		if err != nil {
			return fmt.Errorf("Send: sign message failed %w", err)
		}
		r.Header.Set("HashSHA256", hex.EncodeToString(hash))
	}

	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Real-IP", "172.17.0.20")

	if cfg.CryptoKey != "" {
		r.Header.Set("Content-Encryption", "rsa")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("Send: response get %w", err)
	}

	resp.Body.Close()

	return nil
}

// SendBatch получает все метрики из хранилища и отправляет их по указанному адресу.
func (st *StatStorage) SendBatch(ctx context.Context, cfg *config.AgentConfig) error {
	target := "http://" + cfg.Address + "/updates/"

	// Подготовка данных
	stats := st.GetAll(ctx)

	if err := Send(ctx, target, cfg, stats...); err != nil {
		return fmt.Errorf("SendBatch: send stats error %w", err)
	}

	return nil
}

// SendJSON отправляет отдельно каждую метрику из хранилища в формате JSON
// по указанному адресу.
func (st *StatStorage) SendJSON(ctx context.Context, cfg *config.AgentConfig) error {
	for _, stat := range st.stats {
		target := "http://" + cfg.Address + "/update/"

		req, err := json.Marshal(stat)
		if err != nil {
			return fmt.Errorf("SendJSON: marshal failed %w", err)
		}

		resp, err := http.Post(target, "application/json", bytes.NewBuffer(req))
		if err != nil {
			return fmt.Errorf("SendJSON: response post %w", err)
		}

		resp.Body.Close()
	}
	return nil
}

// SendGZIP отправляет отдельно каждую метрику из хранилища
// по указанному адресу, предварительно сжимая данные.
func (st *StatStorage) SendGZIP(ctx context.Context, cfg *config.AgentConfig) error {

	for _, stat := range st.stats {
		target := "http://" + cfg.Address + "/update/"

		if err := Send(ctx, target, cfg, stat); err != nil {
			return fmt.Errorf("SendBatch: send stats error %w", err)
		}
	}
	return nil
}

func (st *StatStorage) GetAll(ctx context.Context) []entities.Metrics {
	m := []entities.Metrics{}
	st.mu.Lock()
	for _, v := range st.stats {
		m = append(m, v)
	}
	st.mu.Unlock()
	return m
}

// PollCPUstats считывает информацию о занимаемой памяти с указанным интервалом времени
// и обновляет данные в хранилище.
func PollCPUstats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
	interval := time.Duration(cfg.PollInterval) * time.Second

	for {
		v, err := mem.VirtualMemory()
		if err != nil {
			logger.Log.Error("PollGoutilStats: get virtual memory stats failed", zap.Error(err))
		}
		c, err := cpu.PercentWithContext(ctx, 0, false)
		if err != nil {
			logger.Log.Error("PollGoutilStats: get cpu stats failed", zap.Error(err))
		}

		st.Put(ctx, "gauge", "TotalMemory", fmt.Sprintf("%v", v.Total))
		st.Put(ctx, "gauge", "FreeMemory", fmt.Sprintf("%v", v.Free))
		st.Put(ctx, "gauge", "CPUutilization1", fmt.Sprintf("%v", c))

		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(interval)
		}
	}
}

// PollMemStats считывает метрики с указанным интервалом времени
// и обновляет данные в хранилище.
func PollMemStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
	// Runtime метрики
	var memStats runtime.MemStats

	// Дополнительные метрики
	pollCount := 0
	var randomValue float64

	interval := time.Duration(cfg.PollInterval) * time.Second

	for {
		// Обновление метрик
		runtime.ReadMemStats(&memStats)
		pollCount += 1
		randomValue = rand.Float64()

		// Обновление метрик
		if err := st.Update(ctx, memStats, pollCount, randomValue); err != nil {
			logger.Log.Error("PollMemStats: stats update", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(interval)
		}
	}
}
