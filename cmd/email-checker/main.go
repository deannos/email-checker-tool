package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/deannos/email-checker-tool/checker"
)

const workers = 10

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("domain,hasMX,hasSPF,spfRecord,hasDMARC,dmarcRecord")

	jobs := make(chan string)
	results := make(chan checker.Result)

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range jobs {
				results <- checker.CheckDomain(domain)
			}
		}()
	}

	// Feed jobs
	go func() {
		for scanner.Scan() {
			jobs <- scanner.Text()
		}
		close(jobs)
	}()

	// Close results when workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	if err := scanner.Err(); err != nil {
		log.Fatalf("input error: %v", err)
	}

	// Consume results
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
