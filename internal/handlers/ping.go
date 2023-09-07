package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/logger"
	"go.uber.org/zap"
)

func (h *Webhook) HandlePing(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if h.Database == nil {
		logger.Log.Error("HandlePing: database is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.Database.PingContext(ctx); err != nil {
		logger.Log.Error("HandlePing: connection with database is died", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
