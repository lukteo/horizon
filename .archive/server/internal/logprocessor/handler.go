package logprocessor

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luketeo/horizon/server/internal/ocsf"
	"github.com/luketeo/horizon/server/internal/storage"
	"github.com/luketeo/horizon/server/internal/modelz"
)

// Handler handles HTTP requests for log processing
type Handler struct {
	dbService  *storage.DatabaseService
	normalizer *ocsf.Normalizer
}

// NewHandler creates a new instance of Handler
func NewHandler(dbService *storage.DatabaseService, normalizer *ocsf.Normalizer) *Handler {
	return &Handler{
		dbService:  dbService,
		normalizer: normalizer,
	}
}

// IngestLogs handles log ingestion requests
func (h *Handler) IngestLogs(w http.ResponseWriter, r *http.Request) {
	var req models.IngestLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would:
	// 1. Store raw logs in NATS for durability
	// 2. Return immediately to the client
	// 3. Process logs asynchronously

	// For now, just return a success response
	resp := models.APIResponse{
		Success: true,
		Data: models.IngestLogsResponse{
			IngestedCount: len(req.Logs),
			FailedCount:   0,
		},
		Message: "Successfully queued logs for processing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMappings handles getting log mappings
func (h *Handler) GetMappings(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch from the database
	mappings := []models.LogMapping{}

	resp := models.APIResponse{
		Success: true,
		Data:    mappings,
		Message: "Retrieved log mappings",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateMapping handles creating a new log mapping
func (h *Handler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would save to the database
	mapping := models.LogMapping{
		// ID would be generated in the database
		Name:         req.Name,
		SourceType:   req.SourceType,
		MappingConfig: req.MappingConfig,
		// CreatedAt and UpdatedAt would be set in the database
	}

	resp := models.APIResponse{
		Success: true,
		Data:    mapping,
		Message: "Created log mapping",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMapping handles getting a specific log mapping
func (h *Handler) GetMapping(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")  // Using underscore to indicate we're intentionally not using the variable

	// In a real implementation, this would fetch from the database
	mapping := models.LogMapping{
		// ID would come from the database
		Name:        "Sample Mapping",
		SourceType:  "sample",
		MappingConfig: "{}",
		// CreatedAt and UpdatedAt would come from the database
	}

	resp := models.APIResponse{
		Success: true,
		Data:    mapping,
		Message: "Retrieved log mapping",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}