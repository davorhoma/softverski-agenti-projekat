package utils

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/linear"
)

func TrainHeartDiseaseLogisticRegression(trainCSV string, initialParams []float64) ([]float64, error) {
	file, err := os.Open(trainCSV)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read()

	var X [][]float64
	var Y []float64

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var row []float64
		for _, s := range record[0:len(record)-1] {
			f, _ := strconv.ParseFloat(s, 64)
			row = append(row, f)
		}
		X = append(X, row)

		if record[len(record)-1] == "1" {
			Y = append(Y, 1.0)
		} else {
			Y = append(Y, 0.0)
		}
	}


	model := linear.NewLogistic(base.BatchGA, 0.01, 0.0, 1000, X, Y)

	// Postavljanje inicijalnih parametara ako postoje
	if initialParams != nil && len(initialParams) == len(X[0])+1 {
		copy(model.Theta(), initialParams)
	}

	model.Output = io.Discard

	err = model.Learn()
	if err != nil {
		return nil, err
	}

	return model.Theta(), nil
}

func EvaluateHeartDiseaseLogisticRegression(testCSV string, params []float64) (float64, error) {
	file, err := os.Open(testCSV)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read()

	var X [][]float64
	var Y []float64

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		var row []float64
		for _, s := range record[0:len(record)-1] {
			f, _ := strconv.ParseFloat(s, 64)
			row = append(row, f)
		}
		X = append(X, row)

		if record[len(record)-1] == "1" {
			Y = append(Y, 1.0)
		} else {
			Y = append(Y, 0.0)
		}
	}

	model := linear.NewLogistic(base.BatchGA, 0.01, 0.0, 1000, X, Y)
	if params != nil && len(params) == len(model.Theta()) {
		copy(model.Theta(), params)
	}

	correct := 0
	for i, x := range X {
		pred, err := model.Predict(x)
		if err != nil {
			return 0, err
		}
		predictedLabel := 0.0
		if pred[0] >= 0.5 {
			predictedLabel = 1.0
		}
		if predictedLabel == Y[i] {
			correct++
		}
	}

	accuracy := float64(correct) / float64(len(X))
	return accuracy, nil
}
