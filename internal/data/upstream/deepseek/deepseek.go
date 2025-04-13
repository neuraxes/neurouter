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
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/openai/openai-go/packages/ssestream"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type ChatRepo struct {
	config *conf.DeepSeekConfig
	log    *log.Helper
}

func NewDeepSeekChatRepoFactory() repository.UpstreamFactory[conf.DeepSeekConfig] {
	return NewDeepSeekChatRepo
}

func NewDeepSeekChatRepo(config *conf.DeepSeekConfig, logger log.Logger) (repository.ChatRepo, error) {
	// Trim the trailing slash from the base URL to avoid double slashes
	config.BaseUrl = strings.TrimSuffix(config.BaseUrl, "/")

	return &ChatRepo{
		config: config,
		log:    log.NewHelper(logger),
	}, nil
}

func (r *ChatRepo) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	res, err := r.CreateChatCompletion(
		ctx,
		r.convertRequestToDeepSeek(req),
	)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:      req.Id,
		Message: r.convertMessageFromDeepSeek(res.Choices[0].Message),
	}
	resp.Message.Id = res.ID

	if res.Usage.PromptTokens != 0 || res.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(res.Usage.PromptTokens),
				CompletionTokens: int32(res.Usage.CompletionTokens),
			},
		}
	}
	return
}

type deepSeekChatStreamClient struct {
	id       string
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
	resp = &entity.ChatResp{
		Id: c.req.Id,
	}

	if len(chunk.Choices) > 0 {
		resp.Message = &v1.Message{
			Id:   c.id,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Choices[0].Delta.Content,
					},
				},
			},
			ReasoningContents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Choices[0].Delta.ReasoningContent,
					},
				},
			},
		}
		// Clear due to the reuse of the same message struct
		chunk.Choices[0].Delta = nil
	}

	if chunk.Usage != nil && (chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0) {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(chunk.Usage.PromptTokens),
				CompletionTokens: int32(chunk.Usage.CompletionTokens),
			},
		}
	}

	return
}

func (c *deepSeekChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *ChatRepo) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
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
		id:       uuid.NewString(),
		req:      req,
		upstream: ssestream.NewStream[ChatStreamResponse](ssestream.NewDecoder(resp), err),
	}
	return
}
