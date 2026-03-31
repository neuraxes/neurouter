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

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

func convertConfigToOpenAIChat(config *v1.GenerationConfig, req *openai.ChatCompletionNewParams) {
	if config == nil {
		return
	}
	if config.MaxTokens != nil {
		req.MaxCompletionTokens = openai.Opt(*config.MaxTokens)
	}
	if config.Temperature != nil {
		req.Temperature = openai.Opt(float64(*config.Temperature))
	}
	if config.TopP != nil {
		req.TopP = openai.Opt(float64(*config.TopP))
	}
	if config.FrequencyPenalty != nil {
		req.FrequencyPenalty = openai.Opt(float64(*config.FrequencyPenalty))
	}
	if config.PresencePenalty != nil {
		req.PresencePenalty = openai.Opt(float64(*config.PresencePenalty))
	}

	if c := config.ReasoningConfig; c != nil && c.Effort > v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED {
		req.ReasoningEffort = convertEffortToOpenAI(c.Effort)
	}

	// Convert grammar to OpenAI response format
	switch g := config.Grammar.(type) {
	case *v1.GenerationConfig_PresetGrammar:
		if g.PresetGrammar == "json_object" {
			req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
			}
		}
	case *v1.GenerationConfig_JsonSchema:
		var schema map[string]any
		if err := json.Unmarshal([]byte(g.JsonSchema), &schema); err == nil {
			req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
						Name:   "custom_schema",
						Schema: schema,
					},
				},
			}
		}
	case *v1.GenerationConfig_Schema:
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "custom_schema",
					Schema: g.Schema,
				},
			},
		}
	}
}

func (r *upstream) convertMessageToOpenAIChat(message *v1.Message) []openai.ChatCompletionMessageParamUnion {
	plainText := ""
	isPlainText := true

	{
		var sb strings.Builder
		for _, content := range message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				sb.WriteString(c.Text)
			case *v1.Content_Image:
				isPlainText = false
			}
		}
		plainText = sb.String()
	}

	switch message.Role {
	case v1.Role_SYSTEM:
		m := &openai.ChatCompletionSystemMessageParam{}

		if message.Name != "" {
			m.Name = openai.Opt(message.Name)
		}

		if isPlainText && r.config.PreferStringContentForSystem {
			m.Content.OfString = openai.Opt(plainText)
		} else if isPlainText && r.config.PreferSinglePartContent {
			m.Content.OfArrayOfContentParts = append(
				m.Content.OfArrayOfContentParts,
				openai.ChatCompletionContentPartTextParam{Text: plainText},
			)
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					m.Content.OfArrayOfContentParts = append(
						m.Content.OfArrayOfContentParts,
						openai.ChatCompletionContentPartTextParam{Text: c.Text},
					)
				default:
					r.log.Errorf("unsupported content for system: %v", c)
				}
			}
		}

		return []openai.ChatCompletionMessageParamUnion{{OfSystem: m}}
	case v1.Role_USER:
		var result []openai.ChatCompletionMessageParamUnion
		var userContents []*v1.Content

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_ToolResult:
				toolMsg := &openai.ChatCompletionToolMessageParam{
					ToolCallID: c.ToolResult.Id,
				}

				outputText := c.ToolResult.GetTextualOutput()
				if r.config.PreferStringContentForTool {
					toolMsg.Content.OfString = openai.Opt(outputText)
				} else if r.config.PreferSinglePartContent {
					toolMsg.Content.OfArrayOfContentParts = append(
						toolMsg.Content.OfArrayOfContentParts,
						openai.ChatCompletionContentPartTextParam{Text: outputText},
					)
				} else {
					for _, content := range c.ToolResult.Outputs {
						switch c := content.GetOutput().(type) {
						case *v1.ToolResult_Output_Text:
							toolMsg.Content.OfArrayOfContentParts = append(
								toolMsg.Content.OfArrayOfContentParts,
								openai.ChatCompletionContentPartTextParam{Text: c.Text},
							)
						default:
							r.log.Errorf("unsupported content for tool result: %v", c)
						}
					}
				}

				result = append(result, openai.ChatCompletionMessageParamUnion{OfTool: toolMsg})
			default:
				userContents = append(userContents, content)
			}
		}

		if len(result) == 0 || len(userContents) > 0 {
			m := &openai.ChatCompletionUserMessageParam{}

			if message.Name != "" {
				m.Name = openai.Opt(message.Name)
			}

			if isPlainText && r.config.PreferStringContentForUser {
				m.Content.OfString = openai.Opt(plainText)
			} else if isPlainText && r.config.PreferSinglePartContent {
				m.Content.OfArrayOfContentParts = append(
					m.Content.OfArrayOfContentParts,
					openai.TextContentPart(plainText),
				)
			} else {
				for _, content := range userContents {
					switch c := content.GetContent().(type) {
					case *v1.Content_Text:
						m.Content.OfArrayOfContentParts = append(
							m.Content.OfArrayOfContentParts,
							openai.TextContentPart(c.Text),
						)
					case *v1.Content_Image:
						m.Content.OfArrayOfContentParts = append(
							m.Content.OfArrayOfContentParts,
							openai.ImageContentPart(
								openai.ChatCompletionContentPartImageImageURLParam{
									URL: c.Image.GetUrl(),
								},
							),
						)
					default:
						r.log.Errorf("unsupported content for user: %v", c)
					}
				}
			}

			result = append(result, openai.ChatCompletionMessageParamUnion{OfUser: m})
		}

		return result
	case v1.Role_MODEL:
		m := &openai.ChatCompletionAssistantMessageParam{}

		if message.Name != "" {
			m.Name = openai.Opt(message.Name)
		}

		if isPlainText && r.config.PreferStringContentForAssistant {
			m.Content.OfString = openai.Opt(plainText)
		} else if isPlainText && r.config.PreferSinglePartContent {
			m.Content.OfArrayOfContentParts = append(
				m.Content.OfArrayOfContentParts,
				openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
					OfText: &openai.ChatCompletionContentPartTextParam{
						Text: plainText,
					},
				},
			)
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					// Reasoning content should be ignored if it has the reasoning flag
					if content.Reasoning {
						continue
					}
					m.Content.OfArrayOfContentParts = append(
						m.Content.OfArrayOfContentParts,
						openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
							OfText: &openai.ChatCompletionContentPartTextParam{
								Text: c.Text,
							},
						},
					)
				case *v1.Content_ToolUse:
					// Tool calls will be processed later
				default:
					r.log.Errorf("unsupported content for assistant: %v", c)
				}
			}
		}

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_ToolUse:
				m.ToolCalls = append(m.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: c.ToolUse.Id,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      c.ToolUse.Name,
							Arguments: c.ToolUse.GetTextualInput(),
						},
					},
				})
			}
		}

		return []openai.ChatCompletionMessageParamUnion{{OfAssistant: m}}
	default:
		r.log.Errorf("invalid role: %v", message.Role)
		return nil
	}
}

