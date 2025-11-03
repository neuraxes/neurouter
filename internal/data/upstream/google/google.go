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
	"io"
	"iter"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/genai"

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
	cc := &genai.ClientConfig{
		APIKey: config.ApiKey,
	}

	if httpClient != nil {
		cc.HTTPClient = httpClient
	}

	client, err := genai.NewClient(context.Background(), cc)
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
	config := &genai.GenerateContentConfig{
		Tools: convertToolsToGoogle(req.Tools),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: true,
		},
	}
	convertGenerationConfigToGoogle(req.Config, config)

	var messages []*genai.Content
	for _, msg := range req.Messages {
		messages = append(messages, convertMessageToGoogle(msg))
	}

	googleResp, err := r.client.Models.GenerateContent(ctx, req.Model, messages, config)
	if err != nil {
		return
	}

	resp = &entity.ChatResp{
		Id:         req.Id,
		Model:      googleResp.ModelVersion,
		Message:    convertMessageFromGoogle(googleResp.Candidates[0].Content),
		Statistics: convertStatisticsFromGoogle(googleResp.UsageMetadata),
	}
	resp.Message.Id = googleResp.ResponseID

	return
}

type googleChatStreamClient struct {
	req  *entity.ChatReq
	next func() (*genai.GenerateContentResponse, error, bool)
	stop func()
}

func (c *googleChatStreamClient) Recv() (*entity.ChatResp, error) {
	googleResp, err, ok := c.next()
	if !ok {
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
	resp.Message.Id = googleResp.ResponseID

	// Only send for last chunk
	if googleResp.UsageMetadata != nil && googleResp.UsageMetadata.CandidatesTokenCount != 0 {
		resp.Statistics = convertStatisticsFromGoogle(googleResp.UsageMetadata)
	}

	return resp, nil
}

func (c *googleChatStreamClient) Close() error {
	c.stop()
	return nil
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) (repository.ChatStreamClient, error) {
	config := &genai.GenerateContentConfig{
		Tools: convertToolsToGoogle(req.Tools),
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: true,
		},
	}
	convertGenerationConfigToGoogle(req.Config, config)

	var messages []*genai.Content
	for _, msg := range req.Messages {
		messages = append(messages, convertMessageToGoogle(msg))
	}

	it := r.client.Models.GenerateContentStream(ctx, req.Model, messages, config)

	// Adapt iterator to stream client
	next, stop := iter.Pull2(it)

	return &googleChatStreamClient{
		req:  req,
		next: next,
		stop: stop,
	}, nil
}

func (r *upstream) Embed(ctx context.Context, req *entity.EmbedReq) (resp *entity.EmbedResp, err error) {
	var parts []*genai.Part
	for _, content := range req.Contents {
		if part := convertContentToGoogle(content); part != nil {
			parts = append(parts, part)
		}
	}

	googleResp, err := r.client.Models.EmbedContent(ctx, req.Model, []*genai.Content{{Parts: parts}}, &genai.EmbedContentConfig{})
	if err != nil {
		return
	}

	resp = &entity.EmbedResp{
		Id:        req.Id,
		Embedding: googleResp.Embeddings[0].Values,
	}
	return
}
