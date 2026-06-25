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

package openai

import (
	"context"
	"iter"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
	otellog "go.opentelemetry.io/otel/log"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/upstream/shared"
)

type upstream struct {
	config *conf.OpenAIConfig
	client *openai.Client
	log    *log.Helper
}

func NewOpenAIFactory(loggerProvider otellog.LoggerProvider) repository.UpstreamFactory[conf.OpenAIConfig] {
	return func(config *conf.OpenAIConfig, logger log.Logger) (repository.Repo, error) {
		client := shared.NewRecordingClientFromLoggerProvider(loggerProvider, "neurouter.upstream.openai")
		return newOpenAIUpstreamWithClient(config, client, logger)
	}
}

// newOpenAIUpstreamWithClient creates a new OpenAI upstream with a custom HTTP client for testing.
func newOpenAIUpstreamWithClient(config *conf.OpenAIConfig, client option.HTTPClient, logger log.Logger) (repo *upstream, err error) {
	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	for k, v := range config.Headers {
		options = append(options, option.WithHeader(k, v))
	}
	if client != nil {
		options = append(options, option.WithHTTPClient(client))
	}

	repo = &upstream{
		config: config,
		client: new(openai.NewClient(options...)),
		log:    log.NewHelper(logger),
	}
	return repo, nil
}

func (r *upstream) chatWithCompletion(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	openAIReq := r.convertRequestToOpenAIChat(req)

	openAIResp, err := r.client.Chat.Completions.New(ctx, openAIReq)
	if err != nil {
		return
	}

	resp = r.convertResponseFromOpenAIChat(openAIResp)

	return
}

func (r *upstream) chatWithResponses(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	panic("unimplemented")
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	if r.config.UseResponsesApi {
		return r.chatWithResponses(ctx, req)
	}
	return r.chatWithCompletion(ctx, req)
}

type openAIChatStreamClient struct {
	req      *entity.ChatReq
	upstream *ssestream.Stream[openai.ChatCompletionChunk]

	messageID      string
	status         v1.ChatStatus
	messageStarted bool
	stopEmitted    bool
	nextIndex      uint32
	hasOpen        bool
	openIsTool     bool
	openIndex      uint32
	openPhase      v1.ContentPhase
	openToolIndex  int64
}

func (c *openAIChatStreamClient) AsSeq() iter.Seq2[*entity.ChatEvent, error] {
	return func(yield func(*entity.ChatEvent, error) bool) {
		defer c.upstream.Close()
		for {
			if !c.upstream.Next() {
				if err := c.upstream.Err(); err != nil {
					yield(nil, err)
					return
				}
				break
			}

			for _, event := range c.convertStreamChunkFromOpenAIChat(new(c.upstream.Current())) {
				if !yield(event, nil) {
					return
				}
			}
		}

		for _, event := range c.finish() {
			if !yield(event, nil) {
				return
			}
		}
	}
}

func (r *upstream) chatStreamWithCompletion(ctx context.Context, req *entity.ChatReq) iter.Seq2[*entity.ChatEvent, error] {
	openAIReq := r.convertRequestToOpenAIChat(req)
	openAIReq.StreamOptions.IncludeUsage = openai.Opt(true)
	stream := r.client.Chat.Completions.NewStreaming(ctx, openAIReq)

	client := &openAIChatStreamClient{
		req:       req,
		upstream:  stream,
		messageID: uuid.NewString(),
	}

	return client.AsSeq()
}

type openAIResponseStreamClient struct {
	req      *entity.ChatReq
	upstream *ssestream.Stream[responses.ResponseStreamEventUnion]
}

func (c *openAIResponseStreamClient) AsSeq() iter.Seq2[*entity.ChatEvent, error] {
	return func(yield func(*entity.ChatEvent, error) bool) {
		defer c.upstream.Close()
		for {
			if !c.upstream.Next() {
				if err := c.upstream.Err(); err != nil {
					yield(nil, err)
				}
				return
			}

			panic("unimplemented")
		}
	}
}

func (r *upstream) chatStreamWithResponses(ctx context.Context, req *entity.ChatReq) iter.Seq2[*entity.ChatEvent, error] {
	panic("unimplemented")
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) iter.Seq2[*entity.ChatEvent, error] {
	if r.config.UseResponsesApi {
		return r.chatStreamWithResponses(ctx, req)
	}
	return r.chatStreamWithCompletion(ctx, req)
}
