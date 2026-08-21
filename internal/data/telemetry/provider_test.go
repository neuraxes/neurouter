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
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	. "github.com/smartystreets/goconvey/convey"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectormetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"

	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/util"
)

type recordingLogService struct {
	collectorlog.UnimplementedLogsServiceServer
	requests chan *collectorlog.ExportLogsServiceRequest
}

func (s *recordingLogService) Export(
	_ context.Context,
	request *collectorlog.ExportLogsServiceRequest,
) (*collectorlog.ExportLogsServiceResponse, error) {
	s.requests <- request
	return &collectorlog.ExportLogsServiceResponse{}, nil
}

type recordingMetricService struct {
	collectormetric.UnimplementedMetricsServiceServer
	requests chan *collectormetric.ExportMetricsServiceRequest
}

func (s *recordingMetricService) Export(
	_ context.Context,
	request *collectormetric.ExportMetricsServiceRequest,
) (*collectormetric.ExportMetricsServiceResponse, error) {
	s.requests <- request
	return &collectormetric.ExportMetricsServiceResponse{}, nil
}

type recordingTraceService struct {
	collectortrace.UnimplementedTraceServiceServer
	requests chan *collectortrace.ExportTraceServiceRequest
}

func (s *recordingTraceService) Export(
	_ context.Context,
	request *collectortrace.ExportTraceServiceRequest,
) (*collectortrace.ExportTraceServiceResponse, error) {
	s.requests <- request
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func TestNewTracerProviderOTLPDisabled(t *testing.T) {
	Convey("Given OTLP tracing is disabled", t, func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer listener.Close()

		requests := make(chan *collectortrace.ExportTraceServiceRequest, 1)
		server := grpc.NewServer()
		collectortrace.RegisterTraceServiceServer(server, &recordingTraceService{requests: requests})
		go func() {
			_ = server.Serve(listener)
		}()
		defer server.Stop()

		t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://"+listener.Addr().String())
		data := &conf.Data{}
		provider, cleanup, err := NewTracerProvider(data, resource.Empty())

		So(err, ShouldBeNil)
		So(provider, ShouldNotBeNil)
		So(cleanup, ShouldNotBeNil)
		if err != nil {
			return
		}
		defer cleanup()

		_, span := provider.Tracer("test").Start(context.Background(), "test")
		So(span.SpanContext().IsValid(), ShouldBeTrue)
		span.End()
		So(provider.(*sdktrace.TracerProvider).ForceFlush(context.Background()), ShouldBeNil)
		select {
		case <-requests:
			So(true, ShouldBeFalse)
		case <-time.After(100 * time.Millisecond):
		}

		enabledProvider, enabledCleanup, err := NewTracerProvider(
			&conf.Data{EnableOtlpExporter: true},
			resource.Empty(),
		)
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer enabledCleanup()

		_, enabledSpan := enabledProvider.Tracer("test").Start(context.Background(), "exported")
		enabledSpan.End()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		So(enabledProvider.(*sdktrace.TracerProvider).ForceFlush(ctx), ShouldBeNil)
		select {
		case request := <-requests:
			So(request.GetResourceSpans(), ShouldHaveLength, 1)
		case <-ctx.Done():
			So(true, ShouldBeFalse)
		}
	})
}

func TestNewLoggerProviderDisabled(t *testing.T) {
	Convey("Given event logging is disabled", t, func() {
		provider, cleanup, err := NewLoggerProvider(&conf.Data{}, resource.Empty())

		So(err, ShouldBeNil)
		So(provider, ShouldBeNil)
		So(cleanup, ShouldNotBeNil)
		cleanup()
	})
}

func TestNewLoggerProviderOTLPDisabled(t *testing.T) {
	Convey("Given event logging is enabled but OTLP is disabled", t, func() {
		t.Setenv("OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE", "/missing/ca.pem")
		data := &conf.Data{
			EnableEventLog: true,
		}
		provider, cleanup, err := NewLoggerProvider(data, resource.Empty())

		So(err, ShouldBeNil)
		So(provider, ShouldBeNil)
		So(cleanup, ShouldNotBeNil)
		cleanup()
	})
}

