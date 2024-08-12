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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"

	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type PriceFetcher interface {
	GetSnapshotMonthlyCostPerGB(ctx context.Context, regionCode string) (float64, error)
	GetVolumeMonthlyCostPerGB(ctx context.Context, regionCode string, volumeType ec2types.VolumeType) (float64, error)
	GetDataTransferCostPerGB(sourceRegion, destRegion string) (float64, error)
	GetInstancePerHourCost(ctx context.Context, regionCode string, instanceType ec2types.InstanceType, marketOption MarketOption) (float64, error)
}

type PriceFetcherImpl struct {
	pricingClient *pricing.Client
	ec2Client     *ec2.Client
}

// https://docs.aws.amazon.com/AmazonS3/latest/userguide/aws-usage-report-understand.html
var regionCodeToUsageTypeAbbreviation = map[string]string{
	"us-east-1":      "",     // N. virginia
	"us-east-2":      "USE2", // Ohio
	"us-west-1":      "USW1", // N. california
	"us-west-2":      "USW2", // Oregon
	"eu-central-1":   "EUC1", // Frankfurt
	"eu-central-2":   "EUC2", // Zurich
	"eu-south-1":     "EUS1", // Milan
	"eu-west-1":      "EUW1", // Ireland
	"eu-west-2":      "EUW2", // London
	"eu-west-3":      "EUW3", // Paris
	"af-south-1":     "CPT",  // Cape Town
	"ap-east-1":      "APE1", // Hong Kong
	"ap-southeast-3": "APS4", // Jakarta
	"ap-south-2":     "APS5", // Hyderabad
	"ap-southeast-4": "APS6", // Melbourne
	"ca-central-1":   "CAN1", // Canada (Central)
	"eu-north-1":     "EUN1", // Stockholm
	"eu-south-2":     "EUS2", // Spain
	"me-central-1":   "MEC1", // Middle East (UAE)
	"me-south-1":     "MES1", // Middle East (Bahrain)
	"sa-east-1":      "SAE1", // São Paulo
	"us-gov-west-1":  "UGW1", // AWS GovCloud (US-West)
	"us-gov-east-1":  "UGE1", // AWS GovCloud (US-East)
	//"il-central-1": "EUS2", // Tel-Aviv // TODO couldn't find the abbreviation code

}

const (
	usEast1RegionCode = "us-east-1"
)

const (
	estimatedSnapshotCostInCaseOfError = 0.05
	estimatedDataTransferCost          = 0.02
)

func (o *PriceFetcherImpl) GetSnapshotMonthlyCostPerGB(ctx context.Context, regionCode string) (float64, error) {
	var usageType string
	if regionCode == usEast1RegionCode {
		usageType = "EBS:SnapshotUsage"
	} else {
		abb, ok := regionCodeToUsageTypeAbbreviation[regionCode]
		if !ok {
			return 0, fmt.Errorf("failed to find usage type in regionCodeToUsageTypeAbbreviation map with regionCode %v", regionCode)
		}
		// example: USE2-EBS:SnapshotUsage
		usageType = fmt.Sprintf("%v-EBS:SnapshotUsage", abb)
	}

	price, err := o.getPricePerUnit(ctx, usageType, "")
	if err != nil {
		// If we could not find the snapshot price from the api, use an estimated price based on the current pricing (5.9.23)
		// https://aws.amazon.com/ebs/pricing/
		// nolint:nilerr
		return estimatedSnapshotCostInCaseOfError, nil
	}
	return price, nil
}

func (o *PriceFetcherImpl) GetVolumeMonthlyCostPerGB(ctx context.Context, regionCode string, volumeType ec2types.VolumeType) (float64, error) {
	var usageType string
	if regionCode == usEast1RegionCode {
		usageType = fmt.Sprintf("EBS:VolumeUsage.%v", volumeType)
	} else {
		abb, ok := regionCodeToUsageTypeAbbreviation[regionCode]
		if !ok {
			return 0, fmt.Errorf("failed to find usage type in regionCodeToUsageTypeAbbreviation map with regionCode %v", regionCode)
		}
		// example: USE2-EBS:VolumeUsage.gp2
		usageType = fmt.Sprintf("%v-EBS:VolumeUsage.%v", abb, volumeType)
	}

	return o.getPricePerUnit(ctx, usageType, "")
}

