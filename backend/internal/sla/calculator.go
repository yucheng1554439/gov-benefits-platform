package sla

import (
	"time"

	"github.com/govbenefits/platform/internal/domain"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) ComputeDueDate(policy *domain.SLAPolicy, start time.Time) time.Time {
	if policy.BusinessDaysOnly {
		return addBusinessDays(start, policy.TargetDays)
	}
	return start.AddDate(0, 0, policy.TargetDays)
}

func (c *Calculator) ComputeStatus(tracking *domain.CaseSLATracking, policy *domain.SLAPolicy, now time.Time) string {
	if now.After(tracking.DueAt) {
		return "breached"
	}

	total := tracking.DueAt.Sub(now)
	elapsed := policy.TargetDays
	if elapsed > 0 {
		warningPct := float64(policy.WarningThresholdPct) / 100.0
		remainingDays := total.Hours() / 24
		if remainingDays/float64(elapsed) <= (1-warningPct) {
			return "at_risk"
		}
	}
	return "on_track"
}

func addBusinessDays(start time.Time, days int) time.Time {
	current := start
	added := 0
	for added < days {
		current = current.AddDate(0, 0, 1)
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			added++
		}
	}
	return current
}
