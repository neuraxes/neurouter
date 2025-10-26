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
	"context"
	"encoding/json"
	"io"
	"reflect"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

type messageStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx                 context.Context
	httpCtx             http.Context
	messageStarted      bool
	contentBlockStarted bool
	contentIndex        int64
	contentType         reflect.Type
}

func (s *messageStreamServer) Context() context.Context {
	return s.ctx
}

func (s *messageStreamServer) sendMessageStartEvent(resp *v1.ChatResp) {
	event := anthropic.MessageStartEvent{
		Message: anthropic.Message{
			ID:      resp.Message.Id,
			Content: []anthropic.ContentBlockUnion{},
		},
	}
	eventJson, _ := json.Marshal(event)
	_, _ = s.httpCtx.Response().Write([]byte("event: message_start\n"))
	_, _ = s.httpCtx.Response().Write([]byte("data: "))
	_, _ = s.httpCtx.Response().Write(eventJson)
	_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
	s.httpCtx.Response().(http.Flusher).Flush()
}

func (s *messageStreamServer) sendContentBlockStopEvent() {
	stopEvent := anthropic.ContentBlockStopEvent{
		Index: s.contentIndex,
	}
	eventJson, _ := json.Marshal(stopEvent)
	_, _ = s.httpCtx.Response().Write([]byte("event: content_block_stop\n"))
	_, _ = s.httpCtx.Response().Write([]byte("data: "))
	_, _ = s.httpCtx.Response().Write(eventJson)
	_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
	s.httpCtx.Response().(http.Flusher).Flush()
}

func (s *messageStreamServer) sendMessageStopEvent() {
	stopEvent := anthropic.MessageStopEvent{}
	eventJson, _ := json.Marshal(stopEvent)
	_, _ = s.httpCtx.Response().Write([]byte("event: message_stop\n"))
	_, _ = s.httpCtx.Response().Write([]byte("data: "))
	_, _ = s.httpCtx.Response().Write(eventJson)
	_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
	s.httpCtx.Response().(http.Flusher).Flush()
}

