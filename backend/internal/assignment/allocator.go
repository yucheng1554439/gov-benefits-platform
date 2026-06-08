package assignment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type Allocator struct {
	workerRepo *postgres.WorkerRepository
}

func NewAllocator(workerRepo *postgres.WorkerRepository) *Allocator {
	return &Allocator{workerRepo: workerRepo}
}

func (a *Allocator) AssignWorker(ctx context.Context, agencyID, programID uuid.UUID) (uuid.UUID, error) {
	workers, err := a.workerRepo.ListAvailable(ctx, agencyID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(workers) == 0 {
		return uuid.Nil, fmt.Errorf("no available workers")
	}

	programCode, err := a.workerRepo.GetProgramCode(ctx, programID)
	if err != nil {
		return uuid.Nil, err
	}

	var best uuid.UUID
	minCases := int(^uint(0) >> 1)

	for _, w := range workers {
		if !hasSpecialization(w.Specializations, programCode) && len(w.Specializations) > 0 {
			continue
		}
		if w.CurrentCaseCount < minCases {
			minCases = w.CurrentCaseCount
			best = w.UserID
		}
	}

	if best == uuid.Nil {
		best = workers[0].UserID
	}
	return best, nil
}

func hasSpecialization(specs []string, programCode string) bool {
	for _, s := range specs {
		if s == programCode {
			return true
		}
	}
	return false
}
