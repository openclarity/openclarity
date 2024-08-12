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

package cyclonedx_helper // nolint:revive,stylecheck

import (
	"errors"
	"sort"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	purl "github.com/package-url/packageurl-go"
	log "github.com/sirupsen/logrus"
)

func GetComponentHash(component *cdx.Component) (string, error) {
	expectedHashPartsLen := 2

	var lastHash string

	if component == nil {
		return "", errors.New("missing component")
	}

	// In the case of an Image, the BOM version contains the manifestDigest
	//    <component type="container">
	//      <name>poke/test:latest</name>
	//      <version>sha256:81dc3590849e5a1d19f64089979859bb0a62aaa166b46771375a5f6c2d857de3</version>
	//    </component>
	// if there are no hashes, use manifestDigest as a hash by default
	if component.Type == cdx.ComponentTypeContainer {
		if component.Version != "" {
			hashParts := strings.Split(component.Version, "sha256:")
			if len(hashParts) == expectedHashPartsLen {
				lastHash = hashParts[1]
			}
		}
	}

	// There can be multiple hashes in the case of cycloneDX BOM metadata
	// If sha265 hash exist use it, if not use the last one in the list
	// Usually hashes orders in the list are ascending by the algorithm:
	//        <hashes>
	//          <hash alg="MD5">xxxx</hash>
	//          <hash alg="SHA-1">xxxx</hash>
	//          <hash alg="SHA-256">xxxx</hash>
	//          <hash alg="SHA-512">xxxx</hash>
	//        </hashes>
	if component.Hashes != nil {
		hashes := sortHashes(*component.Hashes)
		for _, hash := range hashes {
			lastHash = hash.Value
			if hashParts := strings.Split(hash.Value, ":"); len(hashParts) == expectedHashPartsLen {
				lastHash = hashParts[1]
			}
			if hash.Algorithm == cdx.HashAlgoSHA256 {
				return lastHash, nil
			}
		}
	}

	if lastHash == "" {
		return "", errors.New("no sha256 hash found in component")
	}
	return lastHash, nil
}

func GetComponentLicenses(component cdx.Component) []string {
	//       <licenses>
	//        <license>
	//          <id>MIT</id>
	//          <url>https://spdx.org/licenses/MIT.html</url>
	//        </license>
	//        <expression>GPL-3.0</expression>
	//      </licenses>
	if component.Licenses == nil || len(*component.Licenses) == 0 {
		return nil
	}

	licenses := []string{}
	for _, license := range *component.Licenses {
		// Licenses may be one of either cyclonedx License type or an
		// spdx license expression but not both.
		// https://github.com/CycloneDX/specification/blob/1.4/schema/bom-1.4.xsd#L1398
		// https://spdx.github.io/spdx-spec/v2.3/SPDX-license-expressions/
		if license.License != nil {
			// A license must either have one of ID OR Name specified according to the CDX spec, otherwise its invalid CDX:
			// https://cyclonedx.org/docs/1.4/json/#tab-pane_components_items_licenses_items_oneOf_i0
			if license.License.ID != "" {
				licenses = append(licenses, license.License.ID)
			} else if license.License.Name != "" {
				licenses = append(licenses, license.License.Name)
			}
		}

		if license.Expression != "" {
			// TODO(sambetts) We may need to post-process this
			// expression it can contain multiple licenses like
			// "GPL-3.0 OR LGPL-2.0" etc, see the spdx spec for
			// more details.
			licenses = append(licenses, license.Expression)
		}
	}

	return licenses
}

// nolint:cyclop
func GetComponentLanguage(component cdx.Component) string {
	// Get language from the PackageURL.
	// PackageURL is a mandatory field for Component, so it should exist.
	p, err := purl.FromString(component.PackageURL)
	if err != nil {
		log.Warnf("Failed to convert PURL from string: %v", err)
		return ""
	}
	var lang string
	// Defined languages in package-url
	// https://github.com/package-url/purl-spec/blob/master/PURL-TYPES.rst#package-url-type-definitions
	switch p.Type {
	case "golang":
		lang = "go"
	case "pypi":
		lang = "python"
	case "npm":
		lang = "javascript"
	case "gem":
		lang = "ruby"
	case "cargo":
		lang = "rust"
	case "maven":
		lang = "java"
	case "nuget":
		lang = "c#"
	case "cran":
		lang = "r"
	case "swift":
		lang = "swift"
	case "hackage":
		lang = "haskell"
	case "composer":
		lang = "php"
	case "conan":
		lang = "c/c++"
	case "hex":
		lang = "erlang"
	default:
		// In the case of typ is not a language foe example deb, apk, rpm etc... return empty string
		lang = ""
	}

	return lang
}

func sortHashes(hashes []cdx.Hash) []cdx.Hash {
	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i].Algorithm < hashes[j].Algorithm
	})
	return hashes
}
