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
	"fmt"
	"sync"
	"time"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

const instrumentationName = "github.com/neuraxes/neurouter/internal/biz/observability"

var (
	requestedModelKey = attribute.Key("neurouter.request.model")
	upstreamNameKey   = attribute.Key("neurouter.upstream.name")
	routerModelKey    = attribute.Key("neurouter.router.model")
)

type GenAITarget struct {
	RequestedModel string
	Provider       genaiconv.ProviderNameAttr
	Upstream       string
	RouterModel    string
	UpstreamModel  string
	ServerAddress  string
	ServerPort     int
}

func (t GenAITarget) attributes() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 5)
	if t.Upstream != "" {
		attrs = append(attrs, upstreamNameKey.String(t.Upstream))
	}
	if t.RouterModel != "" {
		attrs = append(attrs, routerModelKey.String(t.RouterModel))
	}
	if t.UpstreamModel != "" {
		attrs = append(attrs, semconv.GenAIRequestModel(t.UpstreamModel))
	}
	if t.ServerAddress != "" {
		attrs = append(attrs, semconv.ServerAddress(t.ServerAddress))
	}
	if t.ServerPort != 0 {
		attrs = append(attrs, semconv.ServerPort(t.ServerPort))
	}
	return attrs
}

type GenAIResult struct {
	ResponseID     string
	ResponseModel  string
	FinishReasons  []string
	Usage          *GenAITokenUsage
	SpanAttributes []attribute.KeyValue
}

type GenAITokenUsage struct {
	Input       int64
	Output      int64
	CachedInput int64
	Reasoning   int64
}

type GenAIInstrumenter struct {
	tracer           trace.Tracer
	operationTime    genaiconv.ClientOperationDuration
	timeToFirstChunk genaiconv.ClientOperationTimeToFirstChunk
	tokenUsage       genaiconv.ClientTokenUsage
}

type GenAIInvocation struct {
	instrumenter   *GenAIInstrumenter
	ctx            context.Context
	span           trace.Span
	operation      genaiconv.OperationNameAttr
	target         GenAITarget
	startedAt      time.Time
	onceFirstChunk sync.Once
	onceEnded      sync.Once
}

func NewGenAIInstrumenter(
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
) (*GenAIInstrumenter, error) {
	if tracerProvider == nil {
		tracerProvider = nooptrace.NewTracerProvider()
	}
	if meterProvider == nil {
		meterProvider = noopmetric.NewMeterProvider()
	}

	meter := meterProvider.Meter(
		instrumentationName,
		metric.WithSchemaURL(semconv.SchemaURL),
	)
	operationTime, err := genaiconv.NewClientOperationDuration(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create operation duration meter: %w", err)
	}
	timeToFirstChunk, err := genaiconv.NewClientOperationTimeToFirstChunk(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create time-to-first-chunk meter: %w", err)
	}
	tokenUsage, err := genaiconv.NewClientTokenUsage(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create token usage meter: %w", err)
	}

	return &GenAIInstrumenter{
		tracer: tracerProvider.Tracer(
			instrumentationName,
			trace.WithSchemaURL(semconv.SchemaURL),
		),
		operationTime:    operationTime,
		timeToFirstChunk: timeToFirstChunk,
		tokenUsage:       tokenUsage,
	}, nil
}

func (i *GenAIInstrumenter) Start(
	ctx context.Context,
	operation genaiconv.OperationNameAttr,
	target GenAITarget,
	attrs ...attribute.KeyValue,
) (context.Context, *GenAIInvocation) {
	startedAt := time.Now()

	spanAttrs := make([]attribute.KeyValue, 0, len(attrs)+8)
	spanAttrs = append(spanAttrs,
		semconv.GenAIOperationNameKey.String(string(operation)),
	)
	if target.RequestedModel != "" {
		spanAttrs = append(spanAttrs, requestedModelKey.String(target.RequestedModel))
	}
	if target.Provider != "" {
		spanAttrs = append(spanAttrs, semconv.GenAIProviderNameKey.String(string(target.Provider)))
	}
	spanAttrs = append(spanAttrs, target.attributes()...)
	spanAttrs = append(spanAttrs, attrs...)

	spanName := string(operation)
	if target.RouterModel != "" {
		spanName += " " + target.RouterModel
	}

	ctx, span := i.tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithTimestamp(startedAt),
		trace.WithAttributes(spanAttrs...),
	)

	return ctx, &GenAIInvocation{
		instrumenter: i,
		ctx:          ctx,
		span:         span,
		operation:    operation,
		target:       target,
		startedAt:    startedAt,
	}
}

