package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/letters"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/storage"
	"github.com/jackc/pgx/v5"
)

type LetterService struct {
	db       *postgres.DB
	letters  *postgres.LetterRepository
	cases    *postgres.CaseRepository
	users    *postgres.UserRepository
	agencies *postgres.AgencyRepository
	benefit  *postgres.BenefitRepository
	pdf      *letters.PDFGenerator
	storage  storage.Provider
	bus      *events.Bus
}

func NewLetterService(
	db *postgres.DB,
	letterRepo *postgres.LetterRepository,
	cases *postgres.CaseRepository,
	users *postgres.UserRepository,
	agencies *postgres.AgencyRepository,
	benefit *postgres.BenefitRepository,
	pdf *letters.PDFGenerator,
	storage storage.Provider,
	bus *events.Bus,
) *LetterService {
	return &LetterService{
		db: db, letters: letterRepo, cases: cases, users: users,
		agencies: agencies, benefit: benefit, pdf: pdf, storage: storage, bus: bus,
	}
}

func (s *LetterService) Generate(ctx context.Context, agencyID, userID, caseID uuid.UUID, letterType string) (*domain.GeneratedLetter, error) {
	return s.generate(ctx, agencyID, userID, caseID, letterType)
}

func (s *LetterService) generate(ctx context.Context, agencyID, userID, caseID uuid.UUID, letterType string) (*domain.GeneratedLetter, error) {
	template, err := s.letters.GetTemplate(ctx, agencyID, letterType)
	if err != nil || template == nil {
		return nil, fmt.Errorf("letter template not found")
	}

	var caseNumber, citizenName, programName string
	var benefitAmount float64

	_ = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		c, _ := s.cases.GetByID(ctx, tx, caseID)
		if c != nil {
			caseNumber = c.CaseNumber
			if p, _ := s.agencies.GetProgramByID(ctx, c.ProgramID); p != nil {
				programName = p.Name
			}
			if profile, _ := s.users.GetProfile(ctx, c.CitizenID); profile != nil {
				citizenName = profile.FirstName + " " + profile.LastName
			}
		}
		if calc, _ := s.benefit.GetLatestCalculation(ctx, tx, caseID); calc != nil {
			benefitAmount = calc.CalculatedAmount
		}
		return nil
	})

	agency, _ := s.agencies.GetByID(ctx, agencyID)
	agencyName := ""
	if agency != nil {
		agencyName = agency.Name
	}

	data := letters.LetterData{
		CitizenName:   citizenName,
		ProgramName:   programName,
		BenefitAmount: fmt.Sprintf("%.2f", benefitAmount),
		CaseNumber:    caseNumber,
		DenialReason:  "Does not meet eligibility requirements",
		AgencyName:    agencyName,
	}

	pdfBytes, err := s.pdf.GenerateLetter(template.Name, template.BodyTemplate, data)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s/letters/%s_%s.pdf", agencyID, caseID, letterType)
	if err := s.storage.Upload(ctx, key, bytes.NewReader(pdfBytes), "application/pdf", int64(len(pdfBytes))); err != nil {
		return nil, err
	}

	var letter domain.GeneratedLetter
	err = postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		letter = domain.GeneratedLetter{
			AgencyID:    agencyID,
			CaseID:      caseID,
			TemplateID:  &template.ID,
			LetterType:  letterType,
			FileKey:     key,
			GeneratedBy: &userID,
		}
		return s.letters.SaveGenerated(ctx, tx, &letter)
	})
	if err != nil {
		return nil, err
	}

	if userID != uuid.Nil {
		actorID := userID
		s.bus.Publish(ctx, events.NewEvent(events.EventLetterGenerated, agencyID, letter.ID, &actorID, map[string]any{
			"case_id": caseID.String(),
		}))
	}
	return &letter, nil
}

func (s *LetterService) GenerateFromJob(ctx context.Context, caseID, agencyID uuid.UUID, letterType string) error {
	_, err := s.generate(ctx, agencyID, uuid.Nil, caseID, letterType)
	return err
}

func (s *LetterService) List(ctx context.Context, agencyID, userID, caseID uuid.UUID) ([]domain.GeneratedLetter, error) {
	var result []domain.GeneratedLetter
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		result, err = s.letters.ListByCase(ctx, tx, caseID)
		return err
	})
	return result, err
}

func (s *LetterService) Download(ctx context.Context, agencyID, userID, letterID uuid.UUID) (io.ReadCloser, *domain.GeneratedLetter, error) {
	var letter *domain.GeneratedLetter
	err := postgres.WithTenant(ctx, s.db, agencyID, userID, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		letter, err = s.letters.GetByID(ctx, tx, letterID)
		return err
	})
	if err != nil || letter == nil {
		return nil, nil, fmt.Errorf("letter not found")
	}
	reader, err := s.storage.Download(ctx, letter.FileKey)
	if err != nil {
		return nil, nil, err
	}
	return reader, letter, nil
}
