package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/deannos/email-checker-tool/internal/checker"
	"github.com/deannos/email-checker-tool/internal/output"
	"golang.org/x/time/rate"
)

// Pool manages the concurrency and rate limiting of email checks.
type Pool struct {
	workerCount int
	jobQueue    chan string
	output      output.Writer
	limiter     *rate.Limiter
	wg          sync.WaitGroup
}

// NewPool creates a new worker pool.
// workerCount: Number of concurrent goroutines.
// queueSize: Buffer size for the job channel (prevents blocking main thread if input is fast).
// w: Output writer (CSV, JSON, etc).
// rps: Requests per second limit (crucial for SMTP reputation).
func NewPool(workerCount int, queueSize int, w output.Writer, rps int) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobQueue:    make(chan string, queueSize),
		output:      w,
		// Allow small bursts (e.g., 5), but enforce rps long-term
		limiter: rate.NewLimiter(rate.Limit(rps), 5),
	}
}

// Start begins processing jobs. It blocks until all jobs are processed or context is cancelled.
func (p *Pool) Start(ctx context.Context) {
	// 1. Start Workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}

	// 2. Wait for all workers to finish
	p.wg.Wait()

	// 3. Ensure any buffered data in the writer is flushed (if applicable)
	if flusher, ok := p.output.(interface{ Flush() error }); ok {
		flusher.Flush()
	}
}

// worker listens for jobs and processes them with rate limiting.
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled (e.g. Ctrl+C), stop working
			return
		case email, ok := <-p.jobQueue:
			if !ok {
				// Channel closed, no more jobs
				return
			}

			// Rate Limiting: Wait here until we are allowed to make a request
			// This blocks the specific worker, not the whole pool
			if err := p.limiter.Wait(ctx); err != nil {
				// Context cancelled while waiting for rate limit
				return
			}

			// Perform the check
			result := checker.Check(ctx, email)

			// Write result immediately (streaming)
			if err := p.output.Write(result); err != nil {
				// In production, you might want to log this error
				// rather than crashing the worker
				fmt.Printf("Error writing result: %v\n", err)
			}
		}
	}
}

// AddJob submits an email to the queue.
// Returns false if the queue is full (or if you prefer to block, remove the 'select/default').
func (p *Pool) AddJob(email string) bool {
	select {
	case p.jobQueue <- email:
		return true
	default:
		// Queue is full.
		// Strategy: Drop the job or Block?
		// For an email checker, we usually want to block or increase buffer.
		// Here we block to ensure no emails are skipped.
		p.jobQueue <- email
		return true
	}
}

// Close signals that no more jobs will be added.
func (p *Pool) Close() {
	close(p.jobQueue)
}
