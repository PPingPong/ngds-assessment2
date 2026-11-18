# NGDS Assessment 2: Predicting gas flaring volume in Algeria with Go

Module 6FTC2177 Next Generation Data Science. A Go program that cleans the NGDS_CO2_Algeria.csv dataset, fits a multiple linear regression (OLS) to predict flaring volume, evaluates it, and compares flaring across years, field types and latitude bands.

**Target (confirmed with the lecturer):** CO2, measured as `Flaring Vol (million m3)`.

## How to run

Requires Go (the version in `go.mod`) and an internet connection the first time, to download modules.

```
go run .
```

It prints the cleaning summary, the model coefficients, in-sample, train and test metrics and the comparison tables. It also writes the processed data, comparison tables and charts into the same folder.

## What the program does

| File | Purpose |
| --- | --- |
| `loader.go` | Reads the CSV with `encoding/csv`, finds columns by name from the header, treats a blank field type as `Unknown`, skips and reports bad rows |
| `processed.go` | Writes the cleaned data to `processed_flaring.csv` |
| `model.go` | Builds the features and fits OLS with `gonum` QR decomposition |
| `eval.go` | Predictions, R2 and RMSE |
| `split.go` | Splits by year (train up to 2020, test after 2020) |
| `comparison.go` | Flaring and model error by year, field type and latitude band |
| `plots.go` | Charts with `gonum.org/v1/plot` |
| `main.go` | Runs the steps in order |

**Model:** flaring volume on Year, Latitude, Longitude and field type (GAS, LNG, OIL as 0/1 columns, Unknown as the baseline). `bcm`, `MMscfd` and `Flare Level` are excluded because they describe the flaring itself and would leak the target.

## Results

2592 rows loaded, 0 skipped.

| Fit | R2 | RMSE |
| --- | --- | --- |
| All rows (in-sample) | 0.0860 | 56.6550 |
| Train, year <= 2020 (n = 1990) | 0.1079 | 57.8456 |
| Test, year > 2020 (n = 602) | -0.0668 | 53.8737 |

The model explains little of the variation. On the later years it does worse than predicting the training mean (baseline RMSE 52.16), so it does not generalise well to 2021 to 2023.

Coefficients (fitted on all rows): Intercept -1412.7550, Year 0.6554, Latitude 3.1835, Longitude 1.6817, GAS 25.9681, LNG 131.1560, OIL 22.7963. The coefficient of each field type is the difference in mean flaring compared with `Unknown`, holding the other variables fixed.

### Comparison across years and groups

- **Years:** mean flaring per site rose from 30.0 (2012) to a peak of 53.8 (2016) and fell to 38.6 (2023). The number of sites reporting changes from year to year (253 in 2012, 159 in 2016, 187 in 2023), so per-site means are fairer than totals.
- **Field type:** LNG has the highest mean (156.7, but only 37 rows), OIL 40.6, GAS 39.3, Unknown 12.5.
- **Latitude bands:** Central (28.7 to 31.7) 44.6, North (above 31.7) 40.9, South (below 28.7) 32.2.
- **Test RMSE by group:** worst for LNG (109.12); about 54 for GAS and OIL; by latitude band South 46.43, North 54.52, Central 59.12; by year 55.45 (2021), 55.26 (2022), 50.43 (2023).

These are descriptive comparisons; no significance tests were run.

## Output files

- `processed_flaring.csv`: cleaned dataset
- `comparison_by_year.csv`, `comparison_by_field_type.csv`, `comparison_by_latitude_band.csv`
- `plot_actual_vs_pred.png`, `plot_residuals.png`, `plot_by_field_type.png`, `plot_mean_by_year.png`, `plot_by_latitude_band.png`

## Checking the results

`python/check_reference.py` fits the same model in Python (numpy and pandas) on the original CSV. Its coefficients and metrics match the Go output.

`docs/go_run_output.png` is a screenshot of `go run .` on my own computer.

## Use of Generative AI

I used a local gemma3 model (Ollama) to learn Go and to draft code, giving it structured prompts and correcting its mistakes. I also used Claude (Anthropic) to review and correct code, to write the Python check, to build the comparison and chart styling, and to explain the method. The full log is in `docs/` (prompt log and the gemma3 chat transcript).

## Limitations

- Low explanatory power (R2 about 0.09 in sample, negative on the test years).
- Location and field type only; no information about the production of each site.
- The number of reporting sites changes by year.
- Standard errors and significance tests were not computed in Go.
