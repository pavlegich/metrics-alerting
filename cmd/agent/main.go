package main

import (
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func main() {
	var StatsStorage = storage.NewStatsStorage()

	var memStats runtime.MemStats
	var pollInterval = time.Duration(2) * time.Second
	// var reportInterval = time.Duration(10) * time.Second
	pollCount := 0
	randomValue := rand.Float64()
	if err := StatsStorage.Update(memStats, pollCount, randomValue); err != nil {
		log.Fatal(err)
	}
	if status := StatsStorage.Send("http://localhost:8080/update"); status != http.StatusOK {
		log.Fatal(status)
	}
	for {
		time.Sleep(pollInterval)
		runtime.ReadMemStats(&memStats)
		pollCount += 1
		randomValue = rand.Float64()
		if err := StatsStorage.Update(memStats, pollCount, randomValue); err != nil {
			log.Fatal(err)
		}
		if pollCount%5 == 0 {
			if status := StatsStorage.Send("http://localhost:8080/update"); status != http.StatusOK {
				log.Fatal(status)
			}
		}
	}
}
