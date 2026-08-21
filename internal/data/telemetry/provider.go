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

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/neuraxes/neurouter/internal/conf"
)

func NewResource() (*resource.Resource, error) {
	return resource.New(
		context.Background(),
		resource.WithAttributes(semconv.ServiceName("neurouter")),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
}

// NewTracerProvider creates an OTel TracerProvider.
func NewTracerProvider(data *conf.Data, res *resource.Resource) (trace.TracerProvider, func(), error) {
	options := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if otlpExporterEnabled(data) {
		exporter, err := otlptracegrpc.New(context.Background())
		if err != nil {
			return nil, nil, err
		}
		options = append(options, sdktrace.WithBatcher(exporter))
	}
	tp := sdktrace.NewTracerProvider(options...)
	cleanup := func() {
		_ = tp.Shutdown(context.Background())
	}
	return tp, cleanup, nil
}

// NewMeterProvider creates an OTel MeterProvider with the configured exporters.
func NewMeterProvider(data *conf.Data, res *resource.Resource) (metric.MeterProvider, func(), error) {
	ctx := context.Background()
	options := []sdkmetric.Option{sdkmetric.WithResource(res)}
	var otlpReader *sdkmetric.PeriodicReader
	if otlpExporterEnabled(data) {
		otlpExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, nil, err
		}
		otlpReader = sdkmetric.NewPeriodicReader(otlpExporter)
		options = append(options, sdkmetric.WithReader(otlpReader))
	}
	if prometheusExporterEnabled(data) {
		prometheusReader, err := prometheus.New()
		if err != nil {
			if otlpReader != nil {
				_ = otlpReader.Shutdown(ctx)
			}
			return nil, nil, err
		}
		options = append(options, sdkmetric.WithReader(prometheusReader))
	}
	mp := sdkmetric.NewMeterProvider(options...)
	cleanup := func() {
		_ = mp.Shutdown(context.Background())
	}
	return mp, cleanup, nil
}

// NewLoggerProvider creates an OTel LoggerProvider.
// Returns nil if event logging or OTLP export is disabled via config.
func NewLoggerProvider(
	data *conf.Data,
	res *resource.Resource,
) (otellog.LoggerProvider, func(), error) {
	if !data.GetEnableEventLog() || !otlpExporterEnabled(data) {
		return nil, func() {}, nil
	}
	exporter, err := otlploggrpc.New(context.Background())
	if err != nil {
		return nil, nil, err
	}
	processor := sdklog.NewBatchProcessor(exporter)
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)
	cleanup := func() {
		_ = lp.Shutdown(context.Background())
	}
	return lp, cleanup, nil
}

func otlpExporterEnabled(data *conf.Data) bool {
	return data.GetEnableOtlpExporter()
}

func prometheusExporterEnabled(data *conf.Data) bool {
	return data.GetEnablePrometheusExporter()
}
