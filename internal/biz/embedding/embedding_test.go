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

package embedding

import (
	"context"
	"io"
	"log/slog"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"
	"go.opentelemetry.io/otel/trace"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/observability"
	"github.com/neuraxes/neurouter/internal/biz/repository"
)

type instrumentedEmbeddingRepo struct {
	sawSpan bool
}

func (r *instrumentedEmbeddingRepo) Embed(ctx context.Context, _ *entity.EmbedRequest) (*entity.EmbedResponse, error) {
	r.sawSpan = trace.SpanContextFromContext(ctx).IsValid()
	return &v1.EmbedResponse{
		Model:     "text-embedding-004-2026",
		Embedding: []float32{0.1, 0.2, 0.3},
	}, nil
}

type instrumentedEmbeddingModel struct {
	repo           repository.EmbeddingRepo
	recordedTokens int64
	closed         bool
}

func (m *instrumentedEmbeddingModel) EmbeddingRepo() repository.EmbeddingRepo {
	return m.repo
}

func (m *instrumentedEmbeddingModel) GenAITarget() observability.GenAITarget {
	return observability.GenAITarget{
		Provider:      genaiconv.ProviderNameGCPGemini,
		Upstream:      "gemini-primary",
		Model:         "embedding-router",
		UpstreamModel: "text-embedding-004",
	}
}

func (m *instrumentedEmbeddingModel) RecordUsage(_ context.Context, tokens int64) {
	m.recordedTokens = tokens
}

func (m *instrumentedEmbeddingModel) Close() {
	m.closed = true
}

type instrumentedEmbeddingElector struct {
	model Model
}

func (e *instrumentedEmbeddingElector) ElectForEmbedding(
	_ context.Context,
	req *v1.EmbedRequest,
) (Model, error) {
	req.Model = "text-embedding-004"
	return e.model, nil
}

func TestEmbedObservability(t *testing.T) {
	Convey("Given an embedding request", t, func() {
		spanRecorder := tracetest.NewSpanRecorder()
		tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
		meterProvider := sdkmetric.NewMeterProvider()
		instrumenter, err := observability.NewGenAIInstrumenter(tracerProvider, meterProvider)
		So(err, ShouldBeNil)

		repo := &instrumentedEmbeddingRepo{}
		model := &instrumentedEmbeddingModel{repo: repo}
		uc := NewUseCase(
			&instrumentedEmbeddingElector{model: model},
			instrumenter,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		)
		req := &v1.EmbedRequest{Model: "semantic-search"}

		resp, err := uc.Embed(context.Background(), req)

		So(err, ShouldBeNil)
		So(resp.GetEmbedding(), ShouldHaveLength, 3)
		So(repo.sawSpan, ShouldBeTrue)
		So(model.closed, ShouldBeTrue)
		So(model.recordedTokens, ShouldEqual, 0)

		spans := spanRecorder.Ended()
		So(spans, ShouldHaveLength, 1)
		span := spans[0]
		So(span.Name(), ShouldEqual, "embeddings text-embedding-004")
		attrs := attribute.NewSet(span.Attributes()...)
		dimensions, ok := attrs.Value(semconv.GenAIEmbeddingsDimensionCountKey)
		So(ok, ShouldBeTrue)
		So(dimensions.AsInt64(), ShouldEqual, 3)
		requestedModel, ok := attrs.Value(attribute.Key("neurouter.request.model"))
		So(ok, ShouldBeTrue)
		So(requestedModel.AsString(), ShouldEqual, "semantic-search")
	})
}
