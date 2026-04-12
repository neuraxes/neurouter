package telemetry

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/neuraxes/neurouter/internal/conf"
)

// NewTracerProvider creates an OTel TracerProvider.
func NewTracerProvider() (trace.TracerProvider, func(), error) {
	tp := sdktrace.NewTracerProvider()
	cleanup := func() {
		_ = tp.Shutdown(context.Background())
	}
	return tp, cleanup, nil
}

// NewMeterProvider creates an OTel MeterProvider backed by a Prometheus exporter.
func NewMeterProvider() (metric.MeterProvider, func(), error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	cleanup := func() {
		_ = mp.Shutdown(context.Background())
	}
	return mp, cleanup, nil
}

// NewLoggerProvider creates an OTel LoggerProvider.
// Returns nil if event logging is disabled via config.
func NewLoggerProvider(data *conf.Data, logger log.Logger) (otellog.LoggerProvider, func(), error) {
	if !data.GetEnableEventLog() {
		return nil, func() {}, nil
	}
	exporter := NewKratosLogExporter(logger)
	processor := sdklog.NewBatchProcessor(exporter)
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(processor))
	cleanup := func() {
		_ = lp.Shutdown(context.Background())
	}
	return lp, cleanup, nil
}
