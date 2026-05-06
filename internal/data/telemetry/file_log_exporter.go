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

		var raw any
		if err := json.Unmarshal(r.Body().AsBytes(), &raw); err != nil {
			// For non-JSON events
			err = os.WriteFile(filepath.Join(e.BaseDir, filename+".txt"), r.Body().AsBytes(), 0o644)
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
