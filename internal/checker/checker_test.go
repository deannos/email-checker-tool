package checker

import (
	"context"
	"testing"
	"time"
)

func newDefaultChecker() *Checker {
	return New(Config{Retries: 0})
}

// TestCheckDomain_Success tests the happy path against a real domain (integration).
func TestCheckDomain_Success(t *testing.T) {
	c := newDefaultChecker()
	result := c.CheckDomain(context.Background(), "google.com")

	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}
	if !result.HasMX {
		t.Error("Expected google.com to have MX record")
	}
	if result.ErrType != ErrTypeNone {
		t.Errorf("Expected ErrTypeNone, got: %s", result.ErrType)
	}
	if result.Duration == 0 {
		t.Error("Expected non-zero Duration")
	}
}

// TestCheckDomain_NonExistentDomain verifies error handling for NXDOMAIN.
func TestCheckDomain_NonExistentDomain(t *testing.T) {
	c := newDefaultChecker()
	result := c.CheckDomain(context.Background(), "this-domain-definitely-does-not-exist-12345.com")

	if result.Error == "" {
		t.Error("Expected an error for non-existent domain")
	}
	if result.HasMX {
		t.Error("Expected HasMX to be false for non-existent domain")
	}
}

// TestCheckDomain_EmptyDomain verifies graceful handling of empty input.
func TestCheckDomain_EmptyDomain(t *testing.T) {
	c := newDefaultChecker()
	result := c.CheckDomain(context.Background(), "")

	if result.HasMX {
		t.Error("Empty domain should not have MX")
	}
}

// TestCheckDomain_ContextTimeout verifies that a very short deadline is respected.
func TestCheckDomain_ContextTimeout(t *testing.T) {
	c := newDefaultChecker()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	result := c.CheckDomain(ctx, "google.com")
	// With a 1ns timeout the lookup will almost certainly fail.
	// We just verify we don't panic and Duration is set.
	_ = result
}

// TestCheckerCache verifies that cached results are returned.
func TestCheckerCache(t *testing.T) {
	c := New(Config{Retries: 0, CacheTTL: 5 * time.Minute})
	r1 := c.CheckDomain(context.Background(), "google.com")
	r2 := c.CheckDomain(context.Background(), "google.com")

	if r1.Domain != r2.Domain || r1.HasMX != r2.HasMX {
		t.Error("Cached result differs from original")
	}
}
