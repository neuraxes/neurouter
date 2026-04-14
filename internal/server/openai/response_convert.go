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
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func convertEffortFromOpenAIResponse(effort shared.ReasoningEffort) v1.ReasoningEffort {
	switch effort {
	case shared.ReasoningEffortNone:
		return v1.ReasoningEffort_REASONING_EFFORT_NONE
	case shared.ReasoningEffortMinimal:
		return v1.ReasoningEffort_REASONING_EFFORT_MINIMAL
	case shared.ReasoningEffortLow:
		return v1.ReasoningEffort_REASONING_EFFORT_LOW
	case shared.ReasoningEffortMedium:
		return v1.ReasoningEffort_REASONING_EFFORT_MEDIUM
	case shared.ReasoningEffortHigh:
		return v1.ReasoningEffort_REASONING_EFFORT_HIGH
	case shared.ReasoningEffortXhigh:
		return v1.ReasoningEffort_REASONING_EFFORT_MAX
	default:
		return v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED
	}
}

func convertConfigFromResponseReq(req *responses.ResponseNewParams) *v1.GenerationConfig {
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
			Effort: convertEffortFromOpenAIResponse(req.Reasoning.Effort),
		}
	}

	if req.Text.Format.OfJSONObject != nil {
		config.Grammar = &v1.GenerationConfig_PresetGrammar{
			PresetGrammar: "json_object",
		}
	} else if req.Text.Format.OfJSONSchema != nil {
		schemaJSON, err := json.Marshal(req.Text.Format.OfJSONSchema.Schema)
		if err == nil {
			config.Grammar = &v1.GenerationConfig_JsonSchema{
				JsonSchema: string(schemaJSON),
			}
		}
	}

	return config
}

func convertEasyInputMessageFromResponse(m *responses.EasyInputMessageParam) *v1.Message {
	var role v1.Role
	switch m.Role {
	case responses.EasyInputMessageRoleSystem, responses.EasyInputMessageRoleDeveloper:
		role = v1.Role_SYSTEM
	case responses.EasyInputMessageRoleUser:
		role = v1.Role_USER
	case responses.EasyInputMessageRoleAssistant:
		role = v1.Role_MODEL
	default:
		role = v1.Role_USER
	}

	var contents []*v1.Content

	if m.Content.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: m.Content.OfString.Value},
		})
	} else {
		for _, part := range m.Content.OfInputItemContentList {
			if part.OfInputText != nil {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Text{Text: part.OfInputText.Text},
				})
			} else if part.OfInputImage != nil {
				if part.OfInputImage.ImageURL.Valid() {
					contents = append(contents, &v1.Content{
						Content: &v1.Content_Image{
							Image: convertImageFromOpenAIURL(part.OfInputImage.ImageURL.Value),
						},
					})
				}
			}
		}
	}

	return &v1.Message{
		Role:     role,
		Contents: contents,
	}
}

func convertOutputMessageFromResponse(m *responses.ResponseOutputMessageParam) *v1.Message {
	var contents []*v1.Content
	for _, part := range m.Content {
		if part.OfOutputText != nil {
			contents = append(contents, &v1.Content{
				Id:      m.ID,
				Content: &v1.Content_Text{Text: part.OfOutputText.Text},
			})
		}
	}
	return &v1.Message{
		Role:     v1.Role_MODEL,
		Contents: contents,
	}
}

func convertFunctionCallFromResponse(fc *responses.ResponseFunctionToolCallParam) *v1.Content {
	c := &v1.Content{
		Content: &v1.Content_ToolUse{
			ToolUse: &v1.ToolUse{
				Id:   fc.CallID,
				Name: fc.Name,
				Inputs: []*v1.ToolUse_Input{{
					Input: &v1.ToolUse_Input_Text{Text: fc.Arguments},
				}},
			},
		},
	}
	if fc.ID.Valid() {
		c.Id = fc.ID.Value
	}
	return c
}

