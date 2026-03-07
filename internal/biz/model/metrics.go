package model

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// metrics holds OTel counter instruments for tracking model usage.
type metrics struct {
	inputTokens       metric.Int64Counter
	outputTokens      metric.Int64Counter
	cachedInputTokens metric.Int64Counter
	requests          metric.Int64Counter
}

// newMetrics creates a new metrics instance from the given MeterProvider.
func newMetrics(meterProvider metric.MeterProvider) (*metrics, error) {
	meter := meterProvider.Meter("neurouter")

	inputTokens, err := meter.Int64Counter("neurouter_input_tokens_total",
		metric.WithDescription("Total number of input tokens processed"),
	)
	if err != nil {
		return nil, err
	}

	outputTokens, err := meter.Int64Counter("neurouter_output_tokens_total",
		metric.WithDescription("Total number of output tokens generated"),
	)
	if err != nil {
		return nil, err
	}

	cachedInputTokens, err := meter.Int64Counter("neurouter_cached_input_tokens_total",
		metric.WithDescription("Total number of cached input tokens"),
	)
	if err != nil {
		return nil, err
	}

	requests, err := meter.Int64Counter("neurouter_requests_total",
		metric.WithDescription("Total number of requests processed"),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		inputTokens:       inputTokens,
		outputTokens:      outputTokens,
		cachedInputTokens: cachedInputTokens,
		requests:          requests,
	}, nil
}

func (m *metrics) recordTokenUsage(ctx context.Context, upstream, model string, input, output, cachedInput int64) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("upstream", upstream),
		attribute.String("model", model),
	)
	if input > 0 {
		m.inputTokens.Add(ctx, input, attrs)
	}
	if output > 0 {
		m.outputTokens.Add(ctx, output, attrs)
	}
	if cachedInput > 0 {
		m.cachedInputTokens.Add(ctx, cachedInput, attrs)
	}
}

func (m *metrics) recordRequest(ctx context.Context, upstream, model string) {
	if m == nil {
		return
	}
	m.requests.Add(ctx, 1, metric.WithAttributes(
		attribute.String("upstream", upstream),
		attribute.String("model", model),
	))
}
