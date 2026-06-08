package fraud

import (
	"github.com/govbenefits/platform/internal/domain"
)

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

type CheckResult struct {
	Flags []domain.FraudFlag
}

func (d *Detector) CheckCase(c *domain.Case, app *domain.Application, existingFlags int) *CheckResult {
	var flags []domain.FraudFlag

	if app != nil && app.AnnualIncome < 0 {
		flags = append(flags, domain.FraudFlag{
			AgencyID: c.AgencyID,
			CaseID:   c.ID,
			FlagType: "negative_income",
			Severity: "high",
			Evidence: map[string]any{"annual_income": app.AnnualIncome},
			Status:   "open",
		})
	}

	if app != nil && app.HouseholdSize > 20 {
		flags = append(flags, domain.FraudFlag{
			AgencyID: c.AgencyID,
			CaseID:   c.ID,
			FlagType: "suspicious_household_size",
			Severity: "medium",
			Evidence: map[string]any{"household_size": app.HouseholdSize},
			Status:   "open",
		})
	}

	if app != nil && app.AnnualIncome > 500000 {
		flags = append(flags, domain.FraudFlag{
			AgencyID: c.AgencyID,
			CaseID:   c.ID,
			FlagType: "high_income_anomaly",
			Severity: "low",
			Evidence: map[string]any{"annual_income": app.AnnualIncome},
			Status:   "open",
		})
	}

	if existingFlags >= 3 {
		flags = append(flags, domain.FraudFlag{
			AgencyID: c.AgencyID,
			CaseID:   c.ID,
			FlagType: "repeat_offender",
			Severity: "high",
			Evidence: map[string]any{"existing_flags": existingFlags},
			Status:   "open",
		})
	}

	return &CheckResult{Flags: flags}
}