func convertFunctionCallOutputFromResponse(fco *responses.ResponseInputItemFunctionCallOutputParam) *v1.Message {
	tr := &v1.ToolResult{
		Id:      fco.CallID,
		Outputs: []*v1.ToolResult_Output{},
	}

	if fco.Output.OfString.Valid() {
		tr.Outputs = append(tr.Outputs, &v1.ToolResult_Output{
			Output: &v1.ToolResult_Output_Text{Text: fco.Output.OfString.Value},
		})
	} else {
		for _, item := range fco.Output.OfResponseFunctionCallOutputItemArray {
			if item.OfInputText != nil {
				tr.Outputs = append(tr.Outputs, &v1.ToolResult_Output{
					Output: &v1.ToolResult_Output_Text{Text: item.OfInputText.Text},
				})
			}
		}
	}

	return &v1.Message{
		Role: v1.Role_USER,
		Contents: []*v1.Content{{
			Content: &v1.Content_ToolResult{ToolResult: tr},
		}},
	}
}

func convertReasoningFromResponse(r *responses.ResponseReasoningItemParam) *v1.Message {
	var contents []*v1.Content
	for _, c := range r.Content {
		ct := &v1.Content{
			Id:        r.ID,
			Reasoning: true,
			Content:   &v1.Content_Text{Text: c.Text},
		}
		if r.EncryptedContent.Valid() {
			ct.SetMeta("encrypted", r.EncryptedContent.Value)
		}
		contents = append(contents, ct)
	}
	for _, s := range r.Summary {
		ct := &v1.Content{
			Id:        r.ID,
			Reasoning: true,
			Content:   &v1.Content_Text{},
		}
		ct.SetMeta("summary", s.Text)
		contents = append(contents, ct)
	}
	if len(contents) == 0 && r.EncryptedContent.Valid() {
		ct := &v1.Content{
			Id:        r.ID,
			Reasoning: true,
			Content:   &v1.Content_Text{},
		}
		ct.SetMeta("encrypted", r.EncryptedContent.Value)
		contents = append(contents, ct)
	}
	return &v1.Message{
		Role:     v1.Role_MODEL,
		Contents: contents,
	}
}

func convertInputItemsFromResponse(items []responses.ResponseInputItemUnionParam) []*v1.Message {
	var messages []*v1.Message
	var pendingModelContents []*v1.Content

	flushModel := func() {
		if len(pendingModelContents) > 0 {
			messages = append(messages, &v1.Message{
				Role:     v1.Role_MODEL,
				Contents: pendingModelContents,
			})
			pendingModelContents = nil
		}
	}

	for i := range items {
		item := &items[i]
		switch {
		case item.OfMessage != nil:
			flushModel()
			messages = append(messages, convertEasyInputMessageFromResponse(item.OfMessage))

		case item.OfOutputMessage != nil:
			flushModel()
			messages = append(messages, convertOutputMessageFromResponse(item.OfOutputMessage))

		case item.OfFunctionCall != nil:
			pendingModelContents = append(pendingModelContents, convertFunctionCallFromResponse(item.OfFunctionCall))

		case item.OfFunctionCallOutput != nil:
			flushModel()
			messages = append(messages, convertFunctionCallOutputFromResponse(item.OfFunctionCallOutput))

		case item.OfReasoning != nil:
			msg := convertReasoningFromResponse(item.OfReasoning)
			pendingModelContents = append(pendingModelContents, msg.Contents...)
		}
	}
	flushModel()
	return messages
}

func convertReqFromResponse(req *responses.ResponseNewParams) *v1.ChatReq {
	config := convertConfigFromResponseReq(req)

	var messages []*v1.Message

	if req.Instructions.Valid() {
		messages = append(messages, &v1.Message{
			Role: v1.Role_SYSTEM,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: req.Instructions.Value},
			}},
		})
	}

	if req.Input.OfString.Valid() {
		messages = append(messages, &v1.Message{
			Role: v1.Role_USER,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: req.Input.OfString.Value},
			}},
		})
	} else if len(req.Input.OfInputItemList) > 0 {
		messages = append(messages, convertInputItemsFromResponse(req.Input.OfInputItemList)...)
	}

	var tools []*v1.Tool
	for _, tool := range req.Tools {
		if tool.OfFunction != nil {
			fn := tool.OfFunction
			var parameters *v1.Schema
			j, _ := json.Marshal(fn.Parameters)
			_ = json.Unmarshal(j, &parameters)

			t := &v1.Tool{
				Tool: &v1.Tool_Function_{
					Function: &v1.Tool_Function{
						Name:       fn.Name,
						Parameters: parameters,
					},
				},
			}
			if fn.Description.Valid() {
				t.GetFunction().Description = fn.Description.Value
			}
			tools = append(tools, t)
		}
	}

	return &v1.ChatReq{
		Model:    string(req.Model),
		Config:   config,
		Messages: messages,
		Tools:    tools,
	}
}

