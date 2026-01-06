package checker

import (
	"context"
	"testing"
)

// TestCheckDomain_Success checks the happy path where MX, SPF, and DMARC exist.
func TestCheckDomain_Success(t *testing.T) {
	// We cannot easily mock net.DefaultResolver.LookupMX globally without
	// changing the architecture significantly.
	// However, for this level of tool, testing against the real network (Integration Test)
	// or accepting that we test live "google.com" is common.
	// Alternatively, we can rely on the fact that CheckDomain calls exported functions.

	// For now, let's test with a known domain (Integration-style test)
	// This ensures our code actually works with real DNS servers.
	result := CheckDomain(context.Background(), "google.com")

	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}
	if !result.HasMX {
		t.Error("Expected google.com to have MX record")
	}
	// Note: SPF/DMARC are harder to assert on live domains as they change frequently
}

// TestCheckDomain_InvalidSyntax checks that we don't crash on bad input.
// Note: Current CheckDomain implementation takes a domain string, not an email.
func TestCheckDomain_NonExistentDomain(t *testing.T) {
	result := CheckDomain(context.Background(), "this-domain-definitely-does-not-exist-12345.com")

	// Should not return panic. Status should indicate error or missing records.
	// Since we capture error in result.Error struct now:
	if result.Error == "" {
		t.Error("Expected an error for non-existent domain, but got nil")
	}
	if result.HasMX {
		t.Error("Expected HasMX to be false for non-existent domain")
	}
}

// TestCheckDomain_EmptyDomain checks behavior with empty input
func TestCheckDomain_EmptyDomain(t *testing.T) {
	result := CheckDomain(context.Background(), "")

	// Should handle gracefully (usually DNS error)
	// We don't want a panic
	if result.HasMX {
		t.Error("Empty domain should not have MX")
	}
}
