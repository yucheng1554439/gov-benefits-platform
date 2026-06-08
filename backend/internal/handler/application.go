package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
)

type ApplicationHandler struct {
	cases    *service.CaseService
	agencies *service.AgencyService
}

func NewApplicationHandler(cases *service.CaseService, agencies *service.AgencyService) *ApplicationHandler {
	return &ApplicationHandler{cases: cases, agencies: agencies}
}

type createApplicationRequest struct {
	AgencyID         string         `json:"agency_id" binding:"required"`
	ProgramID        string         `json:"program_id" binding:"required"`
	HouseholdSize    int            `json:"household_size" binding:"required,min=1"`
	AnnualIncome     float64        `json:"annual_income"`
	EmploymentStatus string         `json:"employment_status"`
	FormData         map[string]any `json:"form_data"`
	ZipCode          string         `json:"zip_code"`
	CensusTract      string         `json:"census_tract"`
	Priority         string         `json:"priority"`
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	var req createApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agencyID, err := uuid.Parse(req.AgencyID)
	if err != nil || agencyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please select a valid agency."})
		return
	}
	programID, err := uuid.Parse(req.ProgramID)
	if err != nil || programID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please select a valid program."})
		return
	}

	enabled, err := h.agencies.IsProgramEnabledForAgency(c.Request.Context(), agencyID, programID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to verify program availability."})
		return
	}
	if !enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The selected program is not available for this agency."})
		return
	}

	created, err := h.cases.CreateApplication(c.Request.Context(), domain.CreateApplicationInput{
		AgencyID: agencyID, CitizenID: middleware.GetUserID(c), ProgramID: programID,
		HouseholdSize: req.HouseholdSize, AnnualIncome: req.AnnualIncome,
		EmploymentStatus: req.EmploymentStatus, FormData: req.FormData,
		ZipCode: req.ZipCode, CensusTract: req.CensusTract, Priority: req.Priority,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to submit your application. Please try again."})
		return
	}
	c.JSON(http.StatusCreated, created)
}
