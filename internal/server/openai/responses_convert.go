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
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3/responses"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

// convertChatRequestFromOpenAIResponses converts a raw Responses API request body.
//
// Everything but the input items is decoded with the OpenAI SDK. The input
// items are read straight from the JSON because the SDK registers
// EasyInputMessage, ResponseInputItemMessage and ResponseOutputMessage under a
// single "message" discriminator and always resolves to the first of them,
// which accepts neither output_text parts nor items that omit "type". Both
// occur whenever a client replays our own output items, and the SDK drops them
// without reporting an error.
func convertChatRequestFromOpenAIResponses(body []byte) (*v1.ChatRequest, error) {
	var req responses.ResponseNewParams
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	chatReq := &v1.ChatRequest{
		Session:  req.PromptCacheKey.Value,
		Model:    string(req.Model),
		Config:   convertGenerationConfigFromOpenAIResponses(&req),
		Tools:    convertToolsFromOpenAIResponses(req.Tools),
		Metadata: map[string]string(req.Metadata),
	}

	if req.Instructions.Valid() {
		appendAdjacentResponsesMessage(&chatReq.Messages, &v1.Message{
			Role: v1.Role_ROLE_SYSTEM,
			Contents: []*v1.Content{{
				Content: v1.NewTextContent(req.Instructions.Value),
			}},
		})
	}

	input := gjson.GetBytes(body, "input")
	if input.Type == gjson.String {
		appendAdjacentResponsesMessage(&chatReq.Messages, &v1.Message{
			Role: v1.Role_ROLE_USER,
			Contents: []*v1.Content{{
				Content: v1.NewTextContent(input.String()),
			}},
		})
	} else {
		for _, item := range input.Array() {
			appendAdjacentResponsesMessage(&chatReq.Messages, convertInputItemFromOpenAIResponses(item))
		}
	}

	return chatReq, nil
}

func convertGenerationConfigFromOpenAIResponses(req *responses.ResponseNewParams) *v1.GenerationConfig {
	config := &v1.GenerationConfig{}

	if req.MaxOutputTokens.Valid() {
		config.MaxTokens = new(req.MaxOutputTokens.Value)
	}
	if req.Temperature.Valid() {
		config.Temperature = new(float32(req.Temperature.Value))
	}
	if req.TopP.Valid() {
		config.TopP = new(float32(req.TopP.Value))
	}

	if req.Reasoning.Effort != "" {
		config.ReasoningConfig = &v1.ReasoningConfig{
			Effort: convertReasoningEffortFromOpenAI(req.Reasoning.Effort),
		}
	}

	if req.Text.Format.OfJSONObject != nil {
		config.Grammar = &v1.GenerationConfig_PresetGrammar{PresetGrammar: "json_object"}
	} else if schemaConfig := req.Text.Format.OfJSONSchema; schemaConfig != nil {
		if schema, err := util.StructFromMap(schemaConfig.Schema); err == nil {
			config.Grammar = &v1.GenerationConfig_Schema{Schema: schema}
		}
	}
	return config
}

func convertToolsFromOpenAIResponses(openAITools []responses.ToolUnionParam) []*v1.Tool {
	var tools []*v1.Tool
	for _, tool := range openAITools {
		if tool.OfFunction == nil {
			continue
		}
		function := tool.OfFunction
		inputSchema, _ := util.StructFromMap(function.Parameters)
		tools = append(tools, &v1.Tool{Tool: &v1.Tool_Function_{Function: &v1.Tool_Function{
			Name:        function.Name,
			Description: function.Description.Value,
			InputSchema: inputSchema,
		}}})
	}
	return tools
}

func convertInputItemFromOpenAIResponses(item gjson.Result) *v1.Message {
	itemType := item.Get("type").String()
	if itemType == "" && item.Get("role").Exists() {
		// "type" is optional on the shorthand message form.
		itemType = "message"
	}

	switch itemType {
	case "message":
		return convertMessageItemFromOpenAIResponses(item)
	case "reasoning":
		return convertReasoningItemFromOpenAIResponses(item)
	case "function_call":
		return &v1.Message{
			Role: v1.Role_ROLE_MODEL,
			Contents: []*v1.Content{{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
				Id:   item.Get("call_id").String(),
				Name: item.Get("name").String(),
				Inputs: []*v1.ToolUse_Input{{
					Input: &v1.ToolUse_Input_Text{Text: item.Get("arguments").String()},
				}},
			}}}},
		}
	case "function_call_output":
		output := item.Get("output")
		var text string
		if output.Type == gjson.String {
			text = output.String()
		} else {
			for _, part := range output.Array() {
				text += part.Get("text").String()
			}
		}
		return &v1.Message{
			Role: v1.Role_ROLE_USER,
			Contents: []*v1.Content{{Content: &v1.Content_ToolResult{ToolResult: &v1.ToolResult{
				Id: item.Get("call_id").String(),
				Outputs: []*v1.ToolResult_Output{{
					Output: &v1.ToolResult_Output_Text{
						Text: text,
					},
				}},
			}}}},
		}
	}
	return nil
}

