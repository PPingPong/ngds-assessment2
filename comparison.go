package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

// Latitude band edges (degrees north), roughly the 33rd and 67th percentiles.
const (
	southEdge = 28.7
	northEdge = 31.7
)

// GroupStat summarises flaring volume for one group (a year, a field type or a latitude band).
type GroupStat struct {
	Name  string
	N     int
	Total float64
	Mean  float64
}

func latBand(lat float64) string {
	switch {
	case lat < southEdge:
		return "South (<28.7)"
	case lat < northEdge:
		return "Central (28.7-31.7)"
	default:
		return "North (>31.7)"
	}
}

// summarise groups rows by key and returns count, total and mean flaring volume per group, sorted by name.
func summarise(rows []FieldData, key func(FieldData) string) []GroupStat {
	byName := map[string]*GroupStat{}
	for _, r := range rows {
		k := key(r)
		g, ok := byName[k]
		if !ok {
			g = &GroupStat{Name: k}
			byName[k] = g
		}
		g.N++
		g.Total += r.FlaringVol
	}
	out := make([]GroupStat, 0, len(byName))
	for _, g := range byName {
		g.Mean = g.Total / float64(g.N)
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// rmseBy returns the model RMSE for each group of rows.
func rmseBy(m *Model, rows []FieldData, key func(FieldData) string) (map[string]float64, error) {
	rss := map[string]float64{}
	cnt := map[string]int{}
	for _, r := range rows {
		p, err := m.Predict(r)
		if err != nil {
			return nil, err
		}
		k := key(r)
		rss[k] += (r.FlaringVol - p) * (r.FlaringVol - p)
		cnt[k]++
	}
	out := make(map[string]float64, len(rss))
	for k, v := range rss {
		out[k] = math.Sqrt(v / float64(cnt[k]))
	}
	return out, nil
}

// writeSummaryCSV saves a group summary table.
func writeSummaryCSV(path, label string, stats []GroupStat) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	if err := w.Write([]string{label, "Rows", "TotalFlaring_million_m3", "MeanFlaring_million_m3"}); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	for _, g := range stats {
		rec := []string{
			g.Name,
			strconv.Itoa(g.N),
			strconv.FormatFloat(g.Total, 'f', 1, 64),
			strconv.FormatFloat(g.Mean, 'f', 1, 64),
		}
		if err := w.Write(rec); err != nil {
			return fmt.Errorf("writing %s: %w", g.Name, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flushing %s: %w", path, err)
	}
	return nil
}

func printSummary(title string, stats []GroupStat) {
	fmt.Println(title)
	for _, g := range stats {
		fmt.Printf("  %-22s n=%4d  total=%8.1f  mean=%7.1f\n", g.Name, g.N, g.Total, g.Mean)
	}
	fmt.Println()
}

func meansOf(stats []GroupStat) (names []string, means []float64) {
	for _, g := range stats {
		names = append(names, g.Name)
		means = append(means, g.Mean)
	}
	return names, means
}

// runComparison compares flaring across years, field types and latitude bands,
// and reports how well the train-only model does on each test group.
func runComparison(m *Model, test, all []FieldData) error {
	byYear := summarise(all, func(r FieldData) string { return strconv.Itoa(r.Year) })
	byType := summarise(all, func(r FieldData) string { return r.FieldType })
	byBand := orderBands(summarise(all, func(r FieldData) string { return latBand(r.Latitude) }))

	printSummary("Flaring by year (all rows)", byYear)
	printSummary("Flaring by field type (all rows)", byType)
	printSummary("Flaring by latitude band (all rows)", byBand)

	if err := writeSummaryCSV("comparison_by_year.csv", "Year", byYear); err != nil {
		return err
	}
	if err := writeSummaryCSV("comparison_by_field_type.csv", "FieldType", byType); err != nil {
		return err
	}
	if err := writeSummaryCSV("comparison_by_latitude_band.csv", "LatitudeBand", byBand); err != nil {
		return err
	}

	// Model error on the test years and test groups.
	for _, c := range []struct {
		title string
		key   func(FieldData) string
	}{
		{"Test RMSE by year", func(r FieldData) string { return strconv.Itoa(r.Year) }},
		{"Test RMSE by field type", func(r FieldData) string { return r.FieldType }},
		{"Test RMSE by latitude band", func(r FieldData) string { return latBand(r.Latitude) }},
	} {
		res, err := rmseBy(m, test, c.key)
		if err != nil {
			return err
		}
		keys := make([]string, 0, len(res))
		for k := range res {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Println(c.title)
		for _, k := range keys {
			fmt.Printf("  %-22s RMSE=%7.2f\n", k, res[k])
		}
		fmt.Println()
	}

	// Charts.
	years := make([]float64, len(byYear))
	yearMeans := make([]float64, len(byYear))
	for i, g := range byYear {
		y, err := strconv.Atoi(g.Name)
		if err != nil {
			return fmt.Errorf("year %q: %w", g.Name, err)
		}
		years[i], yearMeans[i] = float64(y), g.Mean
	}
	if err := linePlot(years, yearMeans, "Mean flaring volume per site by year", "Year", "Mean flaring volume (million m3)", "plot_mean_by_year.png"); err != nil {
		return err
	}
	names, means := meansOf(byBand)
	return barPlot(names, means, "Mean flaring volume by latitude band", "Latitude band (degrees north)", "Mean flaring volume (million m3)", "plot_by_latitude_band.png")
}

// orderBands puts the latitude bands in south-to-north order instead of alphabetical order.
func orderBands(stats []GroupStat) []GroupStat {
	order := map[string]int{latBand(0): 0, latBand(30): 1, latBand(40): 2}
	sort.Slice(stats, func(i, j int) bool { return order[stats[i].Name] < order[stats[j].Name] })
	return stats
}
