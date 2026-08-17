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
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// SlogExporter writes OTel log records to a slog logger.
type SlogExporter struct {
	log *slog.Logger
}

// NewSlogExporter creates an OTel exporter backed by the supplied slog logger.
func NewSlogExporter(logger *slog.Logger) sdklog.Exporter {
	return &SlogExporter{log: logger}
}

func (e *SlogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	for _, record := range records {
		body := record.Body().AsInterface()
		if data, ok := body.([]byte); ok {
			body = string(data)
		}

		attrs := []slog.Attr{
			slog.String("event", record.EventName()),
			slog.Any("body", body),
		}
		if severityText := record.SeverityText(); severityText != "" {
			attrs = append(attrs, slog.String("severity_text", severityText))
		}
		record.WalkAttributes(func(kv attribute.KeyValue) bool {
			attrs = append(attrs, slog.Any(string(kv.Key), kv.Value.AsInterface()))
			return true
		})

		recordContext := ctx
		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    record.TraceID(),
			SpanID:     record.SpanID(),
			TraceFlags: record.TraceFlags(),
		})
		if spanContext.IsValid() {
			recordContext = trace.ContextWithSpanContext(recordContext, spanContext)
		}

		e.log.LogAttrs(recordContext, severityLevel(record.Severity()), "telemetry event", attrs...)
	}
	return nil
}

func severityLevel(severity otellog.Severity) slog.Level {
	switch {
	case severity == otellog.SeverityUndefined:
		return slog.LevelInfo
	case severity < otellog.SeverityInfo:
		return slog.LevelDebug
	case severity < otellog.SeverityWarn:
		return slog.LevelInfo
	case severity < otellog.SeverityError:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func (e *SlogExporter) Shutdown(context.Context) error   { return nil }
func (e *SlogExporter) ForceFlush(context.Context) error { return nil }
