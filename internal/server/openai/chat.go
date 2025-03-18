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

type chatStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx http.Context
}

func (c *chatStreamServer) Context() context.Context {
	return c.ctx
}

func (c *chatStreamServer) Send(resp *v1.ChatResp) error {
	chunk := &openai.ChatCompletionStreamResponse{
		Choices: []openai.ChatCompletionStreamChoice{},
	}

	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		chunk.ID = resp.Message.Id
		chunk.Choices = append(chunk.Choices, openai.ChatCompletionStreamChoice{
			Delta: openai.ChatCompletionStreamChoiceDelta{
				Role:    openai.ChatMessageRoleAssistant,
				Content: resp.Message.Contents[0].GetText(),
			},
		})
	}

	if resp.Statistics != nil {
		chunk.Usage = &openai.Usage{
			PromptTokens:     int(resp.Statistics.Usage.PromptTokens),
			CompletionTokens: int(resp.Statistics.Usage.CompletionTokens),
		}
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	c.ctx.Response().Write([]byte("data: "))
	c.ctx.Response().Write(chunkJson)
	c.ctx.Response().Write([]byte("\n\n"))
	c.ctx.Response().(http.Flusher).Flush()
	return nil
}

func handleChatCompletion(ctx http.Context, svc v1.ChatServer) error {
	requestBody, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	openAIReq := openai.ChatCompletionRequest{}
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertChatReqFromOpenAI(&openAIReq)

	if openAIReq.Stream {
		err = svc.ChatStream(req, &chatStreamServer{ctx: ctx})
	} else {
		m := ctx.Middleware(func(ctx context.Context, req any) (any, error) {
			return svc.Chat(ctx, req.(*v1.ChatReq))
		})
		resp, err := m(ctx, req)
		if err != nil {
			return err
		}

		openAIResp := convertChatRespToOpenAI(resp.(*v1.ChatResp))
		err = ctx.Result(200, openAIResp)
		if err != nil {
			return err
		}
	}

	return err
}
