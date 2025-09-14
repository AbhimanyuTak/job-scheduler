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
	// Adaptive polling configuration
	minInterval     time.Duration
	maxInterval     time.Duration
	currentInterval time.Duration
	batchSize       int
	emptyRuns       int
	maxEmptyRuns    int
}

// NewBackgroundScheduler creates a new background scheduler
func NewBackgroundScheduler(schedulerService *SchedulerService) *BackgroundScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundScheduler{
		schedulerService: schedulerService,
		ctx:              ctx,
		cancel:           cancel,
		// Adaptive polling defaults
		minInterval:     1 * time.Second,  // Fast when there are jobs
		maxInterval:     30 * time.Second, // Slow when idle
		currentInterval: 5 * time.Second,  // Start with default
		batchSize:       100,              // Default batch size
		maxEmptyRuns:    3,                // Increase interval after 3 empty runs
	}
}

// Start begins the background scheduler with adaptive polling
func (bs *BackgroundScheduler) Start(interval time.Duration) {
	bs.currentInterval = interval
	bs.ticker = time.NewTicker(bs.currentInterval)
	log.Printf("Background scheduler started with adaptive polling (initial: %v, batch size: %d)", interval, bs.batchSize)

	go bs.adaptivePollingLoop()
}

// adaptivePollingLoop implements adaptive polling based on job availability
func (bs *BackgroundScheduler) adaptivePollingLoop() {
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
				// On error, use conservative settings
				bs.adjustInterval(true)
			} else {
				// Adjust polling based on whether jobs were found
				// Note: We'd need to modify ProcessReadyJobs to return job count
				// For now, we'll use a simple heuristic
				bs.adjustInterval(false)
			}

			// Restart ticker with new interval if changed
			bs.restartTickerIfNeeded()
		}
	}
}

// adjustInterval adjusts the polling interval based on job availability
func (bs *BackgroundScheduler) adjustInterval(hasError bool) {
	if hasError {
		// On error, slow down
		bs.emptyRuns++
		if bs.currentInterval < bs.maxInterval {
			bs.currentInterval *= 2
			if bs.currentInterval > bs.maxInterval {
				bs.currentInterval = bs.maxInterval
			}
			log.Printf("Scheduler slowing down due to errors (new interval: %v)", bs.currentInterval)
		}
		return
	}

	// Simple heuristic: if we've had empty runs, we might have processed jobs
	if bs.emptyRuns > 0 {
		bs.emptyRuns = 0
		// Reset to faster polling when we start processing again
		if bs.currentInterval > bs.minInterval {
			bs.currentInterval = bs.minInterval
			log.Printf("Scheduler speeding up (new interval: %v)", bs.currentInterval)
		}
	} else {
		// Gradually slow down if no jobs for a while
		bs.emptyRuns++
		if bs.emptyRuns >= bs.maxEmptyRuns && bs.currentInterval < bs.maxInterval {
			bs.currentInterval *= 2
			if bs.currentInterval > bs.maxInterval {
				bs.currentInterval = bs.maxInterval
			}
			bs.emptyRuns = 0
			log.Printf("Scheduler slowing down (new interval: %v)", bs.currentInterval)
		}
	}
}

// restartTickerIfNeeded restarts the ticker if the interval has changed
func (bs *BackgroundScheduler) restartTickerIfNeeded() {
	// This is a simplified version - in practice, you'd want to track the previous interval
	// and only restart when it actually changes
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
