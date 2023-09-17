package middlewares

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/infra/sign"
	"github.com/pavlegich/metrics-alerting/internal/models"
)

func WithSign(h http.Handler) http.Handler {
	signFn := func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("HashSHA256")

		if got != "" && models.Key != "" {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(&buf)
			hash, err := sign.Sign(buf.Bytes(), []byte(models.Key))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			want := hex.EncodeToString(hash)

			if want != got {
				logger.Log.Info("WithSign: hashes not equal")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		sw := sign.SigningResponseWriter{
			ResponseWriter: w,
		}

		h.ServeHTTP(&sw, r)
	}

	return http.HandlerFunc(signFn)
}
