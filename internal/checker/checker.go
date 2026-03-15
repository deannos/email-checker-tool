package checker

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"
)

// ErrorType categorizes DNS lookup failures for better diagnostics.
type ErrorType string

const (
	ErrTypeNone     ErrorType = ""
	ErrTypeTimeout  ErrorType = "timeout"
	ErrTypeNXDomain ErrorType = "nxdomain"
	ErrTypeNetwork  ErrorType = "network"
	ErrTypeUnknown  ErrorType = "unknown"
)

// Result holds the DNS configuration details for a domain.
type Result struct {
	Domain      string
	HasMX       bool
	HasSPF      bool
	SPFRecord   string
	HasDMARC    bool
	DMARCRecord string
	Error       string
	ErrType     ErrorType
	Duration    time.Duration
}

// Config holds configuration for the Checker.
type Config struct {
	// DNSAddr is the custom DNS resolver address (e.g., "8.8.8.8:53").
	// Empty means use the system default resolver.
	DNSAddr string
	// Retries is the number of retry attempts on transient errors (0 = no retries).
	Retries int
	// CacheTTL controls result caching. Zero disables caching.
	CacheTTL time.Duration
}

// Checker performs DNS lookups with a configurable resolver, retry, and caching.
type Checker struct {
	resolver *net.Resolver
	retries  int
	cache    *resultCache
}

// New creates a new Checker with the given configuration.
func New(cfg Config) *Checker {
	var resolver *net.Resolver
	if cfg.DNSAddr != "" {
		addr := cfg.DNSAddr
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", addr)
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	var cache *resultCache
	if cfg.CacheTTL > 0 {
		cache = newCache(cfg.CacheTTL)
	}

	return &Checker{
		resolver: resolver,
		retries:  cfg.Retries,
		cache:    cache,
	}
}

// CheckDomain performs DNS lookups for MX, SPF, and DMARC records.
// It respects the context for cancellation and timeouts.
func (c *Checker) CheckDomain(ctx context.Context, domain string) Result {
	if c.cache != nil {
		if r, ok := c.cache.get(domain); ok {
			return r
		}
	}

	start := time.Now()
	result := c.checkDomain(ctx, domain)
	result.Duration = time.Since(start)

	if c.cache != nil && result.Error == "" {
		c.cache.set(domain, result)
	}

	return result
}

func (c *Checker) checkDomain(ctx context.Context, domain string) Result {
	result := Result{Domain: domain}

	// 1. MX records
	var mx []*net.MX
	if err := c.withRetry(ctx, func() error {
		var e error
		mx, e = c.resolver.LookupMX(ctx, domain)
		return e
	}); err != nil {
		result.Error = err.Error()
		result.ErrType = categorizeError(err)
		return result
	}
	result.HasMX = len(mx) > 0

	// 2. SPF (TXT records on the domain itself)
	var spfTxt []string
	if err := c.withRetry(ctx, func() error {
		var e error
		spfTxt, e = c.resolver.LookupTXT(ctx, domain)
		return e
	}); err == nil {
		for _, r := range spfTxt {
			if strings.HasPrefix(r, "v=spf1") {
				result.HasSPF = true
				result.SPFRecord = r
				break
			}
		}
	}

	// 3. DMARC (TXT records at _dmarc.<domain>)
	var dmarcTxt []string
	if err := c.withRetry(ctx, func() error {
		var e error
		dmarcTxt, e = c.resolver.LookupTXT(ctx, "_dmarc."+domain)
		return e
	}); err == nil {
		for _, r := range dmarcTxt {
			if strings.HasPrefix(r, "v=DMARC1") {
				result.HasDMARC = true
				result.DMARCRecord = r
				break
			}
		}
	}

	return result
}

// withRetry executes fn, retrying on transient errors with exponential backoff.
func (c *Checker) withRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryable(err) {
			return err
		}
	}
	return lastErr
}

// isRetryable returns true for transient DNS errors worth retrying.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary() || dnsErr.Timeout()
	}
	return false
}

// categorizeError maps an error to an ErrorType.
func categorizeError(err error) ErrorType {
	if err == nil {
		return ErrTypeNone
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrTypeTimeout
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsNotFound {
			return ErrTypeNXDomain
		}
		if dnsErr.Timeout() {
			return ErrTypeTimeout
		}
		return ErrTypeNetwork
	}
	return ErrTypeUnknown
}
