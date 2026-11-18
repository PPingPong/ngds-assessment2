package main

import (
	"fmt"
	"os"
	"sort"
)

// trainUntilYear: rows up to this year are training data, later years are test data.
const trainUntilYear = 2020

func main() {
	rows, skipped, err := loadData(csvPath)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d rows, skipped %d\n", len(rows), skipped)

	if err := writeProcessed(rows, processedPath); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("Saved cleaned data to %s\n\n", processedPath)

	// 1. Fit on all rows (in-sample).
	model, err := fitOLS(rows)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	names := []string{"Intercept", "Year", "Latitude", "Longitude", "GAS", "LNG", "OIL"}
	fmt.Println("Model fitted on all rows (target: Flaring Vol, million m3)")
	for i, c := range model.Coef {
		fmt.Printf("  %-10s %12.4f\n", names[i], c)
	}
	r2, rmse, err := evaluate(model, rows)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("In-sample:  R2 = %.4f  RMSE = %.4f  (n = %d)\n\n", r2, rmse, len(rows))

	// 2. Train on early years, test on later years.
	train, test := splitByYear(rows, trainUntilYear)
	tm, err := fitOLS(train)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	trR2, trRMSE, err := evaluate(tm, train)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	teR2, teRMSE, err := evaluate(tm, test)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("Train (<= %d, n = %d): R2 = %.4f  RMSE = %.4f\n", trainUntilYear, len(train), trR2, trRMSE)
	fmt.Printf("Test  (>  %d, n = %d): R2 = %.4f  RMSE = %.4f\n", trainUntilYear, len(test), teR2, teRMSE)

	// 3. Plots (test set predictions from the train-only model).
	if err := makePlots(tm, test, rows); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("Saved plots: plot_actual_vs_pred.png, plot_residuals.png, plot_by_field_type.png")

	// 4. Comparison across years, field types and latitude bands.
	if err := runComparison(tm, test, rows); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("Saved comparison tables (comparison_by_*.csv) and plots (plot_mean_by_year.png, plot_by_latitude_band.png)")
}

// makePlots draws the three charts used in the slide deck.
func makePlots(m *Model, test, all []FieldData) error {
	actual := make([]float64, len(test))
	pred := make([]float64, len(test))
	resid := make([]float64, len(test))
	for i, r := range test {
		p, err := m.Predict(r)
		if err != nil {
			return err
		}
		actual[i], pred[i], resid[i] = r.FlaringVol, p, r.FlaringVol-p
	}
	if err := scatterPlot(actual, pred, "plot_actual_vs_pred.png"); err != nil {
		return err
	}
	if err := histPlot(resid, "plot_residuals.png"); err != nil {
		return err
	}

	// Mean flaring volume by field type (all rows).
	sum := map[string]float64{}
	cnt := map[string]int{}
	for _, r := range all {
		sum[r.FieldType] += r.FlaringVol
		cnt[r.FieldType]++
	}
	types := make([]string, 0, len(sum))
	for t := range sum {
		types = append(types, t)
	}
	sort.Strings(types)
	means := make([]float64, len(types))
	for i, t := range types {
		means[i] = sum[t] / float64(cnt[t])
	}
	return barPlot(types, means, "Mean flaring volume by field type", "Field type", "Mean flaring volume (million m3)", "plot_by_field_type.png")
}
