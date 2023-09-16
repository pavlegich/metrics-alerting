package middlewares

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/infra/sign"
	"github.com/pavlegich/metrics-alerting/internal/models"
)

func WithSign(h http.Handler) http.Handler {
	signFn := func(w http.ResponseWriter, r *http.Request) {
		if models.KEY != "" {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(&buf)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			hash, err := sign.Sign(buf.Bytes(), []byte(models.KEY))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			got := r.Header.Get("HashSHA256")
			want := hex.EncodeToString(hash)

			fmt.Printf("server key: '%s'", models.KEY)
			fmt.Printf("got: '%s'; want: '%s'", got, want)

			if reflect.DeepEqual(want, got) {
				logger.Log.Info("WithSign: hashes not equal")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(signFn)
}
