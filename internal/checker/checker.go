package checker

import "net"

// Injectable DNS functions (for testing)
var lookupMX = net.LookupMX
var lookupTXT = net.LookupTXT

type Result struct {
	Domain      string
	HasMX       bool
	HasSPF      bool
	SPFRecord   string
	HasDMARC    bool
	DMARCRecord string
}

func CheckDomain(domain string) Result {
	hasSPF, spf := getSPF(domain)
	hasDMARC, dmarc := getDMARC(domain)

	return Result{
		Domain:      domain,
		HasMX:       hasMX(domain),
		HasSPF:      hasSPF,
		SPFRecord:   spf,
		HasDMARC:    hasDMARC,
		DMARCRecord: dmarc,
	}
}

func hasMX(domain string) bool {
	mx, err := lookupMX(domain)
	return err == nil && len(mx) > 0
}

func getSPF(domain string) (bool, string) {
	txt, err := lookupTXT(domain)
	if err != nil {
		return false, ""
	}
	for _, r := range txt {
		if len(r) >= 6 && r[:6] == "v=spf1" {
			return true, r
		}
	}
	return false, ""
}

func getDMARC(domain string) (bool, string) {
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
