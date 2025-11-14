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

package anthropic

import (
	"encoding/json"
	"math"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"k8s.io/utils/ptr"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

func (r *upstream) convertGenerationConfigToAnthropic(config *v1.GenerationConfig, req *anthropic.MessageNewParams) {
	if config == nil {
		return
	}
	if config.MaxTokens != nil {
		req.MaxTokens = *config.MaxTokens
	}
	if config.Temperature != nil {
		req.Temperature = anthropic.Opt(math.Round(float64(*config.Temperature)*100) / 100)
	}
	if config.TopP != nil {
		req.TopP = anthropic.Opt(math.Round(float64(*config.TopP)*100) / 100)
	}
	if config.TopK != nil {
		req.TopK = anthropic.Opt(*config.TopK)
	}
}

// convertSystemToAnthropic converts system messages to Anthropic format.
func (r *upstream) convertSystemToAnthropic(messages []*v1.Message) []anthropic.TextBlockParam {
	var parts []anthropic.TextBlockParam
	for _, message := range messages {
		if message.Role != v1.Role_SYSTEM {
			continue
		}
		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				parts = append(parts, anthropic.TextBlockParam{Text: c.Text})
			default:
				r.log.Errorf("unsupport content: %v", c)
			}
		}
	}
	return parts
}

// convertMessageToAnthropic converts a newurouter message to Anthropic format.
func (r *upstream) convertMessageToAnthropic(message *v1.Message) anthropic.MessageParam {
	var parts []anthropic.ContentBlockParamUnion
	for _, content := range message.Contents {
		switch c := content.GetContent().(type) {
		case *v1.Content_Text:
			if content.Reasoning {
				signature := content.Metadata["signature"]
				parts = append(parts, anthropic.NewThinkingBlock(signature, c.Text))
			} else {
				parts = append(parts, anthropic.NewTextBlock(c.Text))
			}
		case *v1.Content_Image:
			parts = append(parts, anthropic.NewImageBlock(
				anthropic.URLImageSourceParam{
					URL: c.Image.GetUrl(),
				},
			))
		case *v1.Content_ToolUse:
			textualInput := c.ToolUse.GetTextualInput()

			var input any
			err := json.Unmarshal([]byte(textualInput), &input)
			if err != nil {
				// Fallback to string
				input = textualInput
				continue
			}

			parts = append(parts, anthropic.NewToolUseBlock(
				c.ToolUse.Id,
				input,
				c.ToolUse.Name,
			))
		case *v1.Content_ToolResult:
			var outputs []anthropic.ToolResultBlockParamContentUnion

			for _, output := range c.ToolResult.Outputs {
				switch o := output.Output.(type) {
				case *v1.ToolResult_Output_Text:
					outputs = append(outputs, anthropic.ToolResultBlockParamContentUnion{
						OfText: &anthropic.TextBlockParam{
							Text: o.Text,
						},
					})
				}
			}

			parts = append(parts, anthropic.ContentBlockParamUnion{
				OfToolResult: &anthropic.ToolResultBlockParam{
					ToolUseID: c.ToolResult.Id,
					Content:   outputs,
				},
			})
		}
	}
	if message.Role == v1.Role_USER || message.Role == v1.Role_SYSTEM {
		return anthropic.NewUserMessage(parts...)
	} else {
		return anthropic.NewAssistantMessage(parts...)
	}
}

func (r *upstream) convertInputSchemaToAnthropic(params *v1.Schema) (schema anthropic.ToolInputSchemaParam) {
	if params == nil {
		return
	}
	schema.Properties = params.Properties
	schema.Required = params.Required
	return
}

// convertRequestToAnthropic converts a newurouter request to Anthropic format.
func (r *upstream) convertRequestToAnthropic(req *entity.ChatReq) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model: anthropic.Model(req.Model),
	}

	r.convertGenerationConfigToAnthropic(req.Config, &params)

	if !r.config.SystemAsUser {
		params.System = r.convertSystemToAnthropic(req.Messages)
	}

	for _, message := range req.Messages {
		if !r.config.SystemAsUser && message.Role == v1.Role_SYSTEM {
			continue
		}
		params.Messages = append(params.Messages, r.convertMessageToAnthropic(message))
	}

	if req.Tools != nil {
		var tools []anthropic.ToolUnionParam
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				at := &anthropic.ToolParam{
					Name:        t.Function.Name,
					InputSchema: r.convertInputSchemaToAnthropic(t.Function.Parameters),
				}
				if t.Function.Description != "" {
					at.Description = anthropic.Opt(t.Function.Description)
				}
				tools = append(tools, anthropic.ToolUnionParam{OfTool: at})
			default:
				r.log.Errorf("unsupport tool: %v", t)
			}
		}
		params.Tools = tools
	}

	return params
}

