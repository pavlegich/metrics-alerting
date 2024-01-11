package middlewares

import (
	"net"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
)

// WithNetworking проверяет соответствие IP адреса клиента указанному доверенному диапозону.
func WithNetworking(network *net.IPNet) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if network == nil {
				h.ServeHTTP(w, r)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			if ip == nil || !network.Contains(ip) {
				logger.Log.Error("WithNetworking: IP not in trusted subnet",
					zap.String("ip", ip.String()))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
