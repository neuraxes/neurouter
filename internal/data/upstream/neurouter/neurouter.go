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

package neurouter

import (
	"context"
	"io"
	"iter"
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"go.opentelemetry.io/otel/trace"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/biz/repository"
	"github.com/neuraxes/neurouter/internal/conf"
)

type upstream struct {
	config          *conf.NeurouterConfig
	chatClient      v1.ChatClient
	embeddingClient v1.EmbeddingClient
	log             *slog.Logger
}

func NewNeurouterFactory(tracerProvider trace.TracerProvider) repository.UpstreamFactory[conf.NeurouterConfig] {
	return func(config *conf.NeurouterConfig, logger *slog.Logger) (repository.Repo, error) {
		return newNeurouterUpstream(config, tracerProvider, logger)
	}
}

func newNeurouterUpstream(
	config *conf.NeurouterConfig,
	tracerProvider trace.TracerProvider,
	logger *slog.Logger,
) (repository.Repo, error) {
	clientTracing := tracing.Client(tracing.WithTracerProvider(tracerProvider))
	conn, err := grpc.NewClient(
		context.Background(),
		grpc.WithEndpoint(config.Endpoint),
		grpc.WithMiddleware(clientTracing),
		grpc.WithStreamMiddleware(clientTracing),
	)
	if err != nil {
		return nil, err
	}

	return &upstream{
		config:          config,
		chatClient:      v1.NewChatClient(conn),
		embeddingClient: v1.NewEmbeddingClient(conn),
		log:             logger,
	}, nil
}

func (r *upstream) Chat(ctx context.Context, req *entity.ChatRequest) (*entity.ChatResponse, error) {
	return r.chatClient.Chat(ctx, req)
}

type neurouterChatStreamClient struct {
	stream v1.Chat_ChatStreamClient
}

func (c *neurouterChatStreamClient) AsSeq() iter.Seq2[*entity.ChatEvent, error] {
	return func(yield func(*entity.ChatEvent, error) bool) {
		for {
			event, err := c.stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				yield(nil, err)
				return
			}

			if !yield(event, nil) {
				return
			}
		}
	}
}

func (r *upstream) ChatStream(ctx context.Context, req *entity.ChatRequest) iter.Seq2[*entity.ChatEvent, error] {
	stream, err := r.chatClient.ChatStream(ctx, req)
	if err != nil {
		return func(yield func(*entity.ChatEvent, error) bool) {
			yield(nil, err)
		}
	}

	client := &neurouterChatStreamClient{
		stream: stream,
	}

	return client.AsSeq()
}

func (r *upstream) Embed(ctx context.Context, req *entity.EmbedRequest) (*entity.EmbedResponse, error) {
	return r.embeddingClient.Embed(ctx, req)
}
