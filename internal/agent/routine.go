package agent

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

func StatsRoutine(st interfaces.StatsStorage, poll time.Duration, report time.Duration, addr string, c chan int) {
	tickerPoll := time.NewTicker(poll)
	tickerReport := time.NewTicker(report)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	// Runtime метрики
	var memStats runtime.MemStats

	// Дополнительные метрики
	pollCount := 0
	var randomValue float64

	for {
		select {
		case <-tickerPoll.C:
			// Обновление метрик
			runtime.ReadMemStats(&memStats)
			pollCount += 1
			randomValue = rand.Float64()

			// Опрос метрик
			if err := st.Update(memStats, pollCount, randomValue); err != nil {
				log.Println(err)
				// close(c)
			}
		case <-tickerReport.C:
			if err := st.SendGZIP(addr); err != nil {
				log.Println(err)
				// close(c)
			}

		}
	}
}
