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
	"reflect"
	"strings"
	"sync"
	"time"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/neuraxes/neurouter/internal/biz/observability"

var (
	upstreamNameKey   = attribute.Key("neurouter.upstream.name")
	routerModelKey    = attribute.Key("neurouter.model.id")
	requestedModelKey = attribute.Key("neurouter.request.model")
)

type GenAITarget struct {
	Provider       genaiconv.ProviderNameAttr
	Upstream       string
	RequestedModel string
	Model          string
	UpstreamModel  string
	ServerAddress  string
	ServerPort     int
}

type GenAITokenUsage struct {
	Input       int64
	Output      int64
	CachedInput int64
	Reasoning   int64
}

type GenAIResult struct {
	ResponseID    string
	ResponseModel string
	FinishReasons []string
	Usage         *GenAITokenUsage
	Attributes    []attribute.KeyValue
}

type GenAIInstrumenter struct {
	tracer           trace.Tracer
	operationTime    genaiconv.ClientOperationDuration
	tokenUsage       genaiconv.ClientTokenUsage
	timeToFirstChunk genaiconv.ClientOperationTimeToFirstChunk
	now              func() time.Time
}

type GenAIInvocation struct {
	instrumenter *GenAIInstrumenter
	ctx          context.Context
	span         trace.Span
	operation    genaiconv.OperationNameAttr
	target       GenAITarget
	startedAt    time.Time
	firstChunk   sync.Once
	ended        sync.Once
}

func NewGenAIInstrumenter(
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
) (*GenAIInstrumenter, error) {
	if tracerProvider == nil {
		tracerProvider = trace.NewNoopTracerProvider()
	}
	if meterProvider == nil {
		meterProvider = metricnoop.NewMeterProvider()
	}

	meter := meterProvider.Meter(
		instrumentationName,
		metric.WithSchemaURL(semconv.SchemaURL),
	)
	operationTime, err := genaiconv.NewClientOperationDuration(meter)
	if err != nil {
		return nil, fmt.Errorf("create GenAI operation duration instrument: %w", err)
	}
	tokenUsage, err := genaiconv.NewClientTokenUsage(meter)
	if err != nil {
		return nil, fmt.Errorf("create GenAI token usage instrument: %w", err)
	}
	timeToFirstChunk, err := genaiconv.NewClientOperationTimeToFirstChunk(meter)
	if err != nil {
		return nil, fmt.Errorf("create GenAI time-to-first-chunk instrument: %w", err)
	}

	return &GenAIInstrumenter{
		tracer: tracerProvider.Tracer(
			instrumentationName,
			trace.WithSchemaURL(semconv.SchemaURL),
		),
		operationTime:    operationTime,
		tokenUsage:       tokenUsage,
		timeToFirstChunk: timeToFirstChunk,
		now:              time.Now,
	}, nil
}

