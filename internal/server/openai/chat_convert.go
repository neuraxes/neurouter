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

	"github.com/openai/openai-go/v3"

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

func convertDeveloperMessageFromOpenAIChat(m *openai.ChatCompletionDeveloperMessageParam) *v1.Message {
	var contents []*v1.Content

	if m.Content.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: m.Content.OfString.Value},
		})
	} else {
		for _, part := range m.Content.OfArrayOfContentParts {
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Text{Text: part.Text},
			})
		}
	}

	msg := &v1.Message{
		Role:     v1.Role_SYSTEM,
		Contents: contents,
	}
	if m.Name.Valid() {
		msg.Name = m.Name.Value
	}
	return msg
}

func convertSystemMessageFromOpenAIChat(m *openai.ChatCompletionSystemMessageParam) *v1.Message {
	var contents []*v1.Content

	if m.Content.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: m.Content.OfString.Value},
		})
	} else {
		for _, part := range m.Content.OfArrayOfContentParts {
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Text{Text: part.Text},
			})
		}
	}

	msg := &v1.Message{
		Role:     v1.Role_SYSTEM,
		Contents: contents,
	}
	if m.Name.Valid() {
		msg.Name = m.Name.Value
	}
	return msg
}

func convertUserMessageFromOpenAIChat(m *openai.ChatCompletionUserMessageParam) *v1.Message {
	var contents []*v1.Content

	if m.Content.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: m.Content.OfString.Value},
		})
	} else {
		for _, part := range m.Content.OfArrayOfContentParts {
			if part.OfText != nil {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Text{Text: part.OfText.Text},
				})
			} else if part.OfImageURL != nil {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Image{
						Image: convertImageFromOpenAIURL(part.OfImageURL.ImageURL.URL),
					},
				})
			}
		}
	}

	msg := &v1.Message{
		Role:     v1.Role_USER,
		Contents: contents,
	}
	if m.Name.Valid() {
		msg.Name = m.Name.Value
	}
	return msg
}

func convertAssistantMessageFromOpenAIChat(m *openai.ChatCompletionAssistantMessageParam) *v1.Message {
	var contents []*v1.Content

	if m.Content.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: m.Content.OfString.Value},
		})
	} else {
		for _, part := range m.Content.OfArrayOfContentParts {
			if part.OfText != nil {
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Text{Text: part.OfText.Text},
				})
			}
		}
	}

	for _, tc := range m.ToolCalls {
		if tc.OfFunction != nil {
			contents = append(contents, &v1.Content{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   tc.OfFunction.ID,
						Name: tc.OfFunction.Function.Name,
						Inputs: []*v1.ToolUse_Input{
							{Input: &v1.ToolUse_Input_Text{Text: tc.OfFunction.Function.Arguments}},
						},
					},
				},
			})
		}
	}

	msg := &v1.Message{
		Role:     v1.Role_MODEL,
		Contents: contents,
	}
	if m.Name.Valid() {
		msg.Name = m.Name.Value
	}
	return msg
}

func convertToolMessageFromOpenAIChat(m *openai.ChatCompletionToolMessageParam) *v1.Message {
	tr := &v1.Content_ToolResult{
		ToolResult: &v1.ToolResult{
			Id:      m.ToolCallID,
			Outputs: []*v1.ToolResult_Output{},
		},
	}

	if m.Content.OfString.Valid() {
		tr.ToolResult.Outputs = append(tr.ToolResult.Outputs, &v1.ToolResult_Output{
			Output: &v1.ToolResult_Output_Text{Text: m.Content.OfString.Value},
		})
	} else {
		for _, part := range m.Content.OfArrayOfContentParts {
			tr.ToolResult.Outputs = append(tr.ToolResult.Outputs, &v1.ToolResult_Output{
				Output: &v1.ToolResult_Output_Text{Text: part.Text},
			})
		}
	}

	return &v1.Message{
		Role:     v1.Role_USER,
		Contents: []*v1.Content{{Content: tr}},
	}
}

func convertChatMessageFromOpenAIChat(msg openai.ChatCompletionMessageParamUnion) *v1.Message {
	if msg.OfDeveloper != nil {
		return convertDeveloperMessageFromOpenAIChat(msg.OfDeveloper)
	}
	if msg.OfSystem != nil {
		return convertSystemMessageFromOpenAIChat(msg.OfSystem)
	}
	if msg.OfUser != nil {
		return convertUserMessageFromOpenAIChat(msg.OfUser)
	}
	if msg.OfAssistant != nil {
		return convertAssistantMessageFromOpenAIChat(msg.OfAssistant)
	}
	if msg.OfTool != nil {
		return convertToolMessageFromOpenAIChat(msg.OfTool)
	}
	return nil
}

