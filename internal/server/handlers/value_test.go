package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func ExampleWebhook_HandleGetMetric() {
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
	h.MemStorage = &storage.MemStorage{
		Metrics: map[string]string{
			"Gauger": "124.4",
		},
	}

	// Запрос к серверу
	url := `http://localhost:8080/value/gauge/Gauger`
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	h.Route(ctx).ServeHTTP(w, req)

	// Получение ответа
	resp := w.Result()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read body failed %w", err)
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 200
	// 124.4
}

func ExampleWebhook_HandlePostValue() {
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
	h.MemStorage = &storage.MemStorage{
		Metrics: map[string]string{
			"Gauger": "124.4",
		},
	}

	// Подготовка данных для запроса
	url := `http://localhost:8080/value/`
	m := entities.Metrics{
		ID:    "Gauger",
		MType: "gauge",
	}
	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("marshal body failed %w", err)
	}

	// Запрос к серверу
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	h.Route(ctx).ServeHTTP(w, req)

	// Получение ответа
	resp := w.Result()
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read body failed %w", err)
	}

	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(resp.StatusCode)
	fmt.Println(string(respBody))

	// Output:
	// application/json
	// 200
	// {"id":"Gauger","type":"gauge","value":124.4}
}
