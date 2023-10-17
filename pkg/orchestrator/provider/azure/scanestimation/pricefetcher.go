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

package scanestimation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

type Item struct {
	CurrencyCode         string  `json:"currencyCode"`
	TierMinimumUnits     float64 `json:"tierMinimumUnits"`
	RetailPrice          float64 `json:"retailPrice"`
	UnitPrice            float64 `json:"unitPrice"`
	ArmRegionName        string  `json:"armRegionName"`
	Location             string  `json:"location"`
	EffectiveStartDate   string  `json:"effectiveStartDate"`
	MeterID              string  `json:"meterId"`
	MeterName            string  `json:"meterName"`
	ProductID            string  `json:"productId"`
	SkuID                string  `json:"skuId"`
	ProductName          string  `json:"productName"`
	SkuName              string  `json:"skuName"`
	ServiceName          string  `json:"serviceName"`
	ServiceID            string  `json:"serviceId"`
	ServiceFamily        string  `json:"serviceFamily"`
	UnitOfMeasure        string  `json:"unitOfMeasure"`
	Type                 string  `json:"type"`
	IsPrimaryMeterRegion bool    `json:"isPrimaryMeterRegion"`
	ArmSkuName           string  `json:"armSkuName"`
}

type Data struct {
	BillingCurrency    string  `json:"BillingCurrency"`
	CustomerEntityID   string  `json:"CustomerEntityID"`
	CustomerEntityType string  `json:"CustomerEntityType"`
	Items              []Item  `json:"Items"`
	NextPageLink       *string `json:"NextPageLink"`
	Count              int     `json:"Count"`
}

const (
	priceListBaseURL = "https://prices.azure.com/api/retail/prices?api-version=2023-01-01-preview"
)

type PriceFetcherImpl struct {
	client *http.Client
}

var StandardSSDSizes = []StandardSSDSize{
	{4, "E1"},
	{8, "E2"},
	{16, "E3"},
	{32, "E4"},
	{64, "E5"},
	{128, "E6"},
	{256, "E10"},
	{512, "E15"},
	{1024, "E20"},
	{1024, "E30"},
	{2048, "E40"},
	{4096, "E50"},
	{8192, "E60"},
	{16384, "E70"},
	{32767, "E80"},
}

type StandardSSDSize struct {
	size   int64
	symbol string
}

func findClosestSSDSizeSymbol(diskSize int64) string {
	for _, s := range StandardSSDSizes {
		if diskSize <= s.size {
			return s.symbol
		}
	}
	return "E80"
}

var diskStorageAccountToFilter = map[armcompute.DiskStorageAccountTypes]string{
	armcompute.DiskStorageAccountTypesStandardSSDLRS: "armRegionName eq '%s' and serviceFamily eq 'Storage' and meterName eq '%v Disks' and skuName eq '%v LRS'",
}

var snapshotStorageAccountToFilters = map[armcompute.SnapshotStorageAccountTypes]string{
	armcompute.SnapshotStorageAccountTypesStandardLRS: "armRegionName eq '%s' and serviceFamily eq 'Storage' and meterName eq 'LRS Snapshots' and skuName eq 'Snapshots LRS' and productName eq 'Standard HDD Managed Disks'",
}

var virtualMachineSizeToFilters = map[armcompute.VirtualMachineSizeTypes]string{
	armcompute.VirtualMachineSizeTypesStandardD2SV3: "armRegionName eq '%s' and serviceFamily eq 'Compute' and contains(armSkuName,'Standard_D2s_v3') and type eq 'Consumption' and meterName eq 'D2s v3' and productName eq 'Virtual Machines DSv3 Series'",
}

var spotVirtualMachineSizeToFilters = map[armcompute.VirtualMachineSizeTypes]string{
	armcompute.VirtualMachineSizeTypesStandardD2SV3: "armRegionName eq '%s' and serviceFamily eq 'Compute' and contains(armSkuName,'Standard_D2s_v3') and type eq 'Consumption' and meterName eq 'D2s v3 Spot' and productName eq 'Virtual Machines DSv3 Series'",
}

var blobStorageAccountToFilter = map[armcompute.StorageAccountTypes]string{
	armcompute.StorageAccountTypesStandardLRS: "armRegionName eq '%s' and serviceFamily eq 'Storage' and meterName eq 'LRS Data Stored' and skuName eq 'Standard LRS' and productName eq 'Standard Page Blob v2'",
}

func (o *PriceFetcherImpl) GetSnapshotGBPerMonthCost(ctx context.Context, region string, storageAccountType armcompute.SnapshotStorageAccountTypes) (float64, error) {
	odataFilter := snapshotStorageAccountToFilters[storageAccountType]
	odataFilter = fmt.Sprintf(odataFilter, region)

	return o.getRetailPrice(ctx, odataFilter)
}

func (o *PriceFetcherImpl) GetManagedDiskMonthlyCost(ctx context.Context, region string, storageAccountType armcompute.DiskStorageAccountTypes, diskSize int64) (float64, error) {
	odataFilter := diskStorageAccountToFilter[storageAccountType]
	symbol := findClosestSSDSizeSymbol(diskSize)
	odataFilter = fmt.Sprintf(odataFilter, region, symbol, symbol)

	return o.getRetailPrice(ctx, odataFilter)
}

func (o *PriceFetcherImpl) GetDataTransferPerGBCost(ctx context.Context, destRegion string) (float64, error) {
	odataFilter := fmt.Sprintf("armRegionName eq '%s' and contains(meterName,'Region Data Transfer')", destRegion)

	return o.getRetailPrice(ctx, odataFilter)
}

func (o *PriceFetcherImpl) GetInstancePerHourCost(ctx context.Context, region string, vmSize armcompute.VirtualMachineSizeTypes, marketOption MarketOption) (float64, error) {
	var odataFilter string

	if marketOption == MarketOptionSpot {
		odataFilter = spotVirtualMachineSizeToFilters[vmSize]
		odataFilter = fmt.Sprintf(odataFilter, region)
	} else {
		odataFilter = virtualMachineSizeToFilters[vmSize]
		odataFilter = fmt.Sprintf(odataFilter, region)
	}

	return o.getRetailPrice(ctx, odataFilter)
}

func (o *PriceFetcherImpl) GetBlobStoragePerGBCost(ctx context.Context, region string, storageAccountType armcompute.StorageAccountTypes) (float64, error) {
	odataFilter := blobStorageAccountToFilter[storageAccountType]
	odataFilter = fmt.Sprintf(odataFilter, region)

	return o.getRetailPrice(ctx, odataFilter)
}

func (o *PriceFetcherImpl) getRetailPrice(ctx context.Context, odataFilter string) (float64, error) {
	urlWithFilter, err := addQueryParamToURL(priceListBaseURL, "$filter", odataFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to add query params to url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, urlWithFilter, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := o.client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code is not 200 when calling url: %v. status code: %v", urlWithFilter, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read body: %w", err)
	}

	var data Data

	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if data.Count != 1 {
		return 0, fmt.Errorf("found %v items from response, excpecting 1. response: %s", data.Count, body)
	}

	return data.Items[0].RetailPrice, nil
}

func addQueryParamToURL(baseURL, key, value string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url %v: %w", baseURL, err)
	}

	encodedParam := fmt.Sprintf("%s=%s", key, url.QueryEscape(value))

	// If there are already parameters, append the new one with '&', otherwise just add the new one
	if u.RawQuery != "" {
		u.RawQuery = u.RawQuery + "&" + encodedParam
	} else {
		u.RawQuery = encodedParam
	}

	return u.String(), nil
}
