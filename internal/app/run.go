// Пакет app содержит основные методы для запуска сервера.
package app

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/pavlegich/metrics-alerting/internal/proto"
	"google.golang.org/grpc"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/database"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	grpcserver "github.com/pavlegich/metrics-alerting/internal/server/grpcserver/handlers"
	"github.com/pavlegich/metrics-alerting/internal/server/httpserver"
	"github.com/pavlegich/metrics-alerting/internal/server/httpserver/handlers"
	"github.com/pavlegich/metrics-alerting/internal/server/httpserver/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"go.uber.org/zap"
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

	// Контроллер
	webhook := handlers.NewWebhook(ctx, memStorage, database, file, cfg)

	// Ключ для хеширования
	if cfg.Key != "" {
		entities.Key = cfg.Key
	}

	// Получение метрик из базы данных или файла
	if cfg.Restore {
		switch {
		case cfg.Database != "":
			if err := database.Load(ctx, webhook.MemStorage); err != nil {
				logger.Log.Error("Run: restore storage from database failed", zap.Error(err))
			}
		case cfg.StoragePath != "":
			if err := file.Load(ctx, webhook.MemStorage); err != nil {
				logger.Log.Error("Run: restore storage from file failed", zap.Error(err))
			}
		}
	}

	// Хранение данных в базе данных или файле
	if cfg.Database != "" || cfg.StoragePath != "" {
		saveFunc := httpserver.SaveToFileRoutine
		if cfg.Database != "" {
			saveFunc = httpserver.SaveToDBRoutine
		}

		wg.Add(1)
		go func() {
			saveFunc(ctx, webhook, storeInterval)
			wg.Done()
		}()
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

	// gRPC
	if cfg.Grpc != "" {
		go func() {
			controller := grpcserver.NewController(ctx, memStorage, database, file)

			var opts []grpc.ServerOption
			// opts = append(opts, )

			serv := grpc.NewServer(opts...)
			pb.RegisterWebhookServer(serv, controller)
			listen, err := net.Listen("tcp", cfg.Grpc)
			if err != nil {
				logger.Log.Error("Run: announce listen failed", zap.Error(err))
			}
			logger.Log.Info("running gRPC server", zap.String("addr", cfg.Grpc))
			err = serv.Serve(listen)
			if err != nil {
				logger.Log.Error("Run: serve grpc failed", zap.Error(err))
			}
		}()
	}

	// Сервер
	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	// Завершение программы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigs
		ctxShutDown, cancelShutDown := context.WithTimeout(ctx, 5*time.Second)
		defer cancelShutDown()

		if err := srv.Shutdown(ctxShutDown); err != nil {
			logger.Log.Error("server shutdown failed",
				zap.Error(err))
		}
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

	logger.Log.Info("running server", zap.String("addr", cfg.Address))

	return srv.ListenAndServe()
}
