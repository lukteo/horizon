package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/luketeo/horizon/server/internal/modelz"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new instance of HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles health check requests
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Message: "Server is running",
	})
}