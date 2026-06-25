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

package chat

import (
	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

// ChatEventReducer rebuilds a complete ChatResp from a stream of ChatEvents.
type ChatEventReducer struct {
	resp   *v1.ChatResp
	blocks map[uint32]*v1.Content
	log    *log.Helper
}

func NewChatEventReducer(logger *log.Helper) *ChatEventReducer {
	if logger == nil {
		logger = log.NewHelper(log.DefaultLogger)
	}

	return &ChatEventReducer{
		resp:   &v1.ChatResp{},
		blocks: map[uint32]*v1.Content{},
		log:    logger,
	}
}

func (r *ChatEventReducer) Resp() *v1.ChatResp {
	return r.resp
}

func (r *ChatEventReducer) ensureMessage() {
	if r.resp.Message == nil {
		r.resp.Message = &v1.Message{Role: v1.Role_MODEL}
	}
}

func (r *ChatEventReducer) Reduce(event *v1.ChatEvent) {
	if event == nil {
		return
	}

	if event.Id != "" {
		r.resp.Id = event.Id
	}
	r.reduceUsage(event.Usage)

	switch e := event.Event.(type) {
	case *v1.ChatEvent_MessageStart:
		r.ensureMessage()
		r.resp.Model = e.MessageStart.GetModel()
		r.resp.Message.Id = e.MessageStart.GetId()
	case *v1.ChatEvent_MessageStop:
		r.resp.Status = e.MessageStop.GetStatus()
	case *v1.ChatEvent_ContentStart:
		r.startContent(e.ContentStart)
	case *v1.ChatEvent_ContentDelta:
		r.applyDelta(e.ContentDelta)
	case *v1.ChatEvent_ContentStop:
		// The block is already materialized; nothing to merge.
	case *v1.ChatEvent_ContentSnapshot:
		r.ensureMessage()
		content := e.ContentSnapshot
		r.resp.Message.Contents = append(r.resp.Message.Contents, content)
		if content.Index != nil {
			r.blocks[*content.Index] = content
		}
	}
}

func (r *ChatEventReducer) startContent(start *v1.ContentStart) {
	r.ensureMessage()

	index := start.GetIndex()
	content := &v1.Content{
		Id:       start.GetId(),
		Index:    new(uint32(index)),
		Phase:    start.GetPhase(),
		Metadata: start.GetMetadata(),
	}

	switch c := start.Content.(type) {
	case *v1.ContentStart_Text:
		content.Content = v1.NewTextContent("")
	case *v1.ContentStart_ToolUse:
		content.Content = &v1.Content_ToolUse{
			ToolUse: &v1.ToolUse{
				Id:   c.ToolUse.GetId(),
				Name: c.ToolUse.GetName(),
			},
		}
	}

	r.resp.Message.Contents = append(r.resp.Message.Contents, content)
	r.blocks[index] = content
}

func (r *ChatEventReducer) applyDelta(delta *v1.ContentDelta) {
	index := delta.GetIndex()
	content := r.blocks[index]
	if content == nil {
		r.log.Errorf("received content delta without content start: index=%d delta=%v", index, delta)
		return
	}

	switch d := delta.Delta.(type) {
	case *v1.ContentDelta_Text:
		if text, ok := content.Content.(*v1.Content_Text); ok {
			if text.Text == nil {
				text.Text = &v1.Text{}
			}
			text.Text.Text += d.Text
		}
	case *v1.ContentDelta_Signature:
		content.Signature += d.Signature
	case *v1.ContentDelta_ToolInputText:
		if toolUse, ok := content.Content.(*v1.Content_ToolUse); ok {
			t := toolUse.ToolUse
			if len(t.Inputs) == 0 {
				t.Inputs = append(t.Inputs, &v1.ToolUse_Input{
					Input: &v1.ToolUse_Input_Text{Text: d.ToolInputText},
				})
				return
			}

			if input, ok := t.Inputs[0].Input.(*v1.ToolUse_Input_Text); ok {
				input.Text += d.ToolInputText
			}
		}
	}
}

func (r *ChatEventReducer) reduceUsage(usage *v1.Usage) {
	if usage == nil {
		return
	}

	if r.resp.Statistics == nil {
		r.resp.Statistics = &v1.Statistics{}
	}
	if r.resp.Statistics.Usage == nil {
		r.resp.Statistics.Usage = &v1.Usage{}
	}

	// The usage snapshot is cumulative, but a single event rarely carries every
	// field. Update only non-zero fields to avoid resetting previous counters.
	if usage.InputTokens != 0 {
		r.resp.Statistics.Usage.InputTokens = usage.InputTokens
	}
	if usage.OutputTokens != 0 {
		r.resp.Statistics.Usage.OutputTokens = usage.OutputTokens
	}
	if usage.CachedInputTokens != 0 {
		r.resp.Statistics.Usage.CachedInputTokens = usage.CachedInputTokens
	}
	if usage.ReasoningTokens != 0 {
		r.resp.Statistics.Usage.ReasoningTokens = usage.ReasoningTokens
	}
}
