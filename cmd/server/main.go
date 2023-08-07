package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/sirupsen/logrus"
)

// функция run запускает сервер
func run() error {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	memStorage := storage.NewMemStorage()
	webhook := handlers.NewWebhook(log, memStorage)

	r := chi.NewRouter()
	r.Mount("/", webhook.Route())

	log.Info("Server is running...")

	return http.ListenAndServe(`:8080`, r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
