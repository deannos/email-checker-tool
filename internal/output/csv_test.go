package output

import (
	"os"
	"strings"
	"testing"

	"github.com/deannos/email-checker-tool/internal/checker"
)

func TestCSVWriter(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	writer, err := NewCSVWriter(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()

	result := checker.Result{
		Domain:   "test.com",
		HasMX:    true,
		HasSPF:   false,
		HasDMARC: true,
	}

	if err := writer.Write(result); err != nil {
		t.Fatalf("Failed to write result: %v", err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	if !strings.Contains(s, "test.com") {
		t.Error("Output missing domain")
	}
	if !strings.Contains(s, "domain,hasMX") {
		t.Error("Output missing header")
	}
	if !strings.Contains(s, "errorType") {
		t.Error("Output missing errorType column")
	}
	if !strings.Contains(s, "durationMs") {
		t.Error("Output missing durationMs column")
	}
}