// convertContentsFromAnthropic converts Anthropic contents to a newurouter message.
func convertContentsFromAnthropic(contents []anthropic.ContentBlockUnion) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	for _, content := range contents {
		switch content.Type {
		case "thinking":
			message.Contents = append(message.Contents, &v1.Content{
				Metadata: map[string]string{
					"signature": content.Signature,
				},
				Reasoning: true,
				Content: &v1.Content_Text{
					Text: content.Thinking,
				},
			})
		case "text":
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: content.Text,
				},
			})
		case "tool_use":
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   content.ID,
						Name: content.Name,
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: gjson.Parse(string(content.Input)).String(),
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

func (c *anthropicChatStreamClient) newResp() *entity.ChatResp {
	return &entity.ChatResp{
		Id:    c.req.Id,
		Model: c.model,
		Message: &v1.Message{
			Id:   c.messageID,
			Role: v1.Role_MODEL,
		},
	}
}

// convertChunkFromAnthropic converts an Anthropic streaming chunk to an internal response.
func (c *anthropicChatStreamClient) convertChunkFromAnthropic(chunk *anthropic.MessageStreamEventUnion) *entity.ChatResp {
	switch chunk.Type {
	case "message_start":
		c.messageID = chunk.Message.ID
		c.model = string(chunk.Message.Model)
		c.inputTokens = uint32(chunk.Message.Usage.InputTokens)
		return nil
	case "content_block_start":
		var resp *entity.ChatResp
		switch chunk.ContentBlock.Type {
		case "text":
			if chunk.ContentBlock.Text != "" {
				resp = c.newResp()
				resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
					Index: ptr.To(uint32(chunk.Index)),
					Content: &v1.Content_Text{
						Text: chunk.ContentBlock.Text,
					},
				})
			}
		case "thinking":
			if chunk.ContentBlock.Thinking != "" {
				resp = c.newResp()
				resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
					Index: ptr.To(uint32(chunk.Index)),
					Metadata: map[string]string{
						"signature": chunk.ContentBlock.Signature,
					},
					Reasoning: true,
					Content: &v1.Content_Text{
						Text: chunk.ContentBlock.Thinking,
					},
				})
			}
		case "tool_use":
			resp = c.newResp()
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Index: ptr.To(uint32(chunk.Index)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   chunk.ContentBlock.ID,
						Name: chunk.ContentBlock.Name,
						// Inputs will be sent in deltas.
					},
				},
			})
		}
		return resp
	case "content_block_delta":
		resp := c.newResp()
		switch chunk.Delta.Type {
		case "thinking_delta":
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Index: ptr.To(uint32(chunk.Index)),
				Metadata: map[string]string{
					"signature": chunk.Delta.Signature,
				},
				Reasoning: true,
				Content: &v1.Content_Text{
					Text: chunk.Delta.Thinking,
				},
			})
		case "text_delta":
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Index: ptr.To(uint32(chunk.Index)),
				Content: &v1.Content_Text{
					Text: chunk.Delta.Text,
				},
			})
		case "input_json_delta":
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Index: ptr.To(uint32(chunk.Index)),
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: chunk.Delta.PartialJSON,
								},
							},
						},
					},
				},
			})
		default:
			return nil
		}
		return resp
	case "message_delta":
		if chunk.Usage.InputTokens != 0 || chunk.Usage.OutputTokens != 0 {
			resp := c.newResp()
			resp.Message = nil
			resp.Statistics = &v1.Statistics{
				Usage: &v1.Statistics_Usage{
					InputTokens:       c.inputTokens + uint32(chunk.Usage.InputTokens),
					OutputTokens:      uint32(chunk.Usage.OutputTokens),
					CachedInputTokens: uint32(chunk.Usage.CacheReadInputTokens),
				},
			}
			return resp
		}
		fallthrough
	default:
		return nil
	}
}

func convertStatisticsFromAnthropic(usage *anthropic.Usage) *v1.Statistics {
	if usage == nil {
		return nil
	}

	if usage.InputTokens == 0 && usage.OutputTokens == 0 {
		return nil
	}

	return &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:       uint32(usage.InputTokens),
			OutputTokens:      uint32(usage.OutputTokens),
			CachedInputTokens: uint32(usage.CacheReadInputTokens),
		},
	}
}
