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

type Registry struct {
	SkipVerifyTLS bool   `yaml:"skip-verify-tls" json:"skip-verify-tls" mapstructure:"skip-verify-tls"`
	UseHTTP       bool   `yaml:"use-http" json:"use-http" mapstructure:"use-http"`
	Auths         []Auth `yaml:"auths" json:"auths" mapstructure:"auths"`
}

type Auth struct {
	Authority string `yaml:"authority" json:"authority" mapstructure:"authority"`
	Username  string `yaml:"-" json:"-" mapstructure:"username"`
	Password  string `yaml:"-" json:"-" mapstructure:"password"`
	Token     string `yaml:"-" json:"-" mapstructure:"token"`
}
