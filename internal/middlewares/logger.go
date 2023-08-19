package middlewares

import (
	"net/http"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/logger"
	"go.uber.org/zap"
)

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &logger.ResponseData{
			Status: 0,
			Size:   0,
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
		)
	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}
