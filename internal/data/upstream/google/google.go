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

package google

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
	"git.xdea.xyz/Turing/neurouter/internal/biz/repository"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type ChatRepo struct {
	config *conf.GoogleConfig
	client *genai.Client
	log    *log.Helper
}

func NewGoogleChatRepoFactory() repository.ChatRepoFactory[conf.GoogleConfig] {
	return NewGoogleChatRepo
}

func NewGoogleChatRepo(config *conf.GoogleConfig, logger log.Logger) (repo repository.ChatRepo, err error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(config.ApiKey))
	if err != nil {
		return
	}

	repo = &ChatRepo{
		config: config,
		client: client,
		log:    log.NewHelper(logger),
	}
	return
}

func (r *ChatRepo) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	messageLen := len(req.Messages)
	if messageLen == 0 {
		return nil, fmt.Errorf("no messages")
	}

	model := r.client.GenerativeModel(req.Model)
	cs := model.StartChat()

	// Add all but last message to history
	for i := 0; i < messageLen-1; i++ {
		cs.History = append(cs.History, convertMessageToGoogle(req.Messages[i]))
	}

	// Send the last message
	lastMsg := convertMessageToGoogle(req.Messages[len(req.Messages)-1])
	res, err := cs.SendMessage(ctx, lastMsg.Parts...)
	if err != nil {
		return
	}

	message := convertMessageFromGoogle(res.Candidates[0].Content)
	resp = &entity.ChatResp{
		Id:      req.Id,
		Message: message,
	}

	if res.UsageMetadata != nil {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     res.UsageMetadata.PromptTokenCount,
				CompletionTokens: res.UsageMetadata.CandidatesTokenCount,
			},
		}
	}
	return
}

type googleChatStreamClient struct {
	id       string
	req      *entity.ChatReq
	upstream *genai.GenerateContentResponseIterator
}

func (c *googleChatStreamClient) Recv() (*entity.ChatResp, error) {
	res, err := c.upstream.Next()
	if errors.Is(err, iterator.Done) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	if len(res.Candidates) == 0 {
		return nil, io.EOF
	}

	message := convertMessageFromGoogle(res.Candidates[0].Content)
	message.Id = c.id

	resp := &entity.ChatResp{
		Id:      c.req.Id,
		Message: message,
	}

	// Only send for last chunk
	if res.UsageMetadata != nil && res.UsageMetadata.CandidatesTokenCount != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     res.UsageMetadata.PromptTokenCount,
				CompletionTokens: res.UsageMetadata.CandidatesTokenCount,
			},
		}
	}
	return resp, nil
}

func (c *googleChatStreamClient) Close() error {
	return nil
}

func (r *ChatRepo) ChatStream(ctx context.Context, req *entity.ChatReq) (repository.ChatStreamClient, error) {
	messageLen := len(req.Messages)
	if messageLen == 0 {
		return nil, io.EOF
	}

	model := r.client.GenerativeModel(req.Model)
	cs := model.StartChat()

	// Add all but last message to history
	for i := 0; i < messageLen-1; i++ {
		cs.History = append(cs.History, convertMessageToGoogle(req.Messages[i]))
	}

	// Send the last message
	lastMsg := convertMessageToGoogle(req.Messages[len(req.Messages)-1])
	iter := cs.SendMessageStream(ctx, lastMsg.Parts...)

	return &googleChatStreamClient{
		id:       uuid.NewString(),
		req:      req,
		upstream: iter,
	}, nil
}
