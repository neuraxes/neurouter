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
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	openaishared "github.com/openai/openai-go/v3/shared"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

func (r *upstream) convertRequestToOpenAIResponses(req *entity.ChatRequest) responses.ResponseNewParams {
	openAIReq := responses.ResponseNewParams{
		Model:    openaishared.ResponsesModel(req.Model),
		Store:    openai.Opt(false),
		Metadata: openaishared.Metadata(req.Metadata),
	}
	if req.Session != "" {
		openAIReq.PromptCacheKey = openai.Opt(req.Session)
	}

	if req.Config != nil {
		r.convertGenerationConfigToOpenAIResponses(req.Config, &openAIReq)
	}

	for _, message := range req.Messages {
		openAIReq.Input.OfInputItemList = append(
			openAIReq.Input.OfInputItemList,
			r.convertMessageToOpenAIResponses(message)...,
		)
	}

	for _, tool := range req.Tools {
		switch t := tool.GetTool().(type) {
		case *v1.Tool_Function_:
			function := t.Function
			openAITool := &responses.FunctionToolParam{
				Name:       function.GetName(),
				Parameters: convertSchemaToOpenAIParameters(function.GetInputSchema()),
				Strict:     openai.Opt(false),
			}
			if function.GetDescription() != "" {
				openAITool.Description = openai.Opt(function.GetDescription())
			}
			openAIReq.Tools = append(openAIReq.Tools, responses.ToolUnionParam{OfFunction: openAITool})
		default:
			r.log.Error("unsupported tool", "tool", t)
		}
	}

	return openAIReq
}

func (r *upstream) convertGenerationConfigToOpenAIResponses(config *v1.GenerationConfig, req *responses.ResponseNewParams) {
	if config.MaxTokens != nil {
		req.MaxOutputTokens = openai.Opt(config.GetMaxTokens())
	}
	if config.Temperature != nil {
		req.Temperature = openai.Opt(float64(config.GetTemperature()))
	}
	if config.TopP != nil {
		req.TopP = openai.Opt(float64(config.GetTopP()))
	}

	if reasoning := config.ReasoningConfig; reasoning != nil && reasoning.Effort > v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED {
		req.Reasoning.Effort = convertReasoningEffortToOpenAI(reasoning.Effort)
		if reasoning.Effort >= v1.ReasoningEffort_REASONING_EFFORT_MINIMAL && !r.config.ResponsesUseRawReasoning {
			req.Reasoning.Summary = openaishared.ReasoningSummaryAuto
		}
	}

	switch grammar := config.Grammar.(type) {
	case *v1.GenerationConfig_PresetGrammar:
		if grammar.PresetGrammar == "json_object" {
			req.Text.Format.OfJSONObject = &openaishared.ResponseFormatJSONObjectParam{}
		}
	case *v1.GenerationConfig_Schema:
		req.Text.Format.OfJSONSchema = &responses.ResponseFormatTextJSONSchemaConfigParam{
			Name:   "custom_schema",
			Schema: convertSchemaToOpenAIParameters(grammar.Schema),
			Strict: openai.Opt(true),
		}
	}
}

