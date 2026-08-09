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
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/openai/openai-go/v3/responses"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

type responsesStreamBlockKind uint8

const (
	responsesStreamBlockText responsesStreamBlockKind = iota
	responsesStreamBlockReasoning
	responsesStreamBlockTool
)

type responsesStreamBlock struct {
	kind        responsesStreamBlockKind
	outputIndex int64
	item        *responsesOutputItem

	contentIndex int
	summaryIndex int
	text         strings.Builder
}

type responsesStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx     context.Context
	httpCtx http.Context
	buffer  *bytes.Buffer

	sequenceNumber  int64
	nextOutputIndex int64
	response        responsesResponse
	usage           *v1.Usage
	blocks          map[uint32]*responsesStreamBlock
	// Finished parts remain pending until the next event determines whether
	// their output item continues.
	pendingText      *responsesStreamBlock
	pendingReasoning *responsesStreamBlock
}

func (s *responsesStreamServer) Context() context.Context {
	return s.ctx
}

func (s *responsesStreamServer) Send(event *v1.ChatEvent) error {
	if event == nil {
		return nil
	}
	s.accumulateUsage(event.Usage)

	switch e := event.Event.(type) {
	case *v1.ChatEvent_MessageStart:
		if err := s.flushPendingItems(); err != nil {
			return err
		}
		return s.handleMessageStart(event.GetId(), e.MessageStart)

	case *v1.ChatEvent_ContentStart:
		return s.handleContentStart(e.ContentStart)

	case *v1.ChatEvent_ContentDelta:
		if err := s.flushPendingItems(); err != nil {
			return err
		}
		return s.handleContentDelta(e.ContentDelta)

	case *v1.ChatEvent_ContentStop:
		if err := s.flushPendingItems(); err != nil {
			return err
		}
		return s.handleContentStop(e.ContentStop)

	case *v1.ChatEvent_ContentSnapshot:
		return s.handleContentSnapshot(e.ContentSnapshot)

	case *v1.ChatEvent_MessageStop:
		if err := s.flushPendingItems(); err != nil {
			return err
		}
		return s.handleMessageStop(e.MessageStop)
	}
	return nil
}

func (s *responsesStreamServer) handleMessageStart(requestID string, start *v1.MessageStart) error {
	responseID := start.GetId()
	if responseID == "" {
		responseID = requestID
	}
	s.response = responsesResponse{
		ID:        responseID,
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Model:     start.GetModel(),
		Status:    "in_progress",
		Output:    make([]*responsesOutputItem, 0),
	}
	if err := s.writeEvent("response.created", map[string]any{"response": &s.response}); err != nil {
		return err
	}
	return s.writeEvent("response.in_progress", map[string]any{"response": &s.response})
}

func (s *responsesStreamServer) handleContentStart(start *v1.ContentStart) error {
	if start == nil {
		return nil
	}

	switch content := start.Content.(type) {
	case *v1.ContentStart_Text:
		if start.Phase == v1.ContentPhase_CONTENT_PHASE_REASONING {
			if err := s.flushPendingText(); err != nil {
				return err
			}
			return s.startReasoningBlock(start)
		}
		if err := s.flushPendingReasoning(); err != nil {
			return err
		}
		return s.startTextBlock(start)
	case *v1.ContentStart_ToolUse:
		if err := s.flushPendingItems(); err != nil {
			return err
		}
		return s.startToolBlock(start, content.ToolUse)
	default:
		return s.flushPendingItems()
	}
}

func (s *responsesStreamServer) startTextBlock(start *v1.ContentStart) error {
	phase := convertContentPhaseToOpenAIResponsesServer(start.Phase)
	if block := s.pendingTextFor(start.GetId(), phase); block != nil {
		// The matching item is active again and must no longer be eligible for flushing.
		s.pendingText = nil
		block.contentIndex = len(block.item.Content)
		block.text.Reset()
		s.setBlock(start.Index, block)
		return s.writeTextContentPartAdded(block)
	}
	if err := s.flushPendingText(); err != nil {
		return err
	}

	outputIndex := s.allocateOutputIndex()
	itemID := start.GetId()
	if itemID == "" {
		itemID = syntheticResponsesItemID("msg", s.response.ID, int(outputIndex))
	}
	item := &responsesOutputItem{
		ID:     itemID,
		Type:   "message",
		Status: "in_progress",
		Role:   "assistant",
		Phase:  phase,
	}
	block := &responsesStreamBlock{
		kind:        responsesStreamBlockText,
		outputIndex: outputIndex,
		item:        item,
	}
	s.setBlock(start.Index, block)
	s.response.Output = append(s.response.Output, item)

	if err := s.writeEvent("response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item":         item,
	}); err != nil {
		return err
	}
	return s.writeTextContentPartAdded(block)
}