func (i *GenAIInstrumenter) Start(
	ctx context.Context,
	operation genaiconv.OperationNameAttr,
	target GenAITarget,
	attrs ...attribute.KeyValue,
) (context.Context, *GenAIInvocation) {
	startedAt := i.now()
	spanAttrs := i.startAttributes(operation, target, attrs)
	spanName := string(operation)
	if target.UpstreamModel != "" {
		spanName += " " + target.UpstreamModel
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

func (i *GenAIInstrumenter) startAttributes(
	operation genaiconv.OperationNameAttr,
	target GenAITarget,
	attrs []attribute.KeyValue,
) []attribute.KeyValue {
	provider := target.Provider
	if provider == "" {
		provider = genaiconv.ProviderNameAttr("unknown")
	}

	result := make([]attribute.KeyValue, 0, len(attrs)+8)
	result = append(result,
		semconv.GenAIOperationNameKey.String(string(operation)),
		semconv.GenAIProviderNameKey.String(string(provider)),
	)
	if target.UpstreamModel != "" {
		result = append(result, semconv.GenAIRequestModel(target.UpstreamModel))
	}
	if target.ServerAddress != "" {
		result = append(result, semconv.ServerAddress(target.ServerAddress))
	}
	if target.ServerPort != 0 {
		result = append(result, semconv.ServerPort(target.ServerPort))
	}
	if target.Upstream != "" {
		result = append(result, upstreamNameKey.String(target.Upstream))
	}
	if target.Model != "" {
		result = append(result, routerModelKey.String(target.Model))
	}
	if target.RequestedModel != "" {
		result = append(result, requestedModelKey.String(target.RequestedModel))
	}
	return append(result, attrs...)
}

func (i *GenAIInvocation) FirstChunk() {
	i.firstChunk.Do(func() {
		duration := i.instrumenter.now().Sub(i.startedAt).Seconds()
		i.instrumenter.timeToFirstChunk.Record(
			i.ctx,
			duration,
			i.operation,
			i.provider(),
			i.metricAttributes(GenAIResult{}, "")...,
		)
	})
}

func (i *GenAIInvocation) End(result GenAIResult, err error) {
	i.ended.Do(func() {
		endedAt := i.instrumenter.now()
		errorType := classifyError(err)
		spanAttrs := make([]attribute.KeyValue, 0, len(result.Attributes)+8)
		if result.ResponseID != "" {
			spanAttrs = append(spanAttrs, semconv.GenAIResponseID(result.ResponseID))
		}
		if result.ResponseModel != "" {
			spanAttrs = append(spanAttrs, semconv.GenAIResponseModel(result.ResponseModel))
		}
		if len(result.FinishReasons) > 0 {
			spanAttrs = append(spanAttrs, semconv.GenAIResponseFinishReasons(result.FinishReasons...))
		}
		if result.Usage != nil {
			spanAttrs = append(spanAttrs,
				semconv.GenAIUsageInputTokensKey.Int64(result.Usage.Input),
				semconv.GenAIUsageOutputTokensKey.Int64(result.Usage.Output),
			)
			if result.Usage.CachedInput > 0 {
				spanAttrs = append(spanAttrs, semconv.GenAIUsageCacheReadInputTokensKey.Int64(result.Usage.CachedInput))
			}
			if result.Usage.Reasoning > 0 {
				spanAttrs = append(spanAttrs, semconv.GenAIUsageReasoningOutputTokensKey.Int64(result.Usage.Reasoning))
			}
		}
		if errorType != "" {
			spanAttrs = append(spanAttrs, semconv.ErrorTypeKey.String(errorType))
			i.span.SetStatus(codes.Error, "")
		}
		spanAttrs = append(spanAttrs, result.Attributes...)
		i.span.SetAttributes(spanAttrs...)

		metricAttrs := i.metricAttributes(result, errorType)
		i.instrumenter.operationTime.Record(
			i.ctx,
			endedAt.Sub(i.startedAt).Seconds(),
			i.operation,
			i.provider(),
			metricAttrs...,
		)
		if result.Usage != nil {
			tokenAttrs := i.metricAttributes(result, "")
			i.instrumenter.tokenUsage.Record(
				i.ctx,
				result.Usage.Input,
				i.operation,
				i.provider(),
				genaiconv.TokenTypeInput,
				tokenAttrs...,
			)
			i.instrumenter.tokenUsage.Record(
				i.ctx,
				result.Usage.Output,
				i.operation,
				i.provider(),
				genaiconv.TokenTypeOutput,
				tokenAttrs...,
			)
		}
		i.span.End(trace.WithTimestamp(endedAt))
	})
}

func (i *GenAIInvocation) provider() genaiconv.ProviderNameAttr {
	if i.target.Provider == "" {
		return genaiconv.ProviderNameAttr("unknown")
	}
	return i.target.Provider
}

func (i *GenAIInvocation) metricAttributes(result GenAIResult, errorType string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 7)
	if i.target.UpstreamModel != "" {
		attrs = append(attrs, semconv.GenAIRequestModel(i.target.UpstreamModel))
	}
	if result.ResponseModel != "" {
		attrs = append(attrs, semconv.GenAIResponseModel(result.ResponseModel))
	}
	if i.target.ServerAddress != "" {
		attrs = append(attrs, semconv.ServerAddress(i.target.ServerAddress))
	}
	if i.target.ServerPort != 0 {
		attrs = append(attrs, semconv.ServerPort(i.target.ServerPort))
	}
	if i.target.Upstream != "" {
		attrs = append(attrs, upstreamNameKey.String(i.target.Upstream))
	}
	if i.target.Model != "" {
		attrs = append(attrs, routerModelKey.String(i.target.Model))
	}
	if errorType != "" {
		attrs = append(attrs, semconv.ErrorTypeKey.String(errorType))
	}
	return attrs
}

func classifyError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "context_canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	if reason := kratoserrors.Reason(err); reason != "" {
		return reason
	}

	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	t := reflect.TypeOf(err)
	if t == nil {
		return string(genaiconv.ErrorTypeOther)
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	name := t.Name()
	if name == "" {
		return string(genaiconv.ErrorTypeOther)
	}
	if pkg := t.PkgPath(); pkg != "" {
		return strings.TrimSpace(pkg + "." + name)
	}
	return name
}