func (r *upstream) convertMessageToOpenAIResponses(message *v1.Message) []responses.ResponseInputItemUnionParam {
	if message == nil {
		return nil
	}

	role, ok := convertRoleToOpenAIResponses(message.Role)
	if !ok {
		r.log.Error("unsupported message role", "role", message.Role)
		return nil
	}

	var items []responses.ResponseInputItemUnionParam

	messagePhase := responses.EasyInputMessagePhase("")
	var messageContent responses.ResponseInputMessageContentListParam

	reasoningItems := make(map[string]*responses.ResponseReasoningItemParam)
	// For servers that don't support item ids.
	var openReasoningItemWithoutID *responses.ResponseReasoningItemParam

	flushMessage := func() {
		if len(messageContent) == 0 {
			return
		}
		var content responses.EasyInputMessageContentUnionParam
		if message.Role == v1.Role_ROLE_MODEL {
			var text string
			for _, part := range messageContent {
				if part.OfInputText != nil {
					text += part.OfInputText.Text
				}
			}
			content.OfString = openai.Opt(text)
		} else {
			content.OfInputItemContentList = messageContent
		}
		items = append(items, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Type:    responses.EasyInputMessageTypeMessage,
				Role:    role,
				Phase:   messagePhase,
				Content: content,
			},
		})
		messagePhase = ""
		messageContent = nil
	}
	reasoningItem := func(id string) *responses.ResponseReasoningItemParam {
		if id == "" {
			if openReasoningItemWithoutID != nil {
				return openReasoningItemWithoutID
			}

			item := &responses.ResponseReasoningItemParam{}
			items = append(items, responses.ResponseInputItemUnionParam{OfReasoning: item})
			openReasoningItemWithoutID = item
			return item
		}

		openReasoningItemWithoutID = nil
		if item := reasoningItems[id]; item != nil {
			return item
		}

		item := &responses.ResponseReasoningItemParam{ID: id}
		reasoningItems[id] = item
		items = append(items, responses.ResponseInputItemUnionParam{OfReasoning: item})
		return item
	}

	for _, content := range message.Contents {
		if content == nil {
			continue
		}

		switch c := content.Content.(type) {
		case *v1.Content_Text:
			if content.Phase == v1.ContentPhase_CONTENT_PHASE_REASONING {
				flushMessage()
				item := reasoningItem(content.GetId())
				if r.config.ResponsesUseRawReasoning {
					item.Content = append(item.Content, responses.ResponseReasoningItemContentParam{Text: c.Text.GetText()})
				} else {
					item.Summary = append(item.Summary, responses.ResponseReasoningItemSummaryParam{Text: c.Text.GetText()})
				}
				continue
			}

			openReasoningItemWithoutID = nil
			phase := convertContentPhaseToOpenAIResponses(message.Role, content.Phase)
			if len(messageContent) > 0 && phase != messagePhase {
				flushMessage()
			}
			messagePhase = phase
			messageContent = append(messageContent, responses.ResponseInputContentParamOfInputText(c.Text.GetText()))

		case *v1.Content_Image:
			openReasoningItemWithoutID = nil
			if message.Role == v1.Role_ROLE_MODEL {
				r.log.Error("unsupported image in model message")
				continue
			}
			messageContent = append(messageContent, responses.ResponseInputContentUnionParam{
				OfInputImage: &responses.ResponseInputImageParam{
					ImageURL: openai.Opt(convertImageToOpenAIURL(c.Image)),
				},
			})

		case *v1.Content_ToolUse:
			openReasoningItemWithoutID = nil
			flushMessage()
			items = append(items, responses.ResponseInputItemParamOfFunctionCall(
				c.ToolUse.GetTextualInput(),
				c.ToolUse.GetId(),
				c.ToolUse.GetName(),
			))

		case *v1.Content_ToolResult:
			openReasoningItemWithoutID = nil
			flushMessage()
			items = append(items, responses.ResponseInputItemParamOfFunctionCallOutput(
				c.ToolResult.GetId(),
				c.ToolResult.GetTextualOutput(),
			))

		case *v1.Content_Opaque:
			if content.Phase != v1.ContentPhase_CONTENT_PHASE_REASONING {
				openReasoningItemWithoutID = nil
				r.log.Error("unsupported non-reasoning opaque content")
				continue
			}
			flushMessage()
			reasoningItem(content.GetId()).EncryptedContent = openai.Opt(c.Opaque)

		default:
			openReasoningItemWithoutID = nil
			r.log.Error("unsupported content", "content", c)
		}
	}

	flushMessage()
	return items
}

func convertRoleToOpenAIResponses(role v1.Role) (responses.EasyInputMessageRole, bool) {
	switch role {
	case v1.Role_ROLE_SYSTEM:
		return responses.EasyInputMessageRoleSystem, true
	case v1.Role_ROLE_USER:
		return responses.EasyInputMessageRoleUser, true
	case v1.Role_ROLE_MODEL:
		return responses.EasyInputMessageRoleAssistant, true
	default:
		return "", false
	}
}

