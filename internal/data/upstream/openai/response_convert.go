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
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/util"
)

// contentPhaseFromOpenAIPhase maps an OpenAI response output-message phase to the internal ContentPhase enum.
func contentPhaseFromOpenAIPhase(phase responses.ResponseOutputMessagePhase) v1.ContentPhase {
	if phase == responses.ResponseOutputMessagePhaseFinalAnswer {
		return v1.ContentPhase_CONTENT_PHASE_OUTCOME
	}
	return v1.ContentPhase_CONTENT_PHASE_NORMAL
}

func convertEffortToOpenAI(effort v1.ReasoningEffort) shared.ReasoningEffort {
	switch effort {
	case v1.ReasoningEffort_REASONING_EFFORT_NONE:
		return shared.ReasoningEffortNone
	case v1.ReasoningEffort_REASONING_EFFORT_MINIMAL:
		return shared.ReasoningEffortMinimal
	case v1.ReasoningEffort_REASONING_EFFORT_LOW:
		return shared.ReasoningEffortLow
	case v1.ReasoningEffort_REASONING_EFFORT_MEDIUM:
		return shared.ReasoningEffortMedium
	case v1.ReasoningEffort_REASONING_EFFORT_HIGH:
		return shared.ReasoningEffortHigh
	case v1.ReasoningEffort_REASONING_EFFORT_EXTRA_HIGH, v1.ReasoningEffort_REASONING_EFFORT_MAX:
		return shared.ReasoningEffortXhigh
	default:
		return ""
	}
}

func convertConfigToOpenAIResponse(config *v1.GenerationConfig, req *responses.ResponseNewParams) {
	if config == nil {
		return
	}
	if config.MaxTokens != nil {
		req.MaxOutputTokens = openai.Opt(*config.MaxTokens)
	}
	if config.Temperature != nil {
		req.Temperature = openai.Opt(float64(*config.Temperature))
	}
	if config.TopP != nil {
		req.TopP = openai.Opt(float64(*config.TopP))
	}

	if c := config.ReasoningConfig; c != nil && c.Effort > v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED {
		req.Reasoning = shared.ReasoningParam{
			Effort:  convertEffortToOpenAI(c.Effort),
			Summary: responses.ReasoningSummaryAuto,
		}
	}

	switch g := config.Grammar.(type) {
	case *v1.GenerationConfig_PresetGrammar:
		if g.PresetGrammar == "json_object" {
			req.Text = responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
				},
			}
		}
	case *v1.GenerationConfig_JsonSchema:
		var schema map[string]any
		if err := json.Unmarshal([]byte(g.JsonSchema), &schema); err == nil {
			req.Text = responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
						Name:   "custom_schema",
						Schema: schema,
					},
				},
			}
		}
	case *v1.GenerationConfig_Schema:
		schemaMap := convertSchemaToMap(g.Schema)
		req.Text = responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "custom_schema",
					Schema: schemaMap,
				},
			},
		}
	}
}

func convertImageToOpenAIResponseInputURL(image *v1.Image) string {
	switch image.Source.(type) {
	case *v1.Image_Url:
		return image.GetUrl()
	case *v1.Image_Data:
		return util.EncodeImageDataToUrl(image.GetData(), image.GetMimeType())
	}
	return ""
}

