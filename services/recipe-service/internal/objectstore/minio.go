package objectstore

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStore struct {
	client *minio.Client
	bucket string
}

func newMinio(cfg Config) (*minioStore, error) {
	c, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, err
	}
	return &minioStore{client: c, bucket: cfg.Bucket}, nil
}

func (s *minioStore) Put(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(body), int64(len(body)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("%w: put %s: %v", ErrUnavailable, key, err)
	}
	return nil
}

func (s *minioStore) Get(ctx context.Context, key string) (Object, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return Object{}, fmt.Errorf("%w: get %s: %v", ErrUnavailable, key, err)
	}
	// GetObject is lazy; Stat surfaces NoSuchKey and connection errors.
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return Object{}, ErrNotFound
		}
		return Object{}, fmt.Errorf("%w: stat %s: %v", ErrUnavailable, key, err)
	}
	return Object{Body: obj, Size: info.Size, ContentType: info.ContentType}, nil
}

func (s *minioStore) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return ErrNotFound
		}
		return fmt.Errorf("%w: delete %s: %v", ErrUnavailable, key, err)
	}
	return nil
}

func (s *minioStore) Ping(ctx context.Context) error {
	ok, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("%w: bucket check: %v", ErrUnavailable, err)
	}
	if !ok {
		return fmt.Errorf("%w: bucket %q does not exist", ErrUnavailable, s.bucket)
	}
	return nil
}
