// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
)

// LogarithmicFormula represents the formula y = a * ln(x) + b.
type LogarithmicFormula struct {
	a float64
	b float64
}

// Evaluate receive an x value and returns the y value of the formula.
func (lf *LogarithmicFormula) Evaluate(x float64) float64 {
	return lf.a*math.Log(x) + lf.b
}

func MustLogarithmicFit(xData, yData []float64) *LogarithmicFormula {
	ret, err := LogarithmicFit(xData, yData)
	if err != nil {
		logrus.Panic(err)
	}
	return ret
}

func LogarithmicFit(xData, yData []float64) (*LogarithmicFormula, error) {
	var err error
	var a, b float64

	a, b, err = logarithmicFit(xData, yData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate a logarithmic fit: %w", err)
	}
	return &LogarithmicFormula{a: a, b: b}, nil
}

// logarithmicFit performs least squares fitting for the logarithmic model y = a * ln(x) + b.
// It returns the a and b constants of the logarithmic model.
func logarithmicFit(xData, yData []float64) (float64, float64, error) {
	n := len(xData)
	if n != len(yData) || n == 0 {
		return 0, 0, fmt.Errorf("input data is invalid. xData: %v, yData: %v", xData, yData)
	}

	lnXData := make([]float64, n)

	for i := 0; i < n; i++ {
		if xData[i] <= 0 {
			return 0, 0, fmt.Errorf("input data contains non-positive x values: %v", xData)
		}
		lnXData[i] = math.Log(xData[i])
	}

	// Calculate the constants of the logarithmic regression model y = a * ln(x) + b
	var sumLnX, sumY, sumLnX2, sumLnXY float64
	for i := 0; i < n; i++ {
		sumLnX += lnXData[i]
		sumY += yData[i]
		sumLnX2 += lnXData[i] * lnXData[i]
		sumLnXY += lnXData[i] * yData[i]
	}

	if ((float64(n))*sumLnX2 - sumLnX*sumLnX) == 0 {
		return 0, 0, fmt.Errorf("zero denominator in calculations")
	}

	a := ((float64(n))*sumLnXY - sumLnX*sumY) / ((float64(n))*sumLnX2 - sumLnX*sumLnX)
	b := (sumY - a*sumLnX) / float64(n)

	return a, b, nil
}