func (i *GenAIInvocation) FirstChunk() {
	i.onceFirstChunk.Do(func() {
		i.instrumenter.timeToFirstChunk.Record(
			i.ctx,
			time.Since(i.startedAt).Seconds(),
			i.operation,
			i.target.Provider,
			i.target.attributes()...,
		)
	})
}

func (i *GenAIInvocation) End(result GenAIResult, err error) {
	i.onceEnded.Do(func() {
		endedAt := time.Now()
		errorType := classifyError(err)

		i.instrumenter.operationTime.Record(
			i.ctx,
			endedAt.Sub(i.startedAt).Seconds(),
			i.operation,
			i.target.Provider,
			i.metricAttributes(result.ResponseModel, errorType)...,
		)
		if usage := result.Usage; usage != nil {
			tokenAttrs := i.metricAttributes(result.ResponseModel, "")
			i.recordTokenUsage(usage.Input, genaiconv.TokenTypeInput, tokenAttrs)
			i.recordTokenUsage(usage.Output, genaiconv.TokenTypeOutput, tokenAttrs)
		}

		spanAttrs := make([]attribute.KeyValue, 0, len(result.SpanAttributes)+8)
		if result.ResponseID != "" {
			spanAttrs = append(spanAttrs, semconv.GenAIResponseID(result.ResponseID))
		}
		if result.ResponseModel != "" {
			spanAttrs = append(spanAttrs, semconv.GenAIResponseModel(result.ResponseModel))
		}
		if len(result.FinishReasons) > 0 {
			spanAttrs = append(spanAttrs, semconv.GenAIResponseFinishReasons(result.FinishReasons...))
		}
		if usage := result.Usage; usage != nil {
			spanAttrs = append(spanAttrs,
				semconv.GenAIUsageInputTokensKey.Int64(usage.Input),
				semconv.GenAIUsageOutputTokensKey.Int64(usage.Output),
			)
			if usage.CachedInput > 0 {
				spanAttrs = append(spanAttrs, semconv.GenAIUsageCacheReadInputTokensKey.Int64(usage.CachedInput))
			}
			if usage.Reasoning > 0 {
				spanAttrs = append(spanAttrs, semconv.GenAIUsageReasoningOutputTokensKey.Int64(usage.Reasoning))
			}
		}
		if errorType != "" {
			i.span.SetStatus(codes.Error, "")
			spanAttrs = append(spanAttrs, semconv.ErrorTypeKey.String(errorType))
		}
		spanAttrs = append(spanAttrs, result.SpanAttributes...)
		i.span.SetAttributes(spanAttrs...)
		i.span.End(trace.WithTimestamp(endedAt))
	})
}

func (i *GenAIInvocation) metricAttributes(responseModel, errorType string) []attribute.KeyValue {
	attrs := i.target.attributes()
	if responseModel != "" {
		attrs = append(attrs, semconv.GenAIResponseModel(responseModel))
	}
	if errorType != "" {
		attrs = append(attrs, semconv.ErrorTypeKey.String(errorType))
	}
	return attrs
}

func (i *GenAIInvocation) recordTokenUsage(
	tokens int64,
	tokenType genaiconv.TokenTypeAttr,
	attrs []attribute.KeyValue,
) {
	i.instrumenter.tokenUsage.Record(
		i.ctx,
		tokens,
		i.operation,
		i.target.Provider,
		tokenType,
		attrs...,
	)
}

func classifyError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "context_canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "context_deadline_exceeded"
	}
	if reason := kratoserrors.Reason(err); reason != "" {
		return reason
	}

	return string(genaiconv.ErrorTypeOther)
}
