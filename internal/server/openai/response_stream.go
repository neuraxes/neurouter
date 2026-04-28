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
	"strconv"
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

// responseStreamServer adapts the internal Chat_ChatStreamServer interface into
// the OpenAI Responses API SSE protocol. It maintains all the state required to
// translate v1.ChatResp deltas into the precisely ordered sequence of
// `response.*` events the OpenAI SDK expects.
type responseStreamServer struct {
	v1.Chat_ChatStreamServer

	ctx     context.Context
	httpCtx http.Context
	buffer  *bytes.Buffer

	responseID string
	model      string

	started        bool
	terminalSent   bool
	sequenceNumber int64
	finalStatus    v1.ChatStatus
	finalUsage     *v1.Statistics_Usage

	outputIndex   int64
	currentKind   outputItemKind
	itemStarted   bool
	currentItemID string

	// outputItems mirrors the items emitted as `response.output_item.done` so
	// that the terminal `response.completed` (or `failed`/`incomplete`) event
	// can echo back the full Response object the way the OpenAI SDK expects.
	outputItems []responseOutputItem

	// Message item state.
	currentRefusal    bool
	contentPartOpened bool
	contentIndex      int64
	accumulatedText   string

	// Function call item state.
	currentCallID   string
	currentFuncName string
	accumulatedArgs string

	// Reasoning item state.
	accumulatedSummaries     []string
	summaryPartOpened        bool
	currentSummaryIndex      int64
	currentSummaryText       string
	reasoningTextOpened      bool
	accumulatedReasoningText string
	encryptedContent         string
}

func (s *responseStreamServer) Context() context.Context {
	return s.ctx
}

func (s *responseStreamServer) write(event string, data []byte) error {
	buf := []byte("event: " + event + "\ndata: ")
	buf = append(buf, data...)
	buf = append(buf, '\n', '\n')

	if s.buffer != nil {
		s.buffer.Write(buf)
	}

	if _, err := s.httpCtx.Response().Write(buf); err != nil {
		return err
	}
	s.httpCtx.Response().(http.Flusher).Flush()
	return nil
}

