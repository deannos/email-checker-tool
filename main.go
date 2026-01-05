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
	var hasMX, hasSPF bool
	var spfRecord string

	mxRecords, _ := net.LookupMX(domain)
	if len(mxRecords) > 0 {
		hasMX = true
	}

	txtRecords, _ := net.LookupTXT(domain)
	for _, record := range txtRecords {
		if len(record) >= 6 && record[:6] == "v=spf1" {
			hasSPF = true
			spfRecord = record
			break
		}
	}

	fmt.Printf("%s,%t,%t,%q\n", domain, hasMX, hasSPF, spfRecord)
}
