package scheduler

import (
	"time"

	"github.com/imkerbos/ACME-Console/internal/logger"
	"github.com/imkerbos/ACME-Console/internal/service"
)

// Scheduler handles periodic tasks
type Scheduler struct {
	notificationSvc *service.NotificationService
	stopChan        chan struct{}
	interval        time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler(notificationSvc *service.NotificationService) *Scheduler {
	return &Scheduler{
		notificationSvc: notificationSvc,
		stopChan:        make(chan struct{}),
		interval:        6 * time.Hour, // Check every 6 hours
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	logger.Info("Starting notification scheduler",
		logger.String("interval", s.interval.String()),
	)

	// Run immediately on start
	go s.runNotificationCheck()

	// Then run periodically
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runNotificationCheck()
			case <-s.stopChan:
				ticker.Stop()
				logger.Info("Notification scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) runNotificationCheck() {
	logger.Info("Running certificate expiry notification check")

	if err := s.notificationSvc.CheckAndNotify(); err != nil {
		logger.Error("Failed to check and notify",
			logger.Err(err),
		)
	} else {
		logger.Info("Certificate expiry notification check completed")
	}
}
