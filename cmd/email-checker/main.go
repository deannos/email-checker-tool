package main

import (
	"context"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deannos/email-checker-tool/internal/output"
	"github.com/deannos/email-checker-tool/internal/worker"
)

var (
	workersFlag = flag.Int("workers", 10, "number of concurrent workers")
	rpsFlag     = flag.Int("rps", 20, "max DNS requests per second (rate limiting)")
	timeoutFlag = flag.Duration("timeout", 5*time.Second, "lookup timeout")
	versionFlag = flag.Bool("version", false, "show version")
	outputFlag  = flag.String("output", "output.csv", "output CSV file path")
)

var version = "v0.1.0"

func main() {
	flag.Parse()

	if *versionFlag {
		log.Println("email-checker", version)
		return
	}

	if flag.NArg() == 0 {
		log.Fatal("usage: email-checker <domains_file.csv>")
	}

	// 1. Setup Context with Cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	defer cancel()

	// Handle Graceful Shutdown (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nShutting down...")
		cancel()
	}()

	// 2. Setup Output
	csvWriter, err := output.NewCSVWriter(*outputFlag)
	if err != nil {
		log.Fatal("Failed to create output file:", err)
	}
	defer csvWriter.Close()

	// 3. Initialize Pool
	pool := worker.NewPool(*workersFlag, 1000, csvWriter, *rpsFlag)

	// 4. Read Input and Feed Jobs (Non-blocking)
	go func() {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Printf("Error opening input file: %v", err)
			pool.Close()
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("Error reading CSV line: %v", err)
				continue
			}
			// Assuming domain is the first column
			if len(record) > 0 && record[0] != "" {
				pool.AddJob(record[0])
			}
		}
		pool.Close() // Signal that all jobs have been submitted
	}()

	// 5. Start Processing (Blocks until finished)
	log.Println("Starting processing...")
	pool.Start(ctx)
	log.Println("Done.")
}
