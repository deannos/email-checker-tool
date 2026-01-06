package output

import (
	"os"
	"strings"
	"testing"

	"github.com/deannos/email-checker-tool/internal/checker"
)

func TestCSVWriter(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	writer, err := NewCSVWriter(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()

	// Write a dummy result
	result := checker.Result{
		Domain:    "test.com",
		HasMX:     true,
		HasSPF:    false,
		SPFRecord: "",
		HasDMARC:  true,
	}

	err = writer.Write(result)
	if err != nil {
		t.Fatalf("Failed to write result: %v", err)
	}

	writer.Flush()

	// Read the file back and check content
	content, _ := os.ReadFile(tmpfile.Name())
	contentStr := string(content)

	if !strings.Contains(contentStr, "test.com") {
		t.Error("Output file does not contain domain")
	}
	if !strings.Contains(contentStr, "domain,hasMX") { // Check header
		t.Error("Output file is missing header")
	}
}
