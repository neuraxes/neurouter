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

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// convertMessageToOpenAI converts an internal message to a message that can be sent to the OpenAI API.
func (r *ChatRepo) convertMessageToOpenAI(message *v1.Message) openai.ChatCompletionMessageParamUnion {
	plainText := ""
	isPlainText := true

	{
		var sb strings.Builder
		for _, content := range message.Contents {
			if textContent, ok := content.GetContent().(*v1.Content_Text); ok {
				sb.WriteString(textContent.Text)
			} else {
				isPlainText = false
			}
		}
		plainText = sb.String()
	}

	switch message.Role {
	case v1.Role_SYSTEM:
		m := openai.ChatCompletionSystemMessageParam{
			Role: openai.F(openai.ChatCompletionSystemMessageParamRoleSystem),
		}

		if message.Name != "" {
			m.Name = openai.F(message.Name)
		}

		var parts []openai.ChatCompletionContentPartTextParam
		if isPlainText && (r.config.PreferStringContentForSystem || r.config.PreferSinglePartContent) {
			parts = append(parts, openai.TextPart(plainText))
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				default:
					r.log.Errorf("unsupported content for system: %v", c)
				}
			}
		}
		m.Content = openai.F(parts)

		return m
	case v1.Role_USER:
		m := openai.ChatCompletionUserMessageParam{
			Role: openai.F(openai.ChatCompletionUserMessageParamRoleUser),
		}

		if message.Name != "" {
			m.Name = openai.F(message.Name)
		}

		var parts []openai.ChatCompletionContentPartUnionParam
		if isPlainText && (r.config.PreferStringContentForUser || r.config.PreferSinglePartContent) {
			parts = append(parts, openai.TextPart(plainText))
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				case *v1.Content_Image_:
					parts = append(parts, openai.ImagePart(c.Image.Url))
				default:
					r.log.Errorf("unsupported content for user: %v", c)
				}
			}
		}
		m.Content = openai.F(parts)

		return m
	case v1.Role_MODEL:
		m := openai.ChatCompletionAssistantMessageParam{
			Role: openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
		}

		if message.Name != "" {
			m.Name = openai.F(message.Name)
		}

		if message.Contents != nil {
			var parts []openai.ChatCompletionAssistantMessageParamContentUnion
			if isPlainText && (r.config.PreferStringContentForAssistant || r.config.PreferSinglePartContent) {
				parts = append(parts, openai.TextPart(plainText))
			} else {
				for _, content := range message.Contents {
					switch c := content.GetContent().(type) {
					case *v1.Content_Text:
						parts = append(parts, openai.TextPart(c.Text))
					default:
						r.log.Errorf("unsupported content for assistant: %v", c)
					}
				}
			}
			m.Content = openai.F(parts)
		}

		if message.ToolCalls != nil {
			var toolCalls []openai.ChatCompletionMessageToolCallParam
			for _, toolCall := range message.ToolCalls {
				switch t := toolCall.Tool.(type) {
				case *v1.ToolCall_Function:
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID:   openai.F(toolCall.Id),
						Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
						Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      openai.F(t.Function.Name),
							Arguments: openai.F(t.Function.Arguments),
						}),
					})
				default:
					r.log.Errorf("unsupported tool call: %v", t)
				}
			}
			m.ToolCalls = openai.F(toolCalls)
		}

		return m
	case v1.Role_TOOL:
		m := openai.ChatCompletionToolMessageParam{
			Role:       openai.F(openai.ChatCompletionToolMessageParamRoleTool),
			ToolCallID: openai.F(message.ToolCallId),
		}

		var parts []openai.ChatCompletionContentPartTextParam
		if isPlainText && (r.config.PreferStringContentForTool || r.config.PreferSinglePartContent) {
			parts = append(parts, openai.TextPart(plainText))
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				default:
					r.log.Errorf("unsupported content for tool: %v", c)
				}
			}
		}
		m.Content = openai.F(parts)

		return m
	default:
		r.log.Errorf("unsupported role: %v", message.Role)
		return nil
	}
}

// convertRequestToOpenAI converts an internal request to a request that can be sent to the OpenAI API.
func (r *ChatRepo) convertRequestToOpenAI(req *entity.ChatReq) openai.ChatCompletionNewParams {
	openAIReq := openai.ChatCompletionNewParams{
		Model: openai.F(req.Model),
	}

	for _, message := range req.Messages {
		m := r.convertMessageToOpenAI(message)
		if m != nil {
			openAIReq.Messages = append(openAIReq.Messages, m)
		}
	}

	if c := req.Config; c != nil {
		if c.MaxTokens != 0 {
			openAIReq.MaxCompletionTokens = openai.Opt(c.MaxTokens)
		}
		openAIReq.Temperature = openai.Opt(float64(c.Temperature))
		if c.TopP != 0 {
			openAIReq.TopP = openai.Opt(float64(c.TopP))
		}
		openAIReq.FrequencyPenalty = openai.Opt(float64(c.FrequencyPenalty))
		openAIReq.PresencePenalty = openai.Opt(float64(c.PresencePenalty))
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

func toolFunctionParametersToOpenAI(parameters *v1.Tool_Function_Parameters) openai.FunctionParameters {
	return map[string]any{
		"type":       parameters.Type,
		"properties": parameters.Properties,
		"required":   parameters.Required,
	}
}

func (r *ChatRepo) convertMessageFromOpenAI(openAIMessage *openai.ChatCompletionMessage) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	if openAIMessage.Content != "" {
		message.Contents = []*v1.Content{
			{
				Content: &v1.Content_Text{
					// The result may contain a leading space, so we need to trim it
					Text: strings.TrimSpace(openAIMessage.Content),
				},
			},
		}
	}

	if openAIMessage.ToolCalls != nil {
		var toolCalls []*v1.ToolCall
		for _, toolCall := range openAIMessage.ToolCalls {
			// Currently only function tool calls are supported by OpenAI
			toolCalls = append(toolCalls, &v1.ToolCall{
				Id: toolCall.ID,
				Tool: &v1.ToolCall_Function{
					Function: &v1.ToolCall_FunctionCall{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				},
			})
		}
		message.ToolCalls = toolCalls
	}

	return message
}

// convertResponseFromOpenAI converts an OpenAI chat completion to an internal response.
func (r *ChatRepo) convertResponseFromOpenAI(res *openai.ChatCompletion) *entity.ChatResp {
	resp := &entity.ChatResp{
		Id:      res.ID,
		Message: r.convertMessageFromOpenAI(&res.Choices[0].Message),
	}

	if res.Usage.PromptTokens != 0 || res.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(res.Usage.PromptTokens),
				CompletionTokens: int32(res.Usage.CompletionTokens),
			},
		}
	}

	return resp
}

// convertChunkFromOpenAI converts an OpenAI chat completion chunk to an internal response.
func convertChunkFromOpenAI(chunk *openai.ChatCompletionChunk, requestID string, messageID string) *entity.ChatResp {
	resp := &entity.ChatResp{
		Id: requestID,
	}

	if len(chunk.Choices) > 0 {
		resp.Message = &v1.Message{
			Id:   messageID,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Choices[0].Delta.Content,
					},
				},
			},
		}
	}

	if chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(chunk.Usage.PromptTokens),
				CompletionTokens: int32(chunk.Usage.CompletionTokens),
			},
		}
	}

	return resp
}
