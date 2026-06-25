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
	"bytes"
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

type messageStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx         context.Context
	httpCtx     http.Context
	buffer      *bytes.Buffer
	blockPhases map[uint32]v1.ContentPhase
}

func (s *messageStreamServer) Context() context.Context {
	return s.ctx
}

func (s *messageStreamServer) sendEvent(event string, eventData []byte) (err error) {
	data := []byte("event: " + event + "\ndata: ")
	data = append(data, eventData...)
	data = append(data, '\n', '\n')

	if s.buffer != nil {
		s.buffer.Write(data)
	}

	_, err = s.httpCtx.Response().Write(data)
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

func (s *messageStreamServer) Send(event *v1.ChatEvent) error {
	switch e := event.Event.(type) {
	case *v1.ChatEvent_MessageStart:
		s.startMessage(e.MessageStart, event.Usage)

	case *v1.ChatEvent_ContentStart:
		s.startContentBlock(e.ContentStart)

	case *v1.ChatEvent_ContentDelta:
		s.deltaContentBlock(e.ContentDelta)

	case *v1.ChatEvent_ContentStop:
		s.sendContentBlockStopEvent(int64(e.ContentStop.GetIndex()))

	case *v1.ChatEvent_ContentSnapshot:
		s.snapshotContentBlock(e.ContentSnapshot)

	case *v1.ChatEvent_MessageStop:
		var usage anthropic.MessageDeltaUsage
		if event.Usage != nil {
			u := convertUsageToAnthropic(event.Usage)
			usage = anthropic.MessageDeltaUsage{
				InputTokens:          u.InputTokens,
				OutputTokens:         u.OutputTokens,
				CacheReadInputTokens: u.CacheReadInputTokens,
			}
		}
		deltaEvent := anthropic.MessageDeltaEvent{
			Delta: anthropic.MessageDeltaEventDelta{
				StopReason: convertStatusToAnthropic(e.MessageStop.GetStatus()),
			},
			Usage: usage,
		}
		s.sendJSONEvent("message_delta", deltaEvent)
	}

	return nil
}

func (s *messageStreamServer) startMessage(start *v1.MessageStart, usage *v1.Usage) {
	event := anthropic.MessageStartEvent{
		Message: anthropic.Message{
			ID:      start.GetId(),
			Model:   anthropic.Model(start.GetModel()),
			Role:    "assistant",
			Content: []anthropic.ContentBlockUnion{},
		},
	}
	if usage != nil {
		event.Message.Usage = convertUsageToAnthropic(usage)
	}
	s.sendJSONEvent("message_start", event)
}

func (s *messageStreamServer) startContentBlock(start *v1.ContentStart) {
	if s.blockPhases == nil {
		s.blockPhases = map[uint32]v1.ContentPhase{}
	}
	s.blockPhases[start.GetIndex()] = start.GetPhase()

	contentBlock := anthropic.ContentBlockStartEventContentBlockUnion{}
	switch c := start.Content.(type) {
	case *v1.ContentStart_Text:
		if start.GetPhase() == v1.ContentPhase_CONTENT_PHASE_REASONING {
			contentBlock.Type = "thinking"
		} else {
			contentBlock.Type = "text"
		}
	case *v1.ContentStart_ToolUse:
		contentBlock.Type = "tool_use"
		contentBlock.ID = c.ToolUse.GetId()
		contentBlock.Name = c.ToolUse.GetName()
	}
	s.sendJSONEvent("content_block_start", anthropic.ContentBlockStartEvent{
		Index:        int64(start.GetIndex()),
		ContentBlock: contentBlock,
	})
}

func (s *messageStreamServer) deltaContentBlock(delta *v1.ContentDelta) {
	index := int64(delta.GetIndex())

	switch d := delta.Delta.(type) {
	case *v1.ContentDelta_Text:
		deltaUnion := anthropic.RawContentBlockDeltaUnion{}
		if s.blockPhases[delta.GetIndex()] == v1.ContentPhase_CONTENT_PHASE_REASONING {
			deltaUnion.Type = "thinking_delta"
			deltaUnion.Thinking = d.Text
		} else {
			deltaUnion.Type = "text_delta"
			deltaUnion.Text = d.Text
		}
		s.sendJSONEvent("content_block_delta", anthropic.ContentBlockDeltaEvent{
			Index: index,
			Delta: deltaUnion,
		})
	case *v1.ContentDelta_Signature:
		s.sendJSONEvent("content_block_delta", anthropic.ContentBlockDeltaEvent{
			Index: index,
			Delta: anthropic.RawContentBlockDeltaUnion{
				Type:      "signature_delta",
				Signature: d.Signature,
			},
		})
	case *v1.ContentDelta_ToolInputText:
		s.sendJSONEvent("content_block_delta", anthropic.ContentBlockDeltaEvent{
			Index: index,
			Delta: anthropic.RawContentBlockDeltaUnion{
				Type:        "input_json_delta",
				PartialJSON: d.ToolInputText,
			},
		})
	}
}

func (s *messageStreamServer) snapshotContentBlock(content *v1.Content) {
	index := int64(content.GetIndex())

	switch c := content.Content.(type) {
	case *v1.Content_Opaque:
		s.sendJSONEvent("content_block_start", anthropic.ContentBlockStartEvent{
			Index: index,
			ContentBlock: anthropic.ContentBlockStartEventContentBlockUnion{
				Type: "redacted_thinking",
				Data: c.Opaque,
			},
		})
	default:
		return
	}

	s.sendContentBlockStopEvent(index)
}

func (s *messageStreamServer) sendContentBlockStopEvent(index int64) {
	if s.blockPhases != nil {
		delete(s.blockPhases, uint32(index))
	}
	event := anthropic.ContentBlockStopEvent{
		Index: index,
	}
	s.sendJSONEvent("content_block_stop", event)
}

func (s *messageStreamServer) sendMessageStopEvent() {
	event := anthropic.MessageStopEvent{}
	s.sendJSONEvent("message_stop", event)
}
