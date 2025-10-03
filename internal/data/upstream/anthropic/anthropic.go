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
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type upstream struct {
	config *conf.AnthropicConfig
	client *anthropic.Client
	log    *log.Helper
}

func NewAnthropicChatRepoFactory() repository.UpstreamFactory[conf.AnthropicConfig] {
	return newAnthropicUpstream
}

func newAnthropicUpstream(config *conf.AnthropicConfig, logger log.Logger) (repository.ChatRepo, error) {
	return newAnthropicUpstreamWithClient(config, nil, logger)
}

func newAnthropicUpstreamWithClient(config *conf.AnthropicConfig, httpClient option.HTTPClient, logger log.Logger) (repo repository.ChatRepo, err error) {
	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	if httpClient != nil {
		options = append(options, option.WithHTTPClient(httpClient))
	}

	client := anthropic.NewClient(options...)

	repo = &upstream{
		config: config,
		client: &client,
		log:    log.NewHelper(logger),
	}
	return
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	anthropicReq := r.convertRequestToAnthropic(req)

	anthropicResp, err := r.client.Messages.New(ctx, anthropicReq)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:      req.Id,
		Model:   string(anthropicReq.Model),
		Message: convertContentsFromAnthropic(anthropicResp.Content),
	}

	if anthropicResp.Usage.InputTokens != 0 || anthropicResp.Usage.OutputTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:       uint32(anthropicResp.Usage.InputTokens),
				CompletionTokens:   uint32(anthropicResp.Usage.OutputTokens),
				CachedPromptTokens: uint32(anthropicResp.Usage.CacheReadInputTokens),
			},
		}
	}

	return
}

type anthropicChatStreamClient struct {
	req         *entity.ChatReq
	upstream    *ssestream.Stream[anthropic.MessageStreamEventUnion]
	messageID   string
	model       string
	inputTokens uint32
}

func (c *anthropicChatStreamClient) Recv() (resp *entity.ChatResp, err error) {
next:
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()
	resp = c.convertChunkFromAnthropic(&chunk)
	if resp == nil {
		// The chunk is ignored, jump to the next one.
		goto next
	}

	return
}

func (c *anthropicChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
	anthropicReq := r.convertRequestToAnthropic(req)
	stream := r.client.Messages.NewStreaming(ctx, anthropicReq)

	client = &anthropicChatStreamClient{
		req:      req,
		upstream: stream,
	}
	return
}
