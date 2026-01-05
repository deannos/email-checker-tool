package output

import (
	"encoding/csv"
	"io"
)

func WriteCSV(w io.Writer, rows [][]string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()
	return writer.WriteAll(rows)
}
