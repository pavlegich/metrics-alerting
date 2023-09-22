package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/pavlegich/metrics-alerting/internal/infra/sign"
	"github.com/pavlegich/metrics-alerting/internal/models"
)

type StatStorage struct {
	stats map[string]models.Metrics
}

func NewStatStorage(ctx context.Context) *StatStorage {
	return &StatStorage{
		stats: make(map[string]models.Metrics),
	}
}

func (st *StatStorage) Put(ctx context.Context, sType string, name string, value string) error {
	switch sType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("Put: parse float64 gauge %w", err)
		}
		st.stats[name] = models.Metrics{
			ID:    name,
			MType: sType,
			Value: &v,
		}
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("Put: parse int64 counter %w", err)
		}
		st.stats[name] = models.Metrics{
			ID:    name,
			MType: sType,
			Delta: &v,
		}
	}

	return nil
}

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

func Send(ctx context.Context, target string, key string, stats ...models.Metrics) error {
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

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, target, buf)
	if err != nil {
		return fmt.Errorf("Send: new post request %w", err)
	}

	if key != "" {
		hash, err := sign.Sign(buf.Bytes(), []byte(key))
		if err != nil {
			return fmt.Errorf("Send: sign message failed %w", err)
		}
		r.Header.Set("HashSHA256", hex.EncodeToString(hash))
	}
	fmt.Printf("HashSHA256: '%s'\n", r.Header.Get("HashSHA256"))

	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")
	r.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("Send: response get %w", err)
	}

	resp.Body.Close()

	return nil
}

func (st *StatStorage) SendBatch(ctx context.Context, url string, key string) error {

	target := "http://" + url + "/updates/"

	// Подготовка данных
	stats := make([]models.Metrics, 0)
	for _, s := range st.stats {
		stats = append(stats, s)
	}

	if err := Send(ctx, target, key, stats...); err != nil {
		return fmt.Errorf("SendBatch: send stats error %w", err)
	}

	return nil
}

func (st *StatStorage) SendJSON(ctx context.Context, url string, key string) error {
	for _, stat := range st.stats {
		target := "http://" + url + "/update/"

		req, err := json.Marshal(stat)
		if err != nil {
			return fmt.Errorf("Send: request marshal %w", err)
		}

		resp, err := http.Post(target, "application/json", bytes.NewBuffer(req))
		if err != nil {
			return fmt.Errorf("Send: response post %w", err)
		}

		resp.Body.Close()
	}
	return nil
}

func (st *StatStorage) SendGZIP(ctx context.Context, url string, key string) error {

	for _, stat := range st.stats {
		target := "http://" + url + "/update/"

		if err := Send(ctx, target, key, stat); err != nil {
			return fmt.Errorf("SendBatch: send stats error %w", err)
		}
	}
	return nil
}
