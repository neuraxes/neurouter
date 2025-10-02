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
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type upstream struct {
	config *conf.GoogleConfig
	client *genai.Client
	log    *log.Helper
}

func NewGoogleFactory() repository.UpstreamFactory[conf.GoogleConfig] {
	return newGoogleUpstream
}

func newGoogleUpstream(config *conf.GoogleConfig, logger log.Logger) (repo repository.ChatRepo, err error) {
	return newGoogleUpstreamWithClient(config, nil, logger)
}

func newGoogleUpstreamWithClient(config *conf.GoogleConfig, httpClient *http.Client, logger log.Logger) (repo repository.ChatRepo, err error) {
	options := []option.ClientOption{
		option.WithAPIKey(config.ApiKey),
	}

	if httpClient != nil {
		options = append(options, option.WithHTTPClient(httpClient))
	}

	client, err := genai.NewClient(context.Background(), options...)
	if err != nil {
		return
	}

	repo = &upstream{
		config: config,
		client: client,
		log:    log.NewHelper(logger),
	}
	return
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	messageLen := len(req.Messages)
	if messageLen == 0 {
		return nil, fmt.Errorf("request has no message")
	}

	model := r.client.GenerativeModel(req.Model)
	model.Tools = convertToolsToGoogle(req.Tools)
	cs := model.StartChat()

	// Add all but last message to history
	for i := 0; i < messageLen-1; i++ {
		cs.History = append(cs.History, convertMessageToGoogle(req.Messages[i]))
	}

	// Send the last message
	lastMsg := convertMessageToGoogle(req.Messages[len(req.Messages)-1])
	googleResp, err := cs.SendMessage(ctx, lastMsg.Parts...)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:         req.Id,
		Model:      req.Model,
		Message:    convertMessageFromGoogle(googleResp.Candidates[0].Content),
		Statistics: convertStatisticsFromGoogle(googleResp.UsageMetadata),
	}

	return
}

type googleChatStreamClient struct {
	req       *entity.ChatReq
	upstream  *genai.GenerateContentResponseIterator
	messageID string
}

func (c *googleChatStreamClient) Recv() (*entity.ChatResp, error) {
	googleResp, err := c.upstream.Next()
	if errors.Is(err, iterator.Done) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	resp := &entity.ChatResp{
		Id:      c.req.Id,
		Model:   c.req.Model,
		Message: convertMessageFromGoogle(googleResp.Candidates[0].Content),
	}
	resp.Message.Id = c.messageID

	// Only send for last chunk
	if googleResp.UsageMetadata != nil && googleResp.UsageMetadata.CandidatesTokenCount != 0 {
		resp.Statistics = convertStatisticsFromGoogle(googleResp.UsageMetadata)
	}

	return resp, nil
}

func (c *googleChatStreamClient) Close() error {
	return nil
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) (repository.ChatStreamClient, error) {
	messageLen := len(req.Messages)
	if messageLen == 0 {
		return nil, io.EOF
	}

	model := r.client.GenerativeModel(req.Model)
	model.Tools = convertToolsToGoogle(req.Tools)
	cs := model.StartChat()

	// Add all but last message to history
	for i := 0; i < messageLen-1; i++ {
		cs.History = append(cs.History, convertMessageToGoogle(req.Messages[i]))
	}

	// Send the last message
	lastMsg := convertMessageToGoogle(req.Messages[len(req.Messages)-1])
	iter := cs.SendMessageStream(ctx, lastMsg.Parts...)

	return &googleChatStreamClient{
		req:       req,
		upstream:  iter,
		messageID: uuid.NewString(),
	}, nil
}

func (r *upstream) Embed(ctx context.Context, req *entity.EmbedReq) (resp *entity.EmbedResp, err error) {
	model := r.client.EmbeddingModel(req.Model)

	var parts []genai.Part
	for _, content := range req.Contents {
		if part := convertContentToGoogle(content); part != nil {
			parts = append(parts, part)
		}
	}

	res, err := model.EmbedContent(ctx, parts...)
	if err != nil {
		return
	}

	resp = &entity.EmbedResp{
		Id:        req.Id,
		Embedding: res.Embedding.Values,
	}
	return
}
