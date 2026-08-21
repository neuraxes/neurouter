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

package shared

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/neuraxes/neurouter/internal/util"
)

type recordingTransport struct {
	base   http.RoundTripper
	logger otellog.Logger
}

// NewRecordingClientFromLoggerProvider creates an instrumented client and enables body event logging when configured.
func NewRecordingClientFromLoggerProvider(
	provider otellog.LoggerProvider,
	tracerProvider trace.TracerProvider,
	scope string,
) *http.Client {
	var base http.RoundTripper = http.DefaultTransport
	if provider != nil {
		base = &recordingTransport{
			base:   base,
			logger: provider.Logger(scope),
		}
	}

	return &http.Client{Transport: otelhttp.NewTransport(
		base,
		otelhttp.WithTracerProvider(tracerProvider),
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
		)),
	)}
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
