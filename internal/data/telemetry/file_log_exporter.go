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

// Set to a directory path to enable writing event to JSON files for debugging.
const fileLogBaseDir = ""

type FileLogExporter struct{}

func NewFileLogExporter() sdklog.Exporter {
	return &FileLogExporter{}
}

func (e *FileLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	if fileLogBaseDir == "" {
		return nil
	}
	err := os.MkdirAll(fileLogBaseDir, 0o755)
	if err != nil {
		return err
	}
	for _, r := range records {
		filename := fmt.Sprintf("%d-%s-%s", time.Now().UnixMilli(), r.TraceID().String(), r.EventName())

		var raw any
		if err := json.Unmarshal(r.Body().AsBytes(), &raw); err != nil {
			// For non-JSON events
			err = os.WriteFile(filepath.Join(fileLogBaseDir, filename+".txt"), r.Body().AsBytes(), 0o644)
			if err != nil {
				return err
			}
			continue
		}

		pretty, err := json.MarshalIndent(raw, "", "    ")
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(fileLogBaseDir, filename+".json"), append(pretty, '\n'), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *FileLogExporter) Shutdown(context.Context) error   { return nil }
func (e *FileLogExporter) ForceFlush(context.Context) error { return nil }
