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

package chat

import (
	"context"
	"io"
	"iter"
	"log/slog"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"
	"go.opentelemetry.io/otel/trace"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/observability"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

type instrumentedChatRepo struct {
	sawSpan bool
}

func (r *instrumentedChatRepo) Chat(ctx context.Context, _ *entity.ChatRequest) (*entity.ChatResponse, error) {
	r.sawSpan = trace.SpanContextFromContext(ctx).IsValid()
	return nil, nil
}

func (r *instrumentedChatRepo) ChatStream(ctx context.Context, req *entity.ChatRequest) iter.Seq2[*entity.ChatEvent, error] {
	r.sawSpan = trace.SpanContextFromContext(ctx).IsValid()
	return func(yield func(*entity.ChatEvent, error) bool) {
		if !yield(&v1.ChatEvent{
			Id:    req.GetId(),
			Event: v1.NewMessageStartEvent("response-1", "gpt-4.1-2026-01-01"),
		}, nil) {
			return
		}
		yield(&v1.ChatEvent{
			Id:    req.GetId(),
			Usage: &v1.Usage{InputTokens: 12, OutputTokens: 7},
			Event: v1.NewMessageStopEvent(v1.ChatStatus_CHAT_STATUS_COMPLETED),
		}, nil)
	}
}

type instrumentedChatModel struct {
	repo          repository.ChatRepo
	recordedUsage *v1.Statistics
	closed        bool
}

func (m *instrumentedChatModel) ChatRepo() repository.ChatRepo {
	return m.repo
}

func (m *instrumentedChatModel) GenAITarget() observability.GenAITarget {
	return observability.GenAITarget{
		Provider:      genaiconv.ProviderNameOpenAI,
		Upstream:      "primary",
		Model:         "router-gpt",
		UpstreamModel: "gpt-4.1",
	}
}

func (m *instrumentedChatModel) RecordUsage(_ context.Context, stats *v1.Statistics) {
	m.recordedUsage = stats
}

func (m *instrumentedChatModel) Close() {
	m.closed = true
}

type instrumentedChatElector struct {
	model Model
}

func (e *instrumentedChatElector) ElectForChat(_ context.Context, req *v1.ChatRequest) (Model, error) {
	req.Model = "gpt-4.1"
	return e.model, nil
}

type collectingChatStreamServer struct {
	events []*entity.ChatEvent
}

func (s *collectingChatStreamServer) Send(event *entity.ChatEvent) error {
	s.events = append(s.events, event)
	return nil
}

func TestChatStreamObservability(t *testing.T) {
	Convey("Given a streaming chat request", t, func() {
		spanRecorder := tracetest.NewSpanRecorder()
		tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
		reader := sdkmetric.NewManualReader()
		meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		instrumenter, err := observability.NewGenAIInstrumenter(tracerProvider, meterProvider)
		So(err, ShouldBeNil)

		repo := &instrumentedChatRepo{}
		model := &instrumentedChatModel{repo: repo}
		uc := NewChatUseCase(
			&instrumentedChatElector{model: model},
			instrumenter,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		)
		server := &collectingChatStreamServer{}
		req := &v1.ChatRequest{
			Id:    "request-1",
			Model: "smart-model",
			Messages: []*v1.Message{{
				Role: v1.Role_ROLE_USER,
				Contents: []*v1.Content{{
					Content: v1.NewTextContent("sensitive prompt"),
				}},
			}},
		}

		err = uc.ChatStream(context.Background(), req, server)

		So(err, ShouldBeNil)
		So(repo.sawSpan, ShouldBeTrue)
		So(model.closed, ShouldBeTrue)
		So(model.recordedUsage.GetUsage().GetInputTokens(), ShouldEqual, 12)
		So(server.events, ShouldHaveLength, 2)

		Convey("the logical span contains routing and response metadata only", func() {
			spans := spanRecorder.Ended()
			So(spans, ShouldHaveLength, 1)
			span := spans[0]
			So(span.Name(), ShouldEqual, "chat gpt-4.1")
			attrs := attribute.NewSet(span.Attributes()...)
			requestedModel, ok := attrs.Value(attribute.Key("neurouter.request.model"))
			So(ok, ShouldBeTrue)
			So(requestedModel.AsString(), ShouldEqual, "smart-model")
			responseID, ok := attrs.Value(attribute.Key("gen_ai.response.id"))
			So(ok, ShouldBeTrue)
			So(responseID.AsString(), ShouldEqual, "response-1")
			_, hasInput := attrs.Value(attribute.Key("gen_ai.input.messages"))
			_, hasOutput := attrs.Value(attribute.Key("gen_ai.output.messages"))
			So(hasInput, ShouldBeFalse)
			So(hasOutput, ShouldBeFalse)
		})

		Convey("the stream records one TTFC measurement", func() {
			var resourceMetrics metricdata.ResourceMetrics
			So(reader.Collect(context.Background(), &resourceMetrics), ShouldBeNil)
			var count uint64
			for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
				for _, metric := range scopeMetrics.Metrics {
					if metric.Name != "gen_ai.client.operation.time_to_first_chunk" {
						continue
					}
					for _, point := range metric.Data.(metricdata.Histogram[float64]).DataPoints {
						count += point.Count
					}
				}
			}
			So(count, ShouldEqual, 1)
		})
	})
}
