package web

import (
	"log/slog"
	"net/http"
	"os"
)

func (h *Handler) GetHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Health check successful"))
	if err != nil {
		slog.Default().Error("failed to write response body", slog.Any("err", err))
		os.Exit(1)
	}

	// add health check here for additional peripherals
	// e.g. pg, redis
}
