// Package csvparser provides a generic CSV parser with transformer support.
// Callers supply a transform function to convert raw CSV data into any desired format.
package csvparser

import (
	"encoding/csv"
	"fmt"
	"os"
)

// Parse reads a CSV file and applies a transformer function to produce type T.
// The transformer receives headers and rows, and returns the desired output format.
func Parse[T any](file string, transform func(headers []string, rows [][]string) T) (T, error) {
	data, err := readCSV(file)
	if err != nil {
		var zero T
		return zero, err
	}

	var headers []string
	var rows [][]string
	if len(data) > 0 {
		headers = data[0]
		rows = data[1:]
	}

	return transform(headers, rows), nil
}

// readCSV opens and parses a CSV file into raw records.
func readCSV(file string) ([][]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}

	return records, nil
}
