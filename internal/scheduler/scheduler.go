package scheduler

import (
	"time"

	"github.com/imkerbos/ACME-Console/internal/logger"
	"github.com/imkerbos/ACME-Console/internal/service"
)

// Scheduler handles periodic tasks
type Scheduler struct {
	notificationSvc *service.NotificationService
	renewalSvc      *service.RenewalService
	stopChan        chan struct{}
	interval        time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler(notificationSvc *service.NotificationService, renewalSvc *service.RenewalService) *Scheduler {
	return &Scheduler{
		notificationSvc: notificationSvc,
		renewalSvc:      renewalSvc,
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
	go s.runRenewalCheck()

	// Then run periodically
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runNotificationCheck()
				s.runRenewalCheck()
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

func (s *Scheduler) runRenewalCheck() {
	if s.renewalSvc == nil {
		return
	}
	logger.Info("Running certificate renewal check")

	if err := s.renewalSvc.CheckAndRenew(); err != nil {
		logger.Error("Failed to check and renew",
			logger.Err(err),
		)
	} else {
		logger.Info("Certificate renewal check completed")
	}
}
