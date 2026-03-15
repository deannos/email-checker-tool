package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/deannos/email-checker-tool/internal/checker"
	"github.com/deannos/email-checker-tool/internal/output"
	"github.com/deannos/email-checker-tool/internal/version"
	"github.com/deannos/email-checker-tool/internal/worker"
)

var (
	workersFlag    = flag.Int("workers", 10, "number of concurrent workers")
	rpsFlag        = flag.Int("rps", 20, "max DNS requests per second")
	timeoutFlag    = flag.Duration("timeout", 30*time.Second, "global operation timeout (e.g. 30s, 5m)")
	outputFlag     = flag.String("output", "output.csv", "output file path")
	formatFlag     = flag.String("format", "csv", "output format: csv or json")
	dnsFlag        = flag.String("dns", "", "custom DNS resolver address (e.g. 8.8.8.8:53)")
	retriesFlag    = flag.Int("retries", 2, "retry attempts on transient DNS errors")
	cacheTTLFlag   = flag.Duration("cache-ttl", 5*time.Minute, "DNS result cache TTL (0 to disable)")
	colFlag        = flag.Int("col", 0, "zero-based column index of the domain in the input CSV")
	skipHeaderFlag = flag.Bool("skip-header", false, "skip the first row of the input CSV")
	noProgressFlag = flag.Bool("no-progress", false, "disable periodic progress stats")
	versionFlag    = flag.Bool("version", false, "show version and exit")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: email-checker [options] <domains.csv>\n\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  email-checker domains.csv\n")
		fmt.Fprintf(os.Stderr, "  email-checker --workers 20 --rps 50 --format json --output results.json domains.csv\n")
		fmt.Fprintf(os.Stderr, "  email-checker --dns 8.8.8.8:53 --retries 3 --cache-ttl 10m domains.csv\n")
		fmt.Fprintf(os.Stderr, "  email-checker --col 1 --skip-header domains.csv\n")
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("email-checker %s\n", version.Version)
		return
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	*formatFlag = strings.ToLower(*formatFlag)
	if *formatFlag != "csv" && *formatFlag != "json" {
		log.Fatalf("Invalid --format %q: must be \"csv\" or \"json\"", *formatFlag)
	}

	// Context with global timeout + signal cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Build the output writer
	out, closer, err := buildWriter(*formatFlag, *outputFlag)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer closer()

	// Build the checker
	c := checker.New(checker.Config{
		DNSAddr:  *dnsFlag,
		Retries:  *retriesFlag,
		CacheTTL: *cacheTTLFlag,
	})

	// Build the worker pool
	pool := worker.NewPool(*workersFlag, 1000, out, *rpsFlag, c.CheckDomain)

	// Feed jobs in a goroutine so Start() can proceed concurrently
	go func() {
		if err := feedJobs(pool, flag.Arg(0), *colFlag, *skipHeaderFlag); err != nil {
			log.Printf("Error reading input: %v", err)
		}
		pool.Close()
	}()

	// Periodic progress reporting
	if !*noProgressFlag {
		startTime := time.Now()
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stats := pool.Stats()
					elapsed := time.Since(startTime).Seconds()
					rate := float64(stats.Processed) / elapsed
					log.Printf("Processed: %d | Errors: %d | Rate: %.1f req/s",
						stats.Processed, stats.Errors, rate)
				}
			}
		}()
	}

	log.Printf("Starting — workers: %d, rps: %d, format: %s, output: %s",
		*workersFlag, *rpsFlag, *formatFlag, *outputFlag)
	pool.Start(ctx)

	stats := pool.Stats()
	log.Printf("Done — processed: %d, errors: %d", stats.Processed, stats.Errors)
}

// feedJobs reads the input CSV and submits domains to the pool.
func feedJobs(pool *worker.Pool, filePath string, col int, skipHeader bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // allow variable columns

	first := true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Skipping malformed row: %v", err)
			continue
		}

		if first && skipHeader {
			first = false
			continue
		}
		first = false

		if col >= len(record) {
			log.Printf("Row has %d columns, --col %d is out of range; skipping", len(record), col)
			continue
		}

		domain := strings.TrimSpace(record[col])
		if domain != "" {
			pool.AddJob(domain)
		}
	}
	return nil
}

// buildWriter returns the appropriate Writer and a close function.
func buildWriter(format, path string) (worker.Writer, func(), error) {
	switch format {
	case "json":
		w, err := output.NewJSONWriter(path)
		if err != nil {
			return nil, nil, err
		}
		return w, func() { w.Close() }, nil
	default:
		w, err := output.NewCSVWriter(path)
		if err != nil {
			return nil, nil, err
		}
		return w, func() { w.Close() }, nil
	}
}
