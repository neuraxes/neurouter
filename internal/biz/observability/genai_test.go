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

package observability

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"
	"go.opentelemetry.io/otel/trace"
)

type instrumenterTestClock struct {
	times []time.Time
	next  int
}

func (c *instrumenterTestClock) now() time.Time {
	value := c.times[c.next]
	c.next++
	return value
}

type collectedGenAIMetrics struct {
	floats map[string][]metricdata.HistogramDataPoint[float64]
	ints   map[string][]metricdata.HistogramDataPoint[int64]
}

func newTestInstrumenter() (*GenAIInstrumenter, *tracetest.SpanRecorder, *sdkmetric.ManualReader) {
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	instrumenter, err := NewGenAIInstrumenter(tracerProvider, meterProvider)
	if err != nil {
		panic(err)
	}
	return instrumenter, spanRecorder, reader
}

func collectGenAIMetrics(reader *sdkmetric.ManualReader) collectedGenAIMetrics {
	var resourceMetrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &resourceMetrics); err != nil {
		panic(err)
	}
	result := collectedGenAIMetrics{
		floats: make(map[string][]metricdata.HistogramDataPoint[float64]),
		ints:   make(map[string][]metricdata.HistogramDataPoint[int64]),
	}
	for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
		for _, m := range scopeMetrics.Metrics {
			switch data := m.Data.(type) {
			case metricdata.Histogram[float64]:
				result.floats[m.Name] = data.DataPoints
			case metricdata.Histogram[int64]:
				result.ints[m.Name] = data.DataPoints
			}
		}
	}
	return result
}

func valueFor(attrs []attribute.KeyValue, key attribute.Key) (attribute.Value, bool) {
	set := attribute.NewSet(attrs...)
	return set.Value(key)
}

