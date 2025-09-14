package services

import (
	"context"
	"log"
	"time"
)

// BackgroundScheduler runs continuously to process scheduled jobs
type BackgroundScheduler struct {
	schedulerService *SchedulerService
	ticker           *time.Ticker
	ctx              context.Context
	cancel           context.CancelFunc
	batchSize        int
}

// NewBackgroundScheduler creates a new background scheduler
func NewBackgroundScheduler(schedulerService *SchedulerService) *BackgroundScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundScheduler{
		schedulerService: schedulerService,
		ctx:              ctx,
		cancel:           cancel,
		batchSize:        100, // Default batch size
	}
}

// Start begins the background scheduler
func (bs *BackgroundScheduler) Start(interval time.Duration) {
	bs.ticker = time.NewTicker(interval)
	log.Printf("Background scheduler started with interval: %v, batch size: %d", interval, bs.batchSize)

	go bs.pollingLoop()
}

// pollingLoop runs the main scheduling loop
func (bs *BackgroundScheduler) pollingLoop() {
	for {
		select {
		case <-bs.ctx.Done():
			log.Println("Background scheduler stopped")
			return
		case <-bs.ticker.C:
			// Process jobs with current batch size
			err := bs.schedulerService.ProcessReadyJobs(bs.ctx, bs.batchSize)
			if err != nil {
				log.Printf("Error processing ready jobs: %v", err)
			}
		}
	}
}

// Stop stops the background scheduler
func (bs *BackgroundScheduler) Stop() {
	if bs.ticker != nil {
		bs.ticker.Stop()
	}
	bs.cancel()
	log.Println("Background scheduler stopped")
}

// IsRunning checks if the scheduler is running
func (bs *BackgroundScheduler) IsRunning() bool {
	select {
	case <-bs.ctx.Done():
		return false
	default:
		return true
	}
}
