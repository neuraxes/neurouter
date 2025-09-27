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

package deepseek

import (
	"strings"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// convertMessageToDeepSeek converts an internal message to a message that can be sent to the DeepSeek API.
func (r *ChatRepo) convertMessageToDeepSeek(message *v1.Message) *Message {
	// Convert role
	var role string
	switch message.Role {
	case v1.Role_SYSTEM:
		role = "system"
	case v1.Role_USER:
		role = "user"
	case v1.Role_MODEL:
		role = "assistant"
	case v1.Role_TOOL:
		role = "tool"
	}

	var content strings.Builder
	var toolCalls []*ToolCall
	for _, c := range message.Contents {
		switch cc := c.GetContent().(type) {
		case *v1.Content_Text:
			content.WriteString(cc.Text)
		case *v1.Content_FunctionCall:
			f := cc.FunctionCall
			toolCalls = append(toolCalls, &ToolCall{
				ID:   f.GetId(),
				Type: "function",
				Function: &FunctionCall{
					Name:      f.GetName(),
					Arguments: f.GetArguments(),
				},
			})
		default:
			r.log.Errorf("unsupported content type: %T", cc)
		}
	}

	// Build DeepSeek message
	deepseekMsg := &Message{
		Role:       role,
		Name:       message.Name,
		ToolCallID: message.ToolCallId,
	}

	if s := content.String(); s != "" {
		deepseekMsg.Content = s
	}
	if len(toolCalls) > 0 {
		deepseekMsg.ToolCalls = toolCalls
	}

	return deepseekMsg
}

// convertRequestToDeepSeek converts an internal request to a request that can be sent to the DeepSeek API.
func (r *ChatRepo) convertRequestToDeepSeek(req *entity.ChatReq) *ChatRequest {
	var messages []*Message
	for _, message := range req.Messages {
		m := r.convertMessageToDeepSeek(message)
		if m != nil {
			messages = append(messages, m)
		}
	}

	deepseekReq := &ChatRequest{
		Model:    req.Model,
		Messages: messages,
	}

	if c := req.Config; c != nil {
		if c.MaxTokens != 0 {
			deepseekReq.MaxTokens = int(c.MaxTokens)
		}
		deepseekReq.Temperature = float64(c.Temperature)
		if c.TopP != 0 {
			deepseekReq.TopP = float64(c.TopP)
		}
		deepseekReq.FrequencyPenalty = float64(c.FrequencyPenalty)
		deepseekReq.PresencePenalty = float64(c.PresencePenalty)
		if c.GetPresetGrammar() == "json_object" {
			deepseekReq.ResponseFormat = &ResponseFormat{
				Type: "json_object",
			}
		}
	}

	if req.Tools != nil {
		var tools []*Tool
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				tools = append(tools, &Tool{
					Type: "function",
					Function: &FunctionDefinition{
						Name:        t.Function.Name,
						Description: t.Function.Description,
						Parameters:  toolFunctionParametersToDeepSeek(t.Function.Parameters),
					},
				})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		deepseekReq.Tools = tools
	}

	return deepseekReq
}

func toolFunctionParametersToDeepSeek(params *v1.Schema) map[string]any {
	return map[string]any{
		"type":       params.Type,
		"properties": params.Properties,
		"required":   params.Required,
	}
}

// convertMessageFromDeepSeek converts a message from the DeepSeek API to an internal message.
func (r *ChatRepo) convertMessageFromDeepSeek(messageID string, deepSeekMessage *Message) *v1.Message {
	// Convert role
	var role v1.Role
	switch deepSeekMessage.Role {
	case "system":
		role = v1.Role_SYSTEM
	case "user":
		role = v1.Role_USER
	case "assistant":
		role = v1.Role_MODEL
	case "tool":
		role = v1.Role_TOOL
	default:
		r.log.Errorf("unsupported role: %v", deepSeekMessage.Role)
		return nil
	}

	message := &v1.Message{
		Id:         messageID,
		Role:       role,
		Name:       deepSeekMessage.Name,
		ToolCallId: deepSeekMessage.ToolCallID,
	}

	if deepSeekMessage.ReasoningContent != "" {
		message.Contents = append(message.Contents, &v1.Content{
			Content: &v1.Content_Thinking{
				Thinking: strings.TrimSpace(deepSeekMessage.ReasoningContent),
			},
		})
	}

	if deepSeekMessage.Content != "" {
		message.Contents = append(message.Contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: strings.TrimSpace(deepSeekMessage.Content),
			},
		})
	}

	if deepSeekMessage.ToolCalls != nil {
		for _, toolCall := range deepSeekMessage.ToolCalls {
			if toolCall.Type == "function" && toolCall.Function != nil {
				message.Contents = append(message.Contents, &v1.Content{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Id:        toolCall.ID,
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					},
				})
			} else {
				r.log.Errorf("unsupported tool call type: %v", toolCall.Type)
			}
		}
	}

	return message
}

func convertStreamRespFromDeepSeek(requestID string, chunk *ChatStreamResponse) *entity.ChatResp {
	resp := &entity.ChatResp{
		Id:    requestID,
		Model: chunk.Model,
	}

	if len(chunk.Choices) > 0 {
		var contents []*v1.Content
		for _, choice := range chunk.Choices {
			if choice.Delta.ReasoningContent != "" {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Thinking{
						Thinking: choice.Delta.ReasoningContent,
					},
				})
			}
			if choice.Delta.Content != "" {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Text{
						Text: choice.Delta.Content,
					},
				})
			}
		}

		resp.Message = &v1.Message{
			Id:       chunk.ID,
			Role:     v1.Role_MODEL,
			Contents: contents,
		}

		// Clear due to the reuse of the same message struct
		chunk.Choices[0].Delta = nil
	}

	// Map usage statistics if present
	resp.Statistics = convertStatisticsFromDeepSeek(chunk.Usage)

	return resp
}

func convertStatisticsFromDeepSeek(usage *Usage) *v1.Statistics {
	if usage == nil {
		return nil
	}

	return &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			PromptTokens:     int32(usage.PromptTokens),
			CompletionTokens: int32(usage.CompletionTokens),
		},
	}
}