func convertStatusToResponse(status v1.ChatStatus) string {
	switch status {
	case v1.ChatStatus_CHAT_COMPLETED, v1.ChatStatus_CHAT_PENDING_TOOL_USE, v1.ChatStatus_CHAT_REFUSED:
		return "completed"
	case v1.ChatStatus_CHAT_FAILED:
		return "failed"
	case v1.ChatStatus_CHAT_CANCELLED:
		return "cancelled"
	case v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT:
		return "incomplete"
	default:
		return "in_progress"
	}
}

func convertUsageToResponse(u *v1.Statistics_Usage) *responseUsage {
	if u == nil {
		return nil
	}
	return &responseUsage{
		InputTokens:  int64(u.InputTokens),
		OutputTokens: int64(u.OutputTokens),
		TotalTokens:  int64(u.InputTokens) + int64(u.OutputTokens),
		InputTokensDetails: responseInputTokensDetails{
			CachedTokens: int64(u.CachedInputTokens),
		},
		OutputTokenDetails: responseOutputTokensDetails{
			ReasoningTokens: int64(u.ReasoningTokens),
		},
	}
}

func convertRespToResponse(resp *v1.ChatResp) *responseObject {
	r := &responseObject{
		ID:     "resp_" + strings.ReplaceAll(resp.Id, "-", ""),
		Object: "response",
		Model:  resp.Model,
		Status: convertStatusToResponse(resp.Status),
	}

	if resp.Message != nil {
		var currentReasoning *responseReasoning
		var currentMessage *responseOutputMessage

		flushReasoning := func() {
			if currentReasoning != nil {
				r.Output = append(r.Output, *currentReasoning)
				currentReasoning = nil
			}
		}

		flushMessage := func() {
			if currentMessage != nil {
				r.Output = append(r.Output, *currentMessage)
				currentMessage = nil
			}
		}

		for _, content := range resp.Message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					flushMessage()
					if currentReasoning == nil || (content.Id != "" && currentReasoning.ID != content.Id) {
						flushReasoning()
						id := content.Id
						if id == "" {
							id = "rs_" + uuid.NewString()[:12]
						}
						currentReasoning = &responseReasoning{
							Type:    "reasoning",
							ID:      id,
							Summary: []responseReasoningSummary{},
						}
					}
					if summary := content.Meta("summary"); summary != "" {
						currentReasoning.Summary = append(currentReasoning.Summary, responseReasoningSummary{
							Type: "summary_text",
							Text: summary,
						})
					}
					if text := c.Text; text != "" {
						currentReasoning.Summary = append(currentReasoning.Summary, responseReasoningSummary{
							Type: "summary_text",
							Text: text,
						})
					}
					if encrypted := content.Meta("encrypted"); encrypted != "" {
						currentReasoning.EncryptedContent = encrypted
					}
					continue
				}

				flushReasoning()
				if currentMessage == nil || (content.Id != "" && currentMessage.ID != content.Id) {
					flushMessage()
					id := content.Id
					if id == "" {
						id = "msg_" + uuid.NewString()[:12]
					}
					currentMessage = &responseOutputMessage{
						Type:    "message",
						ID:      id,
						Role:    "assistant",
						Status:  "completed",
						Content: []responseOutputContent{},
					}
				}
				if resp.Status == v1.ChatStatus_CHAT_REFUSED {
					currentMessage.Content = append(currentMessage.Content, responseRefusal{
						Type:    "refusal",
						Refusal: c.Text,
					})
				} else {
					currentMessage.Content = append(currentMessage.Content, responseOutputText{
						Type:        "output_text",
						Text:        c.Text,
						Annotations: []any{},
					})
				}

			case *v1.Content_ToolUse:
				flushReasoning()
				flushMessage()
				r.Output = append(r.Output, responseFunctionCall{
					Type:      "function_call",
					ID: func() string {
						if content.Id != "" {
							return content.Id
						}
						return "fc_" + uuid.NewString()[:12]
					}(),
					CallID:    c.ToolUse.Id,
					Name:      c.ToolUse.Name,
					Arguments: c.ToolUse.GetTextualInput(),
					Status:    "completed",
				})
			}
		}

		flushReasoning()
		flushMessage()
	}

	if resp.Statistics != nil {
		r.Usage = convertUsageToResponse(resp.Statistics.Usage)
	}

	return r
}
