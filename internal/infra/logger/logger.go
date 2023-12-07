// Пакет logger содержит объекты и методы для логирования событий.
package logger

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type (
	// ResponseData хранит сведения об ответе.
	ResponseData struct {
		Body   *bytes.Buffer
		Status int
		Size   int
	}

	// LoggingResponseWriter представляет собой реализацию http.ResponseWriter.
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

// Log - синглтон логгера событий.
var Log *zap.Logger = zap.NewNop()

// Init инициализирует синглтон логгера с необходимым уровнем логирования.
func Init(ctx context.Context, level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("Init: parse level errors %w", err)
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("Init: logger build error %w", err)
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

// WriteHeader реализует формирование заголовка ответа с захватом кода статуса.
func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode // захватываем код статуса
}

// Write реализует формирование ответа с захватом размера тела
// и самого тела.
func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return size, fmt.Errorf("Write: response write %w", err)
	}
	r.ResponseData.Size += size // захватываем размер
	r.ResponseData.Body.Write(b)
	return size, nil
}
