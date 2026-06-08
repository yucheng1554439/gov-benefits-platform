package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/domain"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/service"
)

type AnalyticsHandler struct {
	analytics *service.AnalyticsService
}

func NewAnalyticsHandler(analytics *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Summary(c *gin.Context) {
	summary, err := h.analytics.Summary(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

type ReportHandler struct {
	reports *service.ReportService
}

func NewReportHandler(reports *service.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

type reportRequest struct {
	ReportType string         `json:"report_type" binding:"required"`
	Params     map[string]any `json:"params"`
}

func (h *ReportHandler) Create(c *gin.Context) {
	var req reportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.reports.Request(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), req.ReportType, req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, report)
}

func (h *ReportHandler) List(c *gin.Context) {
	reports, err := h.reports.List(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reports})
}

type AuditHandler struct {
	audit *service.AuditService
}

func NewAuditHandler(audit *service.AuditService) *AuditHandler {
	return &AuditHandler{audit: audit}
}

func (h *AuditHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filter := postgres.AuditListFilter{
		Action: c.Query("action"),
		Search: c.Query("search"),
		Offset: offset,
		Limit:  limit,
	}
	logs, total, err := h.audit.List(c.Request.Context(), middleware.GetAgencyID(c), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if logs == nil {
		logs = []domain.AuditLog{}
	}
	c.JSON(http.StatusOK, gin.H{"data": logs, "total": total, "offset": offset, "limit": limit})
}

type NotificationHandler struct {
	notifications *service.NotificationService
}

func NewNotificationHandler(notifications *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

func (h *NotificationHandler) List(c *gin.Context) {
	unreadOnly := c.Query("unread") == "true"
	items, err := h.notifications.List(c.Request.Context(), middleware.GetUserID(c), unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.notifications.MarkRead(c.Request.Context(), id, middleware.GetUserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked read"})
}

type FeatureFlagHandler struct {
	flags *service.FeatureFlagService
}

func NewFeatureFlagHandler(flags *service.FeatureFlagService) *FeatureFlagHandler {
	return &FeatureFlagHandler{flags: flags}
}

func (h *FeatureFlagHandler) List(c *gin.Context) {
	flags, err := h.flags.List(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": flags})
}

type upsertFlagRequest struct {
	FlagKey    string `json:"flag_key" binding:"required"`
	IsEnabled  bool   `json:"is_enabled"`
	RolloutPct int    `json:"rollout_pct"`
}

func (h *FeatureFlagHandler) Upsert(c *gin.Context) {
	var req upsertFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	flag := &domain.FeatureFlag{
		AgencyID: middleware.GetAgencyID(c), FlagKey: req.FlagKey,
		IsEnabled: req.IsEnabled, RolloutPct: req.RolloutPct,
	}
	if err := h.flags.Upsert(c.Request.Context(), flag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, flag)
}

type AdminHandler struct {
	retention *service.RetentionService
	workflow  *service.WorkflowService
}

func NewAdminHandler(retention *service.RetentionService, workflow *service.WorkflowService) *AdminHandler {
	return &AdminHandler{retention: retention, workflow: workflow}
}

func (h *AdminHandler) ListRetentionPolicies(c *gin.Context) {
	policies, err := h.retention.ListPolicies(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policies})
}

func (h *AdminHandler) ListWorkflowTransitions(c *gin.Context) {
	transitions, err := h.workflow.ListTransitions(c.Request.Context(), middleware.GetAgencyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": transitions})
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type QueuePinger interface {
	Ping(ctx context.Context) error
}

type StoragePinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db      Pinger
	queue   QueuePinger
	storage StoragePinger
}

func NewHealthHandler(db Pinger, queue QueuePinger, storage StoragePinger) *HealthHandler {
	return &HealthHandler{db: db, queue: queue, storage: storage}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	checks := gin.H{}
	if err := h.db.Ping(ctx); err != nil {
		checks["postgres"] = "unavailable"
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
		return
	}
	checks["postgres"] = "ok"

	if h.queue != nil {
		if err := h.queue.Ping(ctx); err != nil {
			checks["redis"] = "unavailable"
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
			return
		}
		checks["redis"] = "ok"
	}

	if h.storage != nil {
		if err := h.storage.Ping(ctx); err != nil {
			checks["minio"] = "unavailable"
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
			return
		}
		checks["minio"] = "ok"
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready", "checks": checks})
}
