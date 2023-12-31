package middlewares

import (
	"bytes"
	"net/http"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
)

// WithLogging логирует события из обработчиков.
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &logger.ResponseData{
			Status: 0,
			Size:   0,
			Body:   bytes.NewBufferString(""),
		}
		lw := logger.LoggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			ResponseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Info("incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("status", responseData.Status),
			zap.Int("size", responseData.Size),
			// zap.String("body", responseData.Body.String()),
		)
	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}
