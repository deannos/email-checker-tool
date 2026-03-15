package output

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/deannos/email-checker-tool/internal/checker"
)

// JSONWriter writes results as JSONL (one JSON object per line).
// It implements the worker.Writer interface.
type JSONWriter struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

// jsonRecord is the JSON representation of a checker.Result.
type jsonRecord struct {
	Domain      string            `json:"domain"`
	HasMX       bool              `json:"hasMX"`
	HasSPF      bool              `json:"hasSPF"`
	SPFRecord   string            `json:"spfRecord,omitempty"`
	HasDMARC    bool              `json:"hasDMARC"`
	DMARCRecord string            `json:"dmarcRecord,omitempty"`
	Error       string            `json:"error,omitempty"`
	ErrType     checker.ErrorType `json:"errorType,omitempty"`
	DurationMs  int64             `json:"durationMs"`
}

// NewJSONWriter creates a new JSONL output file.
func NewJSONWriter(filePath string) (*JSONWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return &JSONWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Write encodes a result as a JSON line.
func (w *JSONWriter) Write(result checker.Result) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.encoder.Encode(jsonRecord{
		Domain:      result.Domain,
		HasMX:       result.HasMX,
		HasSPF:      result.HasSPF,
		SPFRecord:   result.SPFRecord,
		HasDMARC:    result.HasDMARC,
		DMARCRecord: result.DMARCRecord,
		Error:       result.Error,
		ErrType:     result.ErrType,
		DurationMs:  result.Duration.Milliseconds(),
	})
}

// Flush is a no-op; json.Encoder writes directly to the file.
func (w *JSONWriter) Flush() error {
	return nil
}

// Close closes the underlying file.
func (w *JSONWriter) Close() error {
	return w.file.Close()
}
