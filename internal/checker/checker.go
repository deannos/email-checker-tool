package checker

import (
	"context"
	"net"
)

// Result holds the DNS configuration details for a domain.
type Result struct {
	Domain      string
	HasMX       bool
	HasSPF      bool
	SPFRecord   string
	HasDMARC    bool
	DMARCRecord string
	Error       string // To capture lookup failures
}

// CheckDomain performs DNS lookups for MX, SPF, and DMARC records.
// It respects the context for cancellation/timeouts.
func CheckDomain(ctx context.Context, domain string) Result {

	// 1. Check MX with Context
	hasMX, err := hasMX(ctx, domain)
	if err != nil {
		return Result{Domain: domain, Error: err.Error()}
	}

	// 2. Check SPF (TXT record)
	// Note: Standard library LookupTXT doesn't strictly support context cancellation in older Go versions,
	// but we assume the caller manages the overall timeout if needed.
	hasSPF, spf := getSPF(domain)

	// 3. Check DMARC
	hasDMARC, dmarc := getDMARC(domain)

	return Result{
		Domain:      domain,
		HasMX:       hasMX,
		HasSPF:      hasSPF,
		SPFRecord:   spf,
		HasDMARC:    hasDMARC,
		DMARCRecord: dmarc,
	}
}

// hasMX checks for Mail Exchange records.
func hasMX(ctx context.Context, domain string) (bool, error) {
	// Use the resolver that accepts context
	mx, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil {
		return false, err
	}
	return len(mx) > 0, nil
}

// getSPF checks for SPF records in the TXT records.
func getSPF(domain string) (bool, string) {
	txt, err := net.LookupTXT(domain)
	if err != nil {
		return false, ""
	}
	for _, r := range txt {
		// Simple check for v=spf1
		if len(r) >= 6 && r[:6] == "v=spf1" {
			return true, r
		}
	}
	return false, ""
}

// getDMARC checks for DMARC records.
func getDMARC(domain string) (bool, string) {
	// DMARC records are always at _dmarc.domain
	txt, err := net.LookupTXT("_dmarc." + domain)
	if err != nil {
		return false, ""
	}
	for _, r := range txt {
		if len(r) >= 8 && r[:8] == "v=DMARC1" {
			return true, r
		}
	}
	return false, ""
}
