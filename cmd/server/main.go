package main

import (
	"github.com/pavlegich/metrics-alerting/internal/app"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"go.uber.org/zap"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Error("main: run app error",
			zap.Error(err),
		)
	}
}
