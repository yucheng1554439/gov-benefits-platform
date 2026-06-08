package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
)

type CaseHandler struct {
	cases *service.CaseService
}

func NewCaseHandler(cases *service.CaseService) *CaseHandler {
	return &CaseHandler{cases: cases}
}

func (h *CaseHandler) List(c *gin.Context) {
	filter := domain.CaseListFilter{
		Status: c.Query("status"),
		Limit:  50,
	}
	if middleware.ContainsRole(middleware.GetRoles(c), "citizen") {
		uid := middleware.GetUserID(c)
		filter.CitizenID = &uid
	}

	cases, err := h.cases.List(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cases})
}

func (h *CaseHandler) Get(c *gin.Context) {
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
	c.JSON(http.StatusOK, caseObj)
}

type updateCaseRequest struct {
	Priority    string `json:"priority"`
	ZipCode     string `json:"zip_code"`
	CensusTract string `json:"census_tract"`
}

func (h *CaseHandler) Update(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	var req updateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.cases.Update(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, req.Priority, req.ZipCode, req.CensusTract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

type statusRequest struct {
	ToStatus string `json:"to_status" binding:"required"`
	Reason   string `json:"reason"`
}

func (h *CaseHandler) UpdateStatus(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := middleware.ResolveWorkflowRole(c)
	event, err := h.cases.TransitionStatus(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), domain.StatusTransitionInput{
		CaseID: caseID, ToStatus: req.ToStatus, Role: role, Reason: req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *CaseHandler) AvailableTransitions(c *gin.Context) {
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
	transitions, err := h.cases.GetAvailableTransitions(
		c.Request.Context(),
		middleware.GetAgencyID(c),
		caseObj.Status,
		middleware.ResolveWorkflowRole(c),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": transitions})
}

func (h *CaseHandler) WorkflowHistory(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	events, err := h.cases.GetWorkflowHistory(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}
