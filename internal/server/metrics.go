package server

import (
	"context"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// NewMeterProvider creates an OTel MeterProvider backed by a Prometheus exporter.
// The exporter registers with the default Prometheus registry.
func NewMeterProvider() (metric.MeterProvider, func(), error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	cleanup := func() {
		_ = provider.Shutdown(context.Background())
	}
	return provider, cleanup, nil
}
