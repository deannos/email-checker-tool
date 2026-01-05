package checker

import (
	"net"
	"testing"
)

func TestCheckDomain_WithSPFAndDMARC(t *testing.T) {
	// Save originals
	origMX := lookupMX
	origTXT := lookupTXT
	defer func() {
		lookupMX = origMX
		lookupTXT = origTXT
	}()

	lookupMX = func(domain string) ([]*net.MX, error) {
		return []*net.MX{{Host: "mail.example.com", Pref: 10}}, nil
	}

	lookupTXT = func(domain string) ([]string, error) {
		if domain == "_dmarc.example.com" {
			return []string{"v=DMARC1; p=none"}, nil
		}
		return []string{"v=spf1 include:_spf.example.com"}, nil
	}

	result := CheckDomain("example.com")

	if !result.HasMX {
		t.Error("expected MX record")
	}
	if !result.HasSPF {
		t.Error("expected SPF record")
	}
	if !result.HasDMARC {
		t.Error("expected DMARC record")
	}
}
