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

package deepseek

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/openai/openai-go/packages/ssestream"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type upstream struct {
	client httpClient
	config *conf.DeepSeekConfig
	log    *log.Helper
}

func NewDeepSeekChatRepoFactory() repository.UpstreamFactory[conf.DeepSeekConfig] {
	return NewDeepSeekChatRepo
}

func NewDeepSeekChatRepo(config *conf.DeepSeekConfig, logger log.Logger) (repository.ChatRepo, error) {
	return NewDeepSeekChatRepoWithClient(config, logger, nil)
}

// NewDeepSeekChatRepoWithClient creates a repository.ChatRepo with a custom HTTP client.
func NewDeepSeekChatRepoWithClient(config *conf.DeepSeekConfig, logger log.Logger, client httpClient) (repository.ChatRepo, error) {
	// Trim the trailing slash from the base URL to avoid double slashes
	config.BaseUrl = strings.TrimSuffix(config.BaseUrl, "/")

	if client == nil {
		client = http.DefaultClient
	}

	return &upstream{
		config: config,
		log:    log.NewHelper(logger),
		client: client,
	}, nil
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	deepSeekReq := r.convertRequestToDeepSeek(req)

	deepSeekResp, err := r.CreateChatCompletion(ctx, deepSeekReq)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:         req.Id,
		Model:      deepSeekResp.Model,
		Message:    r.convertMessageFromDeepSeek(deepSeekResp.ID, deepSeekResp.Choices[0].Message),
		Statistics: convertStatisticsFromDeepSeek(deepSeekResp.Usage),
	}

	return
}

type deepSeekChatStreamClient struct {
	req      *entity.ChatReq
	upstream *ssestream.Stream[ChatStreamResponse] // Reuse SSE Stream from OpenAI
}

func (c *deepSeekChatStreamClient) Recv() (resp *entity.ChatResp, err error) {
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()

	return convertStreamRespFromDeepSeek(c.req.Id, &chunk), nil
}

func (c *deepSeekChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
	deepSeekReq := r.convertRequestToDeepSeek(req)
	deepSeekReq.Stream = true
	deepSeekReq.StreamOptions = &StreamOptions{
		IncludeUsage: true,
	}

	resp, err := r.CreateChatCompletionStream(ctx, deepSeekReq)
	if err != nil {
		return
	}

	client = &deepSeekChatStreamClient{
		req:      req,
		upstream: ssestream.NewStream[ChatStreamResponse](ssestream.NewDecoder(resp), err),
	}
	return
}
