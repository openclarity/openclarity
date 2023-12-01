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

package fanout

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"golang.org/x/exp/rand"
)

func TestFanOut(t *testing.T) {
	tests := []struct {
		Name         string
		NumOfReaders int
	}{
		{
			Name:         "FanOut to multiple readers",
			NumOfReaders: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			source := make([]byte, 32*1024)
			_, err := rand.Read(source)
			g.Expect(err).ShouldNot(HaveOccurred())

			// Create source io.Reader
			src := bytes.NewReader(source)

			// Create list of buffers to store data read by consumers
			buffers := []*bytes.Buffer{}

			// Create consumers list
			consumers := []func(io.Reader) error{}
			for i := 0; i < test.NumOfReaders; i++ {
				// Create Buffer
				b := bytes.NewBuffer(nil)
				buffers = append(buffers, b)

				// Create FanOut consumer
				consumers = append(consumers, func(r io.Reader) error {
					if _, err := io.Copy(b, r); err != nil {
						return fmt.Errorf("failed to copy data: %w", err)
					}

					return nil
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Run FanOut
			err = FanOut(ctx, src, consumers)
			g.Expect(err).ShouldNot(HaveOccurred())

			// Check content read by consumers
			for _, buf := range buffers {
				g.Expect(buf.Bytes()).Should(BeEquivalentTo(source))
			}
		})
	}
}
