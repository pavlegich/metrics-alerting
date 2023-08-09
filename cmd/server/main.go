package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/sirupsen/logrus"
)

// функция run запускает сервер
func run() error {
	// Считывание флага адреса и его запись в структуру
	addr := storage.NewAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "HTTP-server endpoint address host:port")
	flag.Parse()

	// Проверяем переменную окружения ADDRESS
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr.Set(envAddr)
	}

	// Создание нового логера
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(log, memStorage)

	r := chi.NewRouter()
	r.Mount("/", webhook.Route())

	log.Info("Server is running on ", addr.String(), "...")

	return http.ListenAndServe(addr.String(), r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
