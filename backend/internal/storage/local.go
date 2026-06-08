package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalProvider struct {
	basePath string
}

func NewLocalProvider(basePath string) (*LocalProvider, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalProvider{basePath: basePath}, nil
}

func (p *LocalProvider) fullPath(key string) string {
	return filepath.Join(p.basePath, filepath.FromSlash(key))
}

func (p *LocalProvider) Upload(ctx context.Context, key string, reader io.Reader, _ string, _ int64) error {
	path := p.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}

func (p *LocalProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(p.fullPath(key))
}

func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	return os.Remove(p.fullPath(key))
}

func (p *LocalProvider) URL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("file://%s", p.fullPath(key)), nil
}

func (p *LocalProvider) Ping(_ context.Context) error {
	if _, err := os.Stat(p.basePath); err != nil {
		return fmt.Errorf("storage unavailable: %w", err)
	}
	return nil
}
