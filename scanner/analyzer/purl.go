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

package analyzer

import (
	"fmt"
	"net/url"
	"strings"
)

type purl struct {
	scheme    string
	typ       string
	namespace string
	name      string
	version   string
	qualifers url.Values
	subpath   string
}

func (p purl) String() string {
	opaque := p.typ
	if p.namespace != "" {
		opaque = fmt.Sprintf("%s/%s", opaque, p.namespace)
	}
	opaque = fmt.Sprintf("%s/%s@%s", opaque, p.name, p.version)

	pURL := url.URL{
		Scheme:   p.scheme,
		Opaque:   opaque,
		RawQuery: p.qualifers.Encode(),
		Fragment: p.subpath,
	}

	return pURL.String()
}

func newPurl() purl {
	return purl{
		qualifers: url.Values{},
	}
}

var validPurlTypes = map[string]struct{}{
	"alpm":        {},
	"apk":         {},
	"bitbucket":   {},
	"cocoapods":   {},
	"cargo":       {},
	"composer":    {},
	"conan":       {},
	"conda":       {},
	"cran":        {},
	"deb":         {},
	"docker":      {},
	"gem":         {},
	"generic":     {},
	"github":      {},
	"golang":      {},
	"hackage":     {},
	"hex":         {},
	"huggingface": {},
	"maven":       {},
	"mlflow":      {},
	"npm":         {},
	"nuget":       {},
	"qpkg":        {},
	"oci":         {},
	"pub":         {},
	"pypi":        {},
	"rpm":         {},
	"swid":        {},
	"swift":       {},
}

func isValidPurlType(t string) bool {
	_, ok := validPurlTypes[t]
	return ok
}

// nolint:mnd
func purlStringToStruct(purlInput string) purl {
	if purlInput == "" {
		return newPurl()
	}

	purlURL, err := url.Parse(purlInput)
	if err != nil {
		// Not a valid purl so return empty
		return newPurl()
	}

	output := purl{
		scheme:    purlURL.Scheme,
		qualifers: purlURL.Query(),
		subpath:   purlURL.Fragment,
	}

	purlParts := strings.Split(purlURL.Opaque, "/")

	// Purls Opaques have 2 valid formats:
	//
	// * type/namespace/name@version
	// * type/name@version
	//
	// Namespace part is optional and type specific, the other fields are
	// required.
	if len(purlParts) == 3 {
		output.typ = purlParts[0]
		output.namespace = purlParts[1]
	} else if len(purlParts) == 2 {
		// Check type is a valid PURL type, if it is then use it.
		//
		// Otherwise, check if the type is a namespace we know belong to
		// one of the types, sometimes the anaylzers goof up and forget
		// the type part, looking at you syft....
		//
		// If it is a recognised namespace then we can correct the PURL,
		// otherwise this is an invalid PURL so return the empty purl
		// struct.
		if isValidPurlType(purlParts[0]) {
			output.typ = purlParts[0]
		} else {
			// Fix known cases otherwise just exclude the type from
			// the purl or error maybe, haven't decided
			switch purlParts[0] {
			case "alpine":
				output.typ = "apk"
				output.namespace = "alpine"
			default:
				return newPurl()
			}
		}
	} else {
		// No other length is valid
		return newPurl()
	}

	// Version is optional, so no need to check if we found the separator
	name, version, _ := strings.Cut(purlParts[len(purlParts)-1], "@")
	output.name = name
	output.version = version

	return output
}

func mergePurl(purlA, purlB purl) purl {
	if purlA.scheme == "" {
		purlA.scheme = purlB.scheme
	}

	if purlA.typ == "" {
		purlA.typ = purlB.typ
	}

	if purlA.namespace == "" {
		purlA.namespace = purlB.namespace
	}

	if purlA.name == "" {
		purlA.name = purlB.name
	}

	if purlA.version == "" {
		purlA.version = purlB.version
	}

	for key := range purlB.qualifers {
		if purlA.qualifers.Has(key) {
			continue
		}
		purlA.qualifers.Set(key, purlB.qualifers.Get(key))
	}

	if purlA.subpath == "" {
		purlA.subpath = purlB.subpath
	}

	return purlA
}

func mergePurlStrings(purlAStr, purlBStr string) string {
	purlA := purlStringToStruct(purlAStr)
	purlB := purlStringToStruct(purlBStr)
	purlC := mergePurl(purlA, purlB)
	return purlC.String()
}
