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

func newOpenAIUpstream(config *conf.OpenAIConfig, logger log.Logger) (repo repository.ChatRepo, err error) {

	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	client := openai.NewClient(options...)

	repo = &upstream{
		config: config,
		client: &client,
		log:    log.NewHelper(logger),
	}
	return repo, nil
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	res, err := r.client.Chat.Completions.New(
		ctx,
		r.convertRequestToOpenAI(req),
	)
	if err != nil {
		return
	}
	resp = r.convertResponseFromOpenAI(res)
	return

}

type openAIChatStreamClient struct {
	id       string
	req      *entity.ChatReq
	upstream *ssestream.Stream[openai.ChatCompletionChunk]
}

func (c openAIChatStreamClient) Recv() (resp *entity.ChatResp, err error) {
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}
	chunk := c.upstream.Current()
	resp = convertChunkFromOpenAI(&chunk, c.req.Id, c.id)
	return
}

func (c openAIChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
	openAIReq := r.convertRequestToOpenAI(req)
	openAIReq.StreamOptions.IncludeUsage = openai.Opt(true)
	stream := r.client.Chat.Completions.NewStreaming(ctx, openAIReq)

	client = &openAIChatStreamClient{
		id:       uuid.NewString(),
		req:      req,
		upstream: stream,
	}
	return
}
