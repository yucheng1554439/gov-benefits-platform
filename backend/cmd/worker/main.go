package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/govbenefits/platform/internal/config"
	"github.com/govbenefits/platform/internal/jobs"
	"github.com/govbenefits/platform/internal/letters"
	"github.com/govbenefits/platform/internal/repository/postgres"
	"github.com/govbenefits/platform/internal/service"
	"github.com/govbenefits/platform/internal/storage"
	"github.com/govbenefits/platform/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Exit(1)
	}
	log := logger.New(cfg.Environment)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	queue, err := jobs.NewQueue(cfg.RedisURL)
	if err != nil {
		log.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer queue.Close()

	store, err := storage.NewProvider(ctx, cfg.StorageDriver, cfg.LocalStoragePath, storage.S3Config{
		Endpoint: cfg.S3Endpoint, Bucket: cfg.S3Bucket,
		AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey, Region: cfg.S3Region,
	})
	if err != nil {
		log.Error("init storage", "error", err)
		os.Exit(1)
	}

	userRepo := postgres.NewUserRepository(db)
	agencyRepo := postgres.NewAgencyRepository(db)
	caseRepo := postgres.NewCaseRepository(db)
	benefitRepo := postgres.NewBenefitRepository(db)
	letterRepo := postgres.NewLetterRepository(db)
	reportRepo := postgres.NewReportRepository(db)

	pdfGen := letters.NewPDFGenerator()

	letterSvc := service.NewLetterService(db, letterRepo, caseRepo, userRepo, agencyRepo, benefitRepo, pdfGen, store, nil)
	reportSvc := service.NewReportService(reportRepo, nil)

	worker := jobs.NewWorker(queue, letterSvc, reportSvc, log)

	go worker.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	worker.Stop()
	log.Info("worker stopped")
}
