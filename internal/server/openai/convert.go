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

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
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
					Content: &v1.Content_Image{
						Image: &v1.Image{
							Source: &v1.Image_Url{
								Url: content.ImageURL.URL,
							},
						},
					},
				})
			}
		}
	}

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
		contents = append(contents, &v1.Content{
			Content: &v1.Content_ToolCall{ToolCall: t},
		})
	}

	return &v1.Message{
		Role:       role,
		Name:       message.Name,
		Contents:   contents,
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

	if resp.Message != nil {
		message := openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleAssistant,
		}

		// Handle contents
		if len(resp.Message.Contents) > 0 {
			if len(resp.Message.Contents) == 1 && resp.Message.Contents[0].GetText() != "" {
				// Single text message
				message.Content = resp.Message.Contents[0].GetText()
			} else {
				// Multipart message
				var multiContent []openai.ChatMessagePart
				var toolCalls []openai.ToolCall
				for _, content := range resp.Message.Contents {
					switch c := content.Content.(type) {
					case *v1.Content_Text:
						multiContent = append(multiContent, openai.ChatMessagePart{
							Type: openai.ChatMessagePartTypeText,
							Text: c.Text,
						})
					case *v1.Content_Image:
						multiContent = append(multiContent, openai.ChatMessagePart{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: c.Image.GetUrl(),
							},
						})
					case *v1.Content_ToolCall:
						t := openai.ToolCall{
							ID:   c.ToolCall.Id,
							Type: openai.ToolTypeFunction,
						}
						if f := c.ToolCall.GetFunction(); f != nil {
							t.Function = openai.FunctionCall{
								Name:      f.Name,
								Arguments: f.Arguments,
							}
						}
						toolCalls = append(toolCalls, t)
					}
				}
				message.MultiContent = multiContent
				message.ToolCalls = toolCalls
			}
		}

		message.ToolCallID = resp.Message.ToolCallId
		openAIResp.Choices = []openai.ChatCompletionChoice{
			{
				Message: message,
			},
		}
	}

	if resp.Statistics != nil {
		openAIResp.Usage.PromptTokens = int(resp.Statistics.Usage.PromptTokens)
		openAIResp.Usage.CompletionTokens = int(resp.Statistics.Usage.CompletionTokens)
	}

	return openAIResp
}

// convertEmbeddingReqFromOpenAI converts an embedding request from OpenAI API to Router API
func convertEmbeddingReqFromOpenAI(req *openai.EmbeddingRequest) *v1.EmbedReq {
	var contents []*v1.Content

	switch input := req.Input.(type) {
	case string:
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: input,
			},
		})
	case []string:
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: input[0],
			},
		})
	}

	return &v1.EmbedReq{
		Model:    string(req.Model),
		Contents: contents,
	}
}

// convertEmbeddingRespToOpenAI converts an embedding response from Router API to OpenAI API
func convertEmbeddingRespToOpenAI(resp *v1.EmbedResp) *openai.EmbeddingResponse {
	return &openai.EmbeddingResponse{
		Object: "list",
		Data: []openai.Embedding{
			{
				Index:     0,
				Object:    "embedding",
				Embedding: resp.Embedding,
			},
		},
	}
}
