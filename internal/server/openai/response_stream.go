// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"strings"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/uuid"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

type outputItemKind int

const (
	outputItemNone outputItemKind = iota
	outputItemMessage
	outputItemReasoning
	outputItemFunctionCall
)

type responseStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx     context.Context
	httpCtx http.Context
	buffer  *bytes.Buffer

	responseID string
	model      string

	started           bool
	outputIndex       int64
	contentIndex      int64
	currentKind       outputItemKind
	itemStarted       bool
	contentPartOpened bool
	terminalSent      bool
	finalStatus       v1.ChatStatus
	finalUsage        *v1.Statistics_Usage

	accumulatedText      string
	accumulatedArgs      string
	accumulatedSummaries []responseReasoningSummary
	currentItemID        string
	currentCallID        string
	currentFuncName      string
	encryptedContent     string
	currentRefusal       bool
}

func (s *responseStreamServer) Context() context.Context {
	return s.ctx
}

func (s *responseStreamServer) sendEvent(event string, data []byte) error {
	buf := []byte("event: " + event + "\ndata: ")
	buf = append(buf, data...)
	buf = append(buf, '\n', '\n')

	if s.buffer != nil {
		s.buffer.Write(buf)
	}

	_, err := s.httpCtx.Response().Write(buf)
	if err != nil {
		return err
	}
	s.httpCtx.Response().(http.Flusher).Flush()
	return nil
}

func (s *responseStreamServer) sendJSONEvent(event string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.sendEvent(event, data)
}

func (s *responseStreamServer) responseSkeleton(status string) *responseObject {
	return &responseObject{
		ID:     s.responseID,
		Object: "response",
		Model:  s.model,
		Status: status,
		Output: []responseOutputItem{},
	}
}

func (s *responseStreamServer) ensureStarted() error {
	if s.started {
		return nil
	}
	s.started = true
	s.responseID = "resp_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	skeleton := s.responseSkeleton("in_progress")
	if err := s.sendJSONEvent("response.created", responseStreamEvent{
		Type:     "response.created",
		Response: skeleton,
	}); err != nil {
		return err
	}
	return s.sendJSONEvent("response.in_progress", responseStreamEvent{
		Type:     "response.in_progress",
		Response: skeleton,
	})
}

func (s *responseStreamServer) closeCurrentItem() error {
	if !s.itemStarted {
		return nil
	}

	switch s.currentKind {
	case outputItemMessage:
		if s.contentPartOpened {
			idx := s.outputIndex
			cidx := s.contentIndex
			part := responseOutputText{
				Type:        "output_text",
				Text:        s.accumulatedText,
				Annotations: []any{},
			}
			if err := s.sendJSONEvent("response.output_text.done", responseStreamEvent{
				Type:         "response.output_text.done",
				OutputIndex:  &idx,
				ContentIndex: &cidx,
				Text:         s.accumulatedText,
			}); err != nil {
				return err
			}
			if err := s.sendJSONEvent("response.content_part.done", responseStreamEvent{
				Type:         "response.content_part.done",
				OutputIndex:  &idx,
				ContentIndex: &cidx,
				Part:         part,
			}); err != nil {
				return err
			}
			s.contentPartOpened = false
		}
		var outputContent responseOutputContent = responseOutputText{
			Type:        "output_text",
			Text:        s.accumulatedText,
			Annotations: []any{},
		}
		if s.currentRefusal {
			outputContent = responseRefusal{
				Type:    "refusal",
				Refusal: s.accumulatedText,
			}
		}
		if err := s.sendJSONEvent("response.output_item.done", responseStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: &s.outputIndex,
			Item: responseOutputMessage{
				Type:   "message",
				ID:     s.currentItemID,
				Role:   "assistant",
				Status: "completed",
				Content: []responseOutputContent{
					outputContent,
				},
			},
		}); err != nil {
			return err
		}

	case outputItemFunctionCall:
		idx := s.outputIndex
		if err := s.sendJSONEvent("response.function_call_arguments.done", responseStreamEvent{
			Type:        "response.function_call_arguments.done",
			OutputIndex: &idx,
			ItemID:      s.currentItemID,
			Arguments:   s.accumulatedArgs,
		}); err != nil {
			return err
		}
		if err := s.sendJSONEvent("response.output_item.done", responseStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: &idx,
			Item: responseFunctionCall{
				Type:      "function_call",
				ID:        s.currentItemID,
				CallID:    s.currentCallID,
				Name:      s.currentFuncName,
				Arguments: s.accumulatedArgs,
				Status:    "completed",
			},
		}); err != nil {
			return err
		}

	case outputItemReasoning:
		if err := s.sendJSONEvent("response.output_item.done", responseStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: &s.outputIndex,
			Item: responseReasoning{
				Type:             "reasoning",
				ID:               s.currentItemID,
				Summary:          s.accumulatedSummaries,
				EncryptedContent: s.encryptedContent,
			},
		}); err != nil {
			return err
		}
	}

	s.itemStarted = false
	s.outputIndex++
	s.contentIndex = 0
	s.accumulatedText = ""
	s.accumulatedArgs = ""
	s.accumulatedSummaries = nil
	s.currentItemID = ""
	s.currentCallID = ""
	s.currentFuncName = ""
	s.encryptedContent = ""
	s.currentRefusal = false
	return nil
}