func convertContentPhaseToOpenAIResponses(role v1.Role, phase v1.ContentPhase) responses.EasyInputMessagePhase {
	if role == v1.Role_ROLE_MODEL && phase == v1.ContentPhase_CONTENT_PHASE_OUTCOME {
		return responses.EasyInputMessagePhaseFinalAnswer
	}
	return ""
}

func (r *upstream) convertResponseFromOpenAIResponses(req *entity.ChatRequest, openAIResp *responses.Response) *entity.ChatResponse {
	resp := &entity.ChatResponse{
		Id:         req.GetId(),
		Model:      string(openAIResp.Model),
		Status:     convertStatusFromOpenAIResponses(openAIResp),
		Statistics: convertStatisticsFromOpenAIResponses(&openAIResp.Usage),
		Message: &v1.Message{
			Id:   openAIResp.ID,
			Role: v1.Role_ROLE_MODEL,
		},
	}

	for _, item := range openAIResp.Output {
		resp.Message.Contents = append(resp.Message.Contents, convertOutputItemFromOpenAIResponses(item)...)
	}
	return resp
}

func convertOutputItemFromOpenAIResponses(item responses.ResponseOutputItemUnion) []*v1.Content {
	switch item.Type {
	case "reasoning":
		return convertReasoningItemFromOpenAIResponses(item.AsReasoning())
	case "message":
		return convertMessageItemFromOpenAIResponses(item.AsMessage())
	case "function_call":
		call := item.AsFunctionCall()
		return []*v1.Content{{
			Id: call.ID,
			Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
				Id:   call.CallID,
				Name: call.Name,
				Inputs: []*v1.ToolUse_Input{{
					Input: &v1.ToolUse_Input_Text{Text: call.Arguments},
				}},
			}},
		}}
	default:
		return nil
	}
}

func convertReasoningItemFromOpenAIResponses(item responses.ResponseReasoningItem) []*v1.Content {
	contents := make([]*v1.Content, 0, len(item.Summary)+len(item.Content)+1)
	for _, summary := range item.Summary {
		contents = append(contents, &v1.Content{
			Id:      item.ID,
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: v1.NewTextContent(summary.Text),
		})
	}
	for _, reasoning := range item.Content {
		contents = append(contents, &v1.Content{
			Id:      item.ID,
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: v1.NewTextContent(reasoning.Text),
		})
	}
	if item.EncryptedContent != "" {
		contents = append(contents, &v1.Content{
			Id:      item.ID,
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: &v1.Content_Opaque{Opaque: item.EncryptedContent},
		})
	}
	return contents
}

func convertMessageItemFromOpenAIResponses(item responses.ResponseOutputMessage) []*v1.Content {
	phase := convertOutputMessagePhaseFromOpenAIResponses(item.Phase)

	var contents []*v1.Content
	for _, part := range item.Content {
		if part.Type != "output_text" {
			continue
		}
		contents = append(contents, &v1.Content{
			Id:      item.ID,
			Phase:   phase,
			Content: v1.NewTextContent(part.Text),
		})
	}
	return contents
}

func convertStatusFromOpenAIResponses(resp *responses.Response) v1.ChatStatus {
	if resp == nil {
		return v1.ChatStatus_CHAT_STATUS_IN_PROGRESS
	}

	switch resp.Status {
	case responses.ResponseStatusCompleted:
		for _, item := range resp.Output {
			if item.Type == "function_call" {
				return v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE
			}
		}
		return v1.ChatStatus_CHAT_STATUS_COMPLETED
	case responses.ResponseStatusIncomplete:
		switch resp.IncompleteDetails.Reason {
		case "max_output_tokens":
			return v1.ChatStatus_CHAT_STATUS_REACHED_TOKEN_LIMIT
		case "content_filter":
			return v1.ChatStatus_CHAT_STATUS_REFUSED
		default:
			return v1.ChatStatus_CHAT_STATUS_FAILED
		}
	case responses.ResponseStatusFailed:
		return v1.ChatStatus_CHAT_STATUS_FAILED
	case responses.ResponseStatusCancelled:
		return v1.ChatStatus_CHAT_STATUS_CANCELLED
	default:
		return v1.ChatStatus_CHAT_STATUS_IN_PROGRESS
	}
}

