package main

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var (
	blue   = color.RGBA{R: 31, G: 119, B: 180, A: 255}
	orange = color.RGBA{R: 255, G: 127, B: 14, A: 255}
	red    = color.RGBA{R: 214, G: 39, B: 40, A: 255}
	grey   = color.RGBA{R: 220, G: 220, B: 220, A: 255}
)

// newPlot makes a plot with a title, axis labels and light grid lines.
func newPlot(title, xLabel, yLabel string) *plot.Plot {
	p := plot.New()
	p.Title.Text = title
	p.Title.TextStyle.Font.Size = vg.Points(14)
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel
	g := plotter.NewGrid()
	g.Horizontal.Color = grey
	g.Vertical.Color = grey
	p.Add(g)
	return p
}

// scatterPlot draws actual vs predicted with a red y = x line (perfect prediction).
func scatterPlot(actual, pred []float64, path string) error {
	if len(actual) != len(pred) {
		return fmt.Errorf("actual (%d) and pred (%d) must have the same length", len(actual), len(pred))
	}
	p := newPlot("Test set: actual vs predicted flaring volume", "Predicted flaring volume (million m3)", "Actual flaring volume (million m3)")

	pts := make(plotter.XYs, len(actual))
	hi := 0.0
	for i := range actual {
		pts[i].X = pred[i]
		pts[i].Y = actual[i]
		hi = math.Max(hi, math.Max(pred[i], actual[i]))
	}
	s, err := plotter.NewScatter(pts)
	if err != nil {
		return fmt.Errorf("error creating scatter plot: %w", err)
	}
	s.GlyphStyle.Color = color.NRGBA{R: 31, G: 119, B: 180, A: 130}
	s.GlyphStyle.Radius = vg.Points(2.5)
	s.GlyphStyle.Shape = draw.CircleGlyph{}

	line, err := plotter.NewLine(plotter.XYs{{X: 0, Y: 0}, {X: hi, Y: hi}})
	if err != nil {
		return fmt.Errorf("error creating y = x line: %w", err)
	}
	line.Color = red
	line.Width = vg.Points(1.5)

	p.Add(s, line)
	p.Legend.Add("Test rows", s)
	p.Legend.Add("Perfect prediction (y = x)", line)
	p.Legend.Top = true
	p.Legend.Left = true
	return p.Save(7*vg.Inch, 5*vg.Inch, path)
}

// histPlot draws a histogram of the residuals (actual - predicted).
func histPlot(v []float64, path string) error {
	p := newPlot("Test set residuals (actual - predicted)", "Residual (million m3)", "Number of rows")
	h, err := plotter.NewHist(plotter.Values(v), 30)
	if err != nil {
		return fmt.Errorf("error creating histogram: %w", err)
	}
	h.FillColor = color.NRGBA{R: 31, G: 119, B: 180, A: 200}
	h.LineStyle.Color = color.RGBA{R: 20, G: 70, B: 110, A: 255}
	p.Add(h)
	return p.Save(7*vg.Inch, 5*vg.Inch, path)
}

// barPlot draws one coloured bar per name with the value written above it.
func barPlot(names []string, vals []float64, title, xLabel, yLabel, path string) error {
	if len(names) != len(vals) {
		return fmt.Errorf("names (%d) and vals (%d) must have the same length", len(names), len(vals))
	}
	p := newPlot(title, xLabel, yLabel)
	b, err := plotter.NewBarChart(plotter.Values(vals), vg.Points(45))
	if err != nil {
		return fmt.Errorf("error creating bar chart: %w", err)
	}
	b.Color = orange
	b.LineStyle.Color = color.RGBA{R: 160, G: 80, B: 0, A: 255}
	p.Add(b)
	p.NominalX(names...)

	pts := make(plotter.XYs, len(vals))
	labels := make([]string, len(vals))
	for i, val := range vals {
		pts[i].X = float64(i)
		pts[i].Y = val
		labels[i] = fmt.Sprintf("%.1f", val)
	}
	lab, err := plotter.NewLabels(plotter.XYLabels{XYs: pts, Labels: labels})
	if err != nil {
		return fmt.Errorf("error creating bar labels: %w", err)
	}
	lab.Offset = vg.Point{X: -vg.Points(10), Y: vg.Points(3)}
	p.Add(lab)
	return p.Save(7*vg.Inch, 5*vg.Inch, path)
}

// linePlot draws a line with a marker at each point (used for a value per year).
func linePlot(xs, ys []float64, title, xLabel, yLabel, path string) error {
	if len(xs) != len(ys) {
		return fmt.Errorf("xs (%d) and ys (%d) must have the same length", len(xs), len(ys))
	}
	p := newPlot(title, xLabel, yLabel)
	pts := make(plotter.XYs, len(xs))
	for i := range xs {
		pts[i].X = xs[i]
		pts[i].Y = ys[i]
	}
	line, points, err := plotter.NewLinePoints(pts)
	if err != nil {
		return fmt.Errorf("error creating line plot: %w", err)
	}
	line.Color = blue
	line.Width = vg.Points(2)
	points.Color = blue
	points.Radius = vg.Points(3.5)
	p.Add(line, points)

	// One labelled tick per x value, so years show as 2012, 2013, ...
	ticks := make([]plot.Tick, len(xs))
	for i, x := range xs {
		ticks[i] = plot.Tick{Value: x, Label: strconv.Itoa(int(x))}
	}
	p.X.Tick.Marker = plot.ConstantTicks(ticks)
	return p.Save(7*vg.Inch, 5*vg.Inch, path)
}
