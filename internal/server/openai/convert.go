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
	"github.com/neuraxes/neurouter/internal/util"
)

func convertImageFromOpenAIURL(u string) *v1.Image {
	if data, mimeType := util.DecodeImageDataFromUrl(u); data != nil {
		return &v1.Image{
			MimeType: mimeType,
			Source:   &v1.Image_Data{Data: data},
		}
	}
	return &v1.Image{
		Source: &v1.Image_Url{Url: u},
	}
}

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
		role = v1.Role_USER
	}

	var contents []*v1.Content

	if message.Role == openai.ChatMessageRoleTool {
		tr := &v1.Content_ToolResult{
			ToolResult: &v1.ToolResult{
				Id:      message.ToolCallID,
				Outputs: []*v1.ToolResult_Output{},
			},
		}
		if message.Content != "" {
			// Single text message
			tr.ToolResult.Outputs = append(tr.ToolResult.Outputs, &v1.ToolResult_Output{
				Output: &v1.ToolResult_Output_Text{
					Text: message.Content,
				},
			})
		} else {
			// Multipart message
			for _, content := range message.MultiContent {
				switch content.Type {
				case openai.ChatMessagePartTypeText:
					tr.ToolResult.Outputs = append(tr.ToolResult.Outputs, &v1.ToolResult_Output{
						Output: &v1.ToolResult_Output_Text{
							Text: content.Text,
						},
					})
				case openai.ChatMessagePartTypeImageURL:
					tr.ToolResult.Outputs = append(tr.ToolResult.Outputs, &v1.ToolResult_Output{
						Output: &v1.ToolResult_Output_Image{
							Image: convertImageFromOpenAIURL(content.ImageURL.URL),
						},
					})
				}
			}
		}
		contents = append(contents, &v1.Content{Content: tr})
	} else {
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
							Image: convertImageFromOpenAIURL(content.ImageURL.URL),
						},
					})
				}
			}
		}

		for _, toolCall := range message.ToolCalls {
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

	return &v1.Message{
		Role:     role,
		Name:     message.Name,
		Contents: contents,
	}
}

// convertChatReqFromOpenAI converts a chat completion request from OpenAI API to Router API
func convertChatReqFromOpenAI(req *openai.ChatCompletionRequest) *v1.ChatReq {
	config := &v1.GenerationConfig{}

	if req.MaxCompletionTokens != 0 {
		config.MaxTokens = new(int64(req.MaxCompletionTokens))
	} else if req.MaxTokens != 0 {
		config.MaxTokens = new(int64(req.MaxTokens))
	}
	if req.Temperature != 0 {
		config.Temperature = &req.Temperature
	}
	if req.TopP != 0 {
		config.TopP = &req.TopP
	}
	if req.FrequencyPenalty != 0 {
		config.FrequencyPenalty = &req.FrequencyPenalty
	}
	if req.PresencePenalty != 0 {
		config.PresencePenalty = &req.PresencePenalty
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
			var parameters *v1.Schema
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
			// Only function tool is supported
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

// convertStatusToOpenAI maps internal chat status to OpenAI finish reason.
func convertStatusToOpenAI(status v1.ChatStatus) openai.FinishReason {
	switch status {
	case v1.ChatStatus_CHAT_COMPLETED:
		return openai.FinishReasonStop
	case v1.ChatStatus_CHAT_REFUSED:
		return openai.FinishReasonContentFilter
	case v1.ChatStatus_CHAT_PENDING_TOOL_USE:
		return openai.FinishReasonToolCalls
	case v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT:
		return openai.FinishReasonLength
	default:
		return ""
	}
}

// convertChatRespToOpenAI converts a chat completion response from Router API to OpenAI API
func convertChatRespToOpenAI(resp *v1.ChatResp) *openai.ChatCompletionResponse {
	openAIResp := &openai.ChatCompletionResponse{
		ID:     resp.Id,
		Object: "chat.completion",
		Model:  resp.Model,
	}

	if resp.Message != nil {
		message := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant}

		for _, content := range resp.Message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					message.ReasoningContent = c.Text
				} else {
					message.Content += c.Text
				}
			case *v1.Content_ToolUse:
				message.ToolCalls = append(message.ToolCalls, openai.ToolCall{
					ID:   c.ToolUse.Id,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      c.ToolUse.Name,
						Arguments: c.ToolUse.GetTextualInput(),
					},
				})
			}
		}

		openAIResp.Choices = []openai.ChatCompletionChoice{
			{
				Message:      message,
				FinishReason: convertStatusToOpenAI(resp.Status),
			},
		}
	}

	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		openAIResp.Usage.PromptTokens = int(resp.Statistics.Usage.InputTokens)
		openAIResp.Usage.CompletionTokens = int(resp.Statistics.Usage.OutputTokens)
		openAIResp.Usage.TotalTokens = openAIResp.Usage.PromptTokens + openAIResp.Usage.CompletionTokens
		openAIResp.Usage.PromptTokensDetails = &openai.PromptTokensDetails{
			CachedTokens: int(resp.Statistics.Usage.CachedInputTokens),
		}
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
		Model:  openai.EmbeddingModel(resp.Model),
		Data: []openai.Embedding{
			{
				Index:     0,
				Object:    "embedding",
				Embedding: resp.Embedding,
			},
		},
	}
}
