package main

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

const (
	GAS     = "GAS"
	LNG     = "LNG"
	OIL     = "OIL"
	Unknown = "Unknown"
)

type Model struct {
	Coef []float64
}

func features(r FieldData) ([]float64, error) {
	if r.FieldType == Unknown {
		return []float64{1, float64(r.Year), r.Latitude, r.Longitude, 0, 0, 0}, nil
	}
	if r.FieldType == GAS {
		return []float64{1, float64(r.Year), r.Latitude, r.Longitude, 1, 0, 0}, nil
	}
	if r.FieldType == LNG {
		return []float64{1, float64(r.Year), r.Latitude, r.Longitude, 0, 1, 0}, nil
	}
	if r.FieldType == OIL {
		return []float64{1, float64(r.Year), r.Latitude, r.Longitude, 0, 0, 1}, nil
	}
	return nil, fmt.Errorf("unsupported field type: %s", r.FieldType)
}

func fitOLS(rows []FieldData) (*Model, error) {
	n := len(rows)
	if n <= 7 {
		return nil, fmt.Errorf("need more than 7 rows to fit, have %d", n)
	}
	X := mat.NewDense(n, 7, nil)
	for i, r := range rows {
		feats, err := features(r)
		if err != nil {
			return nil, err
		}
		X.SetRow(i, feats)
	}
	y := mat.NewVecDense(n, nil)
	for i, r := range rows {
		y.SetVec(i, r.FlaringVol)
	}
	var qr mat.QR
	qr.Factorize(X)
	var beta mat.VecDense
	err := qr.SolveVecTo(&beta, false, y)
	if err != nil {
		return nil, fmt.Errorf("failed to solve for beta: %w", err)
	}
	coef := beta.RawVector().Data
	return &Model{Coef: coef}, nil
}
