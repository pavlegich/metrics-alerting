package main

import (
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/handlers"
)

// функция run запускает сервер
func run() error {
	return http.ListenAndServe(`:8080`, http.HandlerFunc(handlers.Webhook))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
