package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/deannos/email-checker-tool/internal/checker"
)

// MockWriter implements worker.Writer for testing.
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

func (m *MockWriter) Flush() error {
	return nil
}

func TestPool_ProcessJobs(t *testing.T) {
	mockWriter := &MockWriter{}
	// Rate limit high to finish fast
	pool := NewPool(2, 10, mockWriter, 100)

	domains := []string{"example.com", "test.com", "google.com"}

	go func() {
		for _, d := range domains {
			pool.AddJob(d)
		}
		pool.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)

	if len(mockWriter.Results) != len(domains) {
		t.Errorf("Expected %d results, got %d", len(domains), len(mockWriter.Results))
	}
}

func TestPool_RateLimiting(t *testing.T) {
	mockWriter := &MockWriter{}

	// Setup: 1 Worker, 2 Requests Per Second limit.
	pool := NewPool(1, 10, mockWriter, 2)

	startTime := time.Now()

	go func() {
		// Add 10 jobs.
		// At 2 RPS, and burst of 1, we expect:
		// Job 1 (Burst): ~0ms
		// Job 2: Wait 500ms
		// Job 3: Wait 1000ms
		// ...
		// Job 10: Wait 4500ms
		// Total ~ 4.5 seconds minimum.
		for i := 0; i < 10; i++ {
			pool.AddJob("example.com")
		}
		pool.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)

	duration := time.Since(startTime)

	// We expect at least 4 seconds given the burst=1 constraint.
	if duration < 4*time.Second {
		t.Errorf("Rate limiting failed? Expected > 4s, took %v", duration)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	mockWriter := &MockWriter{}
	pool := NewPool(2, 10, mockWriter, 100)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for i := 0; i < 100; i++ {
			pool.AddJob("example.com")
		}
		pool.Close()
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	pool.Start(ctx)

	if len(mockWriter.Results) > 90 {
		t.Errorf("Context cancellation failed? Processed %d jobs, expected fewer", len(mockWriter.Results))
	}
}
