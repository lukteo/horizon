package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// BlobStorageService handles interactions with blob storage (S3-compatible)
type BlobStorageService struct {
	client   *minio.Client
	bucket   string
	endpoint string
}

// NewBlobStorageService creates a new instance of BlobStorageService
func NewBlobStorageService(endpoint, accessKey, secretKey, bucket string, secure bool) (*BlobStorageService, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Check if bucket exists and create if it doesn't
	exists, err := minioClient.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", bucket).Msg("Created blob storage bucket")
	}

	return &BlobStorageService{
		client:   minioClient,
		bucket:   bucket,
		endpoint: endpoint,
	}, nil
}

// UploadRawLog uploads raw log data to blob storage
func (bs *BlobStorageService) UploadRawLog(ctx context.Context, objectName string, data []byte) error {
	reader := bytes.NewReader(data)
	info, err := bs.client.PutObject(ctx, bs.bucket, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("failed to upload to blob storage: %w", err)
	}

	log.Debug().Str("object", objectName).Int64("size", info.Size).Msg("Uploaded raw log to blob storage")
	return nil
}

// DownloadRawLog downloads raw log data from blob storage
func (bs *BlobStorageService) DownloadRawLog(ctx context.Context, objectName string) ([]byte, error) {
	object, err := bs.client.GetObject(ctx, bs.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from blob storage: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object from blob storage: %w", err)
	}

	return data, nil
}

// DeleteRawLog deletes raw log data from blob storage
func (bs *BlobStorageService) DeleteRawLog(ctx context.Context, objectName string) error {
	err := bs.client.RemoveObject(ctx, bs.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from blob storage: %w", err)
	}

	log.Debug().Str("object", objectName).Msg("Deleted raw log from blob storage")
	return nil
}

// HealthCheck checks if the blob storage is accessible
func (bs *BlobStorageService) HealthCheck(ctx context.Context) error {
	// Test by listing objects (even if empty)
	opts := minio.ListObjectsOptions{
		MaxKeys: 1,
	}
	for range bs.client.ListObjects(ctx, bs.bucket, opts) {
		// Just need to make sure we can connect and list
		break
	}
	return nil
}