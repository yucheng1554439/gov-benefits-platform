package storage

import (
	"context"
	"io"
)

type Provider interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) (string, error)
	Ping(ctx context.Context) error
}
