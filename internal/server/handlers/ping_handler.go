package handlers

import (
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
)

// HandlePing обрабатывает запрос для проверки наличия
// соединения с базой данных.
func (h *Webhook) HandlePing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.Database == nil {
		logger.Log.Info("HandlePing: database is not used")
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.Database.Ping(ctx); err != nil {
		logger.Log.Error("HandlePing: connection with database is died", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
