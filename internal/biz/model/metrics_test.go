package model

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// newTestMetrics creates a Metrics instance backed by a ManualReader for testing.
func newTestMetrics() (*metrics, *sdkmetric.ManualReader) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	metrics, err := newMetrics(provider)
	if err != nil {
		panic(err)
	}
	return metrics, reader
}

// collectMetrics collects and returns all metrics from the reader.
func collectMetrics(reader *sdkmetric.ManualReader) map[string][]metricdata.DataPoint[int64] {
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		panic(err)
	}
	result := make(map[string][]metricdata.DataPoint[int64])
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
				result[m.Name] = sum.DataPoints
			}
		}
	}
	return result
}

func TestNewMetrics(t *testing.T) {
	Convey("Test NewMetrics", t, func() {
		Convey("should create metrics from MeterProvider", func() {
			m, _ := newTestMetrics()
			So(m, ShouldNotBeNil)
			So(m.inputTokens, ShouldNotBeNil)
			So(m.outputTokens, ShouldNotBeNil)
			So(m.cachedInputTokens, ShouldNotBeNil)
			So(m.requests, ShouldNotBeNil)
		})

		Convey("should work with noop MeterProvider", func() {
			m, err := newMetrics(noop.NewMeterProvider())
			So(err, ShouldBeNil)
			So(m, ShouldNotBeNil)
			So(m.inputTokens, ShouldNotBeNil)
			So(m.outputTokens, ShouldNotBeNil)
			So(m.cachedInputTokens, ShouldNotBeNil)
			So(m.requests, ShouldNotBeNil)
		})
	})
}

func TestMetrics_RecordTokenUsage(t *testing.T) {
	Convey("Test recordTokenUsage", t, func() {
		Convey("should record token usage with correct attributes", func() {
			m, reader := newTestMetrics()

			m.recordTokenUsage(context.Background(), "openai", "gpt-4", 100, 50, 10)

			data := collectMetrics(reader)

			// Check input tokens
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_input_tokens_total"][0].Value, ShouldEqual, 100)
			attrs := data["neurouter_input_tokens_total"][0].Attributes
			v, ok := attrs.Value(attribute.Key("upstream"))
			So(ok, ShouldBeTrue)
			So(v.AsString(), ShouldEqual, "openai")
			v, ok = attrs.Value(attribute.Key("model"))
			So(ok, ShouldBeTrue)
			So(v.AsString(), ShouldEqual, "gpt-4")

			// Check output tokens
			So(data["neurouter_output_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_output_tokens_total"][0].Value, ShouldEqual, 50)

			// Check cached input tokens
			So(data["neurouter_cached_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_cached_input_tokens_total"][0].Value, ShouldEqual, 10)
		})

		Convey("should accumulate multiple recordings", func() {
			m, reader := newTestMetrics()

			m.recordTokenUsage(context.Background(), "openai", "gpt-4", 100, 50, 10)
			m.recordTokenUsage(context.Background(), "openai", "gpt-4", 200, 100, 20)

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_input_tokens_total"][0].Value, ShouldEqual, 300)
			So(data["neurouter_output_tokens_total"][0].Value, ShouldEqual, 150)
			So(data["neurouter_cached_input_tokens_total"][0].Value, ShouldEqual, 30)
		})

		Convey("should separate metrics by upstream and model", func() {
			m, reader := newTestMetrics()

			m.recordTokenUsage(context.Background(), "openai", "gpt-4", 100, 50, 0)
			m.recordTokenUsage(context.Background(), "anthropic", "claude-3", 200, 100, 0)

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 2)
		})

		Convey("should skip zero-value counters", func() {
			m, reader := newTestMetrics()

			m.recordTokenUsage(context.Background(), "openai", "gpt-4", 100, 0, 0)

			data := collectMetrics(reader)
			So(data["neurouter_input_tokens_total"], ShouldHaveLength, 1)
			So(data["neurouter_input_tokens_total"][0].Value, ShouldEqual, 100)
			// output and cached should not have data points (or have 0 value)
			if pts, ok := data["neurouter_output_tokens_total"]; ok {
				for _, pt := range pts {
					So(pt.Value, ShouldEqual, 0)
				}
			}
		})

		Convey("should be nil-safe", func() {
			var m *metrics
			So(func() { m.recordTokenUsage(context.Background(), "openai", "gpt-4", 100, 50, 10) }, ShouldNotPanic)
		})
	})
}

func TestMetrics_RecordRequest(t *testing.T) {
	Convey("Test recordRequest", t, func() {
		Convey("should record request with correct attributes", func() {
			m, reader := newTestMetrics()

			m.recordRequest(context.Background(), "openai", "gpt-4")

			data := collectMetrics(reader)
			So(data["neurouter_requests_total"], ShouldHaveLength, 1)
			So(data["neurouter_requests_total"][0].Value, ShouldEqual, 1)
			attrs := data["neurouter_requests_total"][0].Attributes
			v, ok := attrs.Value(attribute.Key("upstream"))
			So(ok, ShouldBeTrue)
			So(v.AsString(), ShouldEqual, "openai")
			v, ok = attrs.Value(attribute.Key("model"))
			So(ok, ShouldBeTrue)
			So(v.AsString(), ShouldEqual, "gpt-4")
		})

		Convey("should be nil-safe", func() {
			var m *metrics
			So(func() { m.recordRequest(context.Background(), "openai", "gpt-4") }, ShouldNotPanic)
		})
	})
}
