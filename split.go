package main

// splitByYear puts rows with Year <= lastTrainYear in train and the rest in test.
func splitByYear(rows []FieldData, lastTrainYear int) (train, test []FieldData) {
	for _, r := range rows {
		if r.Year <= lastTrainYear {
			train = append(train, r)
		} else {
			test = append(test, r)
		}
	}
	return train, test
}
