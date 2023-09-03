package handlers

import (
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/logger"
	"go.uber.org/zap"
)

func (h *Webhook) HandlePing(w http.ResponseWriter, r *http.Request) {
	if err := h.Database.Ping(); err != nil {
		logger.Log.Error("HandlePing: connection with database is died", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
