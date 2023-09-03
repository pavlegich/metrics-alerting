package handlers

import (
	"context"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/logger"
	"go.uber.org/zap"
)

func (h *Webhook) HandlePing(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Database.PingContext(ctx); err != nil {
		cancel()
		logger.Log.Error("HandlePing: connection with database is died", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
