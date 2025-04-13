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
	"github.com/google/uuid"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type ChatRepo struct {
	config *conf.AnthropicConfig
	client *anthropic.Client
	log    *log.Helper
}

func NewAnthropicChatRepoFactory() repository.UpstreamFactory[conf.AnthropicConfig] {
	return NewAnthropicChatRepo
}

func NewAnthropicChatRepo(config *conf.AnthropicConfig, logger log.Logger) (repo repository.ChatRepo, err error) {
	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}
	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}
	client := anthropic.NewClient(options...)

	repo = &ChatRepo{
		config: config,
		client: &client,
		log:    log.NewHelper(logger),
	}
	return
}

func (r *ChatRepo) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	res, err := r.client.Messages.New(
		ctx,
		r.convertRequestToAnthropic(req),
	)
	if err != nil {
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id: req.Id,
		Message: &v1.Message{
			Id:   id.String(),
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: res.Content[0].Text,
					},
				},
			},
		},
	}

	if res.Usage.InputTokens != 0 || res.Usage.OutputTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(res.Usage.InputTokens),
				CompletionTokens: int32(res.Usage.OutputTokens),
			},
		}
	}
	return
}

type anthropicChatStreamClient struct {
	id       string
	req      *entity.ChatReq
	upstream *ssestream.Stream[anthropic.MessageStreamEventUnion]
}

func (c anthropicChatStreamClient) Recv() (resp *entity.ChatResp, err error) {
next:
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()
	if chunk.Type != "content_block_delta" {
		goto next
	}

	resp = &entity.ChatResp{
		Id: c.req.Id,
		Message: &v1.Message{
			Id:   c.id,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Delta.Text,
					},
				},
			},
		},
	}

	if chunk.Usage.OutputTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				CompletionTokens: int32(chunk.Usage.OutputTokens),
			},
		}
	}
	return
}

func (c anthropicChatStreamClient) Close() error {
	return c.upstream.Close()
}

func (r *ChatRepo) ChatStream(ctx context.Context, req *entity.ChatReq) (client repository.ChatStreamClient, err error) {
	stream := r.client.Messages.NewStreaming(
		ctx,
		r.convertRequestToAnthropic(req),
	)

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	client = &anthropicChatStreamClient{
		id:       id.String(),
		req:      req,
		upstream: stream,
	}
	return
}
