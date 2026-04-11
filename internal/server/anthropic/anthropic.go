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
	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/service"
)

type AnthropicServer struct {
	chatSvc v1.ChatServer
}

func NewAnthropicServer(svc *service.RouterService) *AnthropicServer {
	return &AnthropicServer{
		chatSvc: svc,
	}
}

func (s *AnthropicServer) RegisterRoutes(srv *http.Server) {
	r := srv.Route("/")

	for _, path := range []string{
		"/messages",
		"/v1/messages",
		"/anthropic/messages",
		"/anthropic/v1/messages",
	} {
		r.POST(path, func(ctx http.Context) error {
			return s.handleMessageCompletion(ctx)
		})
	}
}
