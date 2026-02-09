package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/luketeo/horizon/server/internal/logprocessor"
	"github.com/luketeo/horizon/server/internal/middleware"
	"github.com/luketeo/horizon/server/internal/ocsf"
	"github.com/luketeo/horizon/server/internal/storage"
)

// InitRoutes initializes all application routes
func InitRoutes(r chi.Router, dbService *storage.DatabaseService, normalizer *ocsf.Normalizer) {
	// Create handler instances
	healthHandler := middleware.NewHealthHandler()
	logHandler := logprocessor.NewHandler(dbService, normalizer)

	// Public routes
	r.Get("/health", healthHandler.Health)

	// API routes
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/logs/ingest", logHandler.IngestLogs)
		api.Get("/mappings", logHandler.GetMappings)
		api.Post("/mappings", logHandler.CreateMapping)
		api.Get("/mappings/{id}", logHandler.GetMapping)
	})
}