package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/deannos/email-checker-tool/internal/checker"
	"github.com/deannos/email-checker-tool/internal/output"
	"golang.org/x/time/rate"
)

// Pool manages the concurrent checking of domains.
type Pool struct {
	workerCount int
	jobQueue    chan string
	output      output.CSVWriter // Using concrete type for simplicity, or interface if you prefer
	limiter     *rate.Limiter
	wg          sync.WaitGroup
}

// NewPool initializes the worker pool.
// rps: Requests per second limit (to prevent DNS server throttling).
func NewPool(workerCount int, queueSize int, out output.CSVWriter, rps int) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobQueue:    make(chan string, queueSize),
		output:      out,
		// Limit to RPS requests per second with a burst of 5
		limiter: rate.NewLimiter(rate.Limit(rps), 5),
	}
}

// Start begins the processing loop. It blocks until all jobs are done or context is cancelled.
func (p *Pool) Start(ctx context.Context) {
	// Start Workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}

	// Wait for all workers to finish
	p.wg.Wait()

	// Ensure output is flushed
	p.output.Flush()
}

// worker processes jobs from the queue.
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled
			return
		case domain, ok := <-p.jobQueue:
			if !ok {
				// Queue closed
				return
			}

			// Rate Limit: Wait until we are allowed to proceed
			if err := p.limiter.Wait(ctx); err != nil {
				return
			}

			// Perform Check
			result := checker.CheckDomain(ctx, domain)

			// Write Result immediately
			if err := p.output.Write(result); err != nil {
				fmt.Printf("Error writing to CSV: %v\n", err)
			}
		}
	}
}

// AddJob adds a domain to the queue.
func (p *Pool) AddJob(domain string) {
	p.jobQueue <- domain
}

// Close signals that no more jobs will be added.
func (p *Pool) Close() {
	close(p.jobQueue)
}