func (s *responsesStreamServer) writeTextContentPartAdded(block *responsesStreamBlock) error {
	return s.writeEvent("response.content_part.added", map[string]any{
		"output_index":  block.outputIndex,
		"item_id":       block.item.ID,
		"content_index": block.contentIndex,
		"part": responses.ResponseOutputText{
			Type: "output_text",
			Text: "",
		},
	})
}

func (s *responsesStreamServer) startReasoningBlock(start *v1.ContentStart) error {
	if block := s.pendingReasoningFor(start.GetId()); block != nil {
		s.pendingReasoning = nil
		block.summaryIndex = len(block.item.Summary)
		block.text.Reset()
		s.setBlock(start.Index, block)
		return s.writeReasoningSummaryPartAdded(block)
	}
	if err := s.flushPendingReasoning(); err != nil {
		return err
	}

	outputIndex := s.allocateOutputIndex()
	itemID := start.GetId()
	if itemID == "" {
		itemID = syntheticResponsesItemID("rs", s.response.ID, int(outputIndex))
	}
	item := &responsesOutputItem{
		ID:     itemID,
		Type:   "reasoning",
		Status: "in_progress",
	}
	block := &responsesStreamBlock{
		kind:        responsesStreamBlockReasoning,
		outputIndex: outputIndex,
		item:        item,
	}
	s.setBlock(start.Index, block)
	s.response.Output = append(s.response.Output, item)

	if err := s.writeEvent("response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item":         item,
	}); err != nil {
		return err
	}
	return s.writeReasoningSummaryPartAdded(block)
}

func (s *responsesStreamServer) writeReasoningSummaryPartAdded(block *responsesStreamBlock) error {
	return s.writeEvent("response.reasoning_summary_part.added", map[string]any{
		"output_index":  block.outputIndex,
		"item_id":       block.item.ID,
		"summary_index": block.summaryIndex,
		"part": responses.ResponseReasoningItemSummary{
			Type: "summary_text",
			Text: "",
		},
	})
}

func (s *responsesStreamServer) startToolBlock(start *v1.ContentStart, tool *v1.ToolUseStart) error {
	outputIndex := s.allocateOutputIndex()
	itemID := start.GetId()
	if itemID == "" {
		itemID = syntheticResponsesItemID("fc", s.response.ID, int(outputIndex))
	}
	item := &responsesOutputItem{
		ID:        itemID,
		Type:      "function_call",
		Status:    "in_progress",
		CallID:    tool.GetId(),
		Name:      tool.GetName(),
		Arguments: new(""),
	}
	block := &responsesStreamBlock{
		kind:        responsesStreamBlockTool,
		outputIndex: outputIndex,
		item:        item,
	}
	s.setBlock(start.Index, block)
	s.response.Output = append(s.response.Output, item)
	return s.writeEvent("response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item":         item,
	})
}

func (s *responsesStreamServer) handleContentDelta(delta *v1.ContentDelta) error {
	if delta == nil {
		return nil
	}
	block := s.getBlock(delta.Index)
	if block == nil {
		return nil
	}

	switch value := delta.Delta.(type) {
	case *v1.ContentDelta_Text:
		block.text.WriteString(value.Text)
		if block.kind == responsesStreamBlockReasoning {
			return s.writeEvent("response.reasoning_summary_text.delta", map[string]any{
				"item_id":       block.item.ID,
				"output_index":  block.outputIndex,
				"summary_index": block.summaryIndex,
				"delta":         value.Text,
			})
		}
		return s.writeEvent("response.output_text.delta", map[string]any{
			"item_id":       block.item.ID,
			"output_index":  block.outputIndex,
			"content_index": block.contentIndex,
			"delta":         value.Text,
		})

	case *v1.ContentDelta_ToolInputText:
		block.text.WriteString(value.ToolInputText)
		return s.writeEvent("response.function_call_arguments.delta", map[string]any{
			"output_index": block.outputIndex,
			"item_id":      block.item.ID,
			"delta":        value.ToolInputText,
		})

	case *v1.ContentDelta_Signature:
		return nil
	}
	return nil
}

