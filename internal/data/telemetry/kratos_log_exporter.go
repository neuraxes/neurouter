package telemetry

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// KratosLogExporter writes OTel log records to a Kratos logger.
type KratosLogExporter struct {
	log *log.Helper
}

func NewKratosLogExporter(logger log.Logger) sdklog.Exporter {
	return &KratosLogExporter{log: log.NewHelper(logger)}
}

func (e *KratosLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	for _, r := range records {
		e.log.WithContext(ctx).Log(
			log.LevelDebug, "event", r.EventName(), "body", string(r.Body().AsBytes()),
		)
	}
	return nil
}

func (e *KratosLogExporter) Shutdown(context.Context) error   { return nil }
func (e *KratosLogExporter) ForceFlush(context.Context) error { return nil }
