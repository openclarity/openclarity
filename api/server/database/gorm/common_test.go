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

package gorm

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/openclarity/vmclarity/core/to"
)

type patchObjectTestNestedObject struct {
	NestedString string `json:"nestedString,omitempty"`
}

type patchObjectTestObject struct {
	TestString         *string                        `json:"testString,omitempty"`
	TestInt            *int                           `json:"testInt,omitempty"`
	TestBool           *bool                          `json:"testBool,omitempty"`
	TestNestedObject   *patchObjectTestNestedObject   `json:"testNestedObject,omitempty"`
	TestArrayPrimitive *[]string                      `json:"testArrayPrimitive,omitempty"`
	TestArrayObject    *[]patchObjectTestNestedObject `json:"testArrayObject,omitempty"`
}

func Test_patchObject(t *testing.T) {
	var nilArray []string

	tests := []struct {
		name     string
		existing patchObjectTestObject
		patch    patchObjectTestObject
		want     patchObjectTestObject
		wantErr  bool
	}{
		{
			name:     "patch over unset",
			existing: patchObjectTestObject{},
			patch: patchObjectTestObject{
				TestString: to.Ptr("foo"),
				TestInt:    to.Ptr(1),
				TestBool:   to.Ptr(false),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedfoo",
				},
				TestArrayPrimitive: to.Ptr([]string{"arrayfoo"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedfoo",
					},
				}),
			},
			want: patchObjectTestObject{
				TestString: to.Ptr("foo"),
				TestInt:    to.Ptr(1),
				TestBool:   to.Ptr(false),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedfoo",
				},
				TestArrayPrimitive: to.Ptr([]string{"arrayfoo"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedfoo",
					},
				}),
			},
			wantErr: false,
		},
		{
			name: "patch over already set",
			existing: patchObjectTestObject{
				TestString: to.Ptr("foo"),
				TestInt:    to.Ptr(1),
				TestBool:   to.Ptr(false),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedfoo",
				},
				TestArrayPrimitive: to.Ptr([]string{"arrayfoo"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedfoo",
					},
				}),
			},
			patch: patchObjectTestObject{
				TestString: to.Ptr("bar"),
				TestInt:    to.Ptr(2),
				TestBool:   to.Ptr(true),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedbar",
				},
				TestArrayPrimitive: to.Ptr([]string{"arraybar"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedbar",
					},
				}),
			},
			want: patchObjectTestObject{
				TestString: to.Ptr("bar"),
				TestInt:    to.Ptr(2),
				TestBool:   to.Ptr(true),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedbar",
				},
				TestArrayPrimitive: to.Ptr([]string{"arraybar"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedbar",
					},
				}),
			},
			wantErr: false,
		},
		{
			name: "patch Int, Bool, ArrayPrimitive and ArrayObject, String and NestedObject already set",
			existing: patchObjectTestObject{
				TestString: to.Ptr("foo"),
				TestInt:    to.Ptr(1),
				TestBool:   to.Ptr(false),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedfoo",
				},
				TestArrayPrimitive: to.Ptr([]string{"arrayfoo"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedfoo",
					},
				}),
			},
			patch: patchObjectTestObject{
				TestInt:            to.Ptr(2),
				TestBool:           to.Ptr(true),
				TestArrayPrimitive: to.Ptr([]string{"arraybar"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedbar",
					},
				}),
			},
			want: patchObjectTestObject{
				TestString: to.Ptr("foo"),
				TestInt:    to.Ptr(2),
				TestBool:   to.Ptr(true),
				TestNestedObject: &patchObjectTestNestedObject{
					NestedString: "nestedfoo",
				},
				TestArrayPrimitive: to.Ptr([]string{"arraybar"}),
				TestArrayObject: to.Ptr([]patchObjectTestNestedObject{
					{
						NestedString: "arraynestedbar",
					},
				}),
			},
			wantErr: false,
		},
		{
			name: "patch null to unset a field",
			existing: patchObjectTestObject{
				TestArrayPrimitive: to.Ptr([]string{"arraybar"}),
			},
			patch: patchObjectTestObject{
				TestArrayPrimitive: &nilArray,
			},
			want:    patchObjectTestObject{},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			existing, err := json.Marshal(test.existing)
			if err != nil {
				t.Fatalf("failed to marshal existing object: %v", err)
			}

			updatedBytes, err := patchObject(existing, test.patch)
			if err != nil {
				if !test.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				// Expected this error so return successful
				return
			}

			var updated patchObjectTestObject
			err = json.Unmarshal(updatedBytes, &updated)
			if err != nil {
				t.Fatalf("failed to unmarshal updated bytes: %v", err)
			}

			if diff := cmp.Diff(test.want, updated); diff != "" {
				t.Fatalf("patchObject mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
