package telemetry

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewResource,
	NewTracerProvider,
	NewMeterProvider,
	NewLoggerProvider,
)
