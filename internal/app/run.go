package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	// Установка интервалов
	var storeInterval time.Duration
	if cfg.StoreInterval == 0 {
		storeInterval = time.Duration(1) * time.Second
	} else {
		storeInterval = time.Duration(cfg.StoreInterval) * time.Second
	}

	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Инициализация базы данных
	db, err := storage.NewDatabase(cfg.Database)
	if err != nil {
		logger.Log.Error("Run: database open failed", zap.Error(err))
	}
	defer db.Close()

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(memStorage, db)

	// Загрузка данных из файла
	if cfg.Restore {
		switch {
		case cfg.Database != "":
			if err := storage.LoadFromDB(webhook.Database, webhook.MemStorage); err != nil {
				logger.Log.Error("Run: restore storage from database failed", zap.Error(err))
			}
		case cfg.StoragePath != "":
			if err := storage.LoadFromFile(cfg.StoragePath, webhook.MemStorage); err != nil {
				logger.Log.Error("Run: restore storage from file failed", zap.Error(err))
			}
		}
	}

	// Хранение данных в базе данных или файле
	if cfg.Database != "" {
		go server.SaveToDBRoutine(webhook, storeInterval)
	}
	if cfg.StoragePath != "" {
		go server.SaveToFileRoutine(webhook, storeInterval, cfg.StoragePath)
	}

	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route())

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}
