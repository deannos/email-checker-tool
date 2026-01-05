package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord")

	for scanner.Scan() {
		CheckDomain(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error: could not read from input: %v\n", err)
	}
}

func CheckDomain(domain string) {
	var hasSPF, hasDMARC bool
	var spfRecord, dmarcRecord string

	hasMX := hasMXRecord(domain)
	fmt.Printf("%s,%t\n", domain, hasMX)

	txtRecords, _ := net.LookupTXT(domain)
	for _, record := range txtRecords {
		if len(record) >= 6 && record[:6] == "v=spf1" {
			hasSPF = true
			spfRecord = record
		}
	}

	dmarcRecords, _ := net.LookupTXT("_dmarc." + domain)
	for _, record := range dmarcRecords {
		if len(record) >= 8 && record[:8] == "v=DMARC1" {
			hasDMARC = true
			dmarcRecord = record
		}
	}

	fmt.Printf(
		"%s,%t,%t,%q,%t,%q\n",
		domain,
		hasMX,
		hasSPF,
		spfRecord,
		hasDMARC,
		dmarcRecord,
	)
}

func hasMXRecord(domain string) bool {
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}
