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

	"github.com/sashabaranov/go-openai"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
)

func convertChatMessageFromOpenAI(message *openai.ChatCompletionMessage) *v1.Message {
	var role v1.Role
	switch message.Role {
	case openai.ChatMessageRoleSystem:
		role = v1.Role_SYSTEM
	case openai.ChatMessageRoleUser:
		role = v1.Role_USER
	case openai.ChatMessageRoleAssistant:
		role = v1.Role_MODEL
	case openai.ChatMessageRoleTool:
		role = v1.Role_TOOL
	}

	var contents []*v1.Content
	if message.Content != "" {
		// Single text message
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: message.Content,
			},
		})
	} else {
		// Multipart message
		for _, content := range message.MultiContent {
			switch content.Type {
			case openai.ChatMessagePartTypeText:
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Text{
						Text: content.Text,
					},
				})
			case openai.ChatMessagePartTypeImageURL:
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Image_{
						Image: &v1.Content_Image{
							Url: content.ImageURL.URL,
						},
					},
				})
			}
		}
	}

	var toolCalls []*v1.ToolCall
	for _, toolCall := range message.ToolCalls {
		t := &v1.ToolCall{Id: toolCall.ID}
		switch toolCall.Type {
		case openai.ToolTypeFunction:
			t.Tool = &v1.ToolCall_Function{
				Function: &v1.ToolCall_FunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			}
		}
		toolCalls = append(toolCalls, t)
	}

	return &v1.Message{
		Role:       role,
		Name:       message.Name,
		Contents:   contents,
		ToolCalls:  toolCalls,
		ToolCallId: message.ToolCallID,
	}
}

// convertChatReqFromOpenAI converts a chat completion request from OpenAI API to Router API
func convertChatReqFromOpenAI(req *openai.ChatCompletionRequest) *v1.ChatReq {
	config := &v1.GenerationConfig{
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	if req.MaxCompletionTokens != 0 {
		config.MaxTokens = int64(req.MaxCompletionTokens)
	} else if req.MaxTokens != 0 {
		config.MaxTokens = int64(req.MaxTokens)
	}
	if req.ResponseFormat != nil {
		switch req.ResponseFormat.Type {
		case openai.ChatCompletionResponseFormatTypeJSONObject:
			config.Grammar = &v1.GenerationConfig_PresetGrammar{
				PresetGrammar: "json_object",
			}
		}
	}

	var messages []*v1.Message
	for _, message := range req.Messages {
		messages = append(messages, convertChatMessageFromOpenAI(&message))
	}

	var tools []*v1.Tool
	for _, tool := range req.Tools {
		t := &v1.Tool{}
		switch tool.Type {
		case openai.ToolTypeFunction:
			var parameters *v1.Tool_Function_Parameters
			j, _ := json.Marshal(tool.Function.Parameters)
			_ = json.Unmarshal(j, &parameters)
			t.Tool = &v1.Tool_Function_{
				Function: &v1.Tool_Function{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  parameters,
				},
			}
		default:
			// TODO: Handle other tool types
			continue
		}
		tools = append(tools, t)
	}

	return &v1.ChatReq{
		Model:    req.Model,
		Config:   config,
		Messages: messages,
		Tools:    tools,
	}
}

// convertChatRespToOpenAI converts a chat completion response from Router API to OpenAI API
func convertChatRespToOpenAI(resp *v1.ChatResp) *openai.ChatCompletionResponse {
	openAIResp := &openai.ChatCompletionResponse{
		ID: resp.Message.Id,
	}

	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		openAIResp.Choices = []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: resp.Message.Contents[0].GetText(),
				},
			},
		}
	}

	if resp.Statistics != nil {
		openAIResp.Usage.PromptTokens = int(resp.Statistics.Usage.PromptTokens)
		openAIResp.Usage.CompletionTokens = int(resp.Statistics.Usage.CompletionTokens)
	}

	return openAIResp
}
