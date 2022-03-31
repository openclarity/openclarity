// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package cdx_gomod // nolint:revive,stylecheck

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"golang.org/x/crypto/sha3"
)

const author = "kubeclarity"

var Version = "v0.0.0-unset" // Must be a var so we can set it at build time

func buildToolMetadata() (*cdx.Tool, error) {
	toolExePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get exec path %v", err)
	}

	// Calculate only sha256 hash
	toolHashes, err := calculateFileHashes(toolExePath, cdx.HashAlgoSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tool hashes: %w", err)
	}

	return &cdx.Tool{
		Vendor:  author,
		Name:    AnalyzerName,
		Version: Version,
		Hashes:  &toolHashes,
		ExternalReferences: &[]cdx.ExternalReference{
			{
				Type: cdx.ERTypeVCS,
				URL:  "https://github.com/CycloneDX/cyclonedx-gomod",
			},
			{
				Type: cdx.ERTypeWebsite,
				URL:  "https://cyclonedx.org",
			},
		},
	}, nil
}

// nolint:cyclop
func calculateFileHashes(filePath string, algos ...cdx.HashAlgorithm) ([]cdx.Hash, error) {
	if len(algos) == 0 {
		return make([]cdx.Hash, 0), nil
	}

	hashMap := make(map[cdx.HashAlgorithm]hash.Hash)
	hashWriters := make([]io.Writer, 0)

	for _, algo := range algos {
		var hashWriter hash.Hash

		switch algo { //exhaustive:ignore
		case cdx.HashAlgoSHA256:
			hashWriter = sha256.New()
		case cdx.HashAlgoSHA384:
			hashWriter = sha512.New384()
		case cdx.HashAlgoSHA512:
			hashWriter = sha512.New()
		case cdx.HashAlgoSHA3_256:
			hashWriter = sha3.New256()
		case cdx.HashAlgoSHA3_512:
			hashWriter = sha3.New512()
		default:
			return nil, fmt.Errorf("unsupported hash algorithm: %s", algo)
		}

		hashWriters = append(hashWriters, hashWriter)
		hashMap[algo] = hashWriter
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file=%s: %v", filePath, err)
	}
	defer file.Close()

	// Use multiWriter to copy file to all hashWriter.
	// We can avoid to use _, err = io.Copy(hashMap[algo], file)
	// when iterating over algorithms below
	multiWriter := io.MultiWriter(hashWriters...)
	if _, err = io.Copy(multiWriter, file); err != nil {
		return nil, fmt.Errorf("falied to copy file=%s to hashWriters: %v", filePath, err)
	}

	cdxHashes := make([]cdx.Hash, 0, len(hashMap))
	for _, algo := range algos { // Don't iterate over hashMap, as it doesn't retain order
		// _, err = io.Copy(hashMap[algo], file) was done by multiWriter above
		cdxHashes = append(cdxHashes, cdx.Hash{
			Algorithm: algo,
			Value:     fmt.Sprintf("%x", hashMap[algo].Sum(nil)),
		})
	}

	return cdxHashes, nil
}

// set Licenses in Cyclondex BOM.
func assertLicenses(bom *cdx.BOM) {
	if bom == nil {
		return
	}
	if bom.Metadata != nil {
		assertComponentLicenses(bom.Metadata.Component)
	}
	if bom.Components != nil {
		for i := range *bom.Components {
			assertComponentLicenses(&(*bom.Components)[i])
		}
	}
}

// set Licenses based on Evidence.
func assertComponentLicenses(c *cdx.Component) {
	if c == nil {
		return
	}
	// If the Evidence field contains Licenses:
	//      <evidence>
	//        <licenses>
	//          <license>
	//            <id>MIT</id>
	//          </license>
	//        </licenses>
	//      </evidence>
	// We convert it to simple Licenses:
	//      <licenses>
	//        <license>
	//          <id>Apache-2.0</id>
	//        </license>
	//      </licenses>
	if c.Evidence != nil && c.Evidence.Licenses != nil {
		c.Licenses = c.Evidence.Licenses
		// Keep Evidence if it contains Copyright
		if c.Evidence.Copyright != nil {
			c.Evidence.Licenses = nil
		} else {
			c.Evidence = nil
		}
	}
	// Check if Components contains other Components
	if c.Components != nil {
		for i := range *c.Components {
			assertComponentLicenses(&(*c.Components)[i])
		}
	}
}
