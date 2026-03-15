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

func (m *MockWriter) Flush() error { return nil }

// noopCheckFn returns an empty result without any DNS call.
func noopCheckFn(_ context.Context, domain string) checker.Result {
	return checker.Result{Domain: domain}
}

func TestPool_ProcessJobs(t *testing.T) {
	mockWriter := &MockWriter{}
	pool := NewPool(2, 10, mockWriter, 100, noopCheckFn)

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

func TestPool_Stats(t *testing.T) {
	mockWriter := &MockWriter{}
	errCheckFn := func(_ context.Context, domain string) checker.Result {
		if domain == "bad.com" {
			return checker.Result{Domain: domain, Error: "lookup failed"}
		}
		return checker.Result{Domain: domain}
	}
	pool := NewPool(2, 10, mockWriter, 100, errCheckFn)

	go func() {
		pool.AddJob("good.com")
		pool.AddJob("bad.com")
		pool.AddJob("good2.com")
		pool.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)

	stats := pool.Stats()
	if stats.Processed != 3 {
		t.Errorf("Expected 3 processed, got %d", stats.Processed)
	}
	if stats.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", stats.Errors)
	}
}

func TestPool_RateLimiting(t *testing.T) {
	mockWriter := &MockWriter{}
	pool := NewPool(1, 10, mockWriter, 2, noopCheckFn)

	startTime := time.Now()

	go func() {
		for i := 0; i < 10; i++ {
			pool.AddJob("example.com")
		}
		pool.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool.Start(ctx)

	duration := time.Since(startTime)
	// At 2 RPS with burst=1: job 1 immediate, jobs 2-10 spaced 500ms apart → ~4.5s minimum.
	if duration < 4*time.Second {
		t.Errorf("Rate limiting failed: expected > 4s, took %v", duration)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	mockWriter := &MockWriter{}
	pool := NewPool(2, 100, mockWriter, 1000, noopCheckFn)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for i := 0; i < 100; i++ {
			pool.AddJob("example.com")
		}
		pool.Close()
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	pool.Start(ctx)

	if len(mockWriter.Results) >= 100 {
		t.Errorf("Context cancellation failed: processed all 100 jobs despite early cancel")
	}
}
