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
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type upstream struct {
	config *conf.OpenAIConfig
	client *openai.Client
	log    *log.Helper
}

func NewOpenAIFactory() repository.UpstreamFactory[conf.OpenAIConfig] {
	return newOpenAIUpstream
}

func newOpenAIUpstream(config *conf.OpenAIConfig, logger log.Logger) (repository.ChatRepo, error) {
	return newOpenAIUpstreamWithClient(config, nil, logger)
}

// newOpenAIUpstreamWithClient creates a new OpenAI upstream with a custom HTTP client for testing.
func newOpenAIUpstreamWithClient(config *conf.OpenAIConfig, client option.HTTPClient, logger log.Logger) (repo *upstream, err error) {
	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	if client != nil {
		options = append(options, option.WithHTTPClient(client))
	}

	openaiClient := openai.NewClient(options...)

	repo = &upstream{
		config: config,
		client: &openaiClient,
		log:    log.NewHelper(logger),
	}
	return repo, nil
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	openAIReq := r.convertRequestToOpenAI(req)

	openAIResp, err := r.client.Chat.Completions.New(ctx, openAIReq)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:         openAIResp.ID,
		Model:      openAIResp.Model,
		Message:    r.convertMessageFromOpenAI(&openAIResp.Choices[0].Message),
		Statistics: convertStatisticsFromOpenAI(&openAIResp.Usage),
	}

	return

}

type openAIChatStreamClient struct {
	req       *entity.ChatReq
	upstream  *ssestream.Stream[openai.ChatCompletionChunk]
	messageID string
}

func (c *openAIChatStreamClient) Recv() (resp *entity.ChatResp, err error) {
next:
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()
	resp = c.convertChunkFromOpenAI(&chunk)
	if resp == nil {
		goto next
	}

	if resp.Message != nil {
		resp.Message.Id = c.messageID
	}

	return
}

func (c *openAIChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
	openAIReq := r.convertRequestToOpenAI(req)
	openAIReq.StreamOptions.IncludeUsage = openai.Opt(true)
	stream := r.client.Chat.Completions.NewStreaming(ctx, openAIReq)

	client = &openAIChatStreamClient{
		req:       req,
		upstream:  stream,
		messageID: uuid.NewString(),
	}
	return
}
