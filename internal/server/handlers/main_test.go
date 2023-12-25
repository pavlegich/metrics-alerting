package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func ExampleWebhook_HandleMain() {
	// Контекст
	ctx := context.Background()

	// Хранилище
	ms := storage.NewMemStorage(ctx)

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, nil, nil, cfg)

	// Запрос к серверу
	url := `http://localhost:8080/`
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	h.Route(ctx).ServeHTTP(w, req)

	// Получение ответа
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(resp.StatusCode)

	// Output:
	// text/html; charset=utf-8
	// 200
}

func BenchmarkWebhook_HandleMain(b *testing.B) {
	// Контекст
	ctx := context.Background()

	// Хранилище
	ms := storage.NewMemStorage(ctx)

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, nil, nil, cfg)

	// Запрос к серверу
	url := `http://localhost:8080/`
	req := httptest.NewRequest(http.MethodGet, url, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Route(ctx).ServeHTTP(w, req)
	}
}
