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
	if err := logger.Initialize("Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	// Считывание флагов
	cfg, err := server.ParseFlags()
	if err != nil {
		logger.Log.Info("parse flags error")
	}

	var storeInterval time.Duration
	if cfg.StoreInterval == 0 {
		storeInterval = time.Duration(1) * time.Second
	} else {
		storeInterval = time.Duration(cfg.StoreInterval) * time.Second
	}

	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Загрузка данных из файла
	if cfg.Restore {
		var err error
		memStorage, err = storage.LoadStorage(cfg.StoragePath)
		if err != nil {
			return err
		}
	}

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(memStorage)

	if cfg.StoragePath != "" {
		go storeMetricsRoutine(webhook, storeInterval, cfg.StoragePath)
	}

	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route())

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}

func storeMetricsRoutine(wh *handlers.Webhook, store time.Duration, path string) error {
	for {
		if err := storage.SaveStorage(path, &wh.MemStorage); err != nil {
			return err
		}
		time.Sleep(store)
	}
}
