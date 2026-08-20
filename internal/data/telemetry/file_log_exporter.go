// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// FileLogExporter dumps each event body to its own file under BaseDir. It is a
// debugging aid meant to be wired into NewLoggerProvider by hand while chasing a
// payload issue, so it is intentionally not part of any configured pipeline.
type FileLogExporter struct {
	BaseDir string
}

func NewFileLogExporter(baseDir string) sdklog.Exporter {
	return &FileLogExporter{BaseDir: baseDir}
}

func (e *FileLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	if e.BaseDir == "" {
		return nil
	}
	err := os.MkdirAll(e.BaseDir, 0o755)
	if err != nil {
		return err
	}
	for _, r := range records {
		filename := fmt.Sprintf("%d-%s-%s", time.Now().UnixMilli(), r.TraceID().String(), r.EventName())
		body := []byte(r.Body().AsString())

		var raw any
		if err := json.Unmarshal(body, &raw); err != nil {
			// For non-JSON events
			err = os.WriteFile(filepath.Join(e.BaseDir, filename+".txt"), body, 0o644)
			if err != nil {
				return err
			}
			continue
		}

		pretty, err := json.MarshalIndent(raw, "", "    ")
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(e.BaseDir, filename+".json"), append(pretty, '\n'), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *FileLogExporter) Shutdown(context.Context) error   { return nil }
func (e *FileLogExporter) ForceFlush(context.Context) error { return nil }
