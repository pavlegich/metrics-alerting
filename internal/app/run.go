// Пакет app содержит основные методы для запуска сервера.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
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

// Run инициализирует основные компоненты и запускает сервер.
func Run(done chan bool) error {
	ctx := context.Background()

	// Логгер
	if err := logger.Init(ctx, "Info"); err != nil {
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

	// Ключ для хеширования
	if cfg.Key != "" {
		entities.Key = cfg.Key
	}

	// Получение метрик из базы данных или файла
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

	// Профилирование
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	profile := http.Server{
		Addr:    "localhost:8081",
		Handler: mux,
	}
	go func() {
		profile.ListenAndServe()
	}()

	// Сервер
	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	logger.Log.Info("running server", zap.String("addr", cfg.Address))

	// Завершение программы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Error("server shutdown failed",
				zap.Error(err))
		}
		if err := profile.Shutdown(ctx); err != nil {
			logger.Log.Error("profile shutdown failed",
				zap.Error(err))
		}
		logger.Log.Info("shutting down gracefully",
			zap.String("signal", sig.String()))
		done <- true
	}()

	return srv.ListenAndServe()
}
