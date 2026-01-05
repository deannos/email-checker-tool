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
	hasMX := hasMXRecord(domain)
	hasSPF, spfRecord := getSPFRecord(domain)

	fmt.Printf("%s,%t,%t,%q\n",
		domain,
		hasMX,
		hasSPF,
		spfRecord,
	)
}

func hasMXRecord(domain string) bool {
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}

func getSPFRecord(domain string) (bool, string) {
	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		return false, ""
	}

	for _, record := range txtRecords {
		if len(record) >= 6 && record[:6] == "v=spf1" {
			return true, record
		}
	}
	return false, ""
}
