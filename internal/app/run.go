package app

import (
	"fmt"
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
		return fmt.Errorf("Run: parse flags error %w", err)
	}

	var storeInterval time.Duration
	if cfg.StoreInterval == 0 {
		storeInterval = time.Duration(1) * time.Second
	} else {
		storeInterval = time.Duration(cfg.StoreInterval) * time.Second
	}

	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(memStorage)

	// Загрузка данных из файла
	if cfg.Restore {
		if err := storage.Load(cfg.StoragePath, webhook.MemStorage); err != nil {
			return fmt.Errorf("Run: restore storage from file %w", err)
		}
	}

	if cfg.StoragePath != "" {
		go server.MetricsRoutine(webhook, storeInterval, cfg.StoragePath)
	}

	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route())

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}
