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
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/google/uuid"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

// convertSystemToAnthropic converts system messages to a format that can be sent to the Anthropic API.
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
				r.log.Errorf("unsupported content: %v", c)
			}
		}
	}
	return parts
}

// convertMessageToAnthropic converts an internal message to a message that can be sent to the Anthropic API.
func (r *upstream) convertMessageToAnthropic(message *v1.Message) anthropic.MessageParam {
	var parts []anthropic.ContentBlockParamUnion
	for _, content := range message.Contents {
		switch c := content.GetContent().(type) {
		case *v1.Content_Text:
			parts = append(parts, anthropic.NewTextBlock(c.Text))
		case *v1.Content_Image:
			parts = append(parts, anthropic.ContentBlockParamUnion{
				OfImage: &anthropic.ImageBlockParam{
					Source: anthropic.ImageBlockParamSourceUnion{
						OfURL: &anthropic.URLImageSourceParam{
							URL: c.Image.GetUrl(),
						},
					},
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

// convertRequestToAnthropic converts an internal request to a request that can be sent to the Anthropic API.
func (r *upstream) convertRequestToAnthropic(req *entity.ChatReq) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model: anthropic.Model(req.Model),
	}

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
				tools = append(tools, anthropic.ToolUnionParam{
					OfTool: &anthropic.ToolParam{
						Name:        t.Function.Name,
						Description: anthropic.Opt(t.Function.Description),
						InputSchema: r.convertInputSchemaToAnthropic(t.Function.Parameters),
					},
				})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		params.Tools = tools
	}

	return params
}

// convertContentsFromAnthropic converts an Anthropic message contents to an internal message.
func convertContentsFromAnthropic(contents []anthropic.ContentBlockUnion) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	for _, content := range contents {
		if content.Thinking != "" {
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_Thinking{
					Thinking: content.Thinking,
				},
			})
		}
		if content.Text != "" {
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: content.Text,
				},
			})
		}
	}

	return message
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
		if chunk.ContentBlock.Type == "tool_use" {
			return &entity.ChatResp{
				Id:    c.req.Id,
				Model: c.model,
				Message: &v1.Message{
					Id:   c.messageID,
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{
							Content: &v1.Content_FunctionCall{
								FunctionCall: &v1.FunctionCall{
									Id:   chunk.ContentBlock.ToolUseID,
									Name: chunk.ContentBlock.Name,
								},
							},
						},
					},
				},
			}
		}
		return nil
	case "content_block_delta":
		resp := &entity.ChatResp{
			Id:    c.req.Id,
			Model: c.model,
			Message: &v1.Message{
				Id:   c.messageID,
				Role: v1.Role_MODEL,
			},
		}
		switch chunk.Delta.Type {
		case "thinking_delta":
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Content: &v1.Content_Thinking{
					Thinking: chunk.Delta.Thinking,
				},
			})
		case "text_delta":
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: chunk.Delta.Text,
				},
			})
		case "input_json_delta":
			resp.Message.Contents = append(resp.Message.Contents, &v1.Content{
				Content: &v1.Content_FunctionCall{
					FunctionCall: &v1.FunctionCall{
						Arguments: chunk.Delta.PartialJSON,
					},
				},
			})
		default:
			return nil
		}
		return resp
	case "message_delta":
		if chunk.Usage.OutputTokens != 0 {
			return &entity.ChatResp{
				Id:    c.req.Id,
				Model: c.model,
				Statistics: &v1.Statistics{
					Usage: &v1.Statistics_Usage{
						PromptTokens:     c.inputTokens,
						CompletionTokens: uint32(chunk.Usage.OutputTokens),
					},
				},
			}
		}
		fallthrough
	default:
		return nil
	}
}
