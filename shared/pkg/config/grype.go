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

package config

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	ScannerGrypeMode = "SCANNER_GRYPE_MODE"
)

type Mode string

const (
	ModeLocal  Mode = "LOCAL"
	ModeRemote Mode = "REMOTE"
)

type GrypeConfig struct {
	LocalGrypeConfig  `yaml:"local_grype_config" mapstructure:"local_grype_config"`
	RemoteGrypeConfig `yaml:"remote_grype_config" mapstructure:"remote_grype_config"`
	Mode              Mode `yaml:"mode" mapstructure:"mode"`
}

func LoadGrypeConfig() GrypeConfig {
	setGrypeScannerConfigDefaults()
	return GrypeConfig{
		LocalGrypeConfig:  loadLocalGrypeConfig(),
		RemoteGrypeConfig: loadRemoteGrypeConfig(),
		Mode:              getGrypeMode(viper.GetString(ScannerGrypeMode)),
	}
}

func getGrypeMode(mode string) Mode {
	switch Mode(strings.ToUpper(mode)) {
	case ModeLocal:
		return ModeLocal
	case ModeRemote:
		return ModeRemote
	default:
		log.Fatalf("Unsupported grype mode %q. Supported values (%s, %s)", mode, ModeLocal, ModeRemote)
	}

	return ""
}

func setGrypeScannerConfigDefaults() {
	viper.SetDefault(ScannerGrypeMode, string(ModeLocal))
}