func convertMessageItemFromOpenAIResponses(item gjson.Result) *v1.Message {
	phase := v1.ContentPhase_CONTENT_PHASE_NORMAL
	if item.Get("phase").String() == string(responses.ResponseOutputMessagePhaseFinalAnswer) {
		phase = v1.ContentPhase_CONTENT_PHASE_OUTCOME
	}

	role := v1.Role_ROLE_USER
	switch item.Get("role").String() {
	case "system", "developer":
		role = v1.Role_ROLE_SYSTEM
	case "assistant":
		role = v1.Role_ROLE_MODEL
	}

	message := &v1.Message{Role: role}
	content := item.Get("content")
	if content.Type == gjson.String {
		message.Contents = append(message.Contents, &v1.Content{
			Phase:   phase,
			Content: v1.NewTextContent(content.String()),
		})
		return message
	}

	for _, part := range content.Array() {
		switch part.Get("type").String() {
		// Assistant turns replayed from a previous response carry output_text,
		// while fresh input carries input_text.
		case "input_text", "output_text":
			message.Contents = append(message.Contents, &v1.Content{
				Phase:   phase,
				Content: v1.NewTextContent(part.Get("text").String()),
			})
		case "input_image":
			if url := part.Get("image_url"); url.Type == gjson.String {
				message.Contents = append(message.Contents, &v1.Content{
					Phase: phase,
					Content: &v1.Content_Image{
						Image: convertImageFromOpenAIURL(url.String()),
					},
				})
			}
		}
	}
	return message
}

func convertReasoningItemFromOpenAIResponses(item gjson.Result) *v1.Message {
	id := item.Get("id").String()
	message := &v1.Message{Role: v1.Role_ROLE_MODEL}

	appendText := func(text gjson.Result) {
		message.Contents = append(message.Contents, &v1.Content{
			Id:      id,
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: v1.NewTextContent(text.Get("text").String()),
		})
	}
	for _, summary := range item.Get("summary").Array() {
		appendText(summary)
	}
	for _, reasoning := range item.Get("content").Array() {
		appendText(reasoning)
	}

	if encrypted := item.Get("encrypted_content"); encrypted.Type == gjson.String {
		message.Contents = append(message.Contents, &v1.Content{
			Id:      id,
			Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
			Content: &v1.Content_Opaque{Opaque: encrypted.String()},
		})
	}
	return message
}

func appendAdjacentResponsesMessage(messages *[]*v1.Message, message *v1.Message) {
	if message == nil {
		return
	}
	if len(*messages) > 0 && (*messages)[len(*messages)-1].Role == message.Role {
		last := (*messages)[len(*messages)-1]
		last.Contents = append(last.Contents, message.Contents...)
		return
	}
	*messages = append(*messages, message)
}

func convertChatResponseToOpenAIResponses(resp *v1.ChatResponse) *responsesResponse {
	responseID := resp.GetMessage().GetId()
	if responseID == "" {
		responseID = resp.GetId()
	}

	openAIResp := &responsesResponse{
		ID:        responseID,
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Model:     resp.GetModel(),
		Output:    make([]*responsesOutputItem, 0),
		Usage:     convertUsageToOpenAIResponses(resp.GetStatistics().GetUsage()),
	}
	if resp.Message != nil {
		openAIResp.Output = convertContentsToOpenAIResponses(responseID, resp.Message.Contents)
	}
	openAIResp.Status, openAIResp.IncompleteDetails, openAIResp.Error = convertStatusToOpenAIResponses(resp.GetStatus())

	return openAIResp
}

func convertUsageToOpenAIResponses(usage *v1.Usage) *responses.ResponseUsage {
	if usage == nil {
		return nil
	}
	return &responses.ResponseUsage{
		InputTokens:  int64(usage.InputTokens),
		OutputTokens: int64(usage.OutputTokens),
		TotalTokens:  int64(usage.InputTokens) + int64(usage.OutputTokens),
		InputTokensDetails: responses.ResponseUsageInputTokensDetails{
			CachedTokens: int64(usage.CachedInputTokens),
		},
		OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{
			ReasoningTokens: int64(usage.ReasoningTokens),
		},
	}
}