func convertSchemaToMap(parameters *v1.Schema) openai.FunctionParameters {
	data, err := json.Marshal(parameters)
	if err != nil {
		return nil
	}
	var params openai.FunctionParameters
	if err := json.Unmarshal(data, &params); err != nil {
		return nil
	}
	if parameters.Type == v1.Schema_TYPE_OBJECT && params["properties"] == nil {
		params["properties"] = map[string]any{}
	}
	return params
}

func (r *upstream) convertRequestToOpenAIChat(req *entity.ChatReq) openai.ChatCompletionNewParams {
	openAIReq := openai.ChatCompletionNewParams{
		Model: req.Model,
	}

	if req.Config != nil {
		convertConfigToOpenAIChat(req.Config, &openAIReq)
	}

	for _, message := range req.Messages {
		m := r.convertMessageToOpenAIChat(message)
		openAIReq.Messages = append(openAIReq.Messages, m...)
	}

	if req.Tools != nil {
		var tools []openai.ChatCompletionToolUnionParam
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				ot := openai.FunctionDefinitionParam{
					Name:       t.Function.Name,
					Parameters: convertSchemaToMap(t.Function.Parameters),
				}
				if t.Function.Description != "" {
					ot.Description = openai.Opt(t.Function.Description)
				}
				tools = append(tools, openai.ChatCompletionFunctionTool(ot))
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		openAIReq.Tools = tools
	}

	return openAIReq
}

func convertStatusFromOpenAIChat(finishReason string) v1.ChatStatus {
	switch finishReason {
	case "stop":
		return v1.ChatStatus_CHAT_COMPLETED
	case "length":
		return v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT
	case "tool_calls", "function_call":
		return v1.ChatStatus_CHAT_PENDING_TOOL_USE
	case "content_filter":
		return v1.ChatStatus_CHAT_REFUSED
	default:
		return v1.ChatStatus_CHAT_IN_PROGRESS
	}
}

