package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/deannos/email-checker-tool/internal/checker"
	"golang.org/x/time/rate"
)

// Writer defines the contract for outputting results.
// Any writer (CSV, JSON, etc.) must implement these methods.
type Writer interface {
	Write(result checker.Result) error
	Flush() error
}

// CheckFn is the function signature for domain checking.
type CheckFn func(ctx context.Context, domain string) checker.Result

// Stats holds atomic processing counters.
type Stats struct {
	Processed int64
	Errors    int64
}

// Pool manages the concurrent checking of domains.
type Pool struct {
	workerCount int
	jobQueue    chan string
	output      Writer
	limiter     *rate.Limiter
	wg          sync.WaitGroup
	checkFn     CheckFn
	processed   int64
	errCount    int64
}

// NewPool initializes the worker pool with a custom check function.
func NewPool(workerCount int, queueSize int, out Writer, rps int, checkFn CheckFn) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobQueue:    make(chan string, queueSize),
		output:      out,
		// Burst of 1 ensures strict per-second rate limiting.
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
		checkFn: checkFn,
	}
}

// Stats returns a snapshot of current processing counters.
func (p *Pool) Stats() Stats {
	return Stats{
		Processed: atomic.LoadInt64(&p.processed),
		Errors:    atomic.LoadInt64(&p.errCount),
	}
}

// Start begins the processing loop and blocks until all jobs are done.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}

	p.wg.Wait()

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

			if err := p.limiter.Wait(ctx); err != nil {
				return
			}

			result := p.checkFn(ctx, domain)
			atomic.AddInt64(&p.processed, 1)
			if result.Error != "" {
				atomic.AddInt64(&p.errCount, 1)
			}

			if err := p.output.Write(result); err != nil {
				fmt.Printf("Error writing result: %v\n", err)
			}
		}
	}
}

// AddJob submits a domain to the job queue.
func (p *Pool) AddJob(domain string) {
	p.jobQueue <- domain
}

// Close signals that no more jobs will be submitted.
func (p *Pool) Close() {
	close(p.jobQueue)
}
