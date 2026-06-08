package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/service"
)

type AgencyHandler struct {
	agencies *service.AgencyService
}

func NewAgencyHandler(agencies *service.AgencyService) *AgencyHandler {
	return &AgencyHandler{agencies: agencies}
}

func (h *AgencyHandler) List(c *gin.Context) {
	agencies, err := h.agencies.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agencies})
}

func (h *AgencyHandler) ListPrograms(c *gin.Context) {
	agencyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agency id"})
		return
	}
	agencyPrograms, err := h.agencies.GetPrograms(c.Request.Context(), agencyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	programs := make([]domain.Program, 0, len(agencyPrograms))
	for _, ap := range agencyPrograms {
		if ap.Program != nil {
			programs = append(programs, *ap.Program)
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": programs})
}
