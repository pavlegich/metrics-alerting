package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func ExampleWebhook_HandleMain() {
	// Контекст
	ctx := context.Background()

	// Хранилище
	ms := storage.NewMemStorage(ctx)

	// База данных
	ps := "postgresql://localhost:5432/metrics"
	db, err := sql.Open("pgx", ps)
	if err != nil {
		fmt.Println("database open failed %w", err)
	}
	defer db.Close()

	// Контроллер
	h := NewWebhook(ctx, ms, db)

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

	// База данных
	ps := "postgresql://localhost:5432/metrics"
	db, err := sql.Open("pgx", ps)
	if err != nil {
		fmt.Println("database open failed %w", err)
	}
	defer db.Close()

	// Контроллер
	h := NewWebhook(ctx, ms, db)

	// Запрос к серверу
	url := `http://localhost:8080/`
	req := httptest.NewRequest(http.MethodGet, url, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Route(ctx).ServeHTTP(w, req)
	}
}
