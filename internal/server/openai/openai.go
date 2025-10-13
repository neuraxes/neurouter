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
	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func RegisterOpenAIHTTPServer(s *http.Server, modelSvc v1.ModelServer, chatSvc v1.ChatServer, embedSvc v1.EmbeddingServer) {
	r := s.Route("/")
	r.POST("/chat/completions", func(ctx http.Context) error {
		return handleChatCompletion(ctx, chatSvc)
	})
	r.POST("/v1/chat/completions", func(ctx http.Context) error {
		return handleChatCompletion(ctx, chatSvc)
	})
	r.POST("/embeddings", func(ctx http.Context) error {
		return handleEmbedding(ctx, embedSvc)
	})
	r.POST("/v1/embeddings", func(ctx http.Context) error {
		return handleEmbedding(ctx, embedSvc)
	})
	r.GET("/models", func(ctx http.Context) error {
		return handleListModels(ctx, modelSvc)
	})
}
