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
	"sync"
	"syscall"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/database"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/server"
	"github.com/pavlegich/metrics-alerting/internal/server/grpcserver"
	"github.com/pavlegich/metrics-alerting/internal/server/httpserver"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"go.uber.org/zap"
	_ "google.golang.org/grpc/encoding/gzip"
)

// Run инициализирует основные компоненты и запускает сервер.
func Run(idleConnsClosed chan struct{}) error {
	ctx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()
	wg := &sync.WaitGroup{}

	// Логгер
	if err := logger.Init(ctx, "Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	// Флаги
	cfg, err := config.ServerParseFlags(ctx)
	if err != nil {
		logger.Log.Error("Run: parse flags error", zap.Error(err))
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
	database := storage.NewDatabase(db)

	// Файл
	file := storage.NewFile(cfg.StoragePath)

	// Ключ для хеширования
	if cfg.Key != "" {
		entities.Key = cfg.Key
	}

	// Получение метрик из базы данных или файла
	if cfg.Restore {
		switch {
		case cfg.Database != "":
			if err := database.Load(ctx, memStorage); err != nil {
				logger.Log.Error("Run: restore storage from database failed", zap.Error(err))
			}
		case cfg.StoragePath != "":
			if err := file.Load(ctx, memStorage); err != nil {
				logger.Log.Error("Run: restore storage from file failed", zap.Error(err))
			}
		}
	}

	// Сервер
	var srv interfaces.Server = nil
	if cfg.Grpc != "" {
		srv = grpcserver.NewServer(ctx, memStorage, database, file, cfg)
	} else if cfg.Address != "" {
		srv = httpserver.NewServer(ctx, memStorage, database, file, cfg)
	}

	if srv == nil {
		return fmt.Errorf("Run: server is nil")
	}

	// Хранение данных в базе данных или файле
	if cfg.Database != "" || cfg.StoragePath != "" {
		saveFunc := server.SaveToFileRoutine
		if cfg.Database != "" {
			saveFunc = server.SaveToDBRoutine
		}

		wg.Add(1)
		go func() {
			saveFunc(ctx, memStorage, database, file, storeInterval)
			wg.Done()
		}()
	}

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

	// Завершение программы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigs
		fmt.Println("got sig")
		ctxShutDown, cancelShutDown := context.WithTimeout(ctx, 5*time.Second)
		defer cancelShutDown()

		fmt.Println("start shutdown")
		if err := srv.Shutdown(ctxShutDown); err != nil {
			logger.Log.Error("server shutdown failed",
				zap.Error(err))
		}
		fmt.Println("end shutdown")
		if err := profile.Shutdown(ctxShutDown); err != nil {
			logger.Log.Error("profile shutdown failed",
				zap.Error(err))
		}

		cancelRun()
		logger.Log.Info("shutting down gracefully...",
			zap.String("signal", sig.String()))
		wg.Wait()
		close(idleConnsClosed)
	}()

	logger.Log.Info("running server", zap.String("addr", srv.GetAddress(ctx)))

	return srv.Serve(ctx)
}
