package main

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/storage"
)

var StatsStorage = map[string]storage.Stats{}

func main() {
	var memStats runtime.MemStats
	var pollInterval = time.Duration(2) * time.Second
	// var reportInterval = time.Duration(10) * time.Second
	pollCount := 0
	randomValue := rand.Float64()
	if err := storage.UpdateStats(StatsStorage, memStats, pollCount, randomValue); err != nil {
		log.Fatal(err)
	}
	if err := storage.SendStats(StatsStorage); err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(pollInterval)
		runtime.ReadMemStats(&memStats)
		pollCount += 1
		randomValue = rand.Float64()
		if err := storage.UpdateStats(StatsStorage, memStats, pollCount, randomValue); err != nil {
			log.Fatal(err)
		}
		if pollCount%5 == 0 {
			if err := storage.SendStats(StatsStorage); err != nil {
				log.Fatal(err)
			}
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
