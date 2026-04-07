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

	for _, path := range []string{
		"/chat/completions",
		"/v1/chat/completions",
		"/openai/chat/completions",
		"/openai/v1/chat/completions",
	} {
		r.POST(path, func(ctx http.Context) error { return handleChatCompletion(ctx, chatSvc) })
	}

	for _, path := range []string{
		"/embeddings",
		"/v1/embeddings",
		"/openai/embeddings",
		"/openai/v1/embeddings",
	} {
		r.POST(path, func(ctx http.Context) error { return handleEmbedding(ctx, embedSvc) })
	}

	for _, path := range []string{
		"/models",
		"/openai/models",
		"/openai/v1/models",
	} {
		r.GET(path, func(ctx http.Context) error { return handleListModels(ctx, modelSvc) })
	}
}