func (r *upstream) convertToolResultToFunctionCallOutput(toolResult *v1.ToolResult) responses.ResponseInputItemUnionParam {
	var outputItems responses.ResponseFunctionCallOutputItemListParam
	hasNonText := false

	for _, output := range toolResult.GetOutputs() {
		switch o := output.GetOutput().(type) {
		case *v1.ToolResult_Output_Text:
			outputItems = append(outputItems, responses.ResponseFunctionCallOutputItemUnionParam{
				OfInputText: &responses.ResponseInputTextContentParam{Text: o.Text},
			})
		case *v1.ToolResult_Output_Image:
			hasNonText = true
			outputItems = append(outputItems, responses.ResponseFunctionCallOutputItemUnionParam{
				OfInputImage: &responses.ResponseInputImageContentParam{
					ImageURL: openai.Opt(convertImageToOpenAIResponseInputURL(o.Image)),
					Detail:   responses.ResponseInputImageContentDetailAuto,
				},
			})
		}
	}

	if !hasNonText {
		if r.config.PreferStringContentForTool {
			return responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: toolResult.Id,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfString: openai.Opt(toolResult.GetTextualOutput()),
					},
				},
			}
		}
		if r.config.PreferSinglePartContent && len(outputItems) > 1 {
			return responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: toolResult.Id,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfResponseFunctionCallOutputItemArray: responses.ResponseFunctionCallOutputItemListParam{
							{OfInputText: &responses.ResponseInputTextContentParam{Text: toolResult.GetTextualOutput()}},
						},
					},
				},
			}
		}
	}

	return responses.ResponseInputItemUnionParam{
		OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
			CallID: toolResult.Id,
			Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
				OfResponseFunctionCallOutputItemArray: outputItems,
			},
		},
	}
}

func (r *upstream) convertMessageToOpenAIResponseInput(message *v1.Message) []responses.ResponseInputItemUnionParam {
	plainText := ""
	isPlainText := true

	{
		var sb strings.Builder
		for _, content := range message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				sb.WriteString(c.Text)
			default:
				isPlainText = false
			}
		}
		plainText = sb.String()
	}

	switch message.Role {
	case v1.Role_SYSTEM:
		if isPlainText && r.config.PreferStringContentForSystem {
			return []responses.ResponseInputItemUnionParam{{
				OfMessage: &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(plainText)},
				},
			}}
		} else if isPlainText && r.config.PreferSinglePartContent {
			return []responses.ResponseInputItemUnionParam{{
				OfMessage: &responses.EasyInputMessageParam{
					Role: responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{
						OfInputItemContentList: responses.ResponseInputMessageContentListParam{
							{OfInputText: &responses.ResponseInputTextParam{Text: plainText}},
						},
					},
				},
			}}
		} else {
			var contents responses.ResponseInputMessageContentListParam

			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					contents = append(contents, responses.ResponseInputContentUnionParam{
						OfInputText: &responses.ResponseInputTextParam{Text: c.Text},
					})
				case *v1.Content_Image:
					contents = append(contents, responses.ResponseInputContentUnionParam{
						OfInputImage: &responses.ResponseInputImageParam{
							Detail:   responses.ResponseInputImageDetailAuto,
							ImageURL: openai.Opt(convertImageToOpenAIResponseInputURL(c.Image)),
						},
					})
				default:
					r.log.Errorf("unsupported content for system: %v", c)
				}
			}

			return []responses.ResponseInputItemUnionParam{{
				OfMessage: &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{OfInputItemContentList: contents},
				},
			}}
		}
	case v1.Role_USER:
		var result []responses.ResponseInputItemUnionParam
		var userContents []*v1.Content

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_ToolResult:
				result = append(result, r.convertToolResultToFunctionCallOutput(c.ToolResult))
			default:
				userContents = append(userContents, content)
			}
		}

		if len(result) == 0 || len(userContents) > 0 {
			msg := &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
			}

			if isPlainText && r.config.PreferStringContentForUser {
				msg.Content = responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(plainText)}
			} else if isPlainText && r.config.PreferSinglePartContent {
				msg.Content = responses.EasyInputMessageContentUnionParam{
					OfInputItemContentList: responses.ResponseInputMessageContentListParam{{
						OfInputText: &responses.ResponseInputTextParam{Text: plainText},
					}},
				}
			} else {
				var contents responses.ResponseInputMessageContentListParam

				for _, content := range userContents {
					switch c := content.GetContent().(type) {
					case *v1.Content_Text:
						contents = append(contents, responses.ResponseInputContentUnionParam{
							OfInputText: &responses.ResponseInputTextParam{Text: c.Text},
						})
					case *v1.Content_Image:
						contents = append(contents, responses.ResponseInputContentUnionParam{
							OfInputImage: &responses.ResponseInputImageParam{
								Detail:   responses.ResponseInputImageDetailAuto,
								ImageURL: openai.Opt(convertImageToOpenAIResponseInputURL(c.Image)),
							},
						})
					default:
						r.log.Errorf("unsupported content for user: %v", c)
					}
				}

				msg.Content = responses.EasyInputMessageContentUnionParam{
					OfInputItemContentList: contents,
				}
			}

			result = append(result, responses.ResponseInputItemUnionParam{OfMessage: msg})
		}

		return result
	case v1.Role_MODEL:
		var result []responses.ResponseInputItemUnionParam
		var reasoning *responses.ResponseReasoningItemParam

		flushReasoning := func() {
			if reasoning != nil {
				result = append(result, responses.ResponseInputItemUnionParam{OfReasoning: reasoning})
				reasoning = nil
			}
		}

		ensureReasoning := func(id string) {
			if reasoning != nil && reasoning.ID != id {
				flushReasoning()
			}
			if reasoning == nil {
				reasoning = &responses.ResponseReasoningItemParam{ID: id}
			}
		}

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				switch content.GetPhase() {
				case v1.ContentPhase_CONTENT_PHASE_REASONING_SUMMARY:
					ensureReasoning(content.Id)
					if c.Text != "" {
						reasoning.Summary = append(reasoning.Summary,
							responses.ResponseReasoningItemSummaryParam{Text: c.Text})
					}
				case v1.ContentPhase_CONTENT_PHASE_REASONING:
					ensureReasoning(content.Id)
					if c.Text != "" {
						reasoning.Content = append(reasoning.Content,
							responses.ResponseReasoningItemContentParam{Text: c.Text})
					}
				default:
					flushReasoning()

					msg := &responses.ResponseOutputMessageParam{
						ID: content.Id,
						Content: []responses.ResponseOutputMessageContentUnionParam{{
							OfOutputText: &responses.ResponseOutputTextParam{
								Text: c.Text,
							},
						}},
					}
					if content.GetPhase() == v1.ContentPhase_CONTENT_PHASE_OUTCOME {
						msg.Phase = responses.ResponseOutputMessagePhaseFinalAnswer
					}
					result = append(result, responses.ResponseInputItemUnionParam{OfOutputMessage: msg})
				}
			case *v1.Content_Opaque:
				if content.IsReasoning() {
					ensureReasoning(content.Id)
					reasoning.EncryptedContent = openai.Opt(c.Opaque)
				}
			case *v1.Content_ToolUse:
				flushReasoning()
				fc := &responses.ResponseFunctionToolCallParam{
					CallID:    c.ToolUse.Id,
					Name:      c.ToolUse.Name,
					Arguments: c.ToolUse.GetTextualInput(),
				}
				if content.Id != "" {
					fc.ID = openai.Opt(content.Id)
				}
				result = append(result, responses.ResponseInputItemUnionParam{OfFunctionCall: fc})
			default:
				r.log.Errorf("unsupported content for assistant: %v", c)
			}
		}
		flushReasoning()
		return result
	default:
		r.log.Errorf("invalid role: %v", message.Role)
		return nil
	}
}

