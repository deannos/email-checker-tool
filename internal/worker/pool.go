package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/deannos/email-checker-tool/internal/checker"
	"golang.org/x/time/rate"
)

// Writer defines the contract for outputting results.
// Any writer (CSV, JSON, etc) must implement these methods.
type Writer interface {
	Write(result checker.Result) error
	Flush() error // Critical for production to ensure data is saved
}

// Pool manages the concurrent checking of domains.
type Pool struct {
	workerCount int
	jobQueue    chan string
	output      Writer
	limiter     *rate.Limiter
	wg          sync.WaitGroup
}

// NewPool initializes the worker pool.
func NewPool(workerCount int, queueSize int, out Writer, rps int) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobQueue:    make(chan string, queueSize),
		output:      out,
		// Burst of 1 ensures strict rate limiting.
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
	}
}

// Start begins the processing loop.
func (p *Pool) Start(ctx context.Context) {
	// Start Workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}

	// Wait for all workers to finish
	p.wg.Wait()

	// CRITICAL: Flush output to ensure data integrity on disk
	if err := p.output.Flush(); err != nil {
		fmt.Printf("Warning: Failed to flush output: %v\n", err)
	}
}

// worker processes jobs from the queue.
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case domain, ok := <-p.jobQueue:
			if !ok {
				return
			}

			// Rate Limit: Respect the RPS limit
			if err := p.limiter.Wait(ctx); err != nil {
				return
			}

			// Perform the check
			result := checker.CheckDomain(ctx, domain)

			// Stream the result to output
			if err := p.output.Write(result); err != nil {
				fmt.Printf("Error writing result: %v\n", err)
			}
		}
	}
}

func (p *Pool) AddJob(domain string) {
	p.jobQueue <- domain
}

func (p *Pool) Close() {
	close(p.jobQueue)
}
