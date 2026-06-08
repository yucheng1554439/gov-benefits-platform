package featureflags

import (
	"context"
	"hash/fnv"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/repository/postgres"
)

type Service struct {
	repo *postgres.FeatureFlagRepository
}

func NewService(repo *postgres.FeatureFlagRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IsEnabled(ctx context.Context, agencyID uuid.UUID, flagKey string, userID uuid.UUID) (bool, error) {
	flag, err := s.repo.Get(ctx, agencyID, flagKey)
	if err != nil {
		return false, err
	}
	if flag == nil {
		return false, nil
	}
	if !flag.IsEnabled {
		return false, nil
	}
	if flag.RolloutPct >= 100 {
		return true, nil
	}
	if userID == uuid.Nil {
		return flag.RolloutPct > 0, nil
	}
	return bucket(userID, flagKey) < flag.RolloutPct, nil
}

func (s *Service) List(ctx context.Context, agencyID uuid.UUID) (map[string]bool, error) {
	flags, err := s.repo.List(ctx, agencyID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(flags))
	for _, f := range flags {
		result[f.FlagKey] = f.IsEnabled
	}
	return result, nil
}

func bucket(userID uuid.UUID, flagKey string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(userID.String() + flagKey))
	return int(h.Sum32() % 100)
}