func (s *responseStreamServer) Send(resp *v1.ChatResp) error {
	if resp.Model != "" {
		s.model = resp.Model
	}
	if resp.Status != v1.ChatStatus_CHAT_IN_PROGRESS {
		s.finalStatus = resp.Status
	}
	if resp.Statistics != nil {
		s.finalUsage = resp.Statistics.Usage
	}

	if err := s.ensureStarted(); err != nil {
		return err
	}

	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		for _, content := range resp.Message.Contents {
			kind := outputItemNone
			switch content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					kind = outputItemReasoning
				} else {
					kind = outputItemMessage
				}
			case *v1.Content_ToolUse:
				kind = outputItemFunctionCall
			}

			if s.itemStarted && s.currentKind != kind {
				if err := s.closeCurrentItem(); err != nil {
					return err
				}
			}
			if content.Index != nil {
				newIdx := int64(*content.Index)
				if s.itemStarted && s.outputIndex != newIdx {
					if err := s.closeCurrentItem(); err != nil {
						return err
					}
				}
				s.outputIndex = newIdx
			}

			switch c := content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					if err := s.handleReasoningDelta(content, c); err != nil {
						return err
					}
				} else {
					if err := s.handleTextDelta(resp, c); err != nil {
						return err
					}
				}
			case *v1.Content_ToolUse:
				if err := s.handleToolUseDelta(content, c); err != nil {
					return err
				}
			}
		}
	}

	// Stats-only response (final chunk)
	if resp.Message == nil && resp.Statistics != nil && resp.Statistics.Usage != nil {
		if err := s.closeCurrentItem(); err != nil {
			return err
		}
		return s.sendTerminal(resp.Status, resp.Statistics.Usage)
	}

	return nil
}

