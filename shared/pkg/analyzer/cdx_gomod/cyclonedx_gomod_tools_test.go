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

package cdx_gomod // nolint:revive,stylecheck

// nolint:gci
import (
	"testing"

	"gotest.tools/assert"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

func Test_assertComponentLicenses(t *testing.T) {
	type args struct {
		c *cdx.Component
	}
	tests := []struct {
		name string
		args args
		want *cdx.Component
	}{
		{
			name: "Component is nil",
			want: nil,
		},
		{
			name: "Evidence contains only Licenses",
			args: args{
				c: &cdx.Component{
					Evidence: &cdx.Evidence{
						Licenses: &cdx.Licenses{
							{
								License: &cdx.License{
									ID: "MIT",
								},
							},
						},
					},
				},
			},
			want: &cdx.Component{
				Licenses: &cdx.Licenses{
					{
						License: &cdx.License{
							ID: "MIT",
						},
					},
				},
			},
		},
		{
			name: "Evidence contains Licenses and Copyright",
			args: args{
				c: &cdx.Component{
					Evidence: &cdx.Evidence{
						Licenses: &cdx.Licenses{
							{
								License: &cdx.License{
									ID: "MIT",
								},
							},
						},
						Copyright: &[]cdx.Copyright{
							{
								Text: "Copyright",
							},
						},
					},
				},
			},
			want: &cdx.Component{
				Licenses: &cdx.Licenses{
					{
						License: &cdx.License{
							ID: "MIT",
						},
					},
				},
				Evidence: &cdx.Evidence{
					Copyright: &[]cdx.Copyright{
						{
							Text: "Copyright",
						},
					},
				},
			},
		},
		{
			name: "Component contains other component",
			args: args{
				c: &cdx.Component{
					Evidence: &cdx.Evidence{
						Licenses: &cdx.Licenses{
							{
								License: &cdx.License{
									ID: "MIT",
								},
							},
						},
					},
					Components: &[]cdx.Component{
						{
							Evidence: &cdx.Evidence{
								Licenses: &cdx.Licenses{
									{
										License: &cdx.License{
											ID: "MIT",
										},
									},
								},
							},
						},
					},
				},
			},
			want: &cdx.Component{
				Licenses: &cdx.Licenses{
					{
						License: &cdx.License{
							ID: "MIT",
						},
					},
				},
				Components: &[]cdx.Component{
					{
						Licenses: &cdx.Licenses{
							{
								License: &cdx.License{
									ID: "MIT",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertComponentLicenses(tt.args.c)
			assert.DeepEqual(t, tt.args.c, tt.want)
		})
	}
}

func Test_assertLicenses(t *testing.T) {
	type args struct {
		bom *cdx.BOM
	}
	tests := []struct {
		name string
		args args
		want *cdx.BOM
	}{
		{
			name: "BOM is nil",
			want: nil,
		},
		{
			name: "BOM contains Metadata",
			args: args{
				bom: &cdx.BOM{
					Metadata: &cdx.Metadata{
						Component: &cdx.Component{
							Evidence: &cdx.Evidence{
								Licenses: &cdx.Licenses{
									{
										License: &cdx.License{
											ID: "MIT",
										},
									},
								},
							},
						},
					},
				},
			},
			want: &cdx.BOM{
				Metadata: &cdx.Metadata{
					Component: &cdx.Component{
						Licenses: &cdx.Licenses{
							{
								License: &cdx.License{
									ID: "MIT",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BOM contains Components",
			args: args{
				bom: &cdx.BOM{
					Components: &[]cdx.Component{
						{
							Evidence: &cdx.Evidence{
								Licenses: &cdx.Licenses{
									{
										License: &cdx.License{
											ID: "MIT",
										},
									},
								},
							},
						},
					},
				},
			},
			want: &cdx.BOM{
				Components: &[]cdx.Component{
					{
						Licenses: &cdx.Licenses{
							{
								License: &cdx.License{
									ID: "MIT",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertLicenses(tt.args.bom)
			assert.DeepEqual(t, tt.args.bom, tt.want)
		})
	}
}