func convertStatisticsFromOpenAIResponses(usage *responses.ResponseUsage) *v1.Statistics {
	if usage == nil ||
		(usage.InputTokens == 0 &&
			usage.OutputTokens == 0 &&
			usage.InputTokensDetails.CachedTokens == 0 &&
			usage.OutputTokensDetails.ReasoningTokens == 0) {
		return nil
	}

	return &v1.Statistics{Usage: &v1.Usage{
		InputTokens:       uint32(max(usage.InputTokens, 0)),
		OutputTokens:      uint32(max(usage.OutputTokens, 0)),
		CachedInputTokens: uint32(max(usage.InputTokensDetails.CachedTokens, 0)),
		ReasoningTokens:   uint32(max(usage.OutputTokensDetails.ReasoningTokens, 0)),
	}}
}

func convertOutputMessagePhaseFromOpenAIResponses(phase responses.ResponseOutputMessagePhase) v1.ContentPhase {
	if phase == responses.ResponseOutputMessagePhaseFinalAnswer {
		return v1.ContentPhase_CONTENT_PHASE_OUTCOME
	}
	return v1.ContentPhase_CONTENT_PHASE_NORMAL
}

func (c *openAIResponseStreamClient) convertStreamEventFromOpenAIResponses(
	event responses.ResponseStreamEventUnion,
) ([]*entity.ChatEvent, error) {
	var events []*entity.ChatEvent

	switch event.Type {
	case "response.created":
		c.messageStarted = true
		events = append(events,
			c.newChatEvent(v1.NewMessageStartEvent(event.Response.ID, string(event.Response.Model))))

	case "response.in_progress":

	case "response.output_item.added":
		c.outputPhase[event.OutputIndex] = convertOutputMessagePhaseFromOpenAIResponses(event.Item.Phase)
		if event.Item.Type == "function_call" {
			c.hasFunctionCall = true
			call := event.Item.AsFunctionCall()
			index := c.openBlock(&events, event.OutputIndex)
			events = append(events, c.newChatEvent(v1.NewIdentifiedContentStartToolUseEvent(
				call.ID,
				index,
				call.CallID,
				call.Name,
			)))
		}

	case "response.content_part.added":
		switch event.Part.Type {
		case "output_text":
			index := c.openBlock(&events, event.OutputIndex)
			events = append(events, c.newChatEvent(v1.NewIdentifiedContentStartTextEvent(
				event.ItemID,
				index,
				c.outputPhase[event.OutputIndex],
			)))
		case "reasoning_text":
			index := c.openBlock(&events, event.OutputIndex)
			events = append(events, c.newChatEvent(v1.NewIdentifiedContentStartTextEvent(
				event.ItemID,
				index,
				v1.ContentPhase_CONTENT_PHASE_REASONING,
			)))
		}

	case "response.content_part.done":
		if event.Part.Type == "output_text" || event.Part.Type == "reasoning_text" {
			c.closeOutputBlock(&events, event.OutputIndex)
		}

	case "response.reasoning_summary_part.added":
		if _, ok := c.openContentIndexByOutputIndex[event.OutputIndex]; !ok {
			index := c.openBlock(&events, event.OutputIndex)
			events = append(events, c.newChatEvent(v1.NewIdentifiedContentStartTextEvent(
				event.ItemID,
				index,
				v1.ContentPhase_CONTENT_PHASE_REASONING,
			)))
		}

	case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
		index, ok := c.openContentIndexByOutputIndex[event.OutputIndex]
		if !ok {
			return nil, fmt.Errorf("responses stream sent reasoning text before the reasoning part")
		}
		events = append(events, c.newChatEvent(v1.NewContentDeltaTextEvent(index, event.Delta)))

	case "response.reasoning_summary_part.done":
		c.closeOutputBlock(&events, event.OutputIndex)

	case "response.output_text.delta":
		index, ok := c.openContentIndexByOutputIndex[event.OutputIndex]
		if !ok {
			return nil, fmt.Errorf("responses stream sent output text before the output text part")
		}
		events = append(events, c.newChatEvent(v1.NewContentDeltaTextEvent(index, event.Delta)))

	case "response.function_call_arguments.delta":
		index, ok := c.openContentIndexByOutputIndex[event.OutputIndex]
		if !ok {
			return nil, fmt.Errorf("responses stream sent function arguments before the function item")
		}
		events = append(events, c.newChatEvent(v1.NewContentDeltaToolInputTextEvent(index, event.Delta)))

	case "response.output_item.done":
		item := event.Item
		if item.Type == "function_call" {
			c.hasFunctionCall = true
		}
		c.closeOutputBlock(&events, event.OutputIndex)
		if item.Type == "reasoning" {
			reasoning := item.AsReasoning()
			if reasoning.EncryptedContent != "" {
				index := c.nextIndex
				c.nextIndex++
				events = append(events, c.newChatEvent(v1.NewContentSnapshotEvent(&v1.Content{
					Id:      reasoning.ID,
					Index:   new(index),
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{Opaque: reasoning.EncryptedContent},
				})))
			}
		}

	case "response.completed", "response.incomplete", "response.failed":
		c.closeOpenBlock(&events)
		status := convertStatusFromOpenAIResponses(&event.Response)
		if status == v1.ChatStatus_CHAT_STATUS_COMPLETED && c.hasFunctionCall {
			status = v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE
		}
		stop := c.newChatEvent(v1.NewMessageStopEvent(status))
		if statistics := convertStatisticsFromOpenAIResponses(&event.Response.Usage); statistics != nil {
			stop.Usage = statistics.Usage
		}
		events = append(events, stop)
		c.stopEmitted = true

	case "error":
		return nil, fmt.Errorf("responses stream error %s: %s", event.Code, event.Message)
	}

	return events, nil
}