func (s *responseStreamServer) handleTextDelta(resp *v1.ChatResp, c *v1.Content_Text) error {
	if !s.itemStarted {
		s.itemStarted = true
		s.currentKind = outputItemMessage
		s.currentItemID = "msg_" + uuid.NewString()[:12]
		s.currentRefusal = resp.Status == v1.ChatStatus_CHAT_REFUSED

		idx := s.outputIndex
		if err := s.sendJSONEvent("response.output_item.added", responseStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: &idx,
			Item: responseOutputMessage{
				Type:    "message",
				ID:      s.currentItemID,
				Role:    "assistant",
				Status:  "in_progress",
				Content: []responseOutputContent{},
			},
		}); err != nil {
			return err
		}
	}

	if !s.contentPartOpened {
		s.contentPartOpened = true
		idx := s.outputIndex
		cidx := s.contentIndex
		if err := s.sendJSONEvent("response.content_part.added", responseStreamEvent{
			Type:         "response.content_part.added",
			OutputIndex:  &idx,
			ContentIndex: &cidx,
			Part: responseOutputText{
				Type:        "output_text",
				Text:        "",
				Annotations: []any{},
			},
		}); err != nil {
			return err
		}
	}

	if c.Text != "" {
		s.accumulatedText += c.Text
		idx := s.outputIndex
		cidx := s.contentIndex
		eventType := "response.output_text.delta"
		if s.currentRefusal {
			eventType = "response.refusal.delta"
		}
		if err := s.sendJSONEvent(eventType, responseStreamEvent{
			Type:         eventType,
			OutputIndex:  &idx,
			ContentIndex: &cidx,
			Delta:        c.Text,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *responseStreamServer) handleToolUseDelta(content *v1.Content, c *v1.Content_ToolUse) error {
	if !s.itemStarted {
		s.itemStarted = true
		s.currentKind = outputItemFunctionCall
		s.currentItemID = "fc_" + uuid.NewString()[:12]
		if c.ToolUse.Id != "" {
			s.currentCallID = c.ToolUse.Id
		}
		if c.ToolUse.Name != "" {
			s.currentFuncName = c.ToolUse.Name
		}

		idx := s.outputIndex
		if err := s.sendJSONEvent("response.output_item.added", responseStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: &idx,
			Item: responseFunctionCall{
				Type:      "function_call",
				ID:        s.currentItemID,
				CallID:    s.currentCallID,
				Name:      s.currentFuncName,
				Arguments: "",
				Status:    "in_progress",
			},
		}); err != nil {
			return err
		}
	}

	if c.ToolUse.Id != "" && s.currentCallID == "" {
		s.currentCallID = c.ToolUse.Id
	}
	if c.ToolUse.Name != "" && s.currentFuncName == "" {
		s.currentFuncName = c.ToolUse.Name
	}

	if len(c.ToolUse.Inputs) > 0 {
		delta := c.ToolUse.GetTextualInput()
		s.accumulatedArgs += delta
		idx := s.outputIndex
		if err := s.sendJSONEvent("response.function_call_arguments.delta", responseStreamEvent{
			Type:        "response.function_call_arguments.delta",
			OutputIndex: &idx,
			ItemID:      s.currentItemID,
			Delta:       delta,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *responseStreamServer) handleReasoningDelta(content *v1.Content, c *v1.Content_Text) error {
	if !s.itemStarted {
		s.itemStarted = true
		s.currentKind = outputItemReasoning
		s.currentItemID = "rs_" + uuid.NewString()[:12]

		idx := s.outputIndex
		if err := s.sendJSONEvent("response.output_item.added", responseStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: &idx,
			Item: responseReasoning{
				Type:    "reasoning",
				ID:      s.currentItemID,
				Summary: []responseReasoningSummary{},
			},
		}); err != nil {
			return err
		}
	}

	if summary := content.Meta("summary"); summary != "" {
		summaryIdx := int64(len(s.accumulatedSummaries))
		s.accumulatedSummaries = append(s.accumulatedSummaries, responseReasoningSummary{
			Type: "summary_text",
			Text: summary,
		})
		idx := s.outputIndex
		if err := s.sendJSONEvent("response.reasoning_summary_text.delta", responseStreamEvent{
			Type:         "response.reasoning_summary_text.delta",
			OutputIndex:  &idx,
			ItemID:       s.currentItemID,
			SummaryIndex: &summaryIdx,
			Delta:        summary,
		}); err != nil {
			return err
		}
	}

	if c.Text != "" {
		idx := s.outputIndex
		if err := s.sendJSONEvent("response.reasoning_text.delta", responseStreamEvent{
			Type:        "response.reasoning_text.delta",
			OutputIndex: &idx,
			ItemID:      s.currentItemID,
			Delta:       c.Text,
		}); err != nil {
			return err
		}
	}

	if encrypted := content.Meta("encrypted"); encrypted != "" {
		s.encryptedContent = encrypted
	}

	return nil
}

func (s *responseStreamServer) sendTerminal(status v1.ChatStatus, usage *v1.Statistics_Usage) error {
	if s.terminalSent {
		return nil
	}
	s.terminalSent = true

	eventType := "response.completed"
	switch convertStatusToResponse(status) {
	case "failed":
		eventType = "response.failed"
	case "incomplete":
		eventType = "response.incomplete"
	}

	skeleton := s.responseSkeleton(convertStatusToResponse(status))
	skeleton.Usage = convertUsageToResponse(usage)
	return s.sendJSONEvent(eventType, responseStreamEvent{
		Type:     eventType,
		Response: skeleton,
	})
}

func (s *responseStreamServer) sendDone() error {
	if !s.started {
		return nil
	}
	if err := s.closeCurrentItem(); err != nil {
		return err
	}
	return s.sendTerminal(s.finalStatus, s.finalUsage)
}
