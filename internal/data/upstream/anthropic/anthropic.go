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

package anthropic

import (
	"context"
	"iter"
	"log/slog"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/upstream/shared"
)

type upstream struct {
	config *conf.AnthropicConfig
	client *anthropic.Client
	log    *slog.Logger
}

func NewAnthropicChatRepoFactory(
	loggerProvider otellog.LoggerProvider,
	tracerProvider trace.TracerProvider,
) repository.UpstreamFactory[conf.AnthropicConfig] {
	return func(config *conf.AnthropicConfig, logger *slog.Logger) (repository.Repo, error) {
		client := shared.NewRecordingClientFromLoggerProvider(
			loggerProvider,
			tracerProvider,
			"neurouter.upstream.anthropic",
		)
		return newAnthropicUpstreamWithClient(config, client, logger)
	}
}

func newAnthropicUpstreamWithClient(config *conf.AnthropicConfig, httpClient option.HTTPClient, logger *slog.Logger) (repo repository.ChatRepo, err error) {
	var options []option.RequestOption
	if config.ApiKey != "" {
		options = append(options, option.WithAPIKey(config.ApiKey))
	}
	if config.AuthToken != "" {
		options = append(options, option.WithAuthToken(config.AuthToken))
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	for k, v := range config.Headers {
		options = append(options, option.WithHeader(k, v))
	}
	if httpClient != nil {
		options = append(options, option.WithHTTPClient(httpClient))
	}

	repo = &upstream{
		config: config,
		client: new(anthropic.NewClient(options...)),
		log:    logger,
	}
	return
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatRequest) (resp *entity.ChatResponse, err error) {
	anthropicReq := r.convertRequestToAnthropic(req)

	anthropicResp, err := r.client.Messages.New(ctx, anthropicReq)
	if err != nil {
		return
	}

	resp = &entity.ChatResponse{
		Id:         req.Id,
		Model:      string(anthropicResp.Model),
		Message:    convertMessageFromAnthropic(anthropicResp),
		Statistics: convertStatisticsFromAnthropic(&anthropicResp.Usage),
		Status:     convertStatusFromAnthropic(anthropicResp.StopReason),
	}

	return
}

type anthropicChatStreamClient struct {
	req                  *entity.ChatRequest
	upstream             *ssestream.Stream[anthropic.MessageStreamEventUnion]
	messageID            string
	model                string
	pendingSnapshotStops map[uint32]struct{}
}

func (c *anthropicChatStreamClient) AsSeq() iter.Seq2[*entity.ChatEvent, error] {
	return func(yield func(*entity.ChatEvent, error) bool) {
		defer c.upstream.Close()
		for {
			if !c.upstream.Next() {
				if err := c.upstream.Err(); err != nil {
					yield(nil, err)
				}
				return
			}

			for _, event := range c.convertStreamEventFromAnthropic(new(c.upstream.Current())) {
				if !yield(event, nil) {
					return
				}
			}
		}
	}
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatRequest) iter.Seq2[*entity.ChatEvent, error] {
	anthropicReq := r.convertRequestToAnthropic(req)
	stream := r.client.Messages.NewStreaming(ctx, anthropicReq)

	client := &anthropicChatStreamClient{
		req:      req,
		upstream: stream,
	}

	return client.AsSeq()
}
