package worker

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/deannos/email-checker-tool/internal/checker"
	"github.com/deannos/email-checker-tool/internal/output"
)

// MockWriter is a dummy writer that stores results in memory for testing.
type MockWriter struct {
	mu      sync.Mutex
	Results []checker.Result
}

func (m *MockWriter) Write(r checker.Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Results = append(m.Results, r)
	return nil
}

// MockWriter implements Flush() if needed, but we can just skip it for basic tests
// If the pool requires Flush, we add an empty method.
func (m *MockWriter) Flush() error {
	return nil
}

func TestPool_ProcessJobs(t *testing.T) {
	// Setup a temporary CSV file (though we use MockWriter mostly, we might need the struct)
	mockWriter := &MockWriter{}

	// Create Pool: 2 Workers, Rate limit 100/sec (fast for test)
	pool := NewPool(2, 10, mockWriter, 100)

	// Define jobs
	domains := []string{"example.com", "test.com", "google.com"}

	// Feed jobs in background
	go func() {
		for _, d := range domains {
			pool.AddJob(d)
		}
		pool.Close()
	}()

	// Start pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)

	// Verify results
	if len(mockWriter.Results) != len(domains) {
		t.Errorf("Expected %d results, got %d", len(domains), len(mockWriter.Results))
	}
}

func TestPool_RateLimiting(t *testing.T) {
	mockWriter := &MockWriter{}

	// Create Pool: 1 Worker, Rate limit 1 request per 200ms (5 per sec)
	// This ensures the test is slow enough to measure
	pool := NewPool(1, 10, mockWriter, 5)

	startTime := time.Now()

	go func() {
		// Add 2 jobs. With 5 RPS, these should take at least 200ms (2 * 1/5s)
		pool.AddJob("example.com")
		pool.AddJob("test.com")
		pool.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)

	duration := time.Since(startTime)

	// We expect at least ~200ms (2 jobs * 200ms rate limiter interval)
	// If it finishes in 1ms, rate limiting is broken.
	if duration < 300*time.Millisecond {
		t.Errorf("Rate limiting failed? Expected > 300ms, took %v", duration)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	mockWriter := &MockWriter{}
	pool := NewPool(2, 10, mockWriter, 100)

	// Create a context that cancels after 50ms
	ctx, cancel := context.WithCancel(context.Background())

	// Feed many jobs that take time (simulated by checking real domains)
	go func() {
		for i := 0; i < 100; i++ {
			pool.AddJob("example.com")
		}
		pool.Close()
	}()

	// Cancel context shortly after starting
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	pool.Start(ctx)

	// The pool should have stopped early.
	// We can't assert an exact number of results because of race conditions,
	// but we verify it didn't process all 100.
	if len(mockWriter.Results) > 90 {
		t.Errorf("Context cancellation failed? Processed %d jobs, expected fewer", len(mockWriter.Results))
	}
}
