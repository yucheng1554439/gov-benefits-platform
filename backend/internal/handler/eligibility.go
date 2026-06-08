package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
)

type EligibilityHandler struct {
	eligibility *service.EligibilityService
	cases       *service.CaseService
}

func NewEligibilityHandler(eligibility *service.EligibilityService, cases *service.CaseService) *EligibilityHandler {
	return &EligibilityHandler{eligibility: eligibility, cases: cases}
}

func (h *EligibilityHandler) Evaluate(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	caseObj, err := h.cases.Get(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil || caseObj == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}
	eval, err := h.eligibility.Evaluate(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, caseObj.ProgramID)
	if err != nil {
		if errors.Is(err, service.ErrNoEligibilityRuleVersion) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, eval)
}

func (h *EligibilityHandler) Get(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	eval, err := h.eligibility.GetLatest(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, eval)
}

type BenefitHandler struct {
	benefit *service.BenefitService
	cases   *service.CaseService
}

func NewBenefitHandler(benefit *service.BenefitService, cases *service.CaseService) *BenefitHandler {
	return &BenefitHandler{benefit: benefit, cases: cases}
}

func (h *BenefitHandler) Calculate(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	caseObj, err := h.cases.Get(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil || caseObj == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}
	calc, err := h.benefit.Calculate(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, caseObj.ProgramID)
	if err != nil {
		if errors.Is(err, service.ErrNoBenefitRuleVersion) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, calc)
}

func (h *BenefitHandler) Get(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	calc, err := h.benefit.GetLatest(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, calc)
}
