package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/database"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/server"
	"github.com/pavlegich/metrics-alerting/internal/server/handlers"
	"github.com/pavlegich/metrics-alerting/internal/server/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"go.uber.org/zap"
)

// Run запускает сервер
func Run() error {
	ctx := context.Background()

	// Логгер
	if err := logger.Initialize(ctx, "Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	// Флаги
	cfg, err := server.ParseFlags(ctx)
	if err != nil {
		return fmt.Errorf("Run: parse flags error %w", err)
	}

	// Интервалы
	var storeInterval time.Duration
	if cfg.StoreInterval == 0 {
		storeInterval = time.Duration(1) * time.Second
	} else {
		storeInterval = time.Duration(cfg.StoreInterval) * time.Second
	}

	// Хранилище
	memStorage := storage.NewMemStorage(ctx)

	// База данных
	var db *sql.DB
	if cfg.Database != "" {
		db, err = database.Init(ctx, cfg.Database)
		if err != nil {
			logger.Log.Error("Run: database open failed", zap.Error(err))
		}
		defer db.Close()
	} else {
		db = nil
	}

	// Контроллер
	webhook := handlers.NewWebhook(ctx, memStorage, db)

	// Ключ
	if cfg.Key != "" {
		entities.Key = cfg.Key
	}

	// Файл
	if cfg.Restore {
		switch {
		case cfg.Database != "":
			if err := storage.LoadFromDB(ctx, webhook.Database, webhook.MemStorage); err != nil {
				logger.Log.Error("Run: restore storage from database failed", zap.Error(err))
			}
		case cfg.StoragePath != "":
			if err := storage.LoadFromFile(ctx, cfg.StoragePath, webhook.MemStorage); err != nil {
				logger.Log.Error("Run: restore storage from file failed", zap.Error(err))
			}
		}
	}

	// Хранение данных в базе данных или файле
	switch {
	case cfg.Database != "":
		go server.SaveToDBRoutine(ctx, webhook, storeInterval)
	case cfg.StoragePath != "":
		go server.SaveToFileRoutine(ctx, webhook, storeInterval, cfg.StoragePath)
	}

	// Роутер
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route(ctx))

	logger.Log.Info("running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}
