package util

import (
	"context"

	"go.opentelemetry.io/otel/log"
)

const (
	EventServerReqReceived    = "server_request_received"
	EventServerRespSent       = "server_response_sent"
	EventUpstreamReqSent      = "upstream_request_sent"
	EventUpstreamRespReceived = "upstream_response_received"
)

// EmitEvent sends a telemetry event through the OTel log pipeline.
func EmitEvent(
	ctx context.Context,
	logger log.Logger,
	event string,
	body []byte,
	attrs ...log.KeyValue,
) {
	if logger == nil {
		return
	}
	var record log.Record
	record.SetEventName(event)
	record.SetBody(log.BytesValue(body))
	record.AddAttributes(attrs...)
	logger.Emit(ctx, record)
}
