package storage

import (
	"fmt"
	"net/http"
	"runtime"
)

type (
	Stats struct {
		stype string
		name  string
		value string
	}
)

func UpdateStats(st map[string]Stats, memStats runtime.MemStats, count int, rand float64) error {

	st["Alloc"] = Stats{
		stype: "gauge",
		name:  "Alloc",
		value: fmt.Sprintf("%v", memStats.Alloc),
	}
	st["BuckHashSys"] = Stats{
		stype: "gauge",
		name:  "BuckHashSys",
		value: fmt.Sprintf("%v", memStats.BuckHashSys),
	}
	st["Frees"] = Stats{
		stype: "gauge",
		name:  "Frees",
		value: fmt.Sprintf("%v", memStats.Frees),
	}
	st["GCCPUFraction"] = Stats{
		stype: "gauge",
		name:  "GCCPUFraction",
		value: fmt.Sprintf("%v", memStats.GCCPUFraction),
	}
	st["GCSys"] = Stats{
		stype: "gauge",
		name:  "GCSys",
		value: fmt.Sprintf("%v", memStats.GCSys),
	}
	st["HeapAlloc"] = Stats{
		stype: "gauge",
		name:  "HeapAlloc",
		value: fmt.Sprintf("%v", memStats.HeapAlloc),
	}
	st["HeapIdle"] = Stats{
		stype: "gauge",
		name:  "HeapIdle",
		value: fmt.Sprintf("%v", memStats.HeapIdle),
	}
	st["HeapInuse"] = Stats{
		stype: "gauge",
		name:  "HeapInuse",
		value: fmt.Sprintf("%v", memStats.HeapInuse),
	}
	st["HeapObjects"] = Stats{
		stype: "gauge",
		name:  "HeapObjects",
		value: fmt.Sprintf("%v", memStats.HeapObjects),
	}
	st["HeapReleased"] = Stats{
		stype: "gauge",
		name:  "HeapReleased",
		value: fmt.Sprintf("%v", memStats.HeapReleased),
	}
	st["HeapSys"] = Stats{
		stype: "gauge",
		name:  "HeapSys",
		value: fmt.Sprintf("%v", memStats.HeapSys),
	}
	st["LastGC"] = Stats{
		stype: "gauge",
		name:  "LastGC",
		value: fmt.Sprintf("%v", memStats.LastGC),
	}
	st["Lookups"] = Stats{
		stype: "gauge",
		name:  "Lookups",
		value: fmt.Sprintf("%v", memStats.Lookups),
	}
	st["MCacheInuse"] = Stats{
		stype: "gauge",
		name:  "MCacheInuse",
		value: fmt.Sprintf("%v", memStats.MCacheInuse),
	}
	st["MCacheSys"] = Stats{
		stype: "gauge",
		name:  "MCacheSys",
		value: fmt.Sprintf("%v", memStats.MCacheSys),
	}
	st["MSpanInuse"] = Stats{
		stype: "gauge",
		name:  "MSpanInuse",
		value: fmt.Sprintf("%v", memStats.MSpanInuse),
	}
	st["MSpanSys"] = Stats{
		stype: "gauge",
		name:  "MSpanSys",
		value: fmt.Sprintf("%v", memStats.MSpanSys),
	}
	st["Mallocs"] = Stats{
		stype: "gauge",
		name:  "Mallocs",
		value: fmt.Sprintf("%v", memStats.Mallocs),
	}
	st["NextGC"] = Stats{
		stype: "gauge",
		name:  "NextGC",
		value: fmt.Sprintf("%v", memStats.NextGC),
	}
	st["NumForcedGC"] = Stats{
		stype: "gauge",
		name:  "NumForcedGC",
		value: fmt.Sprintf("%v", memStats.NumForcedGC),
	}
	st["NumGC"] = Stats{
		stype: "gauge",
		name:  "NumGC",
		value: fmt.Sprintf("%v", memStats.NumGC),
	}
	st["OtherSys"] = Stats{
		stype: "gauge",
		name:  "OtherSys",
		value: fmt.Sprintf("%v", memStats.OtherSys),
	}
	st["PauseTotalNs"] = Stats{
		stype: "gauge",
		name:  "PauseTotalNs",
		value: fmt.Sprintf("%v", memStats.PauseTotalNs),
	}
	st["StackInuse"] = Stats{
		stype: "gauge",
		name:  "StackInuse",
		value: fmt.Sprintf("%v", memStats.StackInuse),
	}
	st["StackSys"] = Stats{
		stype: "gauge",
		name:  "StackSys",
		value: fmt.Sprintf("%v", memStats.StackSys),
	}
	st["Sys"] = Stats{
		stype: "gauge",
		name:  "Sys",
		value: fmt.Sprintf("%v", memStats.Sys),
	}
	st["TotalAlloc"] = Stats{
		stype: "gauge",
		name:  "TotalAlloc",
		value: fmt.Sprintf("%v", memStats.TotalAlloc),
	}
	st["PollCount"] = Stats{
		stype: "counter",
		name:  "PollCount",
		value: fmt.Sprintf("%v", count),
	}
	st["RandomValue"] = Stats{
		stype: "gauge",
		name:  "RandomValue",
		value: fmt.Sprintf("%v", rand),
	}

	return nil
}

func SendStats(st map[string]Stats) error {
	target := ""
	for _, stat := range st {
		target = fmt.Sprintf("http://localhost:8080/update/%s/%s/%s", stat.stype, stat.name, stat.value)
		if _, err := http.Post(target, "", nil); err != nil {
			return err
		}
	}
	return nil
}