func (s *responsesStreamServer) handleContentStop(stop *v1.ContentStop) error {
	if stop == nil {
		return nil
	}
	block := s.getBlock(stop.Index)
	if block == nil {
		return nil
	}
	s.deleteBlock(stop.Index)

	switch block.kind {
	case responsesStreamBlockText:
		return s.finishTextBlock(block)
	case responsesStreamBlockReasoning:
		return s.finishReasoningBlock(block)
	case responsesStreamBlockTool:
		return s.finishToolBlock(block)
	default:
		return nil
	}
}

func (s *responsesStreamServer) finishTextBlock(block *responsesStreamBlock) error {
	text := block.text.String()
	part := responses.ResponseOutputText{Type: "output_text", Text: text}
	block.item.Content = append(block.item.Content, part)

	if err := s.writeEvent("response.output_text.done", map[string]any{
		"output_index":  block.outputIndex,
		"item_id":       block.item.ID,
		"content_index": block.contentIndex,
		"text":          text,
	}); err != nil {
		return err
	}
	if err := s.writeEvent("response.content_part.done", map[string]any{
		"output_index":  block.outputIndex,
		"item_id":       block.item.ID,
		"content_index": block.contentIndex,
		"part":          part,
	}); err != nil {
		return err
	}

	// Defer output_item.done so a following text part can reuse this message item.
	s.pendingText = block
	return nil
}

func (s *responsesStreamServer) finishReasoningBlock(block *responsesStreamBlock) error {
	text := block.text.String()
	summary := responses.ResponseReasoningItemSummary{Type: "summary_text", Text: text}
	block.item.Summary = append(block.item.Summary, summary)

	if err := s.writeEvent("response.reasoning_summary_text.done", map[string]any{
		"output_index":  block.outputIndex,
		"item_id":       block.item.ID,
		"summary_index": block.summaryIndex,
		"text":          text,
	}); err != nil {
		return err
	}
	if err := s.writeEvent("response.reasoning_summary_part.done", map[string]any{
		"output_index":  block.outputIndex,
		"item_id":       block.item.ID,
		"summary_index": block.summaryIndex,
		"part":          summary,
	}); err != nil {
		return err
	}

	s.pendingReasoning = block
	return nil
}

func (s *responsesStreamServer) finishToolBlock(block *responsesStreamBlock) error {
	arguments := block.text.String()
	block.item.Arguments = &arguments
	if err := s.writeEvent("response.function_call_arguments.done", map[string]any{
		"output_index": block.outputIndex,
		"item_id":      block.item.ID,
		"arguments":    arguments,
	}); err != nil {
		return err
	}
	block.item.Status = "completed"
	return s.writeOutputItemDone(block)
}

func (s *responsesStreamServer) handleContentSnapshot(content *v1.Content) error {
	if err := s.flushPendingText(); err != nil {
		return err
	}
	if content == nil || content.Phase != v1.ContentPhase_CONTENT_PHASE_REASONING || content.GetOpaque() == "" {
		return s.flushPendingReasoning()
	}
	if block := s.pendingReasoningFor(content.GetId()); block != nil {
		block.item.EncryptedContent = content.GetOpaque()
		return s.flushPendingReasoning()
	}
	if err := s.flushPendingReasoning(); err != nil {
		return err
	}

	outputIndex := s.allocateOutputIndex()
	itemID := content.GetId()
	if itemID == "" {
		itemID = syntheticResponsesItemID("rs", s.response.ID, int(outputIndex))
	}
	item := &responsesOutputItem{
		ID:               itemID,
		Type:             "reasoning",
		Status:           "in_progress",
		EncryptedContent: content.GetOpaque(),
	}
	s.response.Output = append(s.response.Output, item)
	block := &responsesStreamBlock{
		kind:        responsesStreamBlockReasoning,
		outputIndex: outputIndex,
		item:        item,
	}
	if err := s.writeEvent("response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item":         item,
	}); err != nil {
		return err
	}
	item.Status = "completed"
	return s.writeOutputItemDone(block)
}

