package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// csvPath is the file to read. Put the CSV in the same folder as this program.
const csvPath = "NGDS_CO2_Algeria.csv"

// Column names we need, written with single spaces (see normalise below).
const (
	colLat  = "Latitude"
	colLon  = "Longitude"
	colYear = "Year"
	colType = "Field Type"
	colVol  = "Flaring Vol (million m3)"
)

// FieldData is one cleaned row of the CSV.
type FieldData struct {
	Latitude   float64
	Longitude  float64
	Year       int
	FieldType  string
	FlaringVol float64
}

// normalise trims a header name and collapses repeated spaces,
// so "Field  Type" (two spaces) becomes "Field Type".
func normalise(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// columnIndex maps each header name to its position and checks that
// every column we need exists. Returns an error if one is missing.
func columnIndex(header []string, required ...string) (map[string]int, error) {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[normalise(h)] = i
	}
	for _, name := range required {
		if _, ok := idx[name]; !ok {
			return nil, fmt.Errorf("column %q not found in header", name)
		}
	}
	return idx, nil
}

// parseRow converts one CSV record into a FieldData and validates it.
// Any problem is returned as an error so the caller can skip the row.
func parseRow(rec []string, idx map[string]int) (FieldData, error) {
	var d FieldData
	var err error

	if d.Latitude, err = strconv.ParseFloat(strings.TrimSpace(rec[idx[colLat]]), 64); err != nil {
		return d, fmt.Errorf("bad latitude: %w", err)
	}
	if d.Latitude < -90 || d.Latitude > 90 {
		return d, fmt.Errorf("latitude %v out of range", d.Latitude)
	}
	if d.Longitude, err = strconv.ParseFloat(strings.TrimSpace(rec[idx[colLon]]), 64); err != nil {
		return d, fmt.Errorf("bad longitude: %w", err)
	}
	if d.Longitude < -180 || d.Longitude > 180 {
		return d, fmt.Errorf("longitude %v out of range", d.Longitude)
	}
	if d.Year, err = strconv.Atoi(strings.TrimSpace(rec[idx[colYear]])); err != nil {
		return d, fmt.Errorf("bad year: %w", err)
	}
	if d.FlaringVol, err = strconv.ParseFloat(strings.TrimSpace(rec[idx[colVol]]), 64); err != nil {
		return d, fmt.Errorf("bad flaring volume: %w", err)
	}
	if d.FlaringVol < 0 {
		return d, fmt.Errorf("negative flaring volume %v", d.FlaringVol)
	}

	d.FieldType = strings.TrimSpace(rec[idx[colType]])
	if d.FieldType == "" {
		d.FieldType = "Unknown"
	}
	return d, nil
}

// loadData reads the whole CSV. It returns the good rows and how many
// rows were skipped. Each skipped row is reported with its line number.
func loadData(path string) ([]FieldData, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("reading header: %w", err)
	}
	idx, err := columnIndex(header, colLat, colLon, colYear, colType, colVol)
	if err != nil {
		return nil, 0, err
	}

	rows := make([]FieldData, 0, 3000)
	skipped := 0
	for {
		rec, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break // normal end of file
		}
		line, _ := reader.FieldPos(0)
		if err != nil {
			var parseErr *csv.ParseError
			if errors.As(err, &parseErr) {
				fmt.Printf("line %d skipped: %v\n", line, err)
				skipped++
				continue
			}
			return nil, skipped, fmt.Errorf("reading line %d: %w", line, err)
		}
		d, err := parseRow(rec, idx)
		if err != nil {
			fmt.Printf("line %d skipped: %v\n", line, err)
			skipped++
			continue
		}
		rows = append(rows, d)
	}
	return rows, skipped, nil
}
