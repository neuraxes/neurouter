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
func (r *upstream) convertMessageToOpenAI(message *v1.Message) *openai.ChatCompletionMessageParamUnion {
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

		return &openai.ChatCompletionMessageParamUnion{OfSystem: m}
	case v1.Role_USER:
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
			for _, content := range message.Contents {
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

		return &openai.ChatCompletionMessageParamUnion{OfUser: m}
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
				case *v1.Content_FunctionCall:
					// Tool calls will be processed later
				default:
					r.log.Errorf("unsupported content for assistant: %v", c)
				}
			}
		}

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_FunctionCall:
				f := c.FunctionCall
				m.ToolCalls = append(m.ToolCalls, openai.ChatCompletionMessageToolCallParam{
					ID: f.Id,
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      f.Name,
						Arguments: f.Arguments,
					},
				})
			}
		}

		return &openai.ChatCompletionMessageParamUnion{OfAssistant: m}
	case v1.Role_TOOL:
		m := &openai.ChatCompletionToolMessageParam{
			ToolCallID: message.ToolCallId,
		}

		if isPlainText && r.config.PreferStringContentForTool {
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
					r.log.Errorf("unsupported content for tool: %v", c)
				}
			}
		}

		return &openai.ChatCompletionMessageParamUnion{OfTool: m}
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
			openAIReq.Messages = append(openAIReq.Messages, *m)
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

	if openAIMessage.ToolCalls != nil {
		for _, toolCall := range openAIMessage.ToolCalls {
			// Only function tool calls are supported by OpenAI
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_FunctionCall{
					FunctionCall: &v1.FunctionCall{
						Id:        toolCall.ID,
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
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
		if c.Delta.ToolCalls != nil {
			for _, toolCall := range c.Delta.ToolCalls {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Id:        toolCall.ID,
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
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
			PromptTokens:       uint32(usage.PromptTokens),
			CompletionTokens:   uint32(usage.CompletionTokens),
			CachedPromptTokens: uint32(usage.PromptTokensDetails.CachedTokens),
		},
	}
}
