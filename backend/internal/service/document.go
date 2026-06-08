package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/storage"
	"github.com/jackc/pgx/v5"
)

type DocumentService struct {
	db      *postgres.DB
	docs    *postgres.DocumentRepository
	storage storage.Provider
	bus     *events.Bus
}

func NewDocumentService(db *postgres.DB, docs *postgres.DocumentRepository, storage storage.Provider, bus *events.Bus) *DocumentService {
	return &DocumentService{db: db, docs: docs, storage: storage, bus: bus}
}

const maxUploadSize = 10 * 1024 * 1024

var allowedMimeTypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
}

func (s *DocumentService) Upload(ctx context.Context, agencyID, userID, caseID uuid.UUID, originalName, mimeType string, size int64, reader io.Reader) (*domain.Document, error) {
	if size <= 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}
	if size > maxUploadSize {
		return nil, fmt.Errorf("file exceeds maximum size of 10 MB")
	}
	baseName := filepath.Base(originalName)
	if baseName == "." || baseName == "" {
		return nil, fmt.Errorf("invalid file name")
	}
	if strings.Contains(baseName, "..") {
		return nil, fmt.Errorf("invalid file name")
	}
	if !allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("unsupported file type; allowed: PDF, JPEG, PNG, WebP")
	}
	key := fmt.Sprintf("%s/cases/%s/%d_%s", agencyID, caseID, time.Now().UnixNano(), filepath.Base(originalName))
	if err := s.storage.Upload(ctx, key, reader, mimeType, size); err != nil {
		return nil, err
	}

	var doc *domain.Document
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		doc = &domain.Document{
			AgencyID:     agencyID,
			CaseID:       caseID,
			FileKey:      key,
			OriginalName: originalName,
			MimeType:     mimeType,
			FileSize:     size,
		}
		return s.docs.Create(ctx, tx, doc)
	})
	if err != nil {
		return nil, err
	}

	actorID := userID
	s.bus.Publish(ctx, events.NewEvent(events.EventDocumentUploaded, agencyID, doc.ID, &actorID, map[string]any{
		"case_id": caseID.String(),
	}))
	return doc, nil
}

func (s *DocumentService) List(ctx context.Context, agencyID, userID, caseID uuid.UUID) ([]domain.Document, error) {
	var docs []domain.Document
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		docs, err = s.docs.ListByCase(ctx, tx, caseID)
		return err
	})
	return docs, err
}

func (s *DocumentService) Verify(ctx context.Context, agencyID, userID, docID uuid.UUID, status string) error {
	return postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		return s.docs.UpdateVerification(ctx, tx, docID, status, userID)
	})
}

func (s *DocumentService) Download(ctx context.Context, agencyID, userID, docID uuid.UUID) (io.ReadCloser, *domain.Document, error) {
	var doc *domain.Document
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		doc, err = s.docs.GetByID(ctx, tx, docID)
		return err
	})
	if err != nil || doc == nil {
		return nil, nil, fmt.Errorf("document not found")
	}
	reader, err := s.storage.Download(ctx, doc.FileKey)
	return reader, doc, err
}
