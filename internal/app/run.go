package app

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/models"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"go.uber.org/zap"
)

// функция run запускает сервер
func Run() error {
	// Считывание флага адреса и его запись в структуру
	addr := models.NewAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "HTTP-server endpoint address host:port")
	flag.Parse()

	// Проверяем переменную окружения ADDRESS
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr.Set(envAddr)
	}

	if err := logger.Initialize("Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(memStorage)

	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route())

	logger.Log.Info("Running server", zap.String("address", addr.String()))

	return http.ListenAndServe(addr.String(), r)
}