func (c *openAIResponseStreamClient) finish() []*entity.ChatEvent {
	if !c.messageStarted || c.stopEmitted {
		return nil
	}
	var events []*entity.ChatEvent
	c.closeOpenBlock(&events)
	// A well-formed Responses stream always carries a terminal response event.
	// Reaching EOF first means the response was truncated, regardless of the
	// last in-progress status observed in the stream.
	events = append(events, c.newChatEvent(v1.NewMessageStopEvent(v1.ChatStatus_CHAT_STATUS_FAILED)))
	c.stopEmitted = true
	return events
}

func (c *openAIResponseStreamClient) newChatEvent(event v1.ChatEventPayload) *entity.ChatEvent {
	return v1.NewChatEvent(c.req.GetId(), event)
}

func (c *openAIResponseStreamClient) openBlock(events *[]*entity.ChatEvent, outputIndex int64) uint32 {
	if index, ok := c.openContentIndexByOutputIndex[outputIndex]; ok {
		return index
	}
	c.closeOpenBlock(events)

	index := c.nextIndex
	c.nextIndex++
	c.hasOpen = true
	c.openOutputIndex = outputIndex
	c.openContentIndex = index
	c.openContentIndexByOutputIndex[outputIndex] = index
	return index
}

func (c *openAIResponseStreamClient) closeOpenBlock(events *[]*entity.ChatEvent) {
	if !c.hasOpen {
		return
	}
	*events = append(*events, c.newChatEvent(v1.NewContentStopEvent(c.openContentIndex)))
	delete(c.openContentIndexByOutputIndex, c.openOutputIndex)
	c.hasOpen = false
}

func (c *openAIResponseStreamClient) closeOutputBlock(events *[]*entity.ChatEvent, outputIndex int64) {
	index, ok := c.openContentIndexByOutputIndex[outputIndex]
	if !ok {
		return
	}
	*events = append(*events, c.newChatEvent(v1.NewContentStopEvent(index)))
	delete(c.openContentIndexByOutputIndex, outputIndex)
	if c.hasOpen && c.openOutputIndex == outputIndex {
		c.hasOpen = false
	}
}
