package scheduler

import (
	"context"
	"log"
	"log/slog"
	"time"
)

// Job represents a scheduled task
type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Scheduler manages and runs scheduled jobs
type Scheduler struct {
	jobs    []Job
	logger  *slog.Logger
	stopCh  chan struct{}
}

// New creates a new scheduler
func New(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		jobs:   make([]Job, 0),
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// AddJob registers a new job to be scheduled
func (s *Scheduler) AddJob(job Job) {
	s.jobs = append(s.jobs, job)
	s.logger.Info("job registered", "name", job.Name, "interval", job.Interval)
}

// Start begins running all registered jobs
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("scheduler starting", "jobs", len(s.jobs))

	for _, job := range s.jobs {
		go s.runJob(ctx, job)
	}
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("scheduler stopping")
	close(s.stopCh)
}

// runJob executes a single job on its interval
func (s *Scheduler) runJob(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run immediately on start
	s.executeJob(ctx, job)

	for {
		select {
		case <-ticker.C:
			s.executeJob(ctx, job)
		case <-s.stopCh:
			s.logger.Info("job stopped", "name", job.Name)
			return
		case <-ctx.Done():
			s.logger.Info("job stopped due to context cancellation", "name", job.Name)
			return
		}
	}
}

// executeJob runs a single job with error handling
func (s *Scheduler) executeJob(ctx context.Context, job Job) {
	s.logger.Info("job starting", "name", job.Name)

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("job panicked", "name", job.Name, "panic", r)
		}
	}()

	if err := job.Run(ctx); err != nil {
		s.logger.Error("job failed", "name", job.Name, "error", err)
	} else {
		s.logger.Info("job completed successfully", "name", job.Name)
	}
}

// WaitForShutdown blocks until the scheduler is stopped
func (s *Scheduler) WaitForShutdown() {
	<-s.stopCh
	log.Println("scheduler shutdown complete")
}
