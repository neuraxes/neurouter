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
	"iter"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	otellog "go.opentelemetry.io/otel/log"
	"google.golang.org/genai"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/upstream/shared"
)

type upstream struct {
	config *conf.GoogleConfig
	client *genai.Client
	log    *log.Helper
}

func NewGoogleFactory(loggerProvider otellog.LoggerProvider) repository.UpstreamFactory[conf.GoogleConfig] {
	return func(config *conf.GoogleConfig, logger log.Logger) (repository.Repo, error) {
		client := shared.NewRecordingClientFromLoggerProvider(loggerProvider, "neurouter.upstream.google")
		return newGoogleUpstreamWithClient(config, client, logger)
	}
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

// buildGoogleRequest constructs the request to Google API.
func (r *upstream) buildGoogleRequest(req *entity.ChatReq) (messages []*genai.Content, config *genai.GenerateContentConfig) {
	config = &genai.GenerateContentConfig{
		SystemInstruction: r.convertSystemInstructionToGoogle(req.Messages),
		Tools:             convertToolsToGoogle(req.Tools),
	}
	convertGenerationConfigToGoogle(req.Config, config)

	for _, msg := range req.Messages {
		if msg.Role == v1.Role_SYSTEM && !r.config.SystemAsUser {
			continue
		}
		messages = append(messages, convertMessageToGoogle(msg))
	}
	return
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatReq) (resp *entity.ChatResp, err error) {
	messages, config := r.buildGoogleRequest(req)

	googleResp, err := r.client.Models.GenerateContent(ctx, req.Model, messages, config)
	if err != nil {
		return
	}

	if len(googleResp.Candidates) == 0 || googleResp.Candidates[0].Content == nil {
		err = errors.New("no candidates in response")
		return
	}

	resp = &entity.ChatResp{
		Id:         req.Id,
		Model:      googleResp.ModelVersion,
		Status:     convertStatusFromGoogle(googleResp.Candidates[0].FinishReason, googleResp.Candidates[0].Content),
		Message:    convertMessageFromGoogle(googleResp.Candidates[0].Content),
		Statistics: convertStatisticsFromGoogle(googleResp.UsageMetadata),
	}
	resp.Message.Id = googleResp.ResponseID

	return
}

type googleChatStreamClient struct {
	req *entity.ChatReq
	it  iter.Seq2[*genai.GenerateContentResponse, error]
}

func (c *googleChatStreamClient) AsSeq() iter.Seq2[*entity.ChatResp, error] {
	return func(yield func(*entity.ChatResp, error) bool) {
		for googleResp, err := range c.it {
			if err != nil {
				yield(nil, err)
				return
			}

			if len(googleResp.Candidates) == 0 || googleResp.Candidates[0].Content == nil {
				continue
			}

			resp := &entity.ChatResp{
				Id:      c.req.Id,
				Model:   c.req.Model,
				Status:  convertStatusFromGoogle(googleResp.Candidates[0].FinishReason, googleResp.Candidates[0].Content),
				Message: convertMessageFromGoogle(googleResp.Candidates[0].Content),
			}
			resp.Message.Id = googleResp.ResponseID

			if googleResp.UsageMetadata != nil {
				resp.Statistics = convertStatisticsFromGoogle(googleResp.UsageMetadata)
			}

			if !yield(resp, nil) {
				return
			}
		}
	}
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatReq) iter.Seq2[*entity.ChatResp, error] {
	messages, config := r.buildGoogleRequest(req)

	it := r.client.Models.GenerateContentStream(ctx, req.Model, messages, config)

	client := &googleChatStreamClient{
		req: req,
		it:  it,
	}

	return client.AsSeq()
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