func TestGenAIInstrumenter(t *testing.T) {
	Convey("Given a streaming chat invocation", t, func() {
		instrumenter, spanRecorder, reader := newTestInstrumenter()
		start := time.Unix(100, 0)
		clock := &instrumenterTestClock{times: []time.Time{
			start,
			start.Add(500 * time.Millisecond),
			start.Add(2 * time.Second),
		}}
		instrumenter.now = clock.now
		target := GenAITarget{
			Provider:       genaiconv.ProviderNameOpenAI,
			Upstream:       "primary",
			RequestedModel: "smart-model",
			Model:          "gpt-router",
			UpstreamModel:  "gpt-4.1",
			ServerAddress:  "api.openai.com",
			ServerPort:     443,
		}

		ctx, invocation := instrumenter.Start(
			context.Background(),
			genaiconv.OperationNameChat,
			target,
			semconv.GenAIRequestStream(true),
			semconv.GenAIRequestTemperature(0.7),
		)
		So(trace.SpanContextFromContext(ctx).IsValid(), ShouldBeTrue)

		invocation.FirstChunk()
		invocation.FirstChunk()
		invocation.End(GenAIResult{
			ResponseID:    "message-1",
			ResponseModel: "gpt-4.1-2026-01-01",
			FinishReasons: []string{"stop"},
			Usage: &GenAITokenUsage{
				Input:       100,
				Output:      50,
				CachedInput: 10,
				Reasoning:   20,
			},
		}, nil)
		invocation.End(GenAIResult{}, errors.New("ignored duplicate end"))

		Convey("it emits one content-free client span", func() {
			spans := spanRecorder.Ended()
			So(spans, ShouldHaveLength, 1)
			span := spans[0]
			So(span.Name(), ShouldEqual, "chat gpt-4.1")
			So(span.SpanKind(), ShouldEqual, trace.SpanKindClient)
			So(span.Status().Code, ShouldEqual, codes.Unset)
			So(span.EndTime().Sub(span.StartTime()), ShouldEqual, 2*time.Second)

			attrs := span.Attributes()
			value, ok := valueFor(attrs, semconv.GenAIOperationNameKey)
			So(ok, ShouldBeTrue)
			So(value.AsString(), ShouldEqual, "chat")
			value, ok = valueFor(attrs, semconv.GenAIProviderNameKey)
			So(ok, ShouldBeTrue)
			So(value.AsString(), ShouldEqual, "openai")
			value, ok = valueFor(attrs, semconv.GenAIRequestModelKey)
			So(ok, ShouldBeTrue)
			So(value.AsString(), ShouldEqual, "gpt-4.1")
			value, ok = valueFor(attrs, semconv.GenAIResponseIDKey)
			So(ok, ShouldBeTrue)
			So(value.AsString(), ShouldEqual, "message-1")
			value, ok = valueFor(attrs, semconv.GenAIUsageCacheReadInputTokensKey)
			So(ok, ShouldBeTrue)
			So(value.AsInt64(), ShouldEqual, 10)

			_, hasInput := valueFor(attrs, attribute.Key("gen_ai.input.messages"))
			_, hasOutput := valueFor(attrs, attribute.Key("gen_ai.output.messages"))
			_, hasTools := valueFor(attrs, attribute.Key("gen_ai.tool.definitions"))
			So(hasInput, ShouldBeFalse)
			So(hasOutput, ShouldBeFalse)
			So(hasTools, ShouldBeFalse)
		})

		Convey("it emits standard GenAI histograms once", func() {
			metrics := collectGenAIMetrics(reader)
			durations := metrics.floats["gen_ai.client.operation.duration"]
			So(durations, ShouldHaveLength, 1)
			So(durations[0].Count, ShouldEqual, 1)
			So(durations[0].Sum, ShouldEqual, 2.0)

			firstChunks := metrics.floats["gen_ai.client.operation.time_to_first_chunk"]
			So(firstChunks, ShouldHaveLength, 1)
			So(firstChunks[0].Count, ShouldEqual, 1)
			So(firstChunks[0].Sum, ShouldEqual, 0.5)

			tokens := metrics.ints["gen_ai.client.token.usage"]
			So(tokens, ShouldHaveLength, 2)
			for _, point := range tokens {
				tokenType, ok := point.Attributes.Value(attribute.Key("gen_ai.token.type"))
				So(ok, ShouldBeTrue)
				switch tokenType.AsString() {
				case "input":
					So(point.Sum, ShouldEqual, 100)
				case "output":
					So(point.Sum, ShouldEqual, 50)
				default:
					So(tokenType.AsString(), ShouldBeIn, "input", "output")
				}
			}

			So(metrics.ints["neurouter_input_tokens_total"], ShouldBeEmpty)
			So(metrics.ints["neurouter_requests_total"], ShouldBeEmpty)
		})
	})

	Convey("Given a failed invocation", t, func() {
		instrumenter, spanRecorder, reader := newTestInstrumenter()
		start := time.Unix(200, 0)
		clock := &instrumenterTestClock{times: []time.Time{start, start.Add(time.Second)}}
		instrumenter.now = clock.now
		_, invocation := instrumenter.Start(
			context.Background(),
			genaiconv.OperationNameEmbeddings,
			GenAITarget{
				Provider:      genaiconv.ProviderNameGCPGemini,
				UpstreamModel: "text-embedding-004",
			},
		)
		invocation.End(GenAIResult{}, context.DeadlineExceeded)

		Convey("it records a low-cardinality error without an exception event", func() {
			spans := spanRecorder.Ended()
			So(spans, ShouldHaveLength, 1)
			span := spans[0]
			So(span.Status().Code, ShouldEqual, codes.Error)
			So(span.Events(), ShouldBeEmpty)
			value, ok := valueFor(span.Attributes(), semconv.ErrorTypeKey)
			So(ok, ShouldBeTrue)
			So(value.AsString(), ShouldEqual, "deadline_exceeded")

			metrics := collectGenAIMetrics(reader)
			points := metrics.floats["gen_ai.client.operation.duration"]
			So(points, ShouldHaveLength, 1)
			value, ok = points[0].Attributes.Value(semconv.ErrorTypeKey)
			So(ok, ShouldBeTrue)
			So(value.AsString(), ShouldEqual, "deadline_exceeded")
			So(metrics.ints["gen_ai.client.token.usage"], ShouldBeEmpty)
		})
	})
}

func TestNewGenAIInstrumenter(t *testing.T) {
	Convey("It accepts nil providers by using no-op OTel APIs", t, func() {
		instrumenter, err := NewGenAIInstrumenter(nil, nil)
		So(err, ShouldBeNil)
		So(instrumenter, ShouldNotBeNil)
	})
}
