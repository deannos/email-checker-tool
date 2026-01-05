package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deannos/email-checker-tool/internal/checker"
	"github.com/deannos/email-checker-tool/internal/worker"
)

var (
	workersFlag = flag.Int("workers", 10, "number of concurrent workers")
	timeoutFlag = flag.Duration("timeout", 5*time.Second, "DNS lookup timeout")
	versionFlag = flag.Bool("version", false, "show version")
)

var version = "v0.1.0"

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println("email-checker", version)
		return
	}

	if flag.NArg() == 0 {
		log.Fatal("usage: email-checker <domain_file>")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := csv.NewReader(file)
	var domains []string
	for {
		record, err := scanner.Read()
		if err != nil {
			break
		}
		domains = append(domains, record[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	defer cancel()

	results := worker.Run(*workersFlag, domains, func(d string) checker.Result {
		return checker.CheckDomain(ctx, d)
	})

	csvWriter := csv.NewWriter(os.Stdout)
	defer csvWriter.Flush()
	csvWriter.Write([]string{"domain", "hasMX", "hasSPF", "spfRecord", "hasDMARC", "dmarcRecord"})
	for _, r := range results {
		csvWriter.Write([]string{
			r.Domain,
			fmt.Sprintf("%t", r.HasMX),
			fmt.Sprintf("%t", r.HasSPF),
			r.SPFRecord,
			fmt.Sprintf("%t", r.HasDMARC),
			r.DMARCRecord,
		})
	}
}