func (r *upstream) convertRequestToOpenAIResponse(req *entity.ChatReq) responses.ResponseNewParams {
	openAIReq := responses.ResponseNewParams{
		Model: responses.ResponsesModel(req.Model),
		Store: openai.Opt(false),
		Include: []responses.ResponseIncludable{
			responses.ResponseIncludableReasoningEncryptedContent,
		},
	}

	if req.Config != nil {
		convertConfigToOpenAIResponse(req.Config, &openAIReq)
	}

	for _, message := range req.Messages {
		items := r.convertMessageToOpenAIResponseInput(message)
		openAIReq.Input.OfInputItemList = append(openAIReq.Input.OfInputItemList, items...)
	}

	if req.Tools != nil {
		var tools []responses.ToolUnionParam
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				var parameters openai.FunctionParameters
				if t.Function.InputJsonSchema != "" {
					_ = json.Unmarshal([]byte(t.Function.InputJsonSchema), &parameters)
				} else if t.Function.InputSchema != nil {
					parameters = convertSchemaToMap(t.Function.InputSchema)
				}
				ft := &responses.FunctionToolParam{
					Name:       t.Function.Name,
					Parameters: parameters,
				}
				if t.Function.Description != "" {
					ft.Description = openai.Opt(t.Function.Description)
				}
				tools = append(tools, responses.ToolUnionParam{OfFunction: ft})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		openAIReq.Tools = tools
	}

	return openAIReq
}

