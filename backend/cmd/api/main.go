package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/assignment"
	"github.com/govbenefits/platform/internal/benefit"
	"github.com/govbenefits/platform/internal/config"
	"github.com/govbenefits/platform/internal/eligibility"
	"github.com/govbenefits/platform/internal/events"
	"github.com/govbenefits/platform/internal/events/subscribers"
	"github.com/govbenefits/platform/internal/fraud"
	"github.com/govbenefits/platform/internal/handler"
	"github.com/govbenefits/platform/internal/jobs"
	"github.com/govbenefits/platform/internal/letters"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/service"
	"github.com/govbenefits/platform/internal/sla"
	"github.com/govbenefits/platform/internal/storage"
	"github.com/govbenefits/platform/internal/workflow"
	jwtpkg "github.com/govbenefits/platform/pkg/jwt"
	"github.com/govbenefits/platform/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Environment)
	ctx := context.Background()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	migrationsDir := filepath.Join("migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		migrationsDir = filepath.Join("backend", "migrations")
	}
	if err := db.RunMigrations(ctx, migrationsDir); err != nil {
		log.Error("run migrations", "error", err)
		os.Exit(1)
	}

	store, err := storage.NewProvider(ctx, cfg.StorageDriver, cfg.LocalStoragePath, storage.S3Config{
		Endpoint: cfg.S3Endpoint, Bucket: cfg.S3Bucket,
		AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey, Region: cfg.S3Region,
	})
	if err != nil {
		log.Error("init storage", "error", err)
		os.Exit(1)
	}

	queue, err := jobs.NewQueue(cfg.RedisURL)
	if err != nil {
		log.Error("init queue", "error", err)
		os.Exit(1)
	}
	defer queue.Close()

	jwtManager := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	eventBus := events.NewBus(log)

	userRepo := postgres.NewUserRepository(db)
	agencyResolver := middleware.NewAgencyResolver(userRepo)
	agencyRepo := postgres.NewAgencyRepository(db)
	caseRepo := postgres.NewCaseRepository(db)
	appRepo := postgres.NewApplicationRepository(db)
	docRepo := postgres.NewDocumentRepository(db)
	workflowRepo := postgres.NewWorkflowRepository(db)
	eligRepo := postgres.NewEligibilityRepository(db)
	benefitRepo := postgres.NewBenefitRepository(db)
	fraudRepo := postgres.NewFraudRepository(db)
	appealRepo := postgres.NewAppealRepository(db)
	letterRepo := postgres.NewLetterRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	reportRepo := postgres.NewReportRepository(db)
	flagRepo := postgres.NewFeatureFlagRepository(db)
	slaRepo := postgres.NewSLARepository(db)
	workerRepo := postgres.NewWorkerRepository(db)

	auditSub := subscribers.NewAuditSubscriber(auditRepo, log)
	notifSub := subscribers.NewNotificationSubscriber(notificationRepo, log)
	slaSub := subscribers.NewSLASubscriber(slaRepo, log)
	letterSub := subscribers.NewLetterSubscriber(queue, log)

	eventBus.SubscribeAll(auditSub.Handle)
	eventBus.Subscribe(events.EventCaseCreated, notifSub.Handle)
	eventBus.Subscribe(events.EventCaseStatusChanged, notifSub.Handle)
	eventBus.Subscribe(events.EventCaseStatusChanged, slaSub.Handle)
	eventBus.Subscribe(events.EventCaseCreated, slaSub.Handle)
	eventBus.Subscribe(events.EventCaseStatusChanged, letterSub.Handle)

	stateMachine := workflow.NewStateMachine(workflowRepo)
	allocator := assignment.NewAllocator(workerRepo)
	slaCalc := sla.NewCalculator()
	eligEval := eligibility.NewEvaluator()
	benefitCalc := benefit.NewCalculator()
	fraudDetector := fraud.NewDetector()
	pdfGen := letters.NewPDFGenerator()

	authSvc := service.NewAuthService(userRepo, jwtManager)
	agencySvc := service.NewAgencyService(agencyRepo)
	caseSvc := service.NewCaseService(db, caseRepo, appRepo, agencyRepo, userRepo, stateMachine, allocator, workerRepo, slaRepo, slaCalc, eventBus)
	docSvc := service.NewDocumentService(db, docRepo, store, eventBus)
	eligSvc := service.NewEligibilityService(db, eligRepo, appRepo, eligEval, eventBus)
	benefitSvc := service.NewBenefitService(db, benefitRepo, appRepo, benefitCalc, eventBus)
	fraudSvc := service.NewFraudService(db, fraudRepo, caseRepo, appRepo, fraudDetector, eventBus)
	appealSvc := service.NewAppealService(db, appealRepo, caseSvc, eventBus)
	slaSvc := service.NewSLAService(slaRepo, slaCalc)
	letterSvc := service.NewLetterService(db, letterRepo, caseRepo, userRepo, agencyRepo, benefitRepo, pdfGen, store, eventBus)
	analyticsSvc := service.NewAnalyticsService(reportRepo, fraudRepo)
	reportSvc := service.NewReportService(reportRepo, &reportQueueAdapter{queue: queue})
	auditSvc := service.NewAuditService(auditRepo)
	notifSvc := service.NewNotificationService(notificationRepo)
	flagSvc := service.NewFeatureFlagService(flagRepo)
	retentionSvc := service.NewRetentionService(db)
	workflowSvc := service.NewWorkflowService(workflowRepo)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	handler.RegisterRoutes(r, &handler.Dependencies{
		CORSOrigin:     cfg.CORSOrigin,
		DB:             db,
		Queue:          queue,
		Storage:        store,
		MetricsHandler: gin.WrapH(promhttp.Handler()),
		AuthMiddleware: authMiddleware,
		AgencyResolver: agencyResolver,
		Auth:           authSvc,
		Agencies:       agencySvc,
		Cases:          caseSvc,
		Documents:      docSvc,
		Eligibility:    eligSvc,
		Benefit:        benefitSvc,
		Appeals:        appealSvc,
		Fraud:          fraudSvc,
		SLA:            slaSvc,
		Letters:        letterSvc,
		Analytics:      analyticsSvc,
		Reports:        reportSvc,
		Audit:          auditSvc,
		Notifications:  notifSvc,
		FeatureFlags:   flagSvc,
		Retention:      retentionSvc,
		Workflow:       workflowSvc,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("starting api server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

type reportQueueAdapter struct {
	queue *jobs.Queue
}

func (a *reportQueueAdapter) EnqueueReport(ctx context.Context, reportID uuid.UUID) error {
	return a.queue.Enqueue(ctx, jobs.Job{
		Type: jobs.JobGenerateReport,
		Payload: map[string]any{
			"report_id": reportID.String(),
		},
	})
}