func convertContentsToOpenAIResponses(responseID string, contents []*v1.Content) []*responsesOutputItem {
	var output []*responsesOutputItem
	var openMessageItem *responsesOutputItem
	var openReasoningItem *responsesOutputItem
	reasoningItemByID := make(map[string]*responsesOutputItem)

	flushMessage := func() {
		if openMessageItem == nil {
			return
		}
		if openMessageItem.ID == "" {
			openMessageItem.ID = syntheticResponsesItemID("msg", responseID, len(output))
		}
		output = append(output, openMessageItem)
		openMessageItem = nil
	}

	// reasoningItem returns the item that carries the reasoning content with the
	// given id. Contents without an id join the reasoning item that is still
	// being assembled, so that upstreams which do not expose reasoning item ids
	// keep a summary and its encrypted snapshot in a single item.
	reasoningItem := func(id string) *responsesOutputItem {
		if id == "" {
			if openReasoningItem != nil {
				return openReasoningItem
			}
			id = syntheticResponsesItemID("rs", responseID, len(output))
		}
		if item := reasoningItemByID[id]; item != nil {
			openReasoningItem = item
			return item
		}

		item := &responsesOutputItem{
			ID:     id,
			Type:   "reasoning",
			Status: "completed",
		}
		reasoningItemByID[id] = item
		output = append(output, item)
		openReasoningItem = item
		return item
	}

	for _, content := range contents {
		if content == nil {
			continue
		}
		switch c := content.Content.(type) {
		case *v1.Content_Text:
			if content.Phase == v1.ContentPhase_CONTENT_PHASE_REASONING {
				flushMessage()
				item := reasoningItem(content.GetId())
				item.Summary = append(item.Summary, responses.ResponseReasoningItemSummary{
					Type: "summary_text",
					Text: c.Text.GetText(),
				})
				continue
			}

			openReasoningItem = nil
			phase := convertContentPhaseToOpenAIResponsesServer(content.Phase)
			itemID := content.GetId()
			if openMessageItem != nil && (openMessageItem.ID != itemID || openMessageItem.Phase != phase) {
				flushMessage()
			}
			if openMessageItem == nil {
				openMessageItem = &responsesOutputItem{
					ID:     itemID,
					Type:   "message",
					Status: "completed",
					Role:   "assistant",
					Phase:  phase,
				}
			}
			openMessageItem.Content = append(openMessageItem.Content, responses.ResponseOutputText{
				Type: "output_text",
				Text: c.Text.GetText(),
			})

		case *v1.Content_Opaque:
			if content.Phase != v1.ContentPhase_CONTENT_PHASE_REASONING {
				continue
			}
			flushMessage()
			reasoningItem(content.GetId()).EncryptedContent = c.Opaque

		case *v1.Content_ToolUse:
			flushMessage()
			openReasoningItem = nil
			arguments := c.ToolUse.GetTextualInput()
			itemID := content.GetId()
			if itemID == "" {
				itemID = syntheticResponsesItemID("fc", responseID, len(output))
			}
			output = append(output, &responsesOutputItem{
				ID:        itemID,
				Type:      "function_call",
				Status:    "completed",
				CallID:    c.ToolUse.GetId(),
				Name:      c.ToolUse.GetName(),
				Arguments: &arguments,
			})
		}
	}

	flushMessage()
	return output
}

func convertStatusToOpenAIResponses(
	status v1.ChatStatus,
) (string, *responses.ResponseIncompleteDetails, *responses.ResponseError) {
	switch status {
	case v1.ChatStatus_CHAT_STATUS_COMPLETED, v1.ChatStatus_CHAT_STATUS_PENDING_TOOL_USE:
		return "completed", nil, nil
	case v1.ChatStatus_CHAT_STATUS_REACHED_TOKEN_LIMIT:
		return "incomplete", &responses.ResponseIncompleteDetails{Reason: "max_output_tokens"}, nil
	case v1.ChatStatus_CHAT_STATUS_REFUSED:
		return "incomplete", &responses.ResponseIncompleteDetails{Reason: "content_filter"}, nil
	case v1.ChatStatus_CHAT_STATUS_FAILED:
		return "failed", nil, &responses.ResponseError{
			Code:    "server_error",
			Message: "The response failed to generate.",
		}
	case v1.ChatStatus_CHAT_STATUS_CANCELLED:
		// A cancelled response terminates through response.failed, whose error
		// field the protocol requires to be populated.
		return "cancelled", nil, &responses.ResponseError{
			Code:    "cancelled",
			Message: "The response was cancelled.",
		}
	default:
		return "failed", nil, &responses.ResponseError{
			Code:    "server_error",
			Message: "The response ended without a terminal status.",
		}
	}
}

func convertContentPhaseToOpenAIResponsesServer(phase v1.ContentPhase) string {
	if phase == v1.ContentPhase_CONTENT_PHASE_OUTCOME {
		return "final_answer"
	}
	return ""
}

func syntheticResponsesItemID(prefix, responseID string, outputIndex int) string {
	return fmt.Sprintf("%s_%s_%d", prefix, responseID, outputIndex)
}