func convertStatusFromOpenAIResponse(status responses.ResponseStatus) v1.ChatStatus {
	switch status {
	case responses.ResponseStatusCompleted:
		return v1.ChatStatus_CHAT_COMPLETED
	case responses.ResponseStatusFailed:
		return v1.ChatStatus_CHAT_FAILED
	case responses.ResponseStatusCancelled:
		return v1.ChatStatus_CHAT_CANCELLED
	case responses.ResponseStatusIncomplete:
		return v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT
	default:
		return v1.ChatStatus_CHAT_IN_PROGRESS
	}
}

func (r *upstream) convertResponseFromOpenAIResponse(openAIResp *responses.Response) *entity.ChatResp {
	resp := &entity.ChatResp{
		Model:      string(openAIResp.Model),
		Status:     convertStatusFromOpenAIResponse(openAIResp.Status),
		Statistics: convertStatisticsFromOpenAIResponse(&openAIResp.Usage),
	}

	var contents []*v1.Content
	hasFunctionCall := false
	for _, item := range openAIResp.Output {
		switch item.Type {
		case "message":
			for _, content := range item.Content {
				switch content.Type {
				case "output_text":
					contents = append(contents, &v1.Content{
						Id:      item.ID,
						Phase:   contentPhaseFromOpenAIPhase(item.Phase),
						Content: &v1.Content_Text{Text: content.Text},
					})
				case "refusal":
					contents = append(contents, &v1.Content{
						Id:      item.ID,
						Phase:   contentPhaseFromOpenAIPhase(item.Phase),
						Content: &v1.Content_Text{Text: content.Refusal},
					})
					resp.Status = v1.ChatStatus_CHAT_REFUSED
				}
			}
		case "reasoning":
			var reasoningContents []*v1.Content

			if item.EncryptedContent != "" {
				reasoningContents = append(reasoningContents, &v1.Content{
					Id:      item.ID,
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{Opaque: item.EncryptedContent},
				})
			}
			for _, s := range item.Summary {
				reasoningContents = append(reasoningContents, &v1.Content{
					Id:      item.ID,
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING_SUMMARY,
					Content: &v1.Content_Text{Text: s.Text},
				})
			}
			for _, c := range item.Content {
				reasoningContents = append(reasoningContents, &v1.Content{
					Id:      item.ID,
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Text{Text: c.Text},
				})
			}

			// Fallback to an empty reasoning content so the item is preserved.
			if len(reasoningContents) == 0 {
				reasoningContents = append(reasoningContents, &v1.Content{
					Id:      item.ID,
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Text{},
				})
			}

			contents = append(contents, reasoningContents...)
		case "function_call":
			hasFunctionCall = true
			contents = append(contents, &v1.Content{
				Id: item.ID,
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   item.CallID,
						Name: item.Name,
						Inputs: []*v1.ToolUse_Input{{
							Input: &v1.ToolUse_Input_Text{Text: item.Arguments.OfString},
						}},
					},
				},
			})
		}
	}

	if hasFunctionCall && resp.Status == v1.ChatStatus_CHAT_COMPLETED {
		resp.Status = v1.ChatStatus_CHAT_PENDING_TOOL_USE
	}

	if len(contents) > 0 {
		resp.Message = &v1.Message{
			Id:       openAIResp.ID,
			Role:     v1.Role_MODEL,
			Contents: contents,
		}
	}

	return resp
}

