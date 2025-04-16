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
	"context"
	"encoding/json"
	"io"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/sashabaranov/go-openai"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func handleEmbedding(httpCtx http.Context, svc v1.EmbeddingServer) error {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return err
	}

	openAIReq := openai.EmbeddingRequest{}
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertEmbeddingReqFromOpenAI(&openAIReq)

	m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
		return svc.Embed(ctx, req.(*v1.EmbedReq))
	})
	resp, err := m(httpCtx, req)
	if err != nil {
		return err
	}

	openAIResp := convertEmbeddingRespToOpenAI(resp.(*v1.EmbedResp))
	return httpCtx.Result(200, openAIResp)
}
