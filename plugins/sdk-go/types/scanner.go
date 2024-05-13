// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

// TODO(ramizpolic): Document usage and execution flow

// Scanner defines the interface that the plugin scanner developer should
// implement. You should not run Scanner on its own but via plugin.Run.
type Scanner interface {
	Metadata() *Metadata
	GetStatus() *Status
	SetStatus(status *Status)
	Start(config Config)
	Stop(stop Stop)
}
