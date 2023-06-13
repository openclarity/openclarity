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

package common

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

type AddAction struct {
	object TestObject
}

type GetAction struct {
	object TestObject
}

type HasAction struct {
	object TestObject
	has    bool
}

type DoneAction struct {
	object TestObject
}

type TestObject struct {
	ID string
}

func (o TestObject) ToFields() logrus.Fields {
	return logrus.Fields{
		"ID": o.ID,
	}
}

func (o TestObject) String() string {
	return o.ID
}

func (o TestObject) Hash() string {
	return o.ID
}

// nolint:cyclop
func TestQueue(t *testing.T) {
	tests := []struct {
		name    string
		actions []interface{}
	}{
		{
			name: "sanity",
			actions: []interface{}{
				AddAction{TestObject{ID: "foo"}},
				AddAction{TestObject{ID: "bar"}},
				AddAction{TestObject{ID: "baz"}},

				GetAction{TestObject{ID: "foo"}},
				DoneAction{TestObject{ID: "foo"}},
				GetAction{TestObject{ID: "bar"}},
				DoneAction{TestObject{ID: "bar"}},
				GetAction{TestObject{ID: "baz"}},
				DoneAction{TestObject{ID: "baz"}},
			},
		},
		{
			name: "readd the same item before being processed",
			actions: []interface{}{
				AddAction{TestObject{ID: "foo"}},
				AddAction{TestObject{ID: "bar"}},
				AddAction{TestObject{ID: "foo"}},

				GetAction{TestObject{ID: "foo"}},
				AddAction{TestObject{ID: "foo"}},
				DoneAction{TestObject{ID: "foo"}},

				GetAction{TestObject{ID: "bar"}},
				DoneAction{TestObject{ID: "bar"}},
			},
		},
		{
			name: "readd the same item after being processed",
			actions: []interface{}{
				AddAction{TestObject{ID: "foo"}},
				AddAction{TestObject{ID: "bar"}},

				GetAction{TestObject{ID: "foo"}},
				DoneAction{TestObject{ID: "foo"}},

				AddAction{TestObject{ID: "foo"}},

				GetAction{TestObject{ID: "bar"}},
				DoneAction{TestObject{ID: "bar"}},

				GetAction{TestObject{ID: "foo"}},
				DoneAction{TestObject{ID: "foo"}},
			},
		},
		{
			name: "check queue has item",
			actions: []interface{}{
				AddAction{TestObject{ID: "foo"}},
				AddAction{TestObject{ID: "bar"}},
				HasAction{TestObject{ID: "foo"}, true},
				HasAction{TestObject{ID: "bar"}, true},
				HasAction{TestObject{ID: "baz"}, false},
				GetAction{TestObject{ID: "foo"}},
				DoneAction{TestObject{ID: "foo"}},
				GetAction{TestObject{ID: "bar"}},
				DoneAction{TestObject{ID: "bar"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue[TestObject]()

			for _, action := range tt.actions {
				switch ack := action.(type) {
				case AddAction:
					q.Enqueue(ack.object)
				case GetAction:
					jtem, err := q.Dequeue(context.TODO())
					if err != nil {
						t.Fatalf("Unexpected error: %v", err)
					}

					if diff := cmp.Diff(ack.object, jtem); diff != "" {
						t.Fatalf("Dequeue() mismatch (-want, +got):\n%s", diff)
					}
				case DoneAction:
					q.Done(ack.object)
				case HasAction:
					h := q.Has(ack.object)
					if h != ack.has {
						t.Fatalf("Has() mismatch, expected %v got %v", ack.has, h)
					}
				}
				t.Logf("Contents after action: %v", q.queue)
			}

			if q.Length() != 0 {
				t.Fatalf("Expected queue to be empty at the end of the test, Length %d, Contents: %v", q.Length(), q.queue)
			}
		})
	}
}