func convertChatReqFromOpenAIChat(req *openai.ChatCompletionNewParams) *v1.ChatReq {
	config := &v1.GenerationConfig{}

	if req.MaxCompletionTokens.Valid() {
		config.MaxTokens = new(req.MaxCompletionTokens.Value)
	} else if req.MaxTokens.Valid() {
		config.MaxTokens = new(req.MaxTokens.Value)
	}
	if req.Temperature.Valid() {
		config.Temperature = new(float32(req.Temperature.Value))
	}
	if req.TopP.Valid() {
		config.TopP = new(float32(req.TopP.Value))
	}
	if req.FrequencyPenalty.Valid() {
		config.FrequencyPenalty = new(float32(req.FrequencyPenalty.Value))
	}
	if req.PresencePenalty.Valid() {
		config.PresencePenalty = new(float32(req.PresencePenalty.Value))
	}

	if req.ResponseFormat.OfJSONObject != nil {
		config.Grammar = &v1.GenerationConfig_PresetGrammar{
			PresetGrammar: "json_object",
		}
	}

	var messages []*v1.Message
	for _, message := range req.Messages {
		if m := convertChatMessageFromOpenAIChat(message); m != nil {
			messages = append(messages, m)
		}
	}

	var tools []*v1.Tool
	for _, tool := range req.Tools {
		if tool.OfFunction != nil {
			fn := tool.OfFunction.Function
			var parameters *v1.Schema
			j, _ := json.Marshal(fn.Parameters)
			_ = json.Unmarshal(j, &parameters)
			tools = append(tools, &v1.Tool{
				Tool: &v1.Tool_Function_{
					Function: &v1.Tool_Function{
						Name:        fn.Name,
						Description: fn.Description.Value,
						Parameters:  parameters,
					},
				},
			})
		}
	}

	return &v1.ChatReq{
		Model:    string(req.Model),
		Config:   config,
		Messages: messages,
		Tools:    tools,
	}
}

func convertStatusToOpenAIChat(status v1.ChatStatus) string {
	switch status {
	case v1.ChatStatus_CHAT_COMPLETED:
		return "stop"
	case v1.ChatStatus_CHAT_REFUSED:
		return "content_filter"
	case v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT:
		return "length"
	case v1.ChatStatus_CHAT_PENDING_TOOL_USE:
		return "tool_calls"
	default:
		return ""
	}
}

func convertUsageToOpenAIChat(u *v1.Statistics_Usage) *openai.CompletionUsage {
	if u == nil {
		return nil
	}
	return &openai.CompletionUsage{
		PromptTokens:     int64(u.InputTokens),
		CompletionTokens: int64(u.OutputTokens),
		TotalTokens:      int64(u.InputTokens) + int64(u.OutputTokens),
		PromptTokensDetails: openai.CompletionUsagePromptTokensDetails{
			CachedTokens: int64(u.CachedInputTokens),
		},
		CompletionTokensDetails: openai.CompletionUsageCompletionTokensDetails{
			ReasoningTokens: int64(u.ReasoningTokens),
		},
	}
}

func convertChatRespToOpenAIChat(resp *v1.ChatResp) *chatCompletionResponse {
	r := &chatCompletionResponse{
		ID:     resp.Id,
		Object: "chat.completion",
		Model:  resp.Model,
	}

	if resp.Message != nil {
		message := chatCompletionMessage{Role: "assistant"}

		for _, content := range resp.Message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					message.ReasoningContent = c.Text
				} else {
					message.Content += c.Text
				}
			case *v1.Content_ToolUse:
				message.ToolCalls = append(message.ToolCalls, toolCall{
					ID:   c.ToolUse.Id,
					Type: "function",
					Function: functionCall{
						Name:      c.ToolUse.Name,
						Arguments: c.ToolUse.GetTextualInput(),
					},
				})
			}
		}

		r.Choices = []chatCompletionChoice{
			{
				Message:      message,
				FinishReason: convertStatusToOpenAIChat(resp.Status),
			},
		}
	}

	if resp.Statistics != nil {
		r.Usage = convertUsageToOpenAIChat(resp.Statistics.Usage)
	}

	return r
}

func convertEmbeddingReqFromOpenAIChat(req *openai.EmbeddingNewParams) *v1.EmbedReq {
	var contents []*v1.Content

	if req.Input.OfString.Valid() {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: req.Input.OfString.Value},
		})
	} else if len(req.Input.OfArrayOfStrings) > 0 {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{Text: req.Input.OfArrayOfStrings[0]},
		})
	}

	return &v1.EmbedReq{
		Model:    string(req.Model),
		Contents: contents,
	}
}

func convertEmbeddingRespToOpenAIChat(resp *v1.EmbedResp) *embeddingResponse {
	embedding := make([]float64, len(resp.Embedding))
	for i, v := range resp.Embedding {
		embedding[i] = float64(v)
	}
	return &embeddingResponse{
		Object: "list",
		Model:  resp.Model,
		Data: []openai.Embedding{
			{
				Index:     0,
				Embedding: embedding,
			},
		},
	}
}
