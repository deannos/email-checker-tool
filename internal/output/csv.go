package output

import (
	"encoding/csv"
	"os"
	"sync"

	"github.com/deannos/email-checker-tool/internal/checker"
)

// CSVWriter implements the worker.Writer interface.
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
	mu     sync.Mutex
}

// NewCSVWriter creates a new CSV file and writes the header.
func NewCSVWriter(filePath string) (*CSVWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)

	// Write CSV Header
	header := []string{"domain", "hasMX", "hasSPF", "spfRecord", "hasDMARC", "dmarcRecord", "error"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, err
	}
	writer.Flush()

	return &CSVWriter{
		file:   file,
		writer: writer,
	}, nil
}

// Write appends a record to the CSV.
// NOTE: Uses Pointer Receiver (*CSVWriter) to satisfy the interface.
func (w *CSVWriter) Write(result checker.Result) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	record := []string{
		result.Domain,
		boolToStr(result.HasMX),
		boolToStr(result.HasSPF),
		result.SPFRecord,
		boolToStr(result.HasDMARC),
		result.DMARCRecord,
		result.Error,
	}
	return w.writer.Write(record)
}

// Flush forces any buffered data to be written to the file.
// NOTE: Uses Pointer Receiver (*CSVWriter) to satisfy the interface.
func (w *CSVWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writer.Flush()
	return w.writer.Error()
}

// Close closes the underlying file.
func (w *CSVWriter) Close() error {
	return w.file.Close()
}

// Helper function
func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
