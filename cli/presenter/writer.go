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

package presenter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

type Writer interface {
	Write([]byte, string) error
}

type ConsoleWriter struct {
	Output io.Writer
}

func (w *ConsoleWriter) Write(b []byte, prefix string) error {
	buf := bytes.NewBufferString(fmt.Sprintf("%s results:\n", prefix))
	buf.Write(b)
	buf.WriteString("\n=================================================\n")

	_, err := buf.WriteTo(w.Output)
	if err != nil {
		return fmt.Errorf("failed to write to terminal: %w", err)
	}
	return nil
}

type FileWriter struct {
	Path string
}

func (w *FileWriter) Write(b []byte, filename string) error {
	filePath := path.Join(w.Path, filename)
	log.Infof("Writing results to %v...", filePath)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666) // nolint:gomnd,gofumpt
	if err != nil {
		return fmt.Errorf("failed open file %s: %w", filePath, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("failed to close file %v: %v", filePath, err)
		}
	}(file)

	_, err = file.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write bytes to file %s: %w", filePath, err)
	}
	return nil
}
