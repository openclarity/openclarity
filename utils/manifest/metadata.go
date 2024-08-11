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

//nolint:wrapcheck
package manifest

import (
	"bytes"
	"encoding/json"
	"io"
)

// Metadata contains information about the package bundled in Bundle.FS. It partially implements the format defined
// by the Cloud Native Application Bundle specification, however it is not limited to it.
type Metadata struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	License     string       `json:"license,omitempty"`
	Maintainers []Maintainer `json:"maintainers,omitempty"`

	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Maintainer contains the contact information for a maintainer defined in Metadata.
type Maintainer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

func NewMetadataFromRawBytes(b []byte) (*Metadata, error) {
	return NewMetadataFromStream(bytes.NewBuffer(b))
}

func NewMetadataFromStream(r io.Reader) (*Metadata, error) {
	metadata := &Metadata{}

	if err := json.NewDecoder(r).Decode(metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
