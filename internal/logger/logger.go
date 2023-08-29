package logger

import (
	"bytes"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type (
	// берём структуру для хранения сведений об ответе
	ResponseData struct {
		Status int
		Size   int
		Body   *bytes.Buffer
	}

	// добавляем реализацию http.ResponseWriter
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("Initialize: parse level errors %w", err)
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("Initialize: logger build error %w", err)
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode // захватываем код статуса
}

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
