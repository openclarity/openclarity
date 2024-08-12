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
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
)

type ContainerImages[T string | ImageRef] struct {
	APIServer         T `mapstructure:"apiserver_image"`
	Orchestrator      T `mapstructure:"orchestrator_image"`
	UI                T `mapstructure:"ui_image"`
	UIBackend         T `mapstructure:"uibackend_image"`
	Scanner           T `mapstructure:"scanner_image"`
	CRDiscoveryServer T `mapstructure:"cr_discovery_server_image"`
	PluginKics        T `mapstructure:"plugin_kics_image"`
}

func (t ContainerImages[T]) AsSlice() []T {
	return []T{
		t.APIServer,
		t.Orchestrator,
		t.UI,
		t.UIBackend,
		t.Scanner,
		t.CRDiscoveryServer,
		t.PluginKics,
	}
}

func (t ContainerImages[T]) AsStringSlice() ([]string, error) {
	if s, ok := any(t).(ContainerImages[string]); ok {
		return s.AsSlice(), nil
	}

	if s, ok := any(t).(ContainerImages[ImageRef]); ok {
		return []string{
			s.APIServer.String(),
			s.Orchestrator.String(),
			s.UI.String(),
			s.UIBackend.String(),
			s.Scanner.String(),
			s.CRDiscoveryServer.String(),
			s.PluginKics.String(),
		}, nil
	}

	return nil, errors.New("failed to convert to string slice")
}

func NewContainerImages[T string | ImageRef](images map[string]string) (*ContainerImages[T], error) {
	containerImages := &ContainerImages[T]{}

	decoderHooks := mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook:           decoderHooks,
		Result:               containerImages,
		IgnoreUntaggedFields: true,
		MatchName:            nil,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err = decoder.Decode(images); err != nil {
		return nil, fmt.Errorf("failed to unmarshal images: %w", err)
	}

	return containerImages, nil
}
