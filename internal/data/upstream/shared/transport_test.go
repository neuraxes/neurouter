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

package shared

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/neuraxes/neurouter/internal/util"
)

type recordingLogExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *recordingLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return nil
}

func (e *recordingLogExporter) Records() []sdklog.Record {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]sdklog.Record(nil), e.records...)
}

func (*recordingLogExporter) Shutdown(context.Context) error   { return nil }
func (*recordingLogExporter) ForceFlush(context.Context) error { return nil }

func TestInstrumentedHTTPClient(t *testing.T) {
	Convey("Given an upstream HTTP request with a parent span", t, func() {
		var traceparent string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			traceparent = req.Header.Get("traceparent")
			_, _ = io.WriteString(w, "response")
		}))
		defer server.Close()

		spanRecorder := tracetest.NewSpanRecorder()
		tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
		logExporter := new(recordingLogExporter)
		loggerProvider := sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewSimpleProcessor(logExporter)),
		)
		defer func() { _ = loggerProvider.Shutdown(context.Background()) }()
		ctx, parent := tracerProvider.Tracer("test").Start(context.Background(), "gen_ai")
		client := NewRecordingClientFromLoggerProvider(loggerProvider, tracerProvider, "test")
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, strings.NewReader("request"))
		So(err, ShouldBeNil)

		resp, err := client.Do(req)
		So(err, ShouldBeNil)
		_, err = io.Copy(io.Discard, resp.Body)
		So(err, ShouldBeNil)
		So(resp.Body.Close(), ShouldBeNil)
		parent.End()

		So(traceparent, ShouldNotBeBlank)
		spans := spanRecorder.Ended()
		So(spans, ShouldHaveLength, 2)
		var clientSpanID trace.SpanID
		for _, span := range spans {
			if span.SpanKind() == trace.SpanKindClient {
				clientSpanID = span.SpanContext().SpanID()
				So(span.Parent().SpanID(), ShouldEqual, parent.SpanContext().SpanID())
			}
		}
		So(clientSpanID.IsValid(), ShouldBeTrue)

		records := logExporter.Records()
		So(records, ShouldHaveLength, 2)
		for _, record := range records {
			So(record.TraceID(), ShouldEqual, parent.SpanContext().TraceID())
			So(record.SpanID(), ShouldEqual, clientSpanID)
			So(record.SpanID(), ShouldNotEqual, parent.SpanContext().SpanID())
			So(record.EventName(), ShouldBeIn,
				util.EventUpstreamReqSent,
				util.EventUpstreamRespReceived,
			)
		}
	})
}
