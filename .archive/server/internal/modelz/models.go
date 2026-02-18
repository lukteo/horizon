package models

import (
	"time"

	"github.com/google/uuid"
)

// OCSF Base Event - Following OCSF specification
type OcsfEvent struct {
	Timestamp     time.Time                `json:"time" db:"timestamp"`
	EventTypeUID  int                      `json:"type_uid" db:"event_type_uid"`
	EventTypeName string                   `json:"type_name" db:"event_type_name"`
	CategoryUID   int                      `json:"category_uid" db:"category_uid"`
	CategoryName  string                   `json:"category_name" db:"category_name"`
	ClassUID      int                      `json:"class_uid" db:"class_uid"`
	ClassName     string                   `json:"class_name" db:"class_name"`
	SeverityID    int                      `json:"severity_id" db:"severity_id"`
	Severity      string                   `json:"severity" db:"severity"`
	StatusID      *int                     `json:"status_id,omitempty" db:"status_id"`
	Status        *string                  `json:"status,omitempty" db:"status"`
	Message       string                   `json:"message" db:"message"`
	Description   *string                  `json:"description,omitempty" db:"description"`
	Actor         *OcsfActor              `json:"actor,omitempty" db:"actor"`
	Target        *OcsfTarget             `json:"target,omitempty" db:"target"`
	Source        *OcsfSource             `json:"source,omitempty" db:"source"`
	Additional    map[string]interface{}   `json:"additional_fields" db:"additional_fields"`
}

type OcsfActor struct {
	User     *OcsfUser     `json:"user,omitempty"`
	Process  *OcsfProcess  `json:"process,omitempty"`
	Session  *OcsfSession  `json:"session,omitempty"`
}

type OcsfTarget struct {
	User     *OcsfUser     `json:"user,omitempty"`
	Process  *OcsfProcess  `json:"process,omitempty"`
	File     *OcsfFile     `json:"file,omitempty"`
	Resource *OcsfResource `json:"resource,omitempty"`
}

type OcsfSource struct {
	IP   *string   `json:"ip,omitempty"`
	Port *uint16   `json:"port,omitempty"`
	Host *OcsfHost `json:"host,omitempty"`
}

type OcsfUser struct {
	UID   *string     `json:"uid,omitempty"`
	Name  *string     `json:"name,omitempty"`
	Group *OcsfGroup  `json:"group,omitempty"`
}

type OcsfProcess struct {
	PID  *int64    `json:"pid,omitempty"`
	Name *string   `json:"name,omitempty"`
	File *OcsfFile `json:"file,omitempty"`
}

type OcsfFile struct {
	Name *string  `json:"name,omitempty"`
	Path *string  `json:"path,omitempty"`
	Hash *OcsfHash `json:"hash,omitempty"`
}

type OcsfHash struct {
	MD5    *string `json:"md5,omitempty"`
	SHA1   *string `json:"sha1,omitempty"`
	SHA256 *string `json:"sha256,omitempty"`
}

type OcsfGroup struct {
	GID  *string `json:"gid,omitempty"`
	Name *string `json:"name,omitempty"`
}

type OcsfHost struct {
	Hostname *string `json:"hostname,omitempty"`
	IP       *string `json:"ip,omitempty"`
	OS       *OcsfOS `json:"os,omitempty"`
}

type OcsfOS struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

type OcsfResource struct {
	Name  *string `json:"name,omitempty"`
	Type  *string `json:"type,omitempty"`
}

type OcsfSession struct {
	SessionID *string `json:"session_id,omitempty"`
}

// Raw log entry before normalization
type RawLogEntry struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	Timestamp       time.Time     `json:"timestamp" db:"timestamp"`
	Source          string        `json:"source" db:"source"`
	Level           string        `json:"level" db:"level"`
	Message         string        `json:"message" db:"message"`
	RawData         string        `json:"raw_data" db:"raw_data"` // JSON string representation
	ReceivedAt      time.Time     `json:"received_at" db:"received_at"`
	SourceType      string        `json:"source_type" db:"source_type"`
	BlobStoragePath string        `json:"blob_storage_path" db:"blob_storage_path"`
	SizeBytes       int64         `json:"size_bytes" db:"size_bytes"`
	Processed       bool          `json:"processed" db:"processed"`
	ProcessedAt     *time.Time    `json:"processed_at,omitempty" db:"processed_at"`
}

// RawLogReference represents a reference to raw log data stored in blob storage
type RawLogReference struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Source          string    `json:"source" db:"source"`
	SourceType      string    `json:"source_type" db:"source_type"`
	BlobStoragePath string    `json:"blob_storage_path" db:"blob_storage_path"`
	SizeBytes       int64     `json:"size_bytes" db:"size_bytes"`
	ReceivedAt      time.Time `json:"received_at" db:"received_at"`
	Processed       bool      `json:"processed" db:"processed"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty" db:"processed_at"`
}

// Log mapping for OCSF normalization
type LogMapping struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	SourceType  string    `json:"source_type" db:"source_type"`
	MappingConfig string  `json:"mapping_config" db:"mapping_config"` // JSON string representation
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// API Response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Request/Response structures
type IngestLogsRequest struct {
	Logs      []interface{} `json:"logs"`
	SourceType string       `json:"source_type"`
}

type IngestLogsResponse struct {
	IngestedCount int `json:"ingested_count"`
	FailedCount   int `json:"failed_count"`
}

type CreateMappingRequest struct {
	Name         string `json:"name"`
	SourceType   string `json:"source_type"`
	MappingConfig string `json:"mapping_config"`
}