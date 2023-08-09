package storage

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type (
	StatsStorage interface {
		Send() int
		Update() error
	}

	StatStorage struct {
		stats map[string]stat
	}

	stat struct {
		stype string
		name  string
		value string
	}
)

func NewStatsStorage() *StatStorage {
	return &StatStorage{
		stats: make(map[string]stat),
	}
}

func (st *StatStorage) Update(memStats runtime.MemStats, count int, rand float64) error {

	st.stats["Alloc"] = stat{
		stype: "gauge",
		name:  "Alloc",
		value: fmt.Sprintf("%v", memStats.Alloc),
	}
	st.stats["BuckHashSys"] = stat{
		stype: "gauge",
		name:  "BuckHashSys",
		value: fmt.Sprintf("%v", memStats.BuckHashSys),
	}
	st.stats["Frees"] = stat{
		stype: "gauge",
		name:  "Frees",
		value: fmt.Sprintf("%v", memStats.Frees),
	}
	st.stats["GCCPUFraction"] = stat{
		stype: "gauge",
		name:  "GCCPUFraction",
		value: fmt.Sprintf("%v", memStats.GCCPUFraction),
	}
	st.stats["GCSys"] = stat{
		stype: "gauge",
		name:  "GCSys",
		value: fmt.Sprintf("%v", memStats.GCSys),
	}
	st.stats["HeapAlloc"] = stat{
		stype: "gauge",
		name:  "HeapAlloc",
		value: fmt.Sprintf("%v", memStats.HeapAlloc),
	}
	st.stats["HeapIdle"] = stat{
		stype: "gauge",
		name:  "HeapIdle",
		value: fmt.Sprintf("%v", memStats.HeapIdle),
	}
	st.stats["HeapInuse"] = stat{
		stype: "gauge",
		name:  "HeapInuse",
		value: fmt.Sprintf("%v", memStats.HeapInuse),
	}
	st.stats["HeapObjects"] = stat{
		stype: "gauge",
		name:  "HeapObjects",
		value: fmt.Sprintf("%v", memStats.HeapObjects),
	}
	st.stats["HeapReleased"] = stat{
		stype: "gauge",
		name:  "HeapReleased",
		value: fmt.Sprintf("%v", memStats.HeapReleased),
	}
	st.stats["HeapSys"] = stat{
		stype: "gauge",
		name:  "HeapSys",
		value: fmt.Sprintf("%v", memStats.HeapSys),
	}
	st.stats["LastGC"] = stat{
		stype: "gauge",
		name:  "LastGC",
		value: fmt.Sprintf("%v", memStats.LastGC),
	}
	st.stats["Lookups"] = stat{
		stype: "gauge",
		name:  "Lookups",
		value: fmt.Sprintf("%v", memStats.Lookups),
	}
	st.stats["MCacheInuse"] = stat{
		stype: "gauge",
		name:  "MCacheInuse",
		value: fmt.Sprintf("%v", memStats.MCacheInuse),
	}
	st.stats["MCacheSys"] = stat{
		stype: "gauge",
		name:  "MCacheSys",
		value: fmt.Sprintf("%v", memStats.MCacheSys),
	}
	st.stats["MSpanInuse"] = stat{
		stype: "gauge",
		name:  "MSpanInuse",
		value: fmt.Sprintf("%v", memStats.MSpanInuse),
	}
	st.stats["MSpanSys"] = stat{
		stype: "gauge",
		name:  "MSpanSys",
		value: fmt.Sprintf("%v", memStats.MSpanSys),
	}
	st.stats["Mallocs"] = stat{
		stype: "gauge",
		name:  "Mallocs",
		value: fmt.Sprintf("%v", memStats.Mallocs),
	}
	st.stats["NextGC"] = stat{
		stype: "gauge",
		name:  "NextGC",
		value: fmt.Sprintf("%v", memStats.NextGC),
	}
	st.stats["NumForcedGC"] = stat{
		stype: "gauge",
		name:  "NumForcedGC",
		value: fmt.Sprintf("%v", memStats.NumForcedGC),
	}
	st.stats["NumGC"] = stat{
		stype: "gauge",
		name:  "NumGC",
		value: fmt.Sprintf("%v", memStats.NumGC),
	}
	st.stats["OtherSys"] = stat{
		stype: "gauge",
		name:  "OtherSys",
		value: fmt.Sprintf("%v", memStats.OtherSys),
	}
	st.stats["PauseTotalNs"] = stat{
		stype: "gauge",
		name:  "PauseTotalNs",
		value: fmt.Sprintf("%v", memStats.PauseTotalNs),
	}
	st.stats["StackInuse"] = stat{
		stype: "gauge",
		name:  "StackInuse",
		value: fmt.Sprintf("%v", memStats.StackInuse),
	}
	st.stats["StackSys"] = stat{
		stype: "gauge",
		name:  "StackSys",
		value: fmt.Sprintf("%v", memStats.StackSys),
	}
	st.stats["Sys"] = stat{
		stype: "gauge",
		name:  "Sys",
		value: fmt.Sprintf("%v", memStats.Sys),
	}
	st.stats["TotalAlloc"] = stat{
		stype: "gauge",
		name:  "TotalAlloc",
		value: fmt.Sprintf("%v", memStats.TotalAlloc),
	}
	st.stats["PollCount"] = stat{
		stype: "counter",
		name:  "PollCount",
		value: fmt.Sprintf("%v", count),
	}
	st.stats["RandomValue"] = stat{
		stype: "gauge",
		name:  "RandomValue",
		value: fmt.Sprintf("%v", rand),
	}

	return nil
}

func (st *StatStorage) Send(url string) int {

	for _, stat := range st.stats {
		target := strings.Join([]string{url, "update", stat.stype, stat.name, stat.value}, "/")
		url := "http://" + target
		resp, err := http.Post(url, "", nil)
		if err != nil {
			return resp.StatusCode
		}
		resp.Body.Close()
	}
	return http.StatusOK
}
