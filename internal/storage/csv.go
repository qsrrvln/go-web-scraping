package storage

import (
	"encoding/csv"
	"fmt"
	"os"

	"go-scraper-engine/internal/models"
)

// CSVWriter handles writing product data to a CSV file.
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

// NewCSVWriter creates a new CSVWriter and writes the header row.
func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write header row immediately
	header := []string{"ID", "Name", "Price", "URL"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to flush CSV header: %w", err)
	}

	return &CSVWriter{
		file:   file,
		writer: writer,
	}, nil
}

// Write converts the Product struct to a CSV row and writes it.
func (w *CSVWriter) Write(p models.Product) error {
	row := []string{
		p.ID,
		p.Name,
		p.Price,
		p.URL,
	}

	if err := w.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}

	// Flush after each write to ensure data is saved
	w.writer.Flush()

	if err := w.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV row: %w", err)
	}

	return nil
}

// Close flushes remaining data and closes the file.
func (w *CSVWriter) Close() error {
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		w.file.Close()
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}
	return w.file.Close()
}
