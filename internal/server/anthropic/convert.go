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
	"encoding/base64"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func convertEffortFromAnthropic(effort anthropic.OutputConfigEffort) v1.ReasoningEffort {
	switch effort {
	case anthropic.OutputConfigEffortLow:
		return v1.ReasoningEffort_REASONING_EFFORT_LOW
	case anthropic.OutputConfigEffortMedium:
		return v1.ReasoningEffort_REASONING_EFFORT_MEDIUM
	case anthropic.OutputConfigEffortHigh:
		return v1.ReasoningEffort_REASONING_EFFORT_HIGH
	case anthropic.OutputConfigEffortMax:
		return v1.ReasoningEffort_REASONING_EFFORT_MAX
	default:
		return v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED
	}
}

func convertGenerationConfigFromAnthropic(req *anthropic.MessageNewParams) *v1.GenerationConfig {
	config := &v1.GenerationConfig{}
	if req.MaxTokens != 0 {
		config.MaxTokens = &req.MaxTokens
	}
	if req.Temperature.Valid() {
		config.Temperature = new(float32(req.Temperature.Value))
	}
	if req.TopP.Valid() {
		config.TopP = new(float32(req.TopP.Value))
	}
	if req.TopK.Valid() {
		config.TopK = &req.TopK.Value
	}
	if req.Thinking.OfEnabled != nil {
		config.ReasoningConfig = &v1.ReasoningConfig{
			TokenBudget: uint32(req.Thinking.OfEnabled.BudgetTokens),
		}
	} else if req.Thinking.OfDisabled != nil {
		config.ReasoningConfig = &v1.ReasoningConfig{
			Effort: v1.ReasoningEffort_REASONING_EFFORT_NONE,
		}
	} else if req.Thinking.OfAdaptive != nil || req.OutputConfig.Effort != "" {
		config.ReasoningConfig = &v1.ReasoningConfig{
			Effort: convertEffortFromAnthropic(req.OutputConfig.Effort),
		}
	}
	if len(req.OutputConfig.Format.Schema) > 0 {
		jsonSchema, err := json.Marshal(req.OutputConfig.Format.Schema)
		if err == nil {
			config.Grammar = &v1.GenerationConfig_JsonSchema{
				JsonSchema: string(jsonSchema),
			}
		}
	}
	return config
}

func convertSystemFromAnthropic(system []anthropic.TextBlockParam) *v1.Message {
	if len(system) == 0 {
		return nil
	}
	var contents []*v1.Content
	for _, block := range system {
		contents = append(contents, &v1.Content{
			Content: &v1.Content_Text{
				Text: block.Text,
			},
		})
	}
	return &v1.Message{
		Role:     v1.Role_SYSTEM,
		Contents: contents,
	}
}

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
	for _, content := range message.Content {
		switch {
		case content.OfText != nil:
			contents = append(contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: content.OfText.Text,
				},
			})
		case content.OfImage != nil:
			switch {
			case content.OfImage.Source.OfURL != nil:
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Image{
						Image: &v1.Image{
							Source: &v1.Image_Url{
								Url: content.OfImage.Source.OfURL.URL,
							},
						},
					},
				})
			case content.OfImage.Source.OfBase64 != nil:
				data, err := base64.StdEncoding.DecodeString(content.OfImage.Source.OfBase64.Data)
				if err != nil {
					continue
				}
				contents = append(contents, &v1.Content{
					Content: &v1.Content_Image{
						Image: &v1.Image{
							MimeType: string(content.OfImage.Source.OfBase64.MediaType),
							Source: &v1.Image_Data{
								Data: data,
							},
						},
					},
				})
			}
		case content.OfThinking != nil:
			contents = append(contents, &v1.Content{
				Reasoning: true,
				Metadata: map[string]string{
					"signature": content.OfThinking.Signature,
				},
				Content: &v1.Content_Text{
					Text: content.OfThinking.Thinking,
				},
			})
		case content.OfRedactedThinking != nil:
			contents = append(contents, &v1.Content{
				Reasoning: true,
				Metadata: map[string]string{
					"redacted_thinking": content.OfRedactedThinking.Data,
				},
				Content: &v1.Content_Text{Text: ""},
			})
		case content.OfToolUse != nil:
			var args []byte
			if content.OfToolUse.Input != nil {
				args, _ = json.Marshal(content.OfToolUse.Input)
			}
			contents = append(contents, &v1.Content{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   content.OfToolUse.ID,
						Name: content.OfToolUse.Name,
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: string(args),
								},
							},
						},
					},
				},
			})
		case content.OfToolResult != nil:
			tr := &v1.ToolResult{
				Id: content.OfToolResult.ToolUseID,
			}
			for _, output := range content.OfToolResult.Content {
				switch {
				case output.OfText != nil:
					tr.Outputs = append(tr.Outputs, &v1.ToolResult_Output{
						Output: &v1.ToolResult_Output_Text{
							Text: output.OfText.Text,
						},
					})
				case output.OfImage != nil:
					switch {
					case output.OfImage.Source.OfURL != nil:
						tr.Outputs = append(tr.Outputs, &v1.ToolResult_Output{
							Output: &v1.ToolResult_Output_Image{
								Image: &v1.Image{
									Source: &v1.Image_Url{
										Url: output.OfImage.Source.OfURL.URL,
									},
								},
							},
						})
					case output.OfImage.Source.OfBase64 != nil:
						data, err := base64.StdEncoding.DecodeString(output.OfImage.Source.OfBase64.Data)
						if err != nil {
							continue
						}
						tr.Outputs = append(tr.Outputs, &v1.ToolResult_Output{
							Output: &v1.ToolResult_Output_Image{
								Image: &v1.Image{
									MimeType: string(output.OfImage.Source.OfBase64.MediaType),
									Source: &v1.Image_Data{
										Data: data,
									},
								},
							},
						})
					}
				}
			}
			contents = append(contents, &v1.Content{
				Content: &v1.Content_ToolResult{
					ToolResult: tr,
				},
			})
		}
	}

	return &v1.Message{
		Role:     role,
		Contents: contents,
	}
}

