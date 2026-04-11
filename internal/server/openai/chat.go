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
	"github.com/openai/openai-go/v3"
	"github.com/tidwall/gjson"

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
	chunk := &chatCompletionChunk{
		ID:      resp.Id,
		Object:  "chat.completion.chunk",
		Model:   resp.Model,
		Choices: []chatCompletionChunkChoice{},
	}

	if resp.Message != nil {
		var content string
		var toolCalls []toolCall
		for _, c := range resp.Message.Contents {
			switch c := c.Content.(type) {
			case *v1.Content_Text:
				content = c.Text
			case *v1.Content_ToolUse:
				toolCalls = append(toolCalls, toolCall{
					ID:   c.ToolUse.Id,
					Type: "function",
					Function: functionCall{
						Name:      c.ToolUse.GetName(),
						Arguments: c.ToolUse.GetTextualInput(),
					},
				})
			}
		}
		chunk.Choices = append(chunk.Choices, chatCompletionChunkChoice{
			Delta: chatCompletionChunkDelta{
				Role:      "assistant",
				Content:   content,
				ToolCalls: toolCalls,
			},
			FinishReason: convertStatusToOpenAI(resp.Status),
		})
	}

	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		chunk.Usage = convertUsageToOpenAI(resp.Statistics.Usage)
		if len(chunk.Choices) == 0 {
			chunk.Choices = append(chunk.Choices, chatCompletionChunkChoice{
				FinishReason: convertStatusToOpenAI(resp.Status),
			})
		}
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

func (s *OpenAIServer) handleChatCompletion(httpCtx http.Context) (err error) {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return
	}

	var openAIReq openai.ChatCompletionNewParams
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertChatReqFromOpenAI(&openAIReq)

	if gjson.GetBytes(requestBody, "stream").Bool() {
		httpCtx.Response().Header().Set("Content-Type", "text/event-stream")
		httpCtx.Response().Header().Set("Cache-Control", "no-cache")
		httpCtx.Response().Header().Set("Connection", "keep-alive")

		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			err := s.chatSvc.ChatStream(req.(*v1.ChatReq), &chatStreamServer{
				ctx:     ctx,
				httpCtx: httpCtx,
			})
			if err == nil {
				httpCtx.Response().Write([]byte("data: [DONE]\n\n"))
				httpCtx.Response().(http.Flusher).Flush()
			}
			return nil, err
		})
		_, err = m(httpCtx, req)
	} else {
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			return s.chatSvc.Chat(ctx, req.(*v1.ChatReq))
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
