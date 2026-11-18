package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// processedPath is the cleaned dataset the program writes for submission.
const processedPath = "processed_flaring.csv"

// writeProcessed saves the cleaned rows to a CSV: numeric columns,
// blank field types filled with "Unknown", and the one-hot columns used
// by the model (Unknown is the baseline, so it has no column of its own).
func writeProcessed(rows []FieldData, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	header := []string{"Latitude", "Longitude", "Year", "FieldType", "GAS", "LNG", "OIL", "FlaringVol_million_m3"}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	for i, r := range rows {
		f, err := features(r) // [1, Year, Lat, Lon, GAS, LNG, OIL]
		if err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		rec := []string{
			strconv.FormatFloat(r.Latitude, 'f', -1, 64),
			strconv.FormatFloat(r.Longitude, 'f', -1, 64),
			strconv.Itoa(r.Year),
			r.FieldType,
			strconv.Itoa(int(f[4])),
			strconv.Itoa(int(f[5])),
			strconv.Itoa(int(f[6])),
			strconv.FormatFloat(r.FlaringVol, 'f', -1, 64),
		}
		if err := w.Write(rec); err != nil {
			return fmt.Errorf("writing row %d: %w", i, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flushing %s: %w", path, err)
	}
	return nil
}
