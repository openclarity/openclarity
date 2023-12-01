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

package types

import (
	"fmt"

	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
)

var _ imageRef = &ImageRef{}

type imageRef interface {
	reference.NamedTagged
	reference.Canonical
}

type ImageRef struct {
	name   string
	domain string
	path   string
	tag    string
	digest digest.Digest
}

func (i *ImageRef) Name() string {
	return i.name
}

func (i *ImageRef) Domain() string {
	return i.domain
}

func (i *ImageRef) Path() string {
	return i.path
}

func (i *ImageRef) Tag() string {
	return i.tag
}

func (i *ImageRef) Digest() digest.Digest {
	return i.digest
}

func (i *ImageRef) String() string {
	switch {
	case i.tag != "" && i.digest != "":
		return i.Name() + ":" + i.tag + "@" + i.digest.String()
	case i.tag == "" && i.digest != "":
		return i.Name() + "@" + i.digest.String()
	case i.tag != "":
		return i.Name() + ":" + i.tag
	default:
		return i.Name()
	}
}

func (i *ImageRef) UnmarshalText(text []byte) error {
	ref, err := reference.ParseAnyReference(string(text))
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	var name, domain, path string
	if named, ok := ref.(reference.Named); ok {
		name = named.Name()
		domain = reference.Domain(named)
		path = reference.Path(named)
	} else {
		return fmt.Errorf("failed to parse image name: %s", text)
	}

	var tag string
	if tagged, ok := ref.(reference.NamedTagged); ok {
		tag = tagged.Tag()
	}

	var imageDigest digest.Digest
	if digested, ok := ref.(reference.Digested); ok {
		imageDigest = digested.Digest()
	}

	if tag == "" && imageDigest == "" {
		tag = "latest"
	}

	*i = ImageRef{
		name:   name,
		domain: domain,
		path:   path,
		tag:    tag,
		digest: imageDigest,
	}

	return nil
}

func NewImageRef(name, domain, path, tag, imageDigest string) ImageRef {
	return ImageRef{name, domain, path, tag, digest.Digest(imageDigest)}
}

func NewImageRefFrom(s string) (ImageRef, error) {
	ref := &ImageRef{}
	err := ref.UnmarshalText([]byte(s))

	return *ref, err
}
