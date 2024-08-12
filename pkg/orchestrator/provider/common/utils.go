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

	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	familiestypes "github.com/openclarity/vmclarity/pkg/shared/families/types"
)

const (
	MBInGB = 1000
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

// GetScanSize Search in all the families stats and look for the first family (by random order) that has scan size stats for ROOTFS scan.
// nolint:cyclop
func GetScanSize(stats models.AssetScanStats, asset *models.Asset) (int64, error) {
	var scanSizeMB int64
	const half = 2

	sbomStats, ok := findMatchingStatsForInputTypeRootFS(stats.Sbom)
	if ok {
		if sbomStats.Size != nil && *sbomStats.Size > 0 {
			return *sbomStats.Size, nil
		}
	}

	vulStats, ok := findMatchingStatsForInputTypeRootFS(stats.Vulnerabilities)
	if ok {
		if vulStats.Size != nil && *vulStats.Size > 0 {
			return *vulStats.Size, nil
		}
	}

	secretsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Secrets)
	if ok {
		if secretsStats.Size != nil && *secretsStats.Size > 0 {
			return *secretsStats.Size, nil
		}
	}

	exploitsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Exploits)
	if ok {
		if exploitsStats.Size != nil && *exploitsStats.Size > 0 {
			return *exploitsStats.Size, nil
		}
	}

	rootkitsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Rootkits)
	if ok {
		if rootkitsStats.Size != nil && *rootkitsStats.Size > 0 {
			return *rootkitsStats.Size, nil
		}
	}

	misconfigurationsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Misconfigurations)
	if ok {
		if misconfigurationsStats.Size != nil && *misconfigurationsStats.Size > 0 {
			return *misconfigurationsStats.Size, nil
		}
	}

	malwareStats, ok := findMatchingStatsForInputTypeRootFS(stats.Malware)
	if ok {
		if malwareStats.Size != nil && *malwareStats.Size > 0 {
			return *malwareStats.Size, nil
		}
	}

	// if scan size was not found from the previous scan stats, estimate the scan size from the asset root volume size
	vminfo, err := asset.AssetInfo.AsVMInfo()
	if err != nil {
		return 0, fmt.Errorf("failed to use asset info as vminfo: %w", err)
	}
	sourceVolumeSizeMB := int64(vminfo.RootVolume.SizeGB * MBInGB)
	scanSizeMB = sourceVolumeSizeMB / half // Volumes are normally only about 50% full

	return scanSizeMB, nil
}

// findMatchingStatsForInputTypeRootFS will find the first stats for rootfs scan.
func findMatchingStatsForInputTypeRootFS(stats *[]models.AssetScanInputScanStats) (models.AssetScanInputScanStats, bool) {
	if stats == nil {
		return models.AssetScanInputScanStats{}, false
	}
	for i, scanStats := range *stats {
		if *scanStats.Type == string(utils.ROOTFS) {
			ret := *stats
			return ret[i], true
		}
	}
	return models.AssetScanInputScanStats{}, false
}

// nolint:cyclop
func GetScanDuration(stats models.AssetScanStats, familiesConfig *models.ScanFamiliesConfig, scanSizeMB int64, familyScanDurationsMap map[familiestypes.FamilyType]*LogarithmicFormula) int64 {
	var totalScanDuration int64

	scanSizeGB := float64(scanSizeMB) / MBInGB

	if familiesConfig.Sbom.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Sbom)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			// if we didn't find the duration from the stats, take it from our static scan duration map.
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.SBOM].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Vulnerabilities.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Vulnerabilities)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.Vulnerabilities].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Secrets.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Secrets)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.Secrets].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Exploits.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Exploits)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.Exploits].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Rootkits.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Rootkits)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.Rootkits].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Misconfigurations.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Misconfigurations)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.Misconfiguration].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Malware.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Malware)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(familyScanDurationsMap[familiestypes.Malware].Evaluate(scanSizeGB))
		}
	}

	return totalScanDuration
}

func getScanDurationFromStats(stats *[]models.AssetScanInputScanStats) int64 {
	stat, ok := findMatchingStatsForInputTypeRootFS(stats)
	if !ok {
		return 0
	}

	if stat.ScanTime == nil {
		return 0
	}
	if stat.ScanTime.EndTime == nil || stat.ScanTime.StartTime == nil {
		return 0
	}

	dur := stat.ScanTime.EndTime.Sub(*stat.ScanTime.StartTime)

	return int64(dur.Seconds())
}