// emit attaches the next sequence number to event and writes it to the wire.
// Callers MUST set event.Type prior to calling.
func (s *responseStreamServer) emit(event responseStreamEvent) error {
	event.SequenceNumber = s.sequenceNumber
	s.sequenceNumber++
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.write(event.Type, data)
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

// ensureResponseID seeds responseID either from the upstream-provided id (so it
// is stable across `response.created` ... `response.completed`) or, lacking
// that, from a freshly minted UUID.
func (s *responseStreamServer) ensureResponseID(resp *v1.ChatResp) {
	if s.responseID != "" {
		return
	}
	if resp != nil && resp.Id != "" {
		s.responseID = "resp_" + strings.ReplaceAll(resp.Id, "-", "")
	} else {
		s.responseID = "resp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	}
}

func (s *responseStreamServer) ensureStarted(resp *v1.ChatResp) error {
	if s.started {
		return nil
	}
	s.started = true
	s.ensureResponseID(resp)

	skeleton := s.responseSkeleton("in_progress")
	if err := s.emit(responseStreamEvent{
		Type:     "response.created",
		Response: skeleton,
	}); err != nil {
		return err
	}
	return s.emit(responseStreamEvent{
		Type:     "response.in_progress",
		Response: skeleton,
	})
}

// closeCurrentContentPart emits the `*.done` events that pair with a previously
// opened text/refusal content part.
func (s *responseStreamServer) closeCurrentContentPart() error {
	if !s.contentPartOpened {
		return nil
	}
	outIdx := s.outputIndex
	cIdx := s.contentIndex
	var (
		part      responseOutputContent
		doneEvent responseStreamEvent
	)
	if s.currentRefusal {
		part = responseRefusal{Type: "refusal", Refusal: s.accumulatedText}
		doneEvent = responseStreamEvent{
			Type:         "response.refusal.done",
			ItemID:       s.currentItemID,
			OutputIndex:  &outIdx,
			ContentIndex: &cIdx,
			Refusal:      s.accumulatedText,
		}
	} else {
		part = responseOutputText{Type: "output_text", Text: s.accumulatedText, Annotations: []any{}}
		doneEvent = responseStreamEvent{
			Type:         "response.output_text.done",
			ItemID:       s.currentItemID,
			OutputIndex:  &outIdx,
			ContentIndex: &cIdx,
			Text:         s.accumulatedText,
		}
	}
	if err := s.emit(doneEvent); err != nil {
		return err
	}
	if err := s.emit(responseStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       s.currentItemID,
		OutputIndex:  &outIdx,
		ContentIndex: &cIdx,
		Part:         part,
	}); err != nil {
		return err
	}
	s.contentPartOpened = false
	return nil
}

// closeCurrentSummaryPart emits the closing events for an open
// `reasoning_summary_part`, and stashes its accumulated text for later inclusion
// in the `output_item.done` payload.
func (s *responseStreamServer) closeCurrentSummaryPart() error {
	if !s.summaryPartOpened {
		return nil
	}
	outIdx := s.outputIndex
	sIdx := s.currentSummaryIndex
	if err := s.emit(responseStreamEvent{
		Type:         "response.reasoning_summary_text.done",
		ItemID:       s.currentItemID,
		OutputIndex:  &outIdx,
		SummaryIndex: &sIdx,
		Text:         s.currentSummaryText,
	}); err != nil {
		return err
	}
	if err := s.emit(responseStreamEvent{
		Type:         "response.reasoning_summary_part.done",
		ItemID:       s.currentItemID,
		OutputIndex:  &outIdx,
		SummaryIndex: &sIdx,
		Part: responseReasoningSummary{
			Type: "summary_text",
			Text: s.currentSummaryText,
		},
	}); err != nil {
		return err
	}
	s.accumulatedSummaries = append(s.accumulatedSummaries, s.currentSummaryText)
	s.summaryPartOpened = false
	s.currentSummaryText = ""
	return nil
}

// closeCurrentReasoningText emits the closing event for an open reasoning text
// content block (within a reasoning item). Note: the accumulated text is left
// untouched so closeCurrentItem can include it in the final reasoning item.
func (s *responseStreamServer) closeCurrentReasoningText() error {
	if !s.reasoningTextOpened {
		return nil
	}
	outIdx := s.outputIndex
	cIdx := s.contentIndex
	if err := s.emit(responseStreamEvent{
		Type:         "response.reasoning_text.done",
		ItemID:       s.currentItemID,
		OutputIndex:  &outIdx,
		ContentIndex: &cIdx,
		Text:         s.accumulatedReasoningText,
	}); err != nil {
		return err
	}
	s.reasoningTextOpened = false
	return nil
}

// closeCurrentItem closes whichever output item is currently open and emits the
// terminal `output_item.done` event for it.
func (s *responseStreamServer) closeCurrentItem() error {
	if !s.itemStarted {
		return nil
	}

	outIdx := s.outputIndex

	var doneItem responseOutputItem
	switch s.currentKind {
	case outputItemMessage:
		if err := s.closeCurrentContentPart(); err != nil {
			return err
		}
		var content []responseOutputContent
		if s.currentRefusal {
			content = []responseOutputContent{responseRefusal{Type: "refusal", Refusal: s.accumulatedText}}
		} else if s.accumulatedText != "" {
			content = []responseOutputContent{responseOutputText{Type: "output_text", Text: s.accumulatedText, Annotations: []any{}}}
		}
		doneItem = responseOutputMessage{
			Type:    "message",
			ID:      s.currentItemID,
			Role:    "assistant",
			Status:  "completed",
			Content: content,
		}
		if err := s.emit(responseStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: &outIdx,
			Item:        doneItem,
		}); err != nil {
			return err
		}

	case outputItemFunctionCall:
		if err := s.emit(responseStreamEvent{
			Type:        "response.function_call_arguments.done",
			ItemID:      s.currentItemID,
			OutputIndex: &outIdx,
			Arguments:   s.accumulatedArgs,
		}); err != nil {
			return err
		}
		doneItem = responseFunctionCall{
			Type:      "function_call",
			ID:        s.currentItemID,
			CallID:    s.currentCallID,
			Name:      s.currentFuncName,
			Arguments: s.accumulatedArgs,
			Status:    "completed",
		}
		if err := s.emit(responseStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: &outIdx,
			Item:        doneItem,
		}); err != nil {
			return err
		}

	case outputItemReasoning:
		// closeCurrentSummaryPart pushes the in-flight summary text into
		// accumulatedSummaries; collect a snapshot of reasoning text before
		// closeCurrentReasoningText resets it.
		if err := s.closeCurrentSummaryPart(); err != nil {
			return err
		}
		reasoningText := s.accumulatedReasoningText
		if err := s.closeCurrentReasoningText(); err != nil {
			return err
		}
		summaries := make([]responseReasoningSummary, 0, len(s.accumulatedSummaries))
		for _, t := range s.accumulatedSummaries {
			summaries = append(summaries, responseReasoningSummary{Type: "summary_text", Text: t})
		}
		var contents []responseReasoningContent
		if reasoningText != "" {
			contents = []responseReasoningContent{{Type: "reasoning_text", Text: reasoningText}}
		}
		doneItem = responseReasoning{
			Type:             "reasoning",
			ID:               s.currentItemID,
			Summary:          summaries,
			Content:          contents,
			EncryptedContent: s.encryptedContent,
		}
		if err := s.emit(responseStreamEvent{
			Type:        "response.output_item.done",
			OutputIndex: &outIdx,
			Item:        doneItem,
		}); err != nil {
			return err
		}
	}
	if doneItem != nil {
		s.outputItems = append(s.outputItems, doneItem)
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
	s.currentSummaryIndex = 0
	s.currentSummaryText = ""
	s.summaryPartOpened = false
	s.reasoningTextOpened = false
	s.accumulatedReasoningText = ""
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

	if err := s.ensureStarted(resp); err != nil {
		return err
	}

	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		for _, content := range resp.Message.Contents {
			if err := s.handleContent(content); err != nil {
				return err
			}
		}
	}

	// Stats-only response (the explicit terminal chunk).
	if resp.Message == nil && resp.Statistics != nil && resp.Statistics.Usage != nil {
		if err := s.closeCurrentItem(); err != nil {
			return err
		}
		return s.sendTerminal(resp.Status, resp.Statistics.Usage)
	}

	return nil
}

