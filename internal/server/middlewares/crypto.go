package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/pavlegich/metrics-alerting/internal/infra/crypto"
)

// WithDecryption обрабатывает запрос с учётом шифрования сообщения.
func WithDecryption(keyPath string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentEncryption := r.Header.Get("Content-Encryption")
			sendsRSA := strings.Contains(contentEncryption, "rsa")

			if !sendsRSA || keyPath == "" {
				h.ServeHTTP(w, r)
				return
			}

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			msg, err := crypto.Decrypt(buf.Bytes(), keyPath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(msg))
			h.ServeHTTP(w, r)
		})
	}
}
