package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
)

type AppealHandler struct {
	appeals *service.AppealService
}

func NewAppealHandler(appeals *service.AppealService) *AppealHandler {
	return &AppealHandler{appeals: appeals}
}

type fileAppealRequest struct {
	CaseID  string `json:"case_id" binding:"required"`
	Grounds string `json:"grounds" binding:"required"`
}

func (h *AppealHandler) File(c *gin.Context) {
	var req fileAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	caseID, _ := uuid.Parse(req.CaseID)
	appeal, err := h.appeals.File(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, req.Grounds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, appeal)
}

func (h *AppealHandler) List(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	appeals, err := h.appeals.List(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": appeals})
}

func (h *AppealHandler) ListAgency(c *gin.Context) {
	pendingOnly := c.Query("pending") == "true"
	appeals, err := h.appeals.ListAgency(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), pendingOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if appeals == nil {
		appeals = []domain.Appeal{}
	}
	c.JSON(http.StatusOK, gin.H{"data": appeals})
}

type fileAppealOnCaseRequest struct {
	Grounds string `json:"grounds" binding:"required"`
}

func (h *AppealHandler) FileOnCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	var req fileAppealOnCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	appeal, err := h.appeals.File(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, req.Grounds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, appeal)
}

type appealDecisionRequest struct {
	Decision  string `json:"decision" binding:"required"`
	Rationale string `json:"rationale"`
}

func (h *AppealHandler) Decide(c *gin.Context) {
	appealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid appeal id"})
		return
	}
	var req appealDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.appeals.Decide(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), appealID, req.Decision, req.Rationale); err != nil {
		if errors.Is(err, service.ErrAppealAlreadyDecided) {
			c.JSON(http.StatusConflict, gin.H{"error": "This appeal has already been decided."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "decision recorded"})
}

type FraudHandler struct {
	fraud *service.FraudService
}

func NewFraudHandler(fraud *service.FraudService) *FraudHandler {
	return &FraudHandler{fraud: fraud}
}

func (h *FraudHandler) Scan(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	flags, err := h.fraud.ScanCase(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if flags == nil {
		flags = []domain.FraudFlag{}
	}
	c.JSON(http.StatusOK, gin.H{"data": flags, "count": len(flags)})
}

func (h *FraudHandler) List(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	flags, err := h.fraud.ListFlags(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": flags})
}

type reviewFraudRequest struct {
	Disposition string `json:"disposition" binding:"required"`
	Notes       string `json:"notes"`
}

func (h *FraudHandler) Review(c *gin.Context) {
	flagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flag id"})
		return
	}
	var req reviewFraudRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.fraud.ReviewFlag(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), flagID, req.Disposition, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "review recorded"})
}

type SLAHandler struct {
	sla *service.SLAService
}

func NewSLAHandler(sla *service.SLAService) *SLAHandler {
	return &SLAHandler{sla: sla}
}

func (h *SLAHandler) Get(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	tracking, err := h.sla.GetTracking(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tracking)
}

func (h *SLAHandler) ListBreached(c *gin.Context) {
	trackings, err := h.sla.ListBreached(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trackings})
}

type LetterHandler struct {
	letters *service.LetterService
}

func NewLetterHandler(letters *service.LetterService) *LetterHandler {
	return &LetterHandler{letters: letters}
}

type generateLetterRequest struct {
	LetterType string `json:"letter_type" binding:"required"`
}

func (h *LetterHandler) Generate(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	var req generateLetterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	letter, err := h.letters.Generate(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, req.LetterType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, letter)
}

func (h *LetterHandler) List(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	letters, err := h.letters.List(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": letters})
}

func (h *LetterHandler) Download(c *gin.Context) {
	letterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid letter id"})
		return
	}
	reader, letter, err := h.letters.Download(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), letterID)
	if err != nil || letter == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "letter not found"})
		return
	}
	defer reader.Close()

	filename := fmt.Sprintf("%s_%s.pdf", letter.CaseID, letter.LetterType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	_, _ = io.Copy(c.Writer, reader)
}