func (o *PriceFetcherImpl) GetDataTransferCostPerGB(sourceRegion, destRegion string) (float64, error) {
	if destRegion == sourceRegion {
		return 0, nil
	}
	// TODO (erezf) currently I could not find a reliable way to get the data transfer cost from the offer file.
	// 0.02 seems to be a unified price across regions according to https://aws.amazon.com/ec2/pricing/on-demand/
	// This price will probably change in the future, so need to think of a way to keep it updated.
	return estimatedDataTransferCost, nil
}

func (o *PriceFetcherImpl) GetInstancePerHourCost(ctx context.Context, regionCode string, instanceType ec2types.InstanceType, marketOption MarketOption) (float64, error) {
	if marketOption == MarketOptionSpot {
		return o.getSpotInstancePerHourCost(ctx, regionCode, instanceType)
	}

	var usageType string
	if regionCode == usEast1RegionCode {
		usageType = fmt.Sprintf("BoxUsage:%v", instanceType)
	} else {
		abb, ok := regionCodeToUsageTypeAbbreviation[regionCode]
		if !ok {
			return 0, fmt.Errorf("failed to find usage type in regionCodeToUsageTypeAbbreviation map with regionCode %v", regionCode)
		}
		// example: USE2-BoxUsage:t2.large
		usageType = fmt.Sprintf("%v-BoxUsage:%v", abb, instanceType)
	}

	return o.getPricePerUnit(ctx, usageType, "RunInstances")
}

func (o *PriceFetcherImpl) getSpotInstancePerHourCost(ctx context.Context, regionCode string, instanceType ec2types.InstanceType) (float64, error) {
	timeNow := time.Now()
	ret, err := o.ec2Client.DescribeSpotPriceHistory(ctx, &ec2.DescribeSpotPriceHistoryInput{
		Filters: []ec2types.Filter{
			{
				Name:   utils.PointerTo("product-description"),
				Values: []string{"Linux/UNIX"},
			},
		},
		InstanceTypes: []ec2types.InstanceType{instanceType},
		StartTime:     utils.PointerTo(timeNow.Add(-12 * time.Hour)), // last 12 hours should be enough as we only want the latest.
	}, func(options *ec2.Options) {
		options.Region = regionCode
	})
	if err != nil {
		return 0, fmt.Errorf("failed to describe spot price history: %w", err)
	}

	if len(ret.SpotPriceHistory) == 0 {
		return 0, errors.New("failed to find spot instances history")
	}

	// take the latest price (most updated)
	spotPrice := ret.SpotPriceHistory[0].SpotPrice

	price, err := strconv.ParseFloat(*spotPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float %v: %w", *spotPrice, err)
	}

	return price, nil
}

