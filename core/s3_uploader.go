package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Uploader handles uploading files to S3.
type S3Uploader struct {
	Client *s3.Client
	Bucket string
	Prefix string
}

// NewS3Uploader creates a new uploader.
func NewS3Uploader(cfg aws.Config, bucket, prefix string) *S3Uploader {
	return &S3Uploader{
		Client: s3.NewFromConfig(cfg),
		Bucket: bucket,
		Prefix: prefix,
	}
}

// UploadDirectory walks the local directory and uploads all files to S3.
func (u *S3Uploader) UploadDirectory(localDir string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Calculate S3 Key
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		// Ensure forward slashes for S3 keys
		relPath = filepath.ToSlash(relPath)
		key := filepath.Join(u.Prefix, relPath)
		// filepath.Join might use backslash on Windows, so ensure forward slash again
		key = strings.ReplaceAll(key, "\\", "/")
		// Remove leading slash if any
		key = strings.TrimPrefix(key, "/")

		return u.UploadFile(path, key)
	})
}

// UploadFile uploads a single file to S3.
func (u *S3Uploader) UploadFile(localPath, key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	slog.Info("Uploading to S3", "local", localPath, "bucket", u.Bucket, "key", key)

	_, err = u.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(u.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to s3: %w", err)
	}
	return nil
}