func TestNewLoggerProviderExportsOTLP(t *testing.T) {
	Convey("Given an OTLP log collector", t, func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer listener.Close()

		requests := make(chan *collectorlog.ExportLogsServiceRequest, 1)
		service := &recordingLogService{requests: requests}
		server := grpc.NewServer()
		collectorlog.RegisterLogsServiceServer(server, service)
		go func() {
			_ = server.Serve(listener)
		}()
		defer server.Stop()

		t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://"+listener.Addr().String())
		provider, cleanup, err := NewLoggerProvider(
			&conf.Data{EnableEventLog: true, EnableOtlpExporter: true},
			resource.Empty(),
		)
		So(err, ShouldBeNil)
		So(provider, ShouldNotBeNil)
		if err != nil {
			return
		}
		defer cleanup()

		body := []byte(`{"message":"payload"}`)
		util.EmitEvent(context.Background(), provider.Logger("test"), "test.event", body)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		So(provider.(*sdklog.LoggerProvider).ForceFlush(ctx), ShouldBeNil)

		var request *collectorlog.ExportLogsServiceRequest
		select {
		case request = <-requests:
		case <-ctx.Done():
		}
		So(request, ShouldNotBeNil)
		if request == nil {
			return
		}
		So(request.GetResourceLogs(), ShouldHaveLength, 1)
		if len(request.GetResourceLogs()) == 0 || len(request.GetResourceLogs()[0].GetScopeLogs()) == 0 {
			return
		}
		records := request.GetResourceLogs()[0].GetScopeLogs()[0].GetLogRecords()
		So(records, ShouldHaveLength, 1)
		if len(records) == 0 {
			return
		}
		So(records[0].GetEventName(), ShouldEqual, "test.event")
		So(records[0].GetBody().GetStringValue(), ShouldEqual, string(body))
	})
}

func TestNewMeterProviderExporterSwitches(t *testing.T) {
	Convey("Given an OTLP metric collector and isolated Prometheus registries", t, func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		So(err, ShouldBeNil)
		if err != nil {
			return
		}
		defer listener.Close()

		requests := make(chan *collectormetric.ExportMetricsServiceRequest, 1)
		service := &recordingMetricService{requests: requests}
		server := grpc.NewServer()
		collectormetric.RegisterMetricsServiceServer(server, service)
		go func() {
			_ = server.Serve(listener)
		}()
		defer server.Stop()

		originalRegisterer := prometheus.DefaultRegisterer
		originalGatherer := prometheus.DefaultGatherer
		defer func() {
			prometheus.DefaultRegisterer = originalRegisterer
			prometheus.DefaultGatherer = originalGatherer
		}()

		t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://"+listener.Addr().String())
		testCases := []struct {
			name             string
			data             *conf.Data
			expectOTLP       bool
			expectPrometheus bool
		}{
			{name: "unset defaults to both exporters disabled", data: &conf.Data{}},
			{
				name: "both exporters enabled",
				data: &conf.Data{
					EnableOtlpExporter:       true,
					EnablePrometheusExporter: true,
				},
				expectOTLP:       true,
				expectPrometheus: true,
			},
			{
				name: "only OTLP enabled",
				data: &conf.Data{
					EnableOtlpExporter: true,
				},
				expectOTLP: true,
			},
			{
				name: "only Prometheus enabled",
				data: &conf.Data{
					EnablePrometheusExporter: true,
				},
				expectPrometheus: true,
			},
		}

		for _, testCase := range testCases {
			Convey(testCase.name, func() {
				drainMetricRequests(requests)
				registry := prometheus.NewRegistry()
				prometheus.DefaultRegisterer = registry
				prometheus.DefaultGatherer = registry

				provider, cleanup, err := NewMeterProvider(testCase.data, resource.Empty())
				So(err, ShouldBeNil)
				So(provider, ShouldNotBeNil)
				if err != nil {
					return
				}
				defer cleanup()

				counter, err := provider.Meter("test").Int64Counter("neurouter.test.requests")
				So(err, ShouldBeNil)
				counter.Add(context.Background(), 1)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				So(provider.(*sdkmetric.MeterProvider).ForceFlush(ctx), ShouldBeNil)

				if testCase.expectOTLP {
					var request *collectormetric.ExportMetricsServiceRequest
					select {
					case request = <-requests:
					case <-ctx.Done():
					}
					So(request, ShouldNotBeNil)
					if request != nil {
						So(hasOTLPMetric(request, "neurouter.test.requests"), ShouldBeTrue)
					}
				} else {
					select {
					case <-requests:
						So(true, ShouldBeFalse)
					case <-time.After(100 * time.Millisecond):
					}
				}

				families, err := registry.Gather()
				So(err, ShouldBeNil)
				found := false
				for _, family := range families {
					if family.GetName() == "neurouter_test_requests_total" {
						found = true
						break
					}
				}
				So(found, ShouldEqual, testCase.expectPrometheus)
			})
		}
	})
}

func drainMetricRequests(requests chan *collectormetric.ExportMetricsServiceRequest) {
	for {
		select {
		case <-requests:
		default:
			return
		}
	}
}

func hasOTLPMetric(request *collectormetric.ExportMetricsServiceRequest, name string) bool {
	for _, resourceMetrics := range request.GetResourceMetrics() {
		for _, scopeMetrics := range resourceMetrics.GetScopeMetrics() {
			for _, metric := range scopeMetrics.GetMetrics() {
				if metric.GetName() == name {
					return true
				}
			}
		}
	}
	return false
}