// handleContent dispatches a single delta content to the appropriate item
// handler, opening/closing items when the kind or output index changes.
func (s *responseStreamServer) handleContent(content *v1.Content) error {
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
	default:
		return nil
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
			return s.handleReasoningDelta(content, c)
		}
		return s.handleTextDelta(content, c)
	case *v1.Content_ToolUse:
		return s.handleToolUseDelta(content, c)
	}
	return nil
}

// handleTextDelta processes a single non-reasoning text/refusal delta. Refusal
// status is determined per-content via the "refusal=true" metadata flag set by
// the upstream incoming converter; this is more reliable than waiting for the
// terminal `Status == CHAT_REFUSED` since most chunks have an unspecified
// status.
func (s *responseStreamServer) handleTextDelta(content *v1.Content, c *v1.Content_Text) error {
	refusal := content.Meta("refusal") == "true"

	if !s.itemStarted {
		s.itemStarted = true
		s.currentKind = outputItemMessage
		if content.Id != "" {
			s.currentItemID = content.Id
		} else {
			s.currentItemID = "msg_" + uuid.NewString()[:12]
		}
		s.currentRefusal = refusal

		outIdx := s.outputIndex
		if err := s.emit(responseStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: &outIdx,
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
	} else if refusal != s.currentRefusal {
		// Switching between text and refusal within the same message item
		// is allowed; close the current part and open a new one.
		if err := s.closeCurrentContentPart(); err != nil {
			return err
		}
		s.contentIndex++
		s.accumulatedText = ""
		s.currentRefusal = refusal
	}

	if !s.contentPartOpened {
		s.contentPartOpened = true
		outIdx := s.outputIndex
		cIdx := s.contentIndex
		var part responseOutputContent
		if s.currentRefusal {
			part = responseRefusal{Type: "refusal", Refusal: ""}
		} else {
			part = responseOutputText{Type: "output_text", Text: "", Annotations: []any{}}
		}
		if err := s.emit(responseStreamEvent{
			Type:         "response.content_part.added",
			ItemID:       s.currentItemID,
			OutputIndex:  &outIdx,
			ContentIndex: &cIdx,
			Part:         part,
		}); err != nil {
			return err
		}
	}

	if c.Text != "" {
		s.accumulatedText += c.Text
		outIdx := s.outputIndex
		cIdx := s.contentIndex
		eventType := "response.output_text.delta"
		if s.currentRefusal {
			eventType = "response.refusal.delta"
		}
		if err := s.emit(responseStreamEvent{
			Type:         eventType,
			ItemID:       s.currentItemID,
			OutputIndex:  &outIdx,
			ContentIndex: &cIdx,
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
		if content.Id != "" {
			s.currentItemID = content.Id
		} else {
			s.currentItemID = "fc_" + uuid.NewString()[:12]
		}
		if c.ToolUse.Id != "" {
			s.currentCallID = c.ToolUse.Id
		}
		if c.ToolUse.Name != "" {
			s.currentFuncName = c.ToolUse.Name
		}

		outIdx := s.outputIndex
		if err := s.emit(responseStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: &outIdx,
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
		outIdx := s.outputIndex
		if err := s.emit(responseStreamEvent{
			Type:        "response.function_call_arguments.delta",
			ItemID:      s.currentItemID,
			OutputIndex: &outIdx,
			Delta:       delta,
		}); err != nil {
			return err
		}
	}

	return nil
}

// handleReasoningDelta processes reasoning summary and reasoning text deltas
// for a single reasoning item. It honours the "summary_index" metadata
// supplied by the upstream so that consecutive deltas of the same summary
// part are merged into one `summary_text.delta` stream and bracketed with
// `summary_part.added/done` events.
func (s *responseStreamServer) handleReasoningDelta(content *v1.Content, c *v1.Content_Text) error {
	if !s.itemStarted {
		s.itemStarted = true
		s.currentKind = outputItemReasoning
		if content.Id != "" {
			s.currentItemID = content.Id
		} else {
			s.currentItemID = "rs_" + uuid.NewString()[:12]
		}
		// New summary part counter starts before the first part is emitted.
		s.currentSummaryIndex = -1

		outIdx := s.outputIndex
		if err := s.emit(responseStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: &outIdx,
			Item: responseReasoning{
				Type:    "reasoning",
				ID:      s.currentItemID,
				Summary: []responseReasoningSummary{},
			},
		}); err != nil {
			return err
		}
	}

	if encrypted := content.Meta("encrypted"); encrypted != "" {
		s.encryptedContent = encrypted
	}

	if summary := content.Meta("summary"); summary != "" {
		summaryIdx := int64(len(s.accumulatedSummaries))
		if v := content.Meta("summary_index"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				summaryIdx = n
			}
		}
		// A new summary part requires closing the previous one (if any) and any
		// open reasoning text content that may already be in progress.
		if s.summaryPartOpened && summaryIdx != s.currentSummaryIndex {
			if err := s.closeCurrentSummaryPart(); err != nil {
				return err
			}
		}
		if s.reasoningTextOpened {
			if err := s.closeCurrentReasoningText(); err != nil {
				return err
			}
			s.contentIndex++
		}

		if !s.summaryPartOpened {
			s.summaryPartOpened = true
			s.currentSummaryIndex = summaryIdx
			outIdx := s.outputIndex
			sIdx := summaryIdx
			if err := s.emit(responseStreamEvent{
				Type:         "response.reasoning_summary_part.added",
				ItemID:       s.currentItemID,
				OutputIndex:  &outIdx,
				SummaryIndex: &sIdx,
				Part: responseReasoningSummary{
					Type: "summary_text",
					Text: "",
				},
			}); err != nil {
				return err
			}
		}

		s.currentSummaryText += summary
		outIdx := s.outputIndex
		sIdx := s.currentSummaryIndex
		if err := s.emit(responseStreamEvent{
			Type:         "response.reasoning_summary_text.delta",
			ItemID:       s.currentItemID,
			OutputIndex:  &outIdx,
			SummaryIndex: &sIdx,
			Delta:        summary,
		}); err != nil {
			return err
		}
	}

	if c.Text != "" {
		// Reasoning text takes a fresh content block; close any open summary
		// part first so the events stay ordered.
		if s.summaryPartOpened {
			if err := s.closeCurrentSummaryPart(); err != nil {
				return err
			}
		}
		if !s.reasoningTextOpened {
			s.reasoningTextOpened = true
		}
		s.accumulatedReasoningText += c.Text
		outIdx := s.outputIndex
		cIdx := s.contentIndex
		if err := s.emit(responseStreamEvent{
			Type:         "response.reasoning_text.delta",
			ItemID:       s.currentItemID,
			OutputIndex:  &outIdx,
			ContentIndex: &cIdx,
			Delta:        c.Text,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *responseStreamServer) sendTerminal(status v1.ChatStatus, usage *v1.Statistics_Usage) error {
	if s.terminalSent {
		return nil
	}
	s.terminalSent = true

	respStatus := convertStatusToResponse(status)
	eventType := "response.completed"
	switch respStatus {
	case "failed":
		eventType = "response.failed"
	case "incomplete":
		eventType = "response.incomplete"
	case "cancelled":
		// The Responses API has no dedicated cancelled event; surface the state
		// via response.failed with the cancelled status carried in the body.
		eventType = "response.failed"
	}

	skeleton := s.responseSkeleton(respStatus)
	if len(s.outputItems) > 0 {
		skeleton.Output = append([]responseOutputItem{}, s.outputItems...)
	}
	skeleton.Usage = convertUsageToResponse(usage)
	return s.emit(responseStreamEvent{
		Type:     eventType,
		Response: skeleton,
	})
}

// sendDone closes any in-flight item and emits the terminal event consistent
// with the last seen status. Safe to call multiple times.
func (s *responseStreamServer) sendDone() error {
	if !s.started {
		return nil
	}
	if err := s.closeCurrentItem(); err != nil {
		return err
	}
	return s.sendTerminal(s.finalStatus, s.finalUsage)
}

// sendError reports an internal failure to the client by ensuring the SSE
// preamble has been emitted (so the client-side parser stays in a valid
// state) and then emitting `response.failed` for the current response id.
func (s *responseStreamServer) sendError() error {
	if err := s.ensureStarted(nil); err != nil {
		return err
	}
	s.finalStatus = v1.ChatStatus_CHAT_FAILED
	return s.sendDone()
}
