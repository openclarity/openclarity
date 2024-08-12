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
	"io/fs"
	"path/filepath"
)

const DefaultMetadataFile = "bundle.json"

var DefaultMatcher = filepath.Match

type Bundle struct {
	// Prefix contains the prefix for a subtree of the FS.
	Prefix string
	// Matcher is an implementation of PatternMatcher interface used by Bundle.Glob().
	Matcher PatternMatcher
	// Metadata includes optional metadata for content bundled in Bundle.FS.
	Metadata

	FS
}

func (b *Bundle) prefixedPath(path string) string {
	return filepath.ToSlash(filepath.Join(b.Prefix, path))
}

// Open opens the named file. Implements fs.FS interface.
func (b *Bundle) Open(name string) (fs.File, error) {
	return b.FS.Open(b.prefixedPath(name))
}

// ReadDir reads the named directory and returns a list of directory entries sorted by filename.
// Implements fs.ReadDirFS interface.
func (b *Bundle) ReadDir(name string) ([]fs.DirEntry, error) {
	return b.FS.ReadDir(b.prefixedPath(name))
}

// ReadFile reads the named file and returns its contents. Implements fs.ReadFileFS interface.
func (b *Bundle) ReadFile(name string) ([]byte, error) {
	return b.FS.ReadFile(b.prefixedPath(name))
}

func (b *Bundle) WalkDir(root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(b, root, fn)
}

func (b *Bundle) Glob(pattern string) ([]string, error) {
	if b.Matcher == nil {
		b.Matcher = DefaultMatcher
	}

	matches := make([]string, 0)
	matchFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ok, err := b.Matcher(pattern, path)
		if err != nil {
			return err
		}
		if ok {
			matches = append(matches, path)
		}

		return nil
	}

	if err := b.WalkDir(".", matchFn); err != nil {
		return nil, err
	}

	return matches, nil
}

func (b *Bundle) Stat(path string) (fs.FileInfo, error) {
	o, err := b.Open(path)
	if err != nil {
		return nil, err
	}
	defer o.Close()

	return o.Stat()
}

func (b *Bundle) Sub(dir string) (*Bundle, error) {
	_, err := b.Stat(dir)
	if err != nil {
		return nil, err
	}

	return &Bundle{
		FS:     b.FS,
		Prefix: b.prefixedPath(dir),
	}, nil
}

// BundleFn defines transformer function for Bundle.
type BundleFn func(*Bundle) error

func applyBundleWithOpts(b *Bundle, opts ...BundleFn) error {
	if b == nil {
		return nil
	}

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return err
		}
	}

	return nil
}

// New returns a new Bundle object which encapsulates the provided FS and filepath.Match as the default pattern matcher.
// List of BundleFn are applied if provided.
func New(fs FS, opts ...BundleFn) (*Bundle, error) {
	b := &Bundle{
		FS:      fs,
		Matcher: DefaultMatcher,
	}
	if err := applyBundleWithOpts(b, opts...); err != nil {
		return nil, err
	}

	return b, nil
}

// WithPrefix set the filesystem prefix for fs.FS which alters the path used for lookup operations.
// It comes handy in cases where only a subtree of the FS is expected to be access without exposing the whole filesystem structure.
func WithPrefix(prefix string) BundleFn {
	return func(b *Bundle) error {
		if b != nil {
			b.Prefix = prefix
		}

		return nil
	}
}

// PatternMatcher custom file path pattern matcher for use-cases where the default filepath.Match is not sufficient.
type PatternMatcher func(pattern, path string) (bool, error)

// WithMatcher sets the PatternMatcher to be used by Bundle.Glob().
func WithMatcher(m PatternMatcher) BundleFn {
	return func(b *Bundle) error {
		if b != nil {
			b.Matcher = m
		}

		return nil
	}
}

// WithMetadata uses meta to populate Bundle.Metadata.
func WithMetadata(meta Metadata) BundleFn {
	return func(b *Bundle) error {
		if b != nil {
			b.Metadata = meta
		}

		return nil
	}
}

// WithDefaultMetadataFile lookup for bundle.json at path in Bundle.FS to populate Bundle.Metadata.
func WithMetadataFile(path string) BundleFn {
	return func(b *Bundle) error {
		f, err := b.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		metadata, err := NewMetadataFromStream(f)
		if err != nil {
			return err
		}

		b.Metadata = *metadata

		return nil
	}
}

// WithDefaultMetadataFile lookup for bundle.json in root of Bundle.FS to populate Bundle.Metadata.
func WithDefaultMetadataFile() BundleFn {
	return WithMetadataFile(DefaultMetadataFile)
}

// MetadataResolver returns a Metadata object by reading the content of the Bundle. It is meant to be used
// to retrieve metadata for Bundle in formats other than bundle.json supported by default.
type MetadataResolver func(*Bundle) (*Metadata, error)

// WithMetadataResolver invokes the MetadataResolver to populate the Bundle.Metadata.
func WithMetadataResolver(resolver MetadataResolver) BundleFn {
	return func(b *Bundle) error {
		metadata, err := resolver(b)
		if err != nil {
			return err
		}

		b.Metadata = *metadata

		return nil
	}
}
