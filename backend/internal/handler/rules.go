package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
)

type RulesHandler struct {
	eligibility *service.EligibilityService
}

func NewRulesHandler(eligibility *service.EligibilityService) *RulesHandler {
	return &RulesHandler{eligibility: eligibility}
}

func (h *RulesHandler) List(c *gin.Context) {
	rules, err := h.eligibility.ListRules(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rules == nil {
		rules = []domain.EligibilityRuleDetail{}
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

type simulateRuleRequest struct {
	AnnualIncome  *float64 `json:"annual_income"`
	HouseholdSize *int     `json:"household_size"`
}

func (h *RulesHandler) Simulate(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}
	var req simulateRuleRequest
	_ = c.ShouldBindJSON(&req)

	data := map[string]any{}
	if req.AnnualIncome != nil {
		data["annual_income"] = *req.AnnualIncome
	}
	if req.HouseholdSize != nil {
		data["household_size"] = float64(*req.HouseholdSize)
	}
	if len(data) == 0 {
		data = nil
	}

	result, err := h.eligibility.SimulateRule(c.Request.Context(), middleware.GetAgencyID(c), ruleID, data)
	if err != nil {
		if errors.Is(err, service.ErrNoEligibilityRuleVersion) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
