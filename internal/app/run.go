package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/server"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"go.uber.org/zap"
)

// функция run запускает сервер
func Run() error {
	logger.Log.Info("logger initialization")
	if err := logger.Initialize("Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	logger.Log.Info("flags parse")
	// Считывание флагов
	cfg, err := server.ParseFlags()
	if err != nil {
		logger.Log.Info("parse flags error")
	}

	logger.Log.Info("intervals set")
	var storeInterval time.Duration
	if cfg.StoreInterval == 0 {
		storeInterval = time.Duration(1) * time.Second
	} else {
		storeInterval = time.Duration(cfg.StoreInterval) * time.Second
	}

	logger.Log.Info("storage creating")
	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(memStorage)

	logger.Log.Info("load storage")
	// Загрузка данных из файла
	if cfg.Restore {
		var err error
		err = storage.Load(cfg.StoragePath, &webhook.MemStorage)
		if err != nil {
			return err
		}
	}

	logger.Log.Info("metrics routine")
	if cfg.StoragePath != "" {
		go server.MetricsRoutine(webhook, storeInterval, cfg.StoragePath)
	}

	logger.Log.Info("new router")
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route())

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}
