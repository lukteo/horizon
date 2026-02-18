package ocsf

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/luketeo/horizon/server/internal/storage"
	"github.com/rs/zerolog/log"
	"github.com/luketeo/horizon/server/internal/modelz"
)

// Normalizer handles the conversion of raw logs to OCSF format
type Normalizer struct {
	dbService *storage.DatabaseService
}

// NewNormalizer creates a new instance of Normalizer
func NewNormalizer(dbService *storage.DatabaseService) *Normalizer {
	return &Normalizer{
		dbService: dbService,
	}
}

// NormalizeLog normalizes a raw log entry to OCSF format using the appropriate mapping
func (n *Normalizer) NormalizeLog(rawLog models.RawLogEntry) (*models.OcsfEvent, error) {
	// Find the appropriate mapping for this log's source type
	mapping, err := n.getLogMappingBySourceType(rawLog.SourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get log mapping for source type %s: %w", rawLog.SourceType, err)
	}

	// Parse the raw data
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(rawLog.RawData), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse raw log data: %w", err)
	}

	// Apply the mapping to transform the raw data to OCSF format
	ocsfEvent, err := n.applyMapping(rawData, mapping.MappingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to apply mapping: %w", err)
	}

	// Set the timestamp from the raw log
	ocsfEvent.Timestamp = rawLog.Timestamp

	return ocsfEvent, nil
}

// getLogMappingBySourceType retrieves the appropriate log mapping for a given source type
func (n *Normalizer) getLogMappingBySourceType(sourceType string) (*models.LogMapping, error) {
	query := `
		SELECT id, name, source_type, mapping_config, created_at, updated_at
		FROM log_mappings
		WHERE source_type = $1 AND enabled = true
		LIMIT 1
	`

	row := n.dbService.Pool.QueryRow(context.Background(), query, sourceType)

	var mapping models.LogMapping
	var createdAt, updatedAt interface{}
	err := row.Scan(
		&mapping.ID,
		&mapping.Name,
		&mapping.SourceType,
		&mapping.MappingConfig,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		return nil, err
	}

	return &mapping, nil
}

// applyMapping applies the mapping configuration to transform raw data to OCSF format
func (n *Normalizer) applyMapping(rawData map[string]interface{}, mappingConfig string) (*models.OcsfEvent, error) {
	var mapping map[string]interface{}
	if err := json.Unmarshal([]byte(mappingConfig), &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse mapping config: %w", err)
	}

	// Create a default OCSF event
	ocsfEvent := &models.OcsfEvent{
		EventTypeUID:  1, // Default to a generic event type
		EventTypeName: "Generic Event",
		CategoryUID:   1,
		CategoryName:  "Audit",
		ClassUID:      1,
		ClassName:     "System Activity",
		SeverityID:    2, // Default to medium
		Severity:      "Medium",
		Message:       getStringValue(rawData, "message", "Log message not available"),
		Additional:    make(map[string]interface{}),
	}

	// Apply field mappings based on the configuration
	if fieldMappings, ok := mapping["field_mappings"].(map[string]interface{}); ok {
		for ocsfField, rawFieldPath := range fieldMappings {
			rawFieldName, ok := rawFieldPath.(string)
			if !ok {
				continue
			}

			// Extract value from raw data using the field path
			value := extractValueByPath(rawData, rawFieldName)
			
			// Map the value to the appropriate OCSF field
			switch strings.ToLower(ocsfField) {
			case "message":
				if str, ok := value.(string); ok {
					ocsfEvent.Message = str
				}
			case "severity":
				if str, ok := value.(string); ok {
					ocsfEvent.Severity = str
					ocsfEvent.SeverityID = severityToID(str)
				}
			case "category_name":
				if str, ok := value.(string); ok {
					ocsfEvent.CategoryName = str
				}
			case "class_name":
				if str, ok := value.(string); ok {
					ocsfEvent.ClassName = str
				}
			case "actor.user.name":
				if str, ok := value.(string); ok {
					if ocsfEvent.Actor == nil {
						ocsfEvent.Actor = &models.OcsfActor{}
					}
					if ocsfEvent.Actor.User == nil {
						ocsfEvent.Actor.User = &models.OcsfUser{}
					}
					ocsfEvent.Actor.User.Name = &str
				}
			case "source.ip":
				if str, ok := value.(string); ok {
					if ocsfEvent.Source == nil {
						ocsfEvent.Source = &models.OcsfSource{}
					}
					ocsfEvent.Source.IP = &str
				}
			default:
				// Store unmapped fields in additional fields
				ocsfEvent.Additional[ocsfField] = value
			}
		}
	}

	return ocsfEvent, nil
}

// ProcessAndStoreLog processes a raw log and stores the normalized version
func (n *Normalizer) ProcessAndStoreLog(rawLog models.RawLogEntry) error {
	// Normalize the log
	normalizedLog, err := n.NormalizeLog(rawLog)
	if err != nil {
		log.Error().Err(err).Str("source", rawLog.Source).Msg("Failed to normalize log")
		return err
	}

	// Store the normalized log in the database
	if err := n.storeNormalizedLog(normalizedLog, rawLog.ID); err != nil {
		log.Error().Err(err).Msg("Failed to store normalized log")
		return err
	}

	// Update the raw log record to mark it as processed
	if err := n.markRawLogAsProcessed(rawLog.ID); err != nil {
		log.Error().Err(err).Msg("Failed to mark raw log as processed")
		return err
	}

	return nil
}

// storeNormalizedLog stores the normalized log in the database
func (n *Normalizer) storeNormalizedLog(ocsfEvent *models.OcsfEvent, rawLogID interface{}) error {
	// Convert OCSF event to JSON
	eventJSON, err := json.Marshal(ocsfEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal OCSF event: %w", err)
	}

	query := `
		INSERT INTO normalized_logs (ocsf_event, created_at, indexed_at)
		VALUES ($1, NOW(), NOW())
	`

	_, err = n.dbService.Pool.Exec(context.Background(), query, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to insert normalized log: %w", err)
	}

	return nil
}

// markRawLogAsProcessed marks a raw log as processed in the database
func (n *Normalizer) markRawLogAsProcessed(rawLogID interface{}) error {
	query := `
		UPDATE raw_log_references
		SET processed = true, processed_at = NOW()
		WHERE id = $1
	`

	_, err := n.dbService.Pool.Exec(context.Background(), query, rawLogID)
	if err != nil {
		return fmt.Errorf("failed to update raw log: %w", err)
	}

	return nil
}

// extractValueByPath extracts a value from a nested map using dot notation
func extractValueByPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return value
		}

		// Otherwise, continue traversing if it's a map
		if nextMap, ok := value.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return nil
		}
	}

	return nil
}

// getStringValue safely extracts a string value from a map with a fallback
func getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// severityToID converts a severity string to its corresponding ID
func severityToID(severity string) int {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return 4
	case "medium", "med":
		return 3
	case "low":
		return 2
	case "info", "informational":
		return 1
	default:
		return 2 // Default to low
	}
}