func (s *responsesStreamServer) handleMessageStop(stop *v1.MessageStop) error {
	s.response.Status, s.response.IncompleteDetails, s.response.Error = convertStatusToOpenAIResponses(stop.GetStatus())
	s.response.Usage = convertUsageToOpenAIResponses(s.usage)

	eventType := "response.completed"
	switch s.response.Status {
	case "incomplete":
		eventType = "response.incomplete"
	case "failed", "cancelled":
		eventType = "response.failed"
	}
	return s.writeEvent(eventType, map[string]any{"response": &s.response})
}

// pendingReasoningFor returns the reasoning item whose completion is still held
// back and that the given content id continues. Content without an id continues
// whatever item is still open, which is how upstreams that do not expose
// reasoning item ids keep a summary and its encrypted snapshot in one item.
func (s *responsesStreamServer) pendingReasoningFor(itemID string) *responsesStreamBlock {
	if s.pendingReasoning == nil {
		return nil
	}
	if itemID == "" || s.pendingReasoning.item.ID == itemID {
		return s.pendingReasoning
	}
	return nil
}

func (s *responsesStreamServer) flushPendingReasoning() error {
	if s.pendingReasoning == nil {
		return nil
	}
	block := s.pendingReasoning
	s.pendingReasoning = nil
	block.item.Status = "completed"
	return s.writeOutputItemDone(block)
}

func (s *responsesStreamServer) pendingTextFor(itemID, phase string) *responsesStreamBlock {
	if itemID == "" || s.pendingText == nil {
		return nil
	}
	if s.pendingText.item.ID == itemID && s.pendingText.item.Phase == phase {
		return s.pendingText
	}
	return nil
}

func (s *responsesStreamServer) flushPendingText() error {
	if s.pendingText == nil {
		return nil
	}
	block := s.pendingText
	s.pendingText = nil
	block.item.Status = "completed"
	return s.writeOutputItemDone(block)
}

func (s *responsesStreamServer) flushPendingItems() error {
	if err := s.flushPendingReasoning(); err != nil {
		return err
	}
	return s.flushPendingText()
}

func (s *responsesStreamServer) writeOutputItemDone(block *responsesStreamBlock) error {
	return s.writeEvent("response.output_item.done", map[string]any{
		"output_index": block.outputIndex,
		"item":         block.item,
	})
}

func (s *responsesStreamServer) writeEvent(eventType string, payload map[string]any) error {
	payload["type"] = eventType
	payload["sequence_number"] = s.sequenceNumber
	s.sequenceNumber++

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	data := make([]byte, 0, len(eventType)+len(encoded)+16)
	data = append(data, "event: "...)
	data = append(data, eventType...)
	data = append(data, '\n')
	data = append(data, "data: "...)
	data = append(data, encoded...)
	data = append(data, '\n', '\n')
	if s.buffer != nil {
		_, _ = s.buffer.Write(data)
	}
	if _, err = s.httpCtx.Response().Write(data); err != nil {
		return err
	}
	if flusher, ok := s.httpCtx.Response().(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func (s *responsesStreamServer) allocateOutputIndex() int64 {
	index := s.nextOutputIndex
	s.nextOutputIndex++
	return index
}

func (s *responsesStreamServer) accumulateUsage(usage *v1.Usage) {
	if usage == nil {
		return
	}
	if s.usage == nil {
		s.usage = &v1.Usage{}
	}
	if usage.InputTokens != 0 {
		s.usage.InputTokens = usage.InputTokens
	}
	if usage.OutputTokens != 0 {
		s.usage.OutputTokens = usage.OutputTokens
	}
	if usage.CachedInputTokens != 0 {
		s.usage.CachedInputTokens = usage.CachedInputTokens
	}
	if usage.ReasoningTokens != 0 {
		s.usage.ReasoningTokens = usage.ReasoningTokens
	}
}

func (s *responsesStreamServer) getBlock(index uint32) *responsesStreamBlock {
	if s.blocks == nil {
		return nil
	}
	return s.blocks[index]
}

func (s *responsesStreamServer) setBlock(index uint32, block *responsesStreamBlock) {
	if s.blocks == nil {
		s.blocks = make(map[uint32]*responsesStreamBlock)
	}
	s.blocks[index] = block
}

func (s *responsesStreamServer) deleteBlock(index uint32) {
	if s.blocks != nil {
		delete(s.blocks, index)
	}
}
