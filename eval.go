package main

import (
	"fmt"
	"math"
)

func (m *Model) Predict(r FieldData) (float64, error) {
	if len(m.Coef) != 7 {
		return 0, fmt.Errorf("model has %d coefficients, expected 7", len(m.Coef))
	}

	feats, err := features(r)
	if err != nil {
		return 0, err
	}
	var predictedVol float64
	for i := 0; i < len(m.Coef); i++ {
		predictedVol += m.Coef[i] * feats[i]
	}
	return predictedVol, nil
}

func evaluate(m *Model, rows []FieldData) (R2, RMSE float64, err error) {
	if len(rows) == 0 {
		return 0, 0, fmt.Errorf("rows is empty")
	}

	meanFlaringVol := 0.0
	for _, r := range rows {
		meanFlaringVol += r.FlaringVol
	}
	meanFlaringVol /= float64(len(rows))

	TSS := 0.0
	RSS := 0.0
	for _, r := range rows {
		predictedVol, err := m.Predict(r)
		if err != nil {
			return 0, 0, err
		}
		RSS += (r.FlaringVol - predictedVol) * (r.FlaringVol - predictedVol)
		TSS += (r.FlaringVol - meanFlaringVol) * (r.FlaringVol - meanFlaringVol)
	}

	if TSS == 0 {
		R2 = 1.0
	} else {
		R2 = 1 - (RSS / TSS)
	}
	RMSE = math.Sqrt(RSS / float64(len(rows)))

	return R2, RMSE, nil
}