func (o *PriceFetcherImpl) getPricePerUnit(ctx context.Context, usageType, operation string) (float64, error) {
	filters := createGetProductsFilters(usageType, "AmazonEC2", operation)

	products, err := o.pricingClient.GetProducts(ctx, &pricing.GetProductsInput{
		ServiceCode:   utils.PointerTo("AmazonEC2"),
		Filters:       filters,
		FormatVersion: utils.PointerTo("aws_v1"),
	}, func(options *pricing.Options) {
		// the Pricing API is only available on us-east-1 and ap-south-1. for now, we've chosen to use the us-east-1 endpoint.
		options.Region = usEast1RegionCode
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get products. usageType=%v: %w", usageType, err)
	}
	if len(products.PriceList) != 1 {
		return 0, fmt.Errorf("expecting exactly one product in price list, got %v", len(products.PriceList))
	}

	priceStr, err := getPricePerUnitFromJSONPriceList(products.PriceList[0])
	if err != nil {
		return 0, fmt.Errorf("failed to get pricePerUnit from json price list: %w", err)
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float %v: %w", priceStr, err)
	}
	return price, nil
}

func createGetProductsFilters(usageType, serviceCode, operation string) []types.Filter {
	filters := []types.Filter{
		// An AWS SKU uniquely combines product (service code), Usage Type, and Operation for an AWS resource.
		// See https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/procedures.html for more details
		{
			Field: utils.PointerTo("ServiceCode"),
			Type:  "TERM_MATCH",
			Value: &serviceCode,
		},
		{
			Field: utils.PointerTo("usagetype"),
			Type:  "TERM_MATCH",
			Value: &usageType,
		},
	}
	if operation != "" {
		filters = append(filters, types.Filter{
			Field: utils.PointerTo("operation"),
			Type:  "TERM_MATCH",
			Value: &operation,
		})
	}

	return filters
}

// Parse the price list (offer file) which exists in json format, in order to get the pricePerUnit property from it.
// The structure of the price list json is:
/*
{
  "product": {
    "productFamily": "Compute Instance",
    "attributes": {
      .....
    },
    "sku": "7W6DNQ55YG9FCPXZ"
  },
  "serviceCode": "AmazonEC2",
  "terms": {
    "OnDemand": {
      "7W6DNQ55YG9FCPXZ.JRTCKXETXF": {
        "priceDimensions": {
          "7W6DNQ55YG9FCPXZ.JRTCKXETXF.6YS6EN2CT7": {
            "unit": "Hrs",
            "endRange": "Inf",
            "description": "$0.1072 per On Demand Linux t2.large Instance Hour",
            "appliesTo": [

            ],
            "rateCode": "7W6DNQ55YG9FCPXZ.JRTCKXETXF.6YS6EN2CT7",
            "beginRange": "0",
            "pricePerUnit": {
              "USD": "0.1072000000"
            }
          }
        },
        "sku": "7W6DNQ55YG9FCPXZ",
        "effectiveDate": "2023-08-01T00:00:00Z",
        "offerTermCode": "JRTCKXETXF",
        "termAttributes": {

        }
      }
    },
    "Reserved": {
       .....
       Reserved instances offering codes
    }
  },
  "version": "20230817222016",
  "publicationDate": "2023-08-17T22:20:16Z"
}
More details on: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/procedures.html
*/
func getPricePerUnitFromJSONPriceList(jsonPriceList string) (string, error) {
	var productMap map[string]any
	err := json.Unmarshal([]byte(jsonPriceList), &productMap)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal. jsonPriceList=%s: %w", jsonPriceList, err)
	}
	termsMap, ok := productMap["terms"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("terms key was not found in map. productMap=%v", productMap)
	}
	ondemandMap, ok := termsMap["OnDemand"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("OnDemand key was not found in map. termsMap=%v", termsMap)
	}

	// TODO (erezf) The key here is the rate code (7W6DNQ55YG9FCPXZ.JRTCKXETXF.6YS6EN2CT7). The rate code structure is <sku.offerTermCode.rateCode>
	// I can do a map of rate codes per region for each product we use, and by this knowing exactly what the key is according to the region and product.
	// problem is, what will happen when the offer file version will change? will that offer code change or stay the same?
	// since we don't know what the key name is, we will assume this is a map with size of 1. I think this is safe to assume since each product has exactly one OnDemand offering.
	val, err := getFirstValueFromMap(ondemandMap)
	if err != nil {
		return "", fmt.Errorf("failed to get first value from map. ondemandMap=%v", ondemandMap)
	}
	priceDimensionsMap, ok := val["priceDimensions"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("priceDimensions key was not found in map. map=%v", val)
	}
	val, err = getFirstValueFromMap(priceDimensionsMap)
	if err != nil {
		return "", fmt.Errorf("failed to get first value from map. priceDimensionsMap=%v", priceDimensionsMap)
	}
	pricePerUnitMap, ok := val["pricePerUnit"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("pricePerUnit key was not found in map. map=%v", val)
	}

	pricePerUnit, ok := pricePerUnitMap["USD"].(string)
	if !ok {
		return "", fmt.Errorf("USD key was not found in map. pricePerUnitMap=%v", pricePerUnitMap)
	}

	return pricePerUnit, nil
}

func getFirstValueFromMap(m map[string]any) (map[string]any, error) {
	for _, val := range m {
		ret, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("failed to convert value into map[string]any. value=%v", val)
		}

		return ret, nil
	}
	return nil, errors.New("map is empty")
}
