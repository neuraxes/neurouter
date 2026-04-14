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
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/openai/openai-go/v3"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

type chatCompletionStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx     context.Context
	httpCtx http.Context
	buffer  *bytes.Buffer
}

func (c *chatCompletionStreamServer) Context() context.Context {
	return c.ctx
}

func (c *chatCompletionStreamServer) Send(resp *v1.ChatResp) error {
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
			FinishReason: convertStatusToOpenAIChat(resp.Status),
		})
	}

	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		chunk.Usage = convertUsageToOpenAIChat(resp.Statistics.Usage)
		if len(chunk.Choices) == 0 {
			chunk.Choices = append(chunk.Choices, chatCompletionChunkChoice{
				FinishReason: convertStatusToOpenAIChat(resp.Status),
			})
		}
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	data := append([]byte("data: "), chunkJson...)
	data = append(data, '\n', '\n')

	if c.buffer != nil {
		c.buffer.Write(data)
	}

	_, err = c.httpCtx.Response().Write(data)
	if err != nil {
		return err
	}
	c.httpCtx.Response().(http.Flusher).Flush()
	return nil
}

func (c *chatCompletionStreamServer) sendDone() error {
	data := []byte("data: [DONE]\n\n")
	if c.buffer != nil {
		c.buffer.Write(data)
	}
	_, err := c.httpCtx.Response().Write(data)
	if err != nil {
		return err
	}
	c.httpCtx.Response().(http.Flusher).Flush()
	return nil
}

func (s *Server) handleChatCompletion(httpCtx http.Context) (err error) {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return
	}

	var openAIReq openai.ChatCompletionNewParams
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertChatReqFromOpenAIChat(&openAIReq)

	if gjson.GetBytes(requestBody, "stream").Bool() {
		httpCtx.Response().Header().Set("Content-Type", "text/event-stream")
		httpCtx.Response().Header().Set("Cache-Control", "no-cache")
		httpCtx.Response().Header().Set("Connection", "keep-alive")

		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			util.EmitEvent(ctx, s.otelLogger, util.EventServerReqReceived, requestBody)
			streamServer := &chatCompletionStreamServer{
				ctx:     ctx,
				httpCtx: httpCtx,
			}
			if s.otelLogger != nil {
				streamServer.buffer = &bytes.Buffer{}
			}
			err := s.chatSvc.ChatStream(req.(*v1.ChatReq), streamServer)
			if err == nil {
				err = streamServer.sendDone()
			}
			if s.otelLogger != nil {
				util.EmitEvent(ctx, s.otelLogger, util.EventServerRespSent, streamServer.buffer.Bytes())
			}
			return nil, err
		})
		_, err = m(httpCtx, req)
	} else {
		var eventCtx context.Context = httpCtx
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			eventCtx = ctx
			util.EmitEvent(ctx, s.otelLogger, util.EventServerReqReceived, requestBody)
			return s.chatSvc.Chat(ctx, req.(*v1.ChatReq))
		})

		resp, err := m(httpCtx, req)
		if err != nil {
			return err
		}

		openAIResp := convertChatRespToOpenAIChat(resp.(*v1.ChatResp))

		respBytes, err := json.Marshal(openAIResp)
		if err != nil {
			return err
		}

		util.EmitEvent(eventCtx, s.otelLogger, util.EventServerRespSent, respBytes)

		err = httpCtx.Blob(200, "application/json", respBytes)
		if err != nil {
			return err
		}
	}

	return
}
