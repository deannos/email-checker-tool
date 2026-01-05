package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/deannos/email-checker-tool/checker"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord")

	for scanner.Scan() {
		result := checker.CheckDomain(scanner.Text())
		fmt.Printf("%s,%t,%t,%q,%t,%q\n",
			result.Domain,
			result.HasMX,
			result.HasSPF,
			result.SPFRecord,
			result.HasDMARC,
			result.DMARCRecord,
		)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("input error: %v", err)
	}
}
