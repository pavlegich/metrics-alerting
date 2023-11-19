package middlewares

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/hash"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
)

// WithSign обрабатывает запрос с учётом верно полученного хеша.
// Обработчик формирует хеш из полученного тела запроса,
// сравнивает с полученным в заголовке запроса хешем.
// В случае неуспешной проверки прерывает обработку запроса.
func WithSign(h http.Handler) http.Handler {
	signFn := func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("HashSHA256")

		if got != "" && entities.Key != "" {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(&buf)
			hash, err := hash.Sign(buf.Bytes(), []byte(entities.Key))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			want := hex.EncodeToString(hash)

			if want != got {
				logger.Log.Error("WithSign: hashes not equal")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		sw := hash.SigningResponseWriter{
			ResponseWriter: w,
		}

		h.ServeHTTP(&sw, r)
	}

	return http.HandlerFunc(signFn)
}
