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
	"errors"
	"io"
)

type pipe struct {
	In  *io.PipeWriter
	Out *io.PipeReader
}

type fanOutWriter struct {
	readers []pipe
	writer  io.Writer
}

func (r *fanOutWriter) In() io.Writer {
	return r.writer
}

func (r *fanOutWriter) Out() []io.Reader {
	readers := make([]io.Reader, len(r.readers))
	for i, p := range r.readers {
		readers[i] = p.Out
	}

	return readers
}

func (r *fanOutWriter) Close() error {
	var errs []error

	for _, pipe := range r.readers {
		if err := pipe.In.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func newFanOutWriter(readers int) *fanOutWriter {
	pipes := make([]pipe, readers)
	writers := make([]io.Writer, readers)

	for i := range readers {
		out, in := io.Pipe()
		pipes[i] = pipe{
			In:  in,
			Out: out,
		}
		writers[i] = in
	}

	return &fanOutWriter{
		readers: pipes,
		writer:  io.MultiWriter(writers...),
	}
}
