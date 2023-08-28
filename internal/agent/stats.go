package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/pavlegich/metrics-alerting/internal/models"
)

type StatStorage struct {
	stats map[string]models.Metrics
}

func NewStatStorage() *StatStorage {
	return &StatStorage{
		stats: make(map[string]models.Metrics),
	}
}

func (st *StatStorage) Put(sType string, name string, value string) error {
	switch sType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		st.stats[name] = models.Metrics{
			ID:    name,
			MType: sType,
			Value: &v,
		}
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		st.stats[name] = models.Metrics{
			ID:    name,
			MType: sType,
			Delta: &v,
		}
	}

	return nil
}

func (st *StatStorage) Update(memStats runtime.MemStats, count int, rand float64) error {
	st.Put("gauge", "Alloc", fmt.Sprintf("%v", memStats.Alloc))
	st.Put("gauge", "BuckHashSys", fmt.Sprintf("%v", memStats.BuckHashSys))
	st.Put("gauge", "Frees", fmt.Sprintf("%v", memStats.Frees))
	st.Put("gauge", "GCCPUFraction", fmt.Sprintf("%v", memStats.GCCPUFraction))
	st.Put("gauge", "GCSys", fmt.Sprintf("%v", memStats.GCSys))
	st.Put("gauge", "HeapAlloc", fmt.Sprintf("%v", memStats.HeapAlloc))
	st.Put("gauge", "HeapIdle", fmt.Sprintf("%v", memStats.HeapIdle))
	st.Put("gauge", "HeapInuse", fmt.Sprintf("%v", memStats.HeapInuse))
	st.Put("gauge", "HeapObjects", fmt.Sprintf("%v", memStats.HeapObjects))
	st.Put("gauge", "HeapReleased", fmt.Sprintf("%v", memStats.HeapReleased))
	st.Put("gauge", "HeapSys", fmt.Sprintf("%v", memStats.HeapSys))
	st.Put("gauge", "LastGC", fmt.Sprintf("%v", memStats.LastGC))
	st.Put("gauge", "Lookups", fmt.Sprintf("%v", memStats.Lookups))
	st.Put("gauge", "MCacheInuse", fmt.Sprintf("%v", memStats.MCacheInuse))
	st.Put("gauge", "MCacheSys", fmt.Sprintf("%v", memStats.MCacheSys))
	st.Put("gauge", "MSpanInuse", fmt.Sprintf("%v", memStats.MSpanInuse))
	st.Put("gauge", "MSpanSys", fmt.Sprintf("%v", memStats.MSpanSys))
	st.Put("gauge", "Mallocs", fmt.Sprintf("%v", memStats.Mallocs))
	st.Put("gauge", "NextGC", fmt.Sprintf("%v", memStats.NextGC))
	st.Put("gauge", "NumForcedGC", fmt.Sprintf("%v", memStats.NumForcedGC))
	st.Put("gauge", "NumGC", fmt.Sprintf("%v", memStats.NumGC))
	st.Put("gauge", "OtherSys", fmt.Sprintf("%v", memStats.OtherSys))
	st.Put("gauge", "PauseTotalNs", fmt.Sprintf("%v", memStats.PauseTotalNs))
	st.Put("gauge", "StackInuse", fmt.Sprintf("%v", memStats.StackInuse))
	st.Put("gauge", "StackSys", fmt.Sprintf("%v", memStats.StackSys))
	st.Put("gauge", "Sys", fmt.Sprintf("%v", memStats.Sys))
	st.Put("gauge", "TotalAlloc", fmt.Sprintf("%v", memStats.TotalAlloc))
	st.Put("gauge", "RandomValue", fmt.Sprintf("%v", rand))
	st.Put("counter", "PollCount", fmt.Sprintf("%v", count))

	return nil
}

func (st *StatStorage) Send(url string) error {
	for _, stat := range st.stats {
		target := url + "/update/"
		url := "http://" + target

		req, err := json.Marshal(stat)
		if err != nil {
			return err
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
		if err != nil {
			return err
		}

		// respDump, err := httputil.DumpResponse(resp, true)
		// if err != nil {
		// 	log.Println(err)
		// }

		// fmt.Printf("RESPONSE:\n%s", string(respDump))

		resp.Body.Close()
	}
	return nil
}

func (st *StatStorage) SendGZIP(url string) error {
	for _, stat := range st.stats {
		target := url + "/update/"
		url := "http://" + target

		req, err := json.Marshal(stat)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		if _, err := zb.Write(req); err != nil {
			return err
		}
		if err = zb.Close(); err != nil {
			return err
		}

		r, err := http.NewRequest(http.MethodPost, url, buf)
		if err != nil {
			return err
		}

		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			return err
		}

		resp.Body.Close()
	}
	return nil
}
