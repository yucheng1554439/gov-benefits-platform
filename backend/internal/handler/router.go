package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
	"github.com/govbenefits/platform/pkg/metrics"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, statusLabel(c.Writer.Status())).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

func statusLabel(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	default:
		return "2xx"
	}
}

func RegisterRoutes(r *gin.Engine, deps *Dependencies) {
	r.Use(middleware.CORS(deps.CORSOrigin))
	r.Use(middleware.RequestID())
	r.Use(MetricsMiddleware())

	health := NewHealthHandler(deps.DB, deps.Queue, deps.Storage)
	r.GET("/health", health.Health)
	r.GET("/ready", health.Ready)
	r.GET("/metrics", deps.MetricsHandler)

	v1 := r.Group("/api/v1")
	{
		auth := NewAuthHandler(deps.Auth)
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)
		v1.POST("/auth/refresh", auth.Refresh)

		agencies := NewAgencyHandler(deps.Agencies)
		v1.GET("/agencies", agencies.List)
		v1.GET("/agencies/:id/programs", agencies.ListPrograms)

		protected := v1.Group("")
		protected.Use(deps.AuthMiddleware.RequireAuth())
		protected.Use(deps.AgencyResolver.ResolveAgency())
		protected.Use(middleware.TenantContext())
		{
			protected.GET("/auth/me", auth.Me)

			applications := NewApplicationHandler(deps.Cases, deps.Agencies)
			protected.POST("/applications", applications.Create)

			casesHandler := NewCaseHandler(deps.Cases)
			docs := NewDocumentHandler(deps.Documents)
			eligibility := NewEligibilityHandler(deps.Eligibility, deps.Cases)
			benefit := NewBenefitHandler(deps.Benefit, deps.Cases)
			appeals := NewAppealHandler(deps.Appeals)
			fraud := NewFraudHandler(deps.Fraud)
			slaHandler := NewSLAHandler(deps.SLA)
			letters := NewLetterHandler(deps.Letters)

			casesGroup := protected.Group("/cases")
			{
				casesGroup.GET("", casesHandler.List)

				caseItem := casesGroup.Group("/:id")
				{
					caseItem.GET("", casesHandler.Get)
					caseItem.PATCH("", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), casesHandler.Update)
					caseItem.PATCH("/status", casesHandler.UpdateStatus)
					caseItem.GET("/transitions", casesHandler.AvailableTransitions)
					caseItem.GET("/workflow", casesHandler.WorkflowHistory)

					caseItem.GET("/documents", docs.List)
					caseItem.POST("/documents", docs.Upload)

					caseItem.POST("/eligibility/evaluate", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), eligibility.Evaluate)
					caseItem.GET("/eligibility", eligibility.Get)

					caseItem.POST("/benefit/calculate", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), benefit.Calculate)
					caseItem.GET("/benefit", benefit.Get)

					caseItem.GET("/appeals", appeals.List)
					caseItem.POST("/appeal", appeals.FileOnCase)

					caseItem.POST("/fraud/scan", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), fraud.Scan)
					caseItem.GET("/fraud", fraud.List)

					caseItem.GET("/sla", slaHandler.Get)

					caseItem.POST("/letters", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), letters.Generate)
					caseItem.GET("/letters", letters.List)
				}
			}

			protected.GET("/documents/:id/download", docs.Download)
			protected.GET("/letters/:id/download", letters.Download)
			protected.PATCH("/documents/:id/verify", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), docs.Verify)

			protected.POST("/appeals", appeals.File)
			protected.GET("/appeals", middleware.RequireAgencyRoles("worker", "supervisor", "admin", "case_worker"), appeals.ListAgency)
			protected.POST("/appeals/:id/decide", middleware.RequireAgencyRoles("supervisor", "admin"), appeals.Decide)

			protected.POST("/fraud/:id/review", middleware.RequireAgencyRoles("supervisor", "admin"), fraud.Review)

			protected.GET("/sla/breached", middleware.RequireAgencyRoles("supervisor", "admin"), slaHandler.ListBreached)

			analytics := NewAnalyticsHandler(deps.Analytics)
			protected.GET("/analytics/summary", middleware.RequireAgencyRoles("supervisor", "admin"), analytics.Summary)

			reports := NewReportHandler(deps.Reports)
			protected.POST("/reports", middleware.RequireAgencyRoles("supervisor", "admin"), reports.Create)
			protected.GET("/reports", reports.List)

			audit := NewAuditHandler(deps.Audit)
			protected.GET("/audit-logs", middleware.RequireRoles("admin", "supervisor"), audit.List)

			notifications := NewNotificationHandler(deps.Notifications)
			protected.GET("/notifications", notifications.List)
			protected.PATCH("/notifications/:id/read", notifications.MarkRead)

			flags := NewFeatureFlagHandler(deps.FeatureFlags)
			protected.GET("/feature-flags", flags.List)
			protected.PUT("/feature-flags", middleware.RequireRoles("admin"), flags.Upsert)

			admin := NewAdminHandler(deps.Retention, deps.Workflow)
			protected.GET("/admin/retention-policies", middleware.RequireRoles("admin"), admin.ListRetentionPolicies)
			protected.GET("/admin/workflow-transitions", middleware.RequireRoles("admin"), admin.ListWorkflowTransitions)

			rules := NewRulesHandler(deps.Eligibility)
			protected.GET("/admin/eligibility-rules", middleware.RequireRoles("admin"), rules.List)
			protected.POST("/admin/eligibility-rules/:id/simulate", middleware.RequireRoles("admin"), rules.Simulate)
		}
	}
}

type Dependencies struct {
	CORSOrigin     string
	DB             Pinger
	Queue          QueuePinger
	Storage        StoragePinger
	MetricsHandler gin.HandlerFunc
	AuthMiddleware *middleware.AuthMiddleware
	AgencyResolver *middleware.AgencyResolver
	Auth           *service.AuthService
	Agencies       *service.AgencyService
	Cases          *service.CaseService
	Documents      *service.DocumentService
	Eligibility    *service.EligibilityService
	Benefit        *service.BenefitService
	Appeals        *service.AppealService
	Fraud          *service.FraudService
	SLA            *service.SLAService
	Letters        *service.LetterService
	Analytics      *service.AnalyticsService
	Reports        *service.ReportService
	Audit          *service.AuditService
	Notifications  *service.NotificationService
	FeatureFlags   *service.FeatureFlagService
	Retention      *service.RetentionService
	Workflow       *service.WorkflowService
}
