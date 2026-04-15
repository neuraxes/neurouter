package shared

import (
	"bytes"
	"context"
	"io"
	"net/http"

	otellog "go.opentelemetry.io/otel/log"

	"github.com/neuraxes/neurouter/internal/util"
)

type recordingTransport struct {
	base   http.RoundTripper
	logger otellog.Logger
}

// NewRecordingClientFromLoggerProvider creates a recording client using a logger provider and scope name.
// Returns http.DefaultClient if the provider is nil.
func NewRecordingClientFromLoggerProvider(provider otellog.LoggerProvider, scope string) *http.Client {
	if provider == nil {
		return http.DefaultClient
	}

	return NewRecordingClient(provider.Logger(scope), nil)
}

// NewRecordingClient creates an http.Client that captures request and response bodies.
func NewRecordingClient(logger otellog.Logger, base http.RoundTripper) *http.Client {
	if base == nil {
		base = http.DefaultTransport
	}

	return &http.Client{
		Transport: &recordingTransport{
			base:   base,
			logger: logger,
		},
	}
}

type recordingBody struct {
	io.ReadCloser

	buf    bytes.Buffer
	ctx    context.Context
	logger otellog.Logger
}

func (b *recordingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.buf.Write(p[:n])
	}
	return n, err
}

func (b *recordingBody) Close() error {
	err := b.ReadCloser.Close()
	util.EmitEvent(b.ctx, b.logger, util.EventUpstreamRespReceived, b.buf.Bytes())
	return err
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		util.EmitEvent(req.Context(), t.logger, util.EventUpstreamReqSent, body)
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	resp.Body = &recordingBody{
		ReadCloser: resp.Body,
		ctx:        req.Context(),
		logger:     t.logger,
	}

	return resp, nil
}
