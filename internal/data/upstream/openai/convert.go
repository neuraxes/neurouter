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
	"strings"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// convertMessageToOpenAI converts an internal message to a message that can be sent to the OpenAI API.
func (r *upstream) convertMessageToOpenAI(message *v1.Message) []openai.ChatCompletionMessageParamUnion {
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

		// Check if there are any tool results that need to be split
		var toolResults []*v1.ToolResult
		var normalContents []*v1.Content
		for _, content := range message.Contents {
			if tr := content.GetToolResult(); tr != nil {
				toolResults = append(toolResults, tr)
			} else {
				normalContents = append(normalContents, content)
			}
		}

		// First, add tool result messages if any
		for _, toolResult := range toolResults {
			toolMsg := &openai.ChatCompletionToolMessageParam{
				ToolCallID: toolResult.Id,
			}

			outputText := toolResult.GetTextualOutput()
			if r.config.PreferStringContentForTool {
				toolMsg.Content.OfString = openai.Opt(outputText)
			} else if isPlainText && r.config.PreferSinglePartContent {
				toolMsg.Content.OfArrayOfContentParts = append(
					toolMsg.Content.OfArrayOfContentParts,
					openai.ChatCompletionContentPartTextParam{Text: outputText},
				)
			} else {
				for _, content := range toolResult.Outputs {
					switch c := content.GetOutput().(type) {
					case *v1.ToolResult_Output_Text:
						toolMsg.Content.OfArrayOfContentParts = append(
							toolMsg.Content.OfArrayOfContentParts,
							openai.ChatCompletionContentPartTextParam{Text: c.Text},
						)
					default:
						r.log.Errorf("unsupported content for tool: %v", c)
					}
				}

			}

			result = append(result, openai.ChatCompletionMessageParamUnion{OfTool: toolMsg})
		}

		// Then, add user message for normal contents
		// At least one user message should be added
		if len(result) == 0 || len(normalContents) > 0 {
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
				for _, content := range normalContents {
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
				f := c.ToolUse
				m.ToolCalls = append(m.ToolCalls, openai.ChatCompletionMessageToolCallParam{
					ID: f.Id,
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      f.Name,
						Arguments: f.GetTextualInput(),
					},
				})
			}
		}

		return []openai.ChatCompletionMessageParamUnion{{OfAssistant: m}}
	default:
		r.log.Errorf("unsupported role: %v", message.Role)
		return nil
	}
}

// convertRequestToOpenAI converts an internal request to a request that can be sent to the OpenAI API.
func (r *upstream) convertRequestToOpenAI(req *entity.ChatReq) openai.ChatCompletionNewParams {
	openAIReq := openai.ChatCompletionNewParams{
		Model: req.Model,
	}

	for _, message := range req.Messages {
		m := r.convertMessageToOpenAI(message)
		if m != nil {
			openAIReq.Messages = append(openAIReq.Messages, m...)
		}
	}

	if c := req.Config; c != nil {
		if c.MaxTokens != nil {
			openAIReq.MaxCompletionTokens = openai.Opt(*c.MaxTokens)
		}
		if c.Temperature != nil {
			openAIReq.Temperature = openai.Opt(float64(*c.Temperature))
		}
		if c.TopP != nil {
			openAIReq.TopP = openai.Opt(float64(*c.TopP))
		}
		if c.FrequencyPenalty != nil {
			openAIReq.FrequencyPenalty = openai.Opt(float64(*c.FrequencyPenalty))
		}
		if c.PresencePenalty != nil {
			openAIReq.PresencePenalty = openai.Opt(float64(*c.PresencePenalty))
		}
		if c.GetPresetGrammar() == "json_object" {
			openAIReq.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
			}
		}
	}

	if req.Tools != nil {
		var tools []openai.ChatCompletionToolParam
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				// Currently only function tool calls are supported by OpenAI
				tools = append(tools, openai.ChatCompletionToolParam{
					Function: shared.FunctionDefinitionParam{
						Name:        t.Function.Name,
						Description: openai.Opt(t.Function.Description),
						Parameters:  toolFunctionParametersToOpenAI(t.Function.Parameters),
					},
				})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		openAIReq.Tools = tools
	}

	return openAIReq
}

// toolFunctionParametersToOpenAI converts tool function parameters to OpenAI function parameters.
func toolFunctionParametersToOpenAI(parameters *v1.Schema) openai.FunctionParameters {
	return map[string]any{
		"type":       parameters.Type,
		"properties": parameters.Properties,
		"required":   parameters.Required,
	}
}

// convertMessageFromOpenAI converts an OpenAI chat completion message to an internal message.
// The message ID will be generated using UUID.
func (r *upstream) convertMessageFromOpenAI(openAIMessage *openai.ChatCompletionMessage) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	if openAIMessage.Content != "" {
		message.Contents = append(message.Contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: strings.TrimSpace(openAIMessage.Content),
			},
		})
	}

	// Support reasoning content from DeepSeek
	if openAIMessage.JSON.ExtraFields != nil {
		if reasoningContent, ok := openAIMessage.JSON.ExtraFields["reasoning_content"]; ok {
			rc := gjson.Parse(reasoningContent.Raw()).String()
			if rc != "" {
				message.Contents = append(message.Contents, &v1.Content{
					Content: &v1.Content_Reasoning{
						Reasoning: rc,
					},
				})
			}
		}
	}

	if openAIMessage.ToolCalls != nil {
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
	}

	return message
}

// convertChunkFromOpenAI converts an OpenAI chat completion chunk to an internal response.
func convertChunkFromOpenAI(chunk *openai.ChatCompletionChunk) *entity.ChatResp {
	resp := &entity.ChatResp{
		Id:    chunk.ID,
		Model: chunk.Model,
	}

	if len(chunk.Choices) > 0 {
		c := chunk.Choices[0]
		var contents []*v1.Content

		if c.Delta.Content != "" {
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: c.Delta.Content,
				},
			})
		}

		// Support reasoning content from DeepSeek
		if c.Delta.JSON.ExtraFields != nil {
			if reasoningContent, ok := c.Delta.JSON.ExtraFields["reasoning_content"]; ok {
				rc := gjson.Parse(reasoningContent.Raw()).String()
				if rc != "" {
					contents = append(contents, &v1.Content{
						Content: &v1.Content_Reasoning{
							Reasoning: rc,
						},
					})
				}
			}
		}

		if c.Delta.ToolCalls != nil {
			for _, toolCall := range c.Delta.ToolCalls {
				contents = append(contents, &v1.Content{
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
		}

		resp.Message = &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: contents,
		}
	}

	resp.Statistics = convertStatisticsFromOpenAI(&chunk.Usage)

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
