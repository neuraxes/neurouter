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

package anthropic

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func convertMessageFromAnthropic(message *anthropic.MessageParam) *v1.Message {
	var role v1.Role
	switch message.Role {
	case anthropic.MessageParamRoleUser:
		role = v1.Role_USER
	case anthropic.MessageParamRoleAssistant:
		role = v1.Role_MODEL
	default:
		role = v1.Role_USER
	}

	var contents []*v1.Content

	// Handle content from Anthropic message
	for _, content := range message.Content {
		switch {
		case content.OfText != nil:
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: content.OfText.Text,
				},
			})
		case content.OfImage != nil:
			var url string
			if content.OfImage.Source.OfURL != nil {
				url = content.OfImage.Source.OfURL.URL
			}
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Image{
					Image: &v1.Image{
						Source: &v1.Image_Url{
							Url: url,
						},
					},
				},
			})
		case content.OfToolUse != nil:
			var args []byte
			if content.OfToolUse.Input != nil {
				args, _ = json.Marshal(content.OfToolUse.Input)
			}
			contents = append(contents, &v1.Content{
				Content: &v1.Content_ToolCall{
					ToolCall: &v1.ToolCall{
						Id: content.OfToolUse.ID,
						Tool: &v1.ToolCall_Function{
							Function: &v1.ToolCall_FunctionCall{
								Name:      content.OfToolUse.Name,
								Arguments: string(args),
							},
						},
					},
				},
			})
		case content.OfToolResult != nil:
			// Handle tool results
			for _, resultContent := range content.OfToolResult.Content {
				if resultContent.OfText != nil {
					contents = append(contents, &v1.Content{
						Content: &v1.Content_Text{
							Text: resultContent.OfText.Text,
						},
					})
				}
			}
		}
	}

	return &v1.Message{
		Role:     role,
		Contents: contents,
	}
}

// convertChatReqFromAnthropic converts a message request from Anthropic API to Router API
func convertChatReqFromAnthropic(req *anthropic.MessageNewParams) *v1.ChatReq {
	config := &v1.GenerationConfig{
		MaxTokens: req.MaxTokens,
	}

	if req.Temperature.Valid() {
		config.Temperature = float32(req.Temperature.Value)
	}
	if req.TopP.Valid() {
		config.TopP = float32(req.TopP.Value)
	}

	var messages []*v1.Message
	for _, message := range req.Messages {
		messages = append(messages, convertMessageFromAnthropic(&message))
	}

	// Handle system prompts
	if len(req.System) > 0 {
		var systemContent []*v1.Content
		for _, sysBlock := range req.System {
			systemContent = append(systemContent, &v1.Content{
				Content: &v1.Content_Text{
					Text: sysBlock.Text,
				},
			})
		}
		systemMessage := &v1.Message{
			Role:     v1.Role_SYSTEM,
			Contents: systemContent,
		}
		messages = append([]*v1.Message{systemMessage}, messages...)
	}

	var tools []*v1.Tool
	for _, tool := range req.Tools {
		t := &v1.Tool{}
		switch {
		case tool.OfTool != nil:
			// Client tool (function)
			var parameters *v1.Tool_Function_Parameters
			j, _ := json.Marshal(tool.OfTool.InputSchema)
			_ = json.Unmarshal(j, &parameters)

			var description string
			if tool.OfTool.Description.Valid() {
				description = tool.OfTool.Description.Value
			}

			t.Tool = &v1.Tool_Function_{
				Function: &v1.Tool_Function{
					Name:        tool.OfTool.Name,
					Description: description,
					Parameters:  parameters,
				},
			}
		default:
			// Skip unsupported tool types
			continue
		}
		tools = append(tools, t)
	}

	return &v1.ChatReq{
		Model:    string(req.Model),
		Config:   config,
		Messages: messages,
		Tools:    tools,
	}
}

// convertChatRespToAnthropic converts a chat completion response from Router API to Anthropic API
func convertChatRespToAnthropic(resp *v1.ChatResp) *anthropic.Message {
	anthropicResp := &anthropic.Message{
		ID:   resp.Message.Id,
		Role: "assistant",
		Type: "message",
	}

	// Convert content blocks
	var content []anthropic.ContentBlockUnion
	if resp.Message != nil && len(resp.Message.Contents) > 0 {
		for _, c := range resp.Message.Contents {
			switch cont := c.Content.(type) {
			case *v1.Content_Text:
				if cont.Text != "" {
					content = append(content, anthropic.ContentBlockUnion{
						Type: "text",
						Text: cont.Text,
					})
				}
			case *v1.Content_ToolCall:
				if cont.ToolCall != nil {
					var input json.RawMessage
					if func_call := cont.ToolCall.GetFunction(); func_call != nil {
						inputData, _ := json.Marshal(func_call.Arguments)
						input = inputData
						content = append(content, anthropic.ContentBlockUnion{
							Type:  "tool_use",
							ID:    cont.ToolCall.Id,
							Name:  func_call.Name,
							Input: input,
						})
					}
				}
			}
		}
	}
	anthropicResp.Content = content

	// Add usage statistics
	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		anthropicResp.Usage = anthropic.Usage{
			InputTokens:  int64(resp.Statistics.Usage.PromptTokens),
			OutputTokens: int64(resp.Statistics.Usage.CompletionTokens),
		}
	}

	return anthropicResp
}
