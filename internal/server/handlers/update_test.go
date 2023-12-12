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
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func ExampleWebhook_HandlePostUpdates() {
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

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, db, cfg)

	// Запрос к серверу
	url := `http://localhost:8080/updates/`
	v := 214.4
	m := []entities.Metrics{
		{
			ID:    "Gauger1",
			MType: "gauge",
			Value: &v,
		},
		{
			ID:    "Gauger2",
			MType: "gauge",
			Value: &v,
		},
	}
	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("marshal body failed %w", err)
	}
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	h.Route(ctx).ServeHTTP(w, req)

	// Получение ответа
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

func ExampleWebhook_HandlePostMetric() {
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

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, db, cfg)

	// Запрос к серверу
	url := `http://localhost:8080/update/gauge/someMetric/10.1`
	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()
	h.Route(ctx).ServeHTTP(w, req)

	// Получение ответа
	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

func ExampleWebhook_HandlePostUpdate() {
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

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, db, cfg)

	// Подготовка данных для запроса
	url := `http://localhost:8080/update/`
	v := 214.4
	m := entities.Metrics{
		ID:    "Gauger",
		MType: "gauge",
		Value: &v,
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
	// {"id":"Gauger","type":"gauge","value":214.4}
}

func BenchmarkWebhook_HandlePostUpdates(b *testing.B) {
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

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, db, cfg)

	// Запрос к серверу
	url := `http://localhost:8080/updates/`
	v := 214.4
	m := []entities.Metrics{
		{
			ID:    "Gauger1",
			MType: "gauge",
			Value: &v,
		},
		{
			ID:    "Gauger2",
			MType: "gauge",
			Value: &v,
		},
	}
	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("marshal body failed %w", err)
	}
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Route(ctx).ServeHTTP(w, req)
	}
}

func BenchmarkWebhook_HandlePostMetric(b *testing.B) {
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

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, db, cfg)

	// Запрос к серверу
	url := `http://localhost:8080/update/gauge/someMetric/10.1`
	req := httptest.NewRequest(http.MethodPost, url, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Route(ctx).ServeHTTP(w, req)
	}
}

func BenchmarkWebhook_HandlePostUpdate(b *testing.B) {
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

	// Конфиг
	cfg := &config.ServerConfig{}

	// Контроллер
	h := NewWebhook(ctx, ms, db, cfg)

	// Подготовка данных для запроса
	url := `http://localhost:8080/update/`
	v := 214.4
	m := entities.Metrics{
		ID:    "Gauger",
		MType: "gauge",
		Value: &v,
	}
	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("marshal body failed %w", err)
	}

	// Запрос к серверу
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.Route(ctx).ServeHTTP(w, req)
	}
}