func (s *messageStreamServer) Send(resp *v1.ChatResp) error {
	// Send message_start event if this is the first response
	if !s.messageStarted && resp.Message != nil {
		s.messageStarted = true
		s.sendMessageStartEvent(resp)
	}

	// Send content block events
	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		for _, content := range resp.Message.Contents {
			// Switch content block type
			if s.contentType != reflect.TypeOf(content.Content) {
				if s.contentBlockStarted {
					s.sendContentBlockStopEvent()
					s.contentBlockStarted = false
					s.contentIndex += 1
				}
				s.contentType = reflect.TypeOf(content.Content)
			}

			switch content := content.Content.(type) {
			case *v1.Content_Text:
				if !s.contentBlockStarted {
					s.contentBlockStarted = true
					event := anthropic.ContentBlockStartEvent{
						Index: s.contentIndex,
						ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
							Type: "text",
						},
					}
					eventJson, _ := json.Marshal(event)
					_, _ = s.httpCtx.Response().Write([]byte("event: content_block_start\n"))
					_, _ = s.httpCtx.Response().Write([]byte("data: "))
					_, _ = s.httpCtx.Response().Write(eventJson)
					_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
					s.httpCtx.Response().(http.Flusher).Flush()
				}

				// Send content_block_delta event with text_delta
				event := anthropic.ContentBlockDeltaEvent{
					Index: s.contentIndex,
					Delta: anthropic.RawContentBlockDeltaUnion{
						Type: "text_delta",
						Text: content.Text,
					},
				}
				eventJson, _ := json.Marshal(event)
				_, _ = s.httpCtx.Response().Write([]byte("event: content_block_delta\n"))
				_, _ = s.httpCtx.Response().Write([]byte("data: "))
				_, _ = s.httpCtx.Response().Write(eventJson)
				_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
				s.httpCtx.Response().(http.Flusher).Flush()
			case *v1.Content_ToolUse:
				if !s.contentBlockStarted {
					s.contentBlockStarted = true
					event := anthropic.ContentBlockStartEvent{
						Index: s.contentIndex,
						ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
							Type: "tool_use",
							ID:   content.ToolUse.GetId(),
							Name: content.ToolUse.GetName(),
						},
					}
					eventJson, _ := json.Marshal(event)
					_, _ = s.httpCtx.Response().Write([]byte("event: content_block_start\n"))
					_, _ = s.httpCtx.Response().Write([]byte("data: "))
					_, _ = s.httpCtx.Response().Write(eventJson)
					_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
					s.httpCtx.Response().(http.Flusher).Flush()
				}

				// Send content_block_delta event with input_json_delta
				event := anthropic.ContentBlockDeltaEvent{
					Index: s.contentIndex,
					Delta: anthropic.RawContentBlockDeltaUnion{
						Type:        "input_json_delta",
						PartialJSON: content.ToolUse.GetTextualInput(),
					},
				}
				eventJson, _ := json.Marshal(event)
				_, _ = s.httpCtx.Response().Write([]byte("event: content_block_delta\n"))
				_, _ = s.httpCtx.Response().Write([]byte("data: "))
				_, _ = s.httpCtx.Response().Write(eventJson)
				_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
				s.httpCtx.Response().(http.Flusher).Flush()
			}
		}
	}

	// When we receive usage statistics, it means we're at the end
	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		// Send content_block_stop event if we had content
		if s.contentBlockStarted {
			s.sendContentBlockStopEvent()
			s.contentBlockStarted = false
			s.contentIndex += 1
		}

		// Send message_delta event with usage statistics
		deltaEvent := anthropic.MessageDeltaEvent{
			Delta: anthropic.MessageDeltaEventDelta{
				StopReason: "end_turn",
			},
			Usage: anthropic.MessageDeltaUsage{
				OutputTokens: int64(resp.Statistics.Usage.OutputTokens),
			},
		}
		eventJson, _ := json.Marshal(deltaEvent)
		_, _ = s.httpCtx.Response().Write([]byte("event: message_delta\n"))
		_, _ = s.httpCtx.Response().Write([]byte("data: "))
		_, _ = s.httpCtx.Response().Write(eventJson)
		_, _ = s.httpCtx.Response().Write([]byte("\n\n"))
		s.httpCtx.Response().(http.Flusher).Flush()
	}

	return nil
}

func handleMessageCompletion(httpCtx http.Context, svc v1.ChatServer) (err error) {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return
	}

	anthropicReq := anthropic.MessageNewParams{}
	err = json.Unmarshal(requestBody, &anthropicReq)
	if err != nil {
		return err
	}

	req := convertChatReqFromAnthropic(&anthropicReq)

	// Check if streaming is requested
	// Note: Anthropic uses server-sent events for streaming, check for specific headers
	acceptHeader := httpCtx.Request().Header.Get("Accept")
	isStream := acceptHeader == "text/event-stream" ||
		httpCtx.Request().Header.Get("X-Anthropic-Stream") == "true"

	if isStream {
		httpCtx.Response().Header().Set("Content-Type", "text/event-stream")
		httpCtx.Response().Header().Set("Cache-Control", "no-cache")
		httpCtx.Response().Header().Set("Connection", "keep-alive")

		streamServer := &messageStreamServer{
			ctx:     httpCtx,
			httpCtx: httpCtx,
		}

		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			err := svc.ChatStream(req.(*v1.ChatReq), streamServer)
			// Send final events after streaming is complete
			if err == nil {
				streamServer.sendMessageStopEvent()
			}
			return nil, err
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

		anthropicResp := convertChatRespToAnthropic(resp.(*v1.ChatResp))
		err = httpCtx.Result(200, anthropicResp)
		if err != nil {
			return err
		}
	}

	return
}
