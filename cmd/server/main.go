package main

import (
	"fmt"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/app"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// Пример запуска
// go run -ldflags "-X main.buildVersion=v1.0.1 -X main.buildDate=$(date +'%Y/%m/%d') -X main.buildCommit=1d1wdd1f" main.go
func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	done := make(chan bool, 1)
	if err := app.Run(done); err != http.ErrServerClosed {
		logger.Log.Error("main: run app failed",
			zap.Error(err))
	}
	<-done
}
