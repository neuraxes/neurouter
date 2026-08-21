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
	"log/slog"
	"net/http"

	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
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
	log    *slog.Logger
}

func NewGoogleFactory(
	loggerProvider otellog.LoggerProvider,
	tracerProvider trace.TracerProvider,
) repository.UpstreamFactory[conf.GoogleConfig] {
	return func(config *conf.GoogleConfig, logger *slog.Logger) (repository.Repo, error) {
		client := shared.NewRecordingClientFromLoggerProvider(
			loggerProvider,
			tracerProvider,
			"neurouter.upstream.google",
		)
		return newGoogleUpstreamWithClient(config, client, logger)
	}
}

func newGoogleUpstreamWithClient(config *conf.GoogleConfig, httpClient *http.Client, logger *slog.Logger) (repo repository.ChatRepo, err error) {
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
		log:    logger,
	}
	return
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatRequest) (resp *entity.ChatResponse, err error) {
	messages, config := r.convertRequestToGoogle(req)

	googleResp, err := r.client.Models.GenerateContent(ctx, req.Model, messages, config)
	if err != nil {
		return
	}

	if len(googleResp.Candidates) == 0 || googleResp.Candidates[0].Content == nil {
		err = errors.New("no candidates in response")
		return
	}

	resp = &entity.ChatResponse{
		Id:         req.Id,
		Model:      googleResp.ModelVersion,
		Status:     convertStatusFromGoogle(googleResp.Candidates[0].FinishReason, googleResp.Candidates[0].Content),
		Message:    convertMessageFromGoogleContent(googleResp.Candidates[0].Content),
		Statistics: convertStatisticsFromGoogle(googleResp.UsageMetadata),
	}
	resp.Message.Id = googleResp.ResponseID

	return
}

type googleChatStreamClient struct {
	req *entity.ChatRequest
	it  iter.Seq2[*genai.GenerateContentResponse, error]

	messageStarted bool
	nextIndex      uint32
	hasOpen        bool
	openIndex      uint32
	openPhase      v1.ContentPhase
	lastUsage      *v1.Usage
}

func (c *googleChatStreamClient) AsSeq() iter.Seq2[*entity.ChatEvent, error] {
	return func(yield func(*entity.ChatEvent, error) bool) {
		for googleResp, err := range c.it {
			if err != nil {
				yield(nil, err)
				return
			}

			if len(googleResp.Candidates) == 0 || googleResp.Candidates[0].Content == nil {
				continue
			}

			for _, event := range c.convertStreamResponseFromGoogle(googleResp) {
				if !yield(event, nil) {
					return
				}
			}
		}
	}
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatRequest) iter.Seq2[*entity.ChatEvent, error] {
	messages, config := r.convertRequestToGoogle(req)

	it := r.client.Models.GenerateContentStream(ctx, req.Model, messages, config)

	client := &googleChatStreamClient{
		req: req,
		it:  it,
	}

	return client.AsSeq()
}

func (r *upstream) Embed(ctx context.Context, req *entity.EmbedRequest) (resp *entity.EmbedResponse, err error) {
	var parts []*genai.Part
	for _, content := range req.Contents {
		if part := convertContentToGooglePart(content); part != nil {
			parts = append(parts, part)
		}
	}

	googleResp, err := r.client.Models.EmbedContent(ctx, req.Model, []*genai.Content{{Parts: parts}}, &genai.EmbedContentConfig{})
	if err != nil {
		return
	}

	resp = &entity.EmbedResponse{
		Id:        req.Id,
		Embedding: googleResp.Embeddings[0].Values,
	}
	return
}
