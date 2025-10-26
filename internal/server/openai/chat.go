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
	ctx     context.Context
	httpCtx http.Context
}

func (c *chatStreamServer) Context() context.Context {
	return c.ctx
}

func (c *chatStreamServer) Send(resp *v1.ChatResp) error {
	chunk := &openai.ChatCompletionStreamResponse{
		Choices: []openai.ChatCompletionStreamChoice{},
	}

	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		var content string
		var toolCalls []openai.ToolCall
		for _, c := range resp.Message.Contents {
			switch c := c.Content.(type) {
			case *v1.Content_Text:
				content = c.Text
			case *v1.Content_ToolUse:
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   c.ToolUse.Id,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      c.ToolUse.GetName(),
						Arguments: c.ToolUse.GetTextualInput(),
					},
				})
			}
		}
		chunk.ID = resp.Message.Id
		chunk.Choices = append(chunk.Choices, openai.ChatCompletionStreamChoice{
			Delta: openai.ChatCompletionStreamChoiceDelta{
				Role:      openai.ChatMessageRoleAssistant,
				Content:   content,
				ToolCalls: toolCalls,
			},
		})
	}

	if resp.Statistics != nil {
		chunk.Usage = &openai.Usage{
			PromptTokens:     int(resp.Statistics.Usage.InputTokens),
			CompletionTokens: int(resp.Statistics.Usage.OutputTokens),
		}
		chunk.Choices = append(chunk.Choices, openai.ChatCompletionStreamChoice{
			FinishReason: openai.FinishReasonStop,
		})
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	c.httpCtx.Response().Write([]byte("data: "))
	c.httpCtx.Response().Write(chunkJson)
	c.httpCtx.Response().Write([]byte("\n\n"))
	c.httpCtx.Response().(http.Flusher).Flush()
	return nil
}

func handleChatCompletion(httpCtx http.Context, svc v1.ChatServer) (err error) {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return
	}

	openAIReq := openai.ChatCompletionRequest{}
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertChatReqFromOpenAI(&openAIReq)

	if openAIReq.Stream {
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			return nil, svc.ChatStream(req.(*v1.ChatReq), &chatStreamServer{
				ctx:     ctx,
				httpCtx: httpCtx,
			})
		})
		_, err = m(httpCtx, req)
	} else {
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			return svc.Chat(ctx, req.(*v1.ChatReq))
		})
		resp, err := m(httpCtx, req)
		if err != nil {
			return err
		}

		openAIResp := convertChatRespToOpenAI(resp.(*v1.ChatResp))
		err = httpCtx.Result(200, openAIResp)
		if err != nil {
			return err
		}
	}

	return
}
