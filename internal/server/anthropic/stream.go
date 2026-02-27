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

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

// contentBlockKind distinguishes different content block types for streaming.
type contentBlockKind int

const (
	contentBlockNone contentBlockKind = iota
	contentBlockText
	contentBlockThinking
	contentBlockRedactedThinking
	contentBlockToolUse
)

type messageStreamServer struct {
	v1.Chat_ChatStreamServer
	httpCtx             http.Context
	messageStarted      bool
	contentBlockStarted bool
	contentIndex        int64
	contentKind         contentBlockKind
}

func (s *messageStreamServer) Context() context.Context {
	return s.httpCtx
}

func (s *messageStreamServer) sendEvent(event string, data []byte) (err error) {
	_, err = s.httpCtx.Response().Write([]byte("event: " + event + "\n"))
	if err != nil {
		return err
	}
	_, err = s.httpCtx.Response().Write([]byte("data: "))
	if err != nil {
		return err
	}
	_, err = s.httpCtx.Response().Write(data)
	if err != nil {
		return err
	}
	_, err = s.httpCtx.Response().Write([]byte("\n\n"))
	if err != nil {
		return err
	}
	s.httpCtx.Response().(http.Flusher).Flush()
	return
}

func (s *messageStreamServer) sendJSONEvent(event string, v any) (err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.sendEvent(event, data)
}

func (s *messageStreamServer) sendMessageStartEvent(resp *v1.ChatResp) {
	event := anthropic.MessageStartEvent{
		Message: anthropic.Message{
			ID:      resp.Message.Id,
			Model:   anthropic.Model(resp.Model),
			Role:    "assistant",
			Content: []anthropic.ContentBlockUnion{},
		},
	}
	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		event.Message.Usage = anthropic.Usage{
			InputTokens:          int64(resp.Statistics.Usage.InputTokens),
			OutputTokens:         int64(resp.Statistics.Usage.OutputTokens),
			CacheReadInputTokens: int64(resp.Statistics.Usage.CachedInputTokens),
		}
	}
	s.sendJSONEvent("message_start", event)
}

func (s *messageStreamServer) sendContentBlockStartEvent(contentBlockType string) {
	event := anthropic.ContentBlockStartEvent{
		Index: s.contentIndex,
		ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
			Type: contentBlockType,
		},
	}
	s.sendJSONEvent("content_block_start", event)
}

func (s *messageStreamServer) sendContentBlockStopEvent() {
	event := anthropic.ContentBlockStopEvent{
		Index: s.contentIndex,
	}
	s.sendJSONEvent("content_block_stop", event)
}

func (s *messageStreamServer) sendMessageStopEvent() {
	event := anthropic.MessageStopEvent{}
	s.sendJSONEvent("message_stop", event)
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
			// Determine the kind of the current content block
			currentKind := contentBlockNone
			switch content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					if content.Metadata["redacted_thinking"] != "" {
						currentKind = contentBlockRedactedThinking
					} else {
						currentKind = contentBlockThinking
					}
				} else {
					currentKind = contentBlockText
				}
			case *v1.Content_ToolUse:
				currentKind = contentBlockToolUse
			}

			// Switch content block type
			if content.Index != nil {
				currentContentIndex := int64(*content.Index)
				if s.contentBlockStarted && s.contentIndex != currentContentIndex {
					s.sendContentBlockStopEvent()
					s.contentBlockStarted = false
				}
				s.contentIndex = currentContentIndex
			} else {
				if s.contentBlockStarted && s.contentKind != currentKind {
					s.sendContentBlockStopEvent()
					s.contentBlockStarted = false
					s.contentIndex += 1
				}
			}
			s.contentKind = currentKind

			switch c := content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					if content.Metadata["redacted_thinking"] != "" {
						// Redacted thinking: emit as a self-contained content block
						if !s.contentBlockStarted {
							s.contentBlockStarted = true
							event := anthropic.ContentBlockStartEvent{
								Index: s.contentIndex,
								ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
									Type: "redacted_thinking",
									Data: content.Metadata["redacted_thinking"],
								},
							}
							s.sendJSONEvent("content_block_start", event)
						}
					} else {
						if !s.contentBlockStarted {
							s.contentBlockStarted = true
							s.sendContentBlockStartEvent("thinking")
						}
						if c.Text != "" {
							event := anthropic.ContentBlockDeltaEvent{
								Index: s.contentIndex,
								Delta: anthropic.RawContentBlockDeltaUnion{
									Type:     "thinking_delta",
									Thinking: c.Text,
								},
							}
							s.sendJSONEvent("content_block_delta", event)
						}
						if content.Metadata["signature"] != "" {
							event := anthropic.ContentBlockDeltaEvent{
								Index: s.contentIndex,
								Delta: anthropic.RawContentBlockDeltaUnion{
									Type:      "signature_delta",
									Signature: content.Metadata["signature"],
								},
							}
							s.sendJSONEvent("content_block_delta", event)
						}
					}
				} else {
					if !s.contentBlockStarted {
						s.contentBlockStarted = true
						s.sendContentBlockStartEvent("text")
					}
					event := anthropic.ContentBlockDeltaEvent{
						Index: s.contentIndex,
						Delta: anthropic.RawContentBlockDeltaUnion{
							Type: "text_delta",
							Text: c.Text,
						},
					}
					s.sendJSONEvent("content_block_delta", event)
				}
			case *v1.Content_ToolUse:
				if !s.contentBlockStarted {
					s.contentBlockStarted = true
					event := anthropic.ContentBlockStartEvent{
						Index: s.contentIndex,
						ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
							Type: "tool_use",
							ID:   c.ToolUse.GetId(),
							Name: c.ToolUse.GetName(),
						},
					}
					s.sendJSONEvent("content_block_start", event)
				}
				if len(c.ToolUse.Inputs) > 0 {
					event := anthropic.ContentBlockDeltaEvent{
						Index: s.contentIndex,
						Delta: anthropic.RawContentBlockDeltaUnion{
							Type:        "input_json_delta",
							PartialJSON: c.ToolUse.GetTextualInput(),
						},
					}
					s.sendJSONEvent("content_block_delta", event)
				}
			}
		}
	}

	// When we receive a stats-only response (no message), it means we're at the end
	if resp.Message == nil && resp.Statistics != nil && resp.Statistics.Usage != nil {
		if s.contentBlockStarted {
			s.sendContentBlockStopEvent()
			s.contentBlockStarted = false
			s.contentIndex += 1
		}

		deltaEvent := anthropic.MessageDeltaEvent{
			Delta: anthropic.MessageDeltaEventDelta{
				StopReason: convertStatusToAnthropic(resp.Status),
			},
			Usage: anthropic.MessageDeltaUsage{
				InputTokens:          int64(resp.Statistics.Usage.InputTokens),
				OutputTokens:         int64(resp.Statistics.Usage.OutputTokens),
				CacheReadInputTokens: int64(resp.Statistics.Usage.CachedInputTokens),
			},
		}
		s.sendJSONEvent("message_delta", deltaEvent)
	}

	return nil
}