func convertStatisticsFromOpenAIResponse(usage *responses.ResponseUsage) *v1.Statistics {
	if usage == nil ||
		(usage.InputTokens == 0 &&
			usage.OutputTokens == 0 &&
			usage.InputTokensDetails.CachedTokens == 0 &&
			usage.OutputTokensDetails.ReasoningTokens == 0) {
		return nil
	}

	return &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:       uint32(max(usage.InputTokens, 0)),
			OutputTokens:      uint32(max(usage.OutputTokens, 0)),
			CachedInputTokens: uint32(max(usage.InputTokensDetails.CachedTokens, 0)),
			ReasoningTokens:   uint32(max(usage.OutputTokensDetails.ReasoningTokens, 0)),
		},
	}
}

func (c *openAIResponseStreamClient) convertStreamEventFromOpenAIResponse(event *responses.ResponseStreamEventUnion) *entity.ChatResp {
	resp := &entity.ChatResp{
		Id:    c.req.Id,
		Model: c.respModel,
	}

	switch event.Type {
	case "response.created":
		c.respModel = event.Response.Model
		c.messageID = event.Response.ID
		return nil

	case "response.output_item.added":
		switch event.Item.Type {
		case "message":
			phase := contentPhaseFromOpenAIPhase(event.Item.Phase)
			c.currentOutputItemPhase = phase
			resp.Message = &v1.Message{
				Id:   c.messageID,
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{{
					Id:      event.Item.ID,
					Index:   new(uint32(event.OutputIndex)),
					Phase:   phase,
					Content: &v1.Content_Text{},
				}},
			}

		case "function_call":
			c.hasToolCall = true
			resp.Message = &v1.Message{
				Id:   c.messageID,
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{{
					Id:    event.Item.ID,
					Index: new(uint32(event.OutputIndex)),
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   event.Item.CallID,
							Name: event.Item.Name,
						},
					},
				}},
			}

		default:
			return nil
		}

	case "response.output_text.delta":
		resp.Message = &v1.Message{
			Id:   c.messageID,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:   new(uint32(event.OutputIndex)),
				Phase:   c.currentOutputItemPhase,
				Content: &v1.Content_Text{Text: event.Delta},
			}},
		}

	case "response.refusal.delta":
		c.hasRefused = true
		resp.Message = &v1.Message{
			Id:   c.messageID,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Index:   new(uint32(event.OutputIndex)),
				Phase:   c.currentOutputItemPhase,
				Content: &v1.Content_Text{Text: event.Delta},
			}},
		}

	case "response.reasoning_summary_text.delta":
		resp.Message = &v1.Message{
			Id:   c.messageID,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Id:      event.ItemID,
				Index:   new(uint32(event.OutputIndex)),
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING_SUMMARY,
				Content: &v1.Content_Text{Text: event.Delta},
			}},
		}

	case "response.reasoning_text.delta":
		resp.Message = &v1.Message{
			Id:   c.messageID,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Id:      event.ItemID,
				Index:   new(uint32(event.OutputIndex)),
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Text{Text: event.Delta},
			}},
		}

	case "response.output_item.done":
		if event.Item.Type == "reasoning" && event.Item.EncryptedContent != "" {
			resp.Message = &v1.Message{
				Id:   c.messageID,
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{{
					Id:      event.Item.ID,
					Index:   new(uint32(event.OutputIndex)),
					Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{Opaque: event.Item.EncryptedContent},
				}},
			}
		}

	case "response.function_call_arguments.delta":
		resp.Message = &v1.Message{
			Id:   c.messageID,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Id:    event.ItemID,
				Index: new(uint32(event.OutputIndex)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{{
							Input: &v1.ToolUse_Input_Text{Text: event.Delta},
						}},
					},
				},
			}},
		}

	case "response.completed", "response.failed", "response.incomplete":
		resp.Status = convertStatusFromOpenAIResponse(event.Response.Status)
		if c.hasRefused {
			resp.Status = v1.ChatStatus_CHAT_REFUSED
		} else if c.hasToolCall && resp.Status == v1.ChatStatus_CHAT_COMPLETED {
			resp.Status = v1.ChatStatus_CHAT_PENDING_TOOL_USE
		}
		resp.Statistics = convertStatisticsFromOpenAIResponse(&event.Response.Usage)

	default:
		return nil
	}

	if resp.Message == nil && resp.Statistics == nil {
		return nil
	}

	return resp
}