func convertChatReqFromAnthropic(req *anthropic.MessageNewParams) *v1.ChatReq {
	var messages []*v1.Message

	system := convertSystemFromAnthropic(req.System)
	if system != nil {
		messages = append(messages, system)
	}

	for _, message := range req.Messages {
		messages = append(messages, convertMessageFromAnthropic(&message))
	}

	var tools []*v1.Tool
	for _, tool := range req.Tools {
		t := &v1.Tool{}
		switch {
		case tool.OfTool != nil:
			var parameters *v1.Schema
			j, err := json.Marshal(tool.OfTool.InputSchema)
			if err != nil {
				log.Errorf("failed to marshal anthropic tool schema: %s", err.Error())
				continue
			}
			err = json.Unmarshal(j, &parameters)
			if err != nil {
				log.Errorf("failed to unmarshal anthropic tool schema: %s", err.Error())
				continue
			}

			t.Tool = &v1.Tool_Function_{
				Function: &v1.Tool_Function{
					Name:        tool.OfTool.Name,
					Description: tool.OfTool.Description.Value,
					Parameters:  parameters,
				},
			}
		default:
			log.Errorf("unsupported anthropic tool: %v", tool)
			continue
		}
		tools = append(tools, t)
	}

	var metadata map[string]string
	if req.Metadata.UserID.Valid() {
		metadata = map[string]string{
			"user_id": req.Metadata.UserID.Value,
		}
	}

	return &v1.ChatReq{
		Model:    string(req.Model),
		Config:   convertGenerationConfigFromAnthropic(req),
		Messages: messages,
		Tools:    tools,
		Metadata: metadata,
	}
}

// convertStatusToAnthropic maps internal chat status to Anthropic stop reason.
func convertStatusToAnthropic(status v1.ChatStatus) anthropic.StopReason {
	switch status {
	case v1.ChatStatus_CHAT_COMPLETED:
		return anthropic.StopReasonEndTurn
	case v1.ChatStatus_CHAT_REFUSED:
		return anthropic.StopReasonRefusal
	case v1.ChatStatus_CHAT_PENDING_TOOL_USE:
		return anthropic.StopReasonToolUse
	case v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT:
		return anthropic.StopReasonMaxTokens
	default:
		return anthropic.StopReasonEndTurn
	}
}

func convertChatRespToAnthropic(resp *v1.ChatResp) *anthropic.Message {
	anthropicResp := &anthropic.Message{
		Type:       "message",
		Model:      anthropic.Model(resp.Model),
		Role:       "assistant",
		StopReason: convertStatusToAnthropic(resp.Status),
	}

	if resp.Message != nil {
		anthropicResp.ID = resp.Message.Id
		for _, content := range resp.Message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					if content.Metadata["redacted_thinking"] != "" {
						// Redacted thinking block
						anthropicResp.Content = append(anthropicResp.Content, anthropic.ContentBlockUnion{
							Type: "redacted_thinking",
							Data: content.Metadata["redacted_thinking"],
						})
					} else {
						// Thinking block
						anthropicResp.Content = append(anthropicResp.Content, anthropic.ContentBlockUnion{
							Type:      "thinking",
							Thinking:  c.Text,
							Signature: content.Metadata["signature"],
						})
					}
				} else {
					if c.Text != "" {
						anthropicResp.Content = append(anthropicResp.Content, anthropic.ContentBlockUnion{
							Type: "text",
							Text: c.Text,
						})
					}
				}
			case *v1.Content_ToolUse:
				f := c.ToolUse
				anthropicResp.Content = append(anthropicResp.Content, anthropic.ContentBlockUnion{
					Type:  "tool_use",
					ID:    f.Id,
					Name:  f.Name,
					Input: json.RawMessage(f.GetTextualInput()),
				})
			}
		}
	}

	if resp.Statistics != nil && resp.Statistics.Usage != nil {
		anthropicResp.Usage = convertStatisticsToAnthropic(resp.Statistics)
	}

	return anthropicResp
}

func convertStatisticsToAnthropic(stats *v1.Statistics) anthropic.Usage {
	return anthropic.Usage{
		InputTokens:          max(int64(stats.Usage.InputTokens)-int64(stats.Usage.CachedInputTokens), 0),
		OutputTokens:         int64(stats.Usage.OutputTokens),
		CacheReadInputTokens: int64(stats.Usage.CachedInputTokens),
	}
}