// convertMessageFromOpenAIChat converts an OpenAI chat completion message to an internal message.
// The message ID will be generated using UUID.
func (r *upstream) convertMessageFromOpenAIChat(openAIMessage *openai.ChatCompletionMessage) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	if openAIMessage.Content != "" {
		message.Contents = append(message.Contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: openAIMessage.Content,
			},
		})
	}

	// Support reasoning
	if openAIMessage.JSON.ExtraFields != nil {
		reasoning, ok := openAIMessage.JSON.ExtraFields["reasoning_content"] // DeepSeek
		if !ok {
			reasoning, ok = openAIMessage.JSON.ExtraFields["reasoning"] // OpenRouter
		}
		if ok {
			rc := gjson.Parse(reasoning.Raw()).String()
			if rc != "" {
				message.Contents = append(message.Contents, &v1.Content{
					Reasoning: true,
					Content: &v1.Content_Text{
						Text: rc,
					},
				})
			}
		}
	}

	for _, toolCall := range openAIMessage.ToolCalls {
		// Only function tool calls are supported by OpenAI
		message.Contents = append(message.Contents, &v1.Content{
			Content: &v1.Content_ToolUse{
				ToolUse: &v1.ToolUse{
					Id:   toolCall.ID,
					Name: toolCall.Function.Name,
					Inputs: []*v1.ToolUse_Input{
						{
							Input: &v1.ToolUse_Input_Text{
								Text: toolCall.Function.Arguments,
							},
						},
					},
				},
			},
		})
	}

	return message
}

func (r *upstream) convertResponseFromOpenAIChat(openAIResp *openai.ChatCompletion) (resp *entity.ChatResp) {
	resp = &entity.ChatResp{
		Id:         openAIResp.ID,
		Model:      openAIResp.Model,
		Statistics: convertStatisticsFromOpenAI(&openAIResp.Usage),
	}

	if len(openAIResp.Choices) > 0 {
		resp.Status = convertStatusFromOpenAIChat(openAIResp.Choices[0].FinishReason)
		resp.Message = r.convertMessageFromOpenAIChat(&openAIResp.Choices[0].Message)
	}

	return
}

func (c *openAIChatStreamClient) convertChunkFromOpenAIChat(chunk *openai.ChatCompletionChunk) *entity.ChatResp {
	resp := &entity.ChatResp{
		Id:    chunk.ID,
		Model: chunk.Model,
	}

	if len(chunk.Choices) > 0 {
		msg := chunk.Choices[0]
		var contents []*v1.Content

		if msg.Delta.Content != "" {
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: msg.Delta.Content,
				},
			})
		}

		// Support reasoning content from DeepSeek / OpenRouter
		if msg.Delta.JSON.ExtraFields != nil {
			reasoning, ok := msg.Delta.JSON.ExtraFields["reasoning_content"] // DeepSeek
			if !ok {
				reasoning, ok = msg.Delta.JSON.ExtraFields["reasoning"] // OpenRouter
			}
			if ok {
				rc := gjson.Parse(reasoning.Raw()).String()
				if rc != "" {
					contents = append(contents, &v1.Content{
						Reasoning: true,
						Content: &v1.Content_Text{
							Text: rc,
						},
					})
				}
			}
		}

		for _, toolCall := range msg.Delta.ToolCalls {
			contents = append(contents, &v1.Content{
				Index: new(uint32(toolCall.Index)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   toolCall.ID,
						Name: toolCall.Function.Name,
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: toolCall.Function.Arguments,
								},
							},
						},
					},
				},
			})
		}

		c.status = convertStatusFromOpenAIChat(msg.FinishReason)
		resp.Status = c.status
		resp.Message = &v1.Message{
			Id:       c.messageID,
			Role:     v1.Role_MODEL,
			Contents: contents,
		}
	} else {
		resp.Status = c.status // Keep previous status
	}

	resp.Statistics = convertStatisticsFromOpenAI(&chunk.Usage)

	if resp.Message == nil && resp.Statistics == nil {
		return nil
	}

	return resp
}

func convertStatisticsFromOpenAI(usage *openai.CompletionUsage) *v1.Statistics {
	if usage == nil || (usage.PromptTokens == 0 && usage.CompletionTokens == 0 && usage.PromptTokensDetails.CachedTokens == 0) {
		return nil
	}

	return &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:       uint32(usage.PromptTokens),
			OutputTokens:      uint32(usage.CompletionTokens),
			CachedInputTokens: uint32(usage.PromptTokensDetails.CachedTokens),
		},
	}
}
