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

	"github.com/google/uuid"

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

	// Concatenate text contents
	var content strings.Builder
	for _, c := range message.Contents {
		if textContent, ok := c.GetContent().(*v1.Content_Text); ok {
			content.WriteString(textContent.Text)
		} else {
			r.log.Errorf("unsupported content type: %v", c)
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

	if message.ToolCalls != nil {
		var toolCalls []*ToolCall
		for _, toolCall := range message.ToolCalls {
			switch t := toolCall.Tool.(type) {
			case *v1.ToolCall_Function:
				toolCalls = append(toolCalls, &ToolCall{
					ID:   toolCall.Id,
					Type: "function",
					Function: &FunctionCall{
						Name:      t.Function.Name,
						Arguments: t.Function.Arguments,
					},
				})
			default:
				r.log.Errorf("unsupported tool call: %v", t)
			}
		}
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

func toolFunctionParametersToDeepSeek(params *v1.Tool_Function_Parameters) map[string]any {
	return map[string]any{
		"type":       params.Type,
		"properties": params.Properties,
		"required":   params.Required,
	}
}

// convertMessageFromDeepSeek converts a message from the DeepSeek API to an internal message.
func (r *ChatRepo) convertMessageFromDeepSeek(deepSeekMessage *Message) *v1.Message {

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
		Id:         uuid.NewString(),
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
		var toolCalls []*v1.ToolCall
		for _, toolCall := range deepSeekMessage.ToolCalls {
			if toolCall.Type == "function" {
				toolCalls = append(toolCalls, &v1.ToolCall{
					Id: toolCall.ID,
					Tool: &v1.ToolCall_Function{
						Function: &v1.ToolCall_FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					},
				})
			} else {
				r.log.Errorf("unsupported tool call type: %v", toolCall.Type)
			}
		}
		message.ToolCalls = toolCalls
	}

	return message
}
