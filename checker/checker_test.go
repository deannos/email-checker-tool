package checker

import "testing"

func TestCheckDomain_WithSPFAndDMARC(t *testing.T) {
	lookupMX = func(domain string) ([]*struct {
		Host string
		Pref uint16
	}, error) {
		return []*struct {
			Host string
			Pref uint16
		}{{}}, nil
	}

	lookupTXT = func(domain string) ([]string, error) {
		if domain == "_dmarc.example.com" {
			return []string{"v=DMARC1; p=none"}, nil
		}
		return []string{"v=spf1 include:_spf.example.com"}, nil
	}

	result := CheckDomain("example.com")

	if !result.HasMX || !result.HasSPF || !result.HasDMARC {
		t.Fatal("expected MX, SPF, and DMARC to be true")
	}
}
