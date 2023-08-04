package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type (
	stats struct {
		stype string
		name  string
		value string
	}
)

var StatsStorage = map[string]stats{}

func updateStats(memStats runtime.MemStats, count int, rand float64) {

	StatsStorage["Alloc"] = stats{
		stype: "gauge",
		name:  "Alloc",
		value: fmt.Sprintf("%v", memStats.Alloc),
	}
	StatsStorage["BuckHashSys"] = stats{
		stype: "gauge",
		name:  "BuckHashSys",
		value: fmt.Sprintf("%v", memStats.BuckHashSys),
	}
	StatsStorage["Frees"] = stats{
		stype: "gauge",
		name:  "Frees",
		value: fmt.Sprintf("%v", memStats.Frees),
	}
	StatsStorage["GCCPUFraction"] = stats{
		stype: "gauge",
		name:  "GCCPUFraction",
		value: fmt.Sprintf("%v", memStats.GCCPUFraction),
	}
	StatsStorage["GCSys"] = stats{
		stype: "gauge",
		name:  "GCSys",
		value: fmt.Sprintf("%v", memStats.GCSys),
	}
	StatsStorage["HeapAlloc"] = stats{
		stype: "gauge",
		name:  "HeapAlloc",
		value: fmt.Sprintf("%v", memStats.HeapAlloc),
	}
	StatsStorage["HeapIdle"] = stats{
		stype: "gauge",
		name:  "HeapIdle",
		value: fmt.Sprintf("%v", memStats.HeapIdle),
	}
	StatsStorage["HeapInuse"] = stats{
		stype: "gauge",
		name:  "HeapInuse",
		value: fmt.Sprintf("%v", memStats.HeapInuse),
	}
	StatsStorage["HeapObjects"] = stats{
		stype: "gauge",
		name:  "HeapObjects",
		value: fmt.Sprintf("%v", memStats.HeapObjects),
	}
	StatsStorage["HeapReleased"] = stats{
		stype: "gauge",
		name:  "HeapReleased",
		value: fmt.Sprintf("%v", memStats.HeapReleased),
	}
	StatsStorage["HeapSys"] = stats{
		stype: "gauge",
		name:  "HeapSys",
		value: fmt.Sprintf("%v", memStats.HeapSys),
	}
	StatsStorage["LastGC"] = stats{
		stype: "gauge",
		name:  "LastGC",
		value: fmt.Sprintf("%v", memStats.LastGC),
	}
	StatsStorage["Lookups"] = stats{
		stype: "gauge",
		name:  "Lookups",
		value: fmt.Sprintf("%v", memStats.Lookups),
	}
	StatsStorage["MCacheInuse"] = stats{
		stype: "gauge",
		name:  "MCacheInuse",
		value: fmt.Sprintf("%v", memStats.MCacheInuse),
	}
	StatsStorage["MCacheSys"] = stats{
		stype: "gauge",
		name:  "MCacheSys",
		value: fmt.Sprintf("%v", memStats.MCacheSys),
	}
	StatsStorage["MSpanInuse"] = stats{
		stype: "gauge",
		name:  "MSpanInuse",
		value: fmt.Sprintf("%v", memStats.MSpanInuse),
	}
	StatsStorage["MSpanSys"] = stats{
		stype: "gauge",
		name:  "MSpanSys",
		value: fmt.Sprintf("%v", memStats.MSpanSys),
	}
	StatsStorage["Mallocs"] = stats{
		stype: "gauge",
		name:  "Mallocs",
		value: fmt.Sprintf("%v", memStats.Mallocs),
	}
	StatsStorage["NextGC"] = stats{
		stype: "gauge",
		name:  "NextGC",
		value: fmt.Sprintf("%v", memStats.NextGC),
	}
	StatsStorage["NumForcedGC"] = stats{
		stype: "gauge",
		name:  "NumForcedGC",
		value: fmt.Sprintf("%v", memStats.NumForcedGC),
	}
	StatsStorage["NumGC"] = stats{
		stype: "gauge",
		name:  "NumGC",
		value: fmt.Sprintf("%v", memStats.NumGC),
	}
	StatsStorage["OtherSys"] = stats{
		stype: "gauge",
		name:  "OtherSys",
		value: fmt.Sprintf("%v", memStats.OtherSys),
	}
	StatsStorage["PauseTotalNs"] = stats{
		stype: "gauge",
		name:  "PauseTotalNs",
		value: fmt.Sprintf("%v", memStats.PauseTotalNs),
	}
	StatsStorage["StackInuse"] = stats{
		stype: "gauge",
		name:  "StackInuse",
		value: fmt.Sprintf("%v", memStats.StackInuse),
	}
	StatsStorage["StackSys"] = stats{
		stype: "gauge",
		name:  "StackSys",
		value: fmt.Sprintf("%v", memStats.StackSys),
	}
	StatsStorage["Sys"] = stats{
		stype: "gauge",
		name:  "Sys",
		value: fmt.Sprintf("%v", memStats.Sys),
	}
	StatsStorage["TotalAlloc"] = stats{
		stype: "gauge",
		name:  "TotalAlloc",
		value: fmt.Sprintf("%v", memStats.TotalAlloc),
	}
	StatsStorage["PollCount"] = stats{
		stype: "counter",
		name:  "PollCount",
		value: fmt.Sprintf("%v", count),
	}
	StatsStorage["RandomValue"] = stats{
		stype: "gauge",
		name:  "RandomValue",
		value: fmt.Sprintf("%v", rand),
	}
}

func sendStats() {
	target := ""
	for _, stat := range StatsStorage {
		target = fmt.Sprintf("http://localhost:8080/update/%s/%s/%s", stat.stype, stat.name, stat.value)
		http.Post(target, "", nil)
	}
}

func main() {
	var memStats runtime.MemStats
	var pollInterval = time.Duration(2) * time.Second
	// var reportInterval = time.Duration(10) * time.Second
	pollCount := 0
	randomValue := rand.Float64()
	updateStats(memStats, pollCount, randomValue)
	sendStats()

	for {
		time.Sleep(pollInterval)
		runtime.ReadMemStats(&memStats)
		pollCount += 1
		randomValue = rand.Float64()
		updateStats(memStats, pollCount, randomValue)

		if pollCount%5 == 0 {
			sendStats()
		}
	}

	// metricsUint32 := []uint32{
	// 	memStats.NumForcedGC,
	// 	memStats.NumGC,
	// }

	// metricsFloat64 := []float64{
	// 	memStats.GCCPUFraction,
	// }
}
