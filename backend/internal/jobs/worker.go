package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/govbenefits/platform/pkg/metrics"
)

type LetterService interface {
	GenerateFromJob(ctx context.Context, caseID, agencyID uuid.UUID, letterType string) error
}

type ReportService interface {
	ProcessFromJob(ctx context.Context, reportID uuid.UUID) error
}

type Worker struct {
	queue    *Queue
	letters  LetterService
	reports  ReportService
	log      *slog.Logger
	stopCh   chan struct{}
}

func NewWorker(queue *Queue, letters LetterService, reports ReportService, log *slog.Logger) *Worker {
	return &Worker{
		queue:   queue,
		letters: letters,
		reports: reports,
		log:     log,
		stopCh:  make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.log.Info("job worker started")
	for {
		select {
		case <-ctx.Done():
			w.log.Info("job worker stopping")
			return
		case <-w.stopCh:
			return
		default:
			job, err := w.queue.Dequeue(ctx, 5*time.Second)
			if err != nil {
				w.log.Error("dequeue job", "error", err)
				continue
			}
			if job == nil {
				continue
			}
			w.process(ctx, job)
		}
	}
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

func (w *Worker) process(ctx context.Context, job *Job) {
	w.log.Info("processing job", "type", job.Type, "id", job.ID)
	var err error

	switch job.Type {
	case JobGenerateLetter:
		caseID, _ := uuid.Parse(stringVal(job.Payload["case_id"]))
		agencyID, _ := uuid.Parse(stringVal(job.Payload["agency_id"]))
		letterType := stringVal(job.Payload["letter_type"])
		err = w.letters.GenerateFromJob(ctx, caseID, agencyID, letterType)
	case JobGenerateReport:
		reportID, _ := uuid.Parse(stringVal(job.Payload["report_id"]))
		err = w.reports.ProcessFromJob(ctx, reportID)
	default:
		w.log.Warn("unknown job type", "type", job.Type)
		metrics.JobsProcessedTotal.WithLabelValues(string(job.Type), "unknown").Inc()
		return
	}

	status := "success"
	if err != nil {
		status = "error"
		w.log.Error("job failed", "type", job.Type, "error", err)
	}
	metrics.JobsProcessedTotal.WithLabelValues(string(job.Type), status).Inc()
}

func stringVal(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
