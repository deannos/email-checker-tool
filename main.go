package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/deannos/email-checker-tool/checker"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord")

	results := make(chan checker.Result)
	var wg sync.WaitGroup

	for scanner.Scan() {
		domain := scanner.Text()
		wg.Add(1)

		go func(d string) {
			defer wg.Done()
			results <- checker.CheckDomain(d)
		}(domain)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	if err := scanner.Err(); err != nil {
		log.Fatalf("input error: %v", err)
	}

	for r := range results {
		fmt.Printf("%s,%t,%t,%q,%t,%q\n",
			r.Domain,
			r.HasMX,
			r.HasSPF,
			r.SPFRecord,
			r.HasDMARC,
			r.DMARCRecord,
		)
	}
}
