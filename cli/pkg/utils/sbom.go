package utils

import (
	"fmt"
	"os"

	"github.com/openclarity/kubeclarity/shared/pkg/converter"
	"github.com/openclarity/kubeclarity/shared/pkg/formatter"
)

func ConvertInputSBOMIfNeeded(inputSBOMFile, outputFormat string) ([]byte, error) {
	inputSBOM, err := os.ReadFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file %s: %v", inputSBOMFile, err)
	}
	inputSBOMFormat := converter.DetermineCycloneDXFormat(inputSBOM)
	if inputSBOMFormat == outputFormat {
		return inputSBOM, nil
	}

	// Create cycloneDX formatter to convert input SBOM to the defined output format.
	cdxFormatter := formatter.New(inputSBOMFormat, inputSBOM)
	if err = cdxFormatter.Decode(inputSBOMFormat); err != nil {
		return nil, fmt.Errorf("failed to decode input SBOM %s: %v", inputSBOMFile, err)
	}
	if err := cdxFormatter.Encode(outputFormat); err != nil {
		return nil, fmt.Errorf("failed to encode input SBOM: %v", err)
	}

	return cdxFormatter.GetSBOMBytes(), nil
}
