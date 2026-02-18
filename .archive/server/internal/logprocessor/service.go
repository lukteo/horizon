package logprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/luketeo/horizon/server/internal/messaging"
	"github.com/luketeo/horizon/server/internal/ocsf"
	"github.com/luketeo/horizon/server/internal/storage"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/luketeo/horizon/server/internal/modelz"
)

// Processor handles consuming logs from NATS and processing them
type Processor struct {
	natsService     *messaging.NatsService
	normalizer      *ocsf.Normalizer
	dbService       *storage.DatabaseService
	blobStorage     *storage.BlobStorageService
}

// NewProcessor creates a new instance of Processor
func NewProcessor(
	natsService *messaging.NatsService,
	normalizer *ocsf.Normalizer,
	dbService *storage.DatabaseService,
	blobStorage *storage.BlobStorageService,
) *Processor {
	return &Processor{
		natsService: natsService,
		normalizer:  normalizer,
		dbService:   dbService,
		blobStorage: blobStorage,
	}
}

// Start begins processing logs from NATS
func (lp *Processor) Start(ctx context.Context) error {
	log.Info().Msg("Starting log processor...")

	// Subscribe to raw logs from NATS
	sub, err := lp.natsService.SubscribeToRawLogs("logs.raw", lp.handleRawLog)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	log.Info().Msg("Log processor started, waiting for messages...")

	// Wait for context cancellation
	<-ctx.Done()
	log.Info().Msg("Log processor shutting down...")
	return nil
}

// handleRawLog processes a single raw log message from NATS
func (lp *Processor) handleRawLog(msg *nats.Msg) {
	log.Debug().Str("subject", msg.Subject).Int("size", len(msg.Data)).Msg("Received raw log")

	// Generate a unique ID for this log
	logID := uuid.New()

	// Create blob storage path
	blobPath := fmt.Sprintf("raw-logs/%s/%s/%s.json", 
		time.Now().Format("2006/01/02"), // Date-based folder structure
		time.Now().Format("15-04-05"),   // Hour-minute-second
		logID.String())

	// Upload raw log data to blob storage
	if err := lp.blobStorage.UploadRawLog(context.Background(), blobPath, msg.Data); err != nil {
		log.Error().Err(err).Msg("Failed to upload raw log to blob storage")
		// Don't acknowledge the message to allow for retry
		return
	}

	// Parse the raw log data to extract metadata
	var rawLogData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &rawLogData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal raw log data for metadata extraction")
		// Acknowledge the message to prevent redelivery of malformed data
		msg.Ack()
		return
	}

	// Convert to RawLogReference model
	sourceType := getStringValueFromMap(rawLogData, "source_type", "unknown")
	logRef := models.RawLogReference{
		ID:              logID,
		Source:          getStringValueFromMap(rawLogData, "source", "unknown"),
		SourceType:      sourceType,
		BlobStoragePath: blobPath,
		SizeBytes:       int64(len(msg.Data)),
		ReceivedAt:      time.Now(),
		Processed:       false,
	}

	// Store log reference in database
	if err := lp.storeRawLogReference(logRef); err != nil {
		log.Error().Err(err).Msg("Failed to store raw log reference")
		// Don't acknowledge the message to allow for retry
		return
	}

	// Process the log (download from blob, normalize, and store)
	if err := lp.processLogFromBlob(logRef); err != nil {
		log.Error().Err(err).Msg("Failed to process and store normalized log")
		// Don't acknowledge the message to allow for retry
		return
	}

	// Acknowledge the message after successful processing
	if err := msg.Ack(); err != nil {
		log.Error().Err(err).Msg("Failed to acknowledge message")
	}
}

// storeRawLogReference stores a raw log reference in the database
func (lp *Processor) storeRawLogReference(logRef models.RawLogReference) error {
	query := `
		INSERT INTO raw_log_references (id, source, source_type, blob_storage_path, size_bytes, received_at, processed)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := lp.dbService.Pool.Exec(context.Background(), query,
		logRef.ID,
		logRef.Source,
		logRef.SourceType,
		logRef.BlobStoragePath,
		logRef.SizeBytes,
		logRef.ReceivedAt,
		logRef.Processed,
	)

	if err != nil {
		return err
	}

	log.Debug().Interface("id", logRef.ID).Str("path", logRef.BlobStoragePath).Msg("Stored raw log reference with ID")

	return nil
}

// processLogFromBlob downloads a log from blob storage, normalizes it, and stores the result
func (lp *Processor) processLogFromBlob(logRef models.RawLogReference) error {
	// Download raw log data from blob storage
	rawData, err := lp.blobStorage.DownloadRawLog(context.Background(), logRef.BlobStoragePath)
	if err != nil {
		return fmt.Errorf("failed to download raw log from blob storage: %w", err)
	}

	// Parse the raw log data
	var rawLogData map[string]interface{}
	if err := json.Unmarshal(rawData, &rawLogData); err != nil {
		return fmt.Errorf("failed to unmarshal raw log data: %w", err)
	}

	// Convert to RawLogEntry model for normalization
	rawLog := models.RawLogEntry{
		ID:              logRef.ID,
		Source:          logRef.Source,
		SourceType:      logRef.SourceType,
		RawData:         string(rawData),
		BlobStoragePath: logRef.BlobStoragePath,
		SizeBytes:       logRef.SizeBytes,
		ReceivedAt:      logRef.ReceivedAt,
		Timestamp:       logRef.ReceivedAt, // Use received time as timestamp for now
		Level:           getStringValueFromMap(rawLogData, "level", "INFO"),
		Message:         getStringValueFromMap(rawLogData, "message", ""),
	}

	// Normalize the log
	if err := lp.normalizer.ProcessAndStoreLog(rawLog); err != nil {
		return fmt.Errorf("failed to process and store normalized log: %w", err)
	}

	// Mark the log reference as processed in the database
	if err := lp.markLogReferenceAsProcessed(logRef.ID); err != nil {
		return fmt.Errorf("failed to mark log reference as processed: %w", err)
	}

	return nil
}

// markLogReferenceAsProcessed marks a raw log reference as processed in the database
func (lp *Processor) markLogReferenceAsProcessed(logID uuid.UUID) error {
	query := `
		UPDATE raw_log_references
		SET processed = true, processed_at = NOW()
		WHERE id = $1
	`

	_, err := lp.dbService.Pool.Exec(context.Background(), query, logID)
	if err != nil {
		return fmt.Errorf("failed to update raw log reference: %w", err)
	}

	return nil
}

// getStringValueFromMap safely extracts a string value from a map with a fallback
func getStringValueFromMap(data map[string]interface{}, key, defaultValue string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}