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

package containerrootfs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/openclarity/vmclarity/core/log"
)

type Rootfs interface {
	Dir() string
	Cleanup() error
}

func ToTempDirectory(ctx context.Context, src string) (Rootfs, error) {
	cache, exists := GetCacheFromContext(ctx)
	if exists && cache != nil {
		return cache.ToTempDirectory(ctx, src)
	}
	return toTempDirectory(ctx, src)
}

type tempDirRootfs struct {
	dir string
}

func (tdrf *tempDirRootfs) Dir() string {
	return tdrf.dir
}

func (tdrf *tempDirRootfs) Cleanup() error {
	if tdrf.dir == "" {
		return nil
	}

	err := os.RemoveAll(tdrf.dir)
	if err != nil {
		return fmt.Errorf("unable to remove temp directory: %w", err)
	}
	return nil
}

func toTempDirectory(ctx context.Context, src string) (Rootfs, error) {
	successful := false
	tdrf := &tempDirRootfs{}
	defer func() {
		// If we're successful then it is the responsibility of the caller
		// to defer the cleanup, if we error during this function then
		// we need to handle it.
		if !successful {
			if err := tdrf.Cleanup(); err != nil {
				log.GetLoggerFromContextOrDefault(ctx).WithError(err).Error("failed to clean up container rootfs")
			}
		}
	}()

	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp directory: %w", err)
	}
	tdrf.dir = tmpdir

	image, cleanup, err := GetImageWithCleanup(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("unable to get image: %w", err)
	}
	defer cleanup()

	err = ToDirectory(ctx, image, tmpdir)
	if err != nil {
		return nil, fmt.Errorf("unable to output squashed image to directory: %w", err)
	}

	successful = true
	return tdrf, nil
}

type cachedRootfs struct {
	complete chan struct{}
	err      error
	rootfs   Rootfs
}

func (crfs *cachedRootfs) Dir() string {
	return crfs.rootfs.Dir()
}

func (crfs *cachedRootfs) Cleanup() error {
	return nil
}

func (crfs *cachedRootfs) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("wait aborted: %w", ctx.Err())
	case <-crfs.complete:
		return crfs.err
	}
}

func (crfs *cachedRootfs) Done(rootfs Rootfs, err error) {
	crfs.err = err
	crfs.rootfs = rootfs
	close(crfs.complete)
}

func newCachedRootfs() *cachedRootfs {
	return &cachedRootfs{
		complete: make(chan struct{}),
		err:      nil,
	}
}

type Cache struct {
	mu sync.Mutex
	m  map[string]*cachedRootfs
}

func NewCache() *Cache {
	return &Cache{
		m: make(map[string]*cachedRootfs),
	}
}

func (cache *Cache) ToTempDirectory(ctx context.Context, src string) (Rootfs, error) {
	if cache == nil {
		return nil, errors.New("uninitialized cache")
	}

	cache.mu.Lock()
	entry, ok := cache.m[src]
	if ok {
		cache.mu.Unlock()
		err := entry.Wait(ctx)
		return entry, err
	}
	entry = newCachedRootfs()
	cache.m[src] = entry
	cache.mu.Unlock()

	rootfs, err := toTempDirectory(ctx, src)
	entry.Done(rootfs, err)
	return entry, entry.err
}

func (cache *Cache) CleanupAll() error {
	if cache == nil {
		return errors.New("uninitialized cache")
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()

	errs := make([]error, 0, len(cache.m))
	for key, entry := range cache.m {
		if entry.rootfs != nil {
			errs = append(errs, entry.rootfs.Cleanup())
		}
		delete(cache.m, key)
	}
	return errors.Join(errs...)
}

type CacheContextKeyType string

const CacheContextKey CacheContextKeyType = "VMClarityContainerRootfsCacheKey"

func GetCacheFromContext(ctx context.Context) (*Cache, bool) {
	cache, ok := ctx.Value(CacheContextKey).(*Cache)
	return cache, ok
}

func SetCacheForContext(ctx context.Context, l *Cache) context.Context {
	return context.WithValue(ctx, CacheContextKey, l)
}
