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
	"encoding/base64"
	"encoding/json"
	"math"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/structpb"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/util"
)

func (r *upstream) convertRequestToAnthropic(req *entity.ChatReq) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model: anthropic.Model(req.Model),
	}

	r.convertGenerationConfigToAnthropic(req.Config, &params)

	if !r.config.SystemAsUser {
		params.System = r.convertSystemMessagesToAnthropic(req.Messages)
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
					InputSchema: convertSchemaToAnthropicToolInputSchema(t.Function.InputSchema),
				}
				if t.Function.Description != "" {
					at.Description = anthropic.Opt(t.Function.Description)
				}
				tools = append(tools, anthropic.ToolUnionParam{OfTool: at})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		params.Tools = tools
	}

	if req.Metadata != nil {
		if userID, ok := req.Metadata["user_id"]; ok {
			params.Metadata.UserID = anthropic.Opt(userID)
		}
	}

	return params
}

func (r *upstream) convertGenerationConfigToAnthropic(config *v1.GenerationConfig, req *anthropic.MessageNewParams) {
	if config == nil {
		return
	}
	if config.MaxTokens != nil && *config.MaxTokens > 0 {
		req.MaxTokens = *config.MaxTokens
	} else {
		req.MaxTokens = 8192
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
	if c := config.ReasoningConfig; c != nil {
		switch {
		case c.Effort == v1.ReasoningEffort_REASONING_EFFORT_NONE:
			req.Thinking.OfDisabled = &anthropic.ThinkingConfigDisabledParam{}
		case c.TokenBudget > 0:
			budgetTokens := int64(1024)
			if c.TokenBudget > 1024 {
				budgetTokens = int64(c.TokenBudget)
			}
			req.Thinking.OfEnabled = &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: budgetTokens,
			}
		default:
			req.Thinking.OfAdaptive = &anthropic.ThinkingConfigAdaptiveParam{}
			req.OutputConfig.Effort = convertReasoningEffortToAnthropic(c.Effort)
		}
	}
	if len(config.StopSequences) > 0 {
		req.StopSequences = config.StopSequences
	}
	switch g := config.Grammar.(type) {
	case *v1.GenerationConfig_Schema:
		if g.Schema != nil {
			req.OutputConfig.Format.Schema = g.Schema.AsMap()
		}
	}
}

func (r *upstream) convertSystemMessagesToAnthropic(messages []*v1.Message) []anthropic.TextBlockParam {
	var parts []anthropic.TextBlockParam
	for _, message := range messages {
		if message.Role != v1.Role_SYSTEM {
			continue
		}
		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				parts = append(parts, anthropic.TextBlockParam{Text: c.Text.GetText()})
			default:
				r.log.Errorf("unsupported content: %v", c)
			}
		}
	}
	return parts
}

func (r *upstream) convertMessageToAnthropic(message *v1.Message) anthropic.MessageParam {
	var parts []anthropic.ContentBlockParamUnion
	for _, content := range message.Contents {
		switch c := content.GetContent().(type) {
		case *v1.Content_Text:
			if content.Phase == v1.ContentPhase_CONTENT_PHASE_REASONING {
				parts = append(parts, anthropic.NewThinkingBlock(content.Signature, c.Text.GetText()))
			} else {
				parts = append(parts, anthropic.NewTextBlock(c.Text.GetText()))
			}
		case *v1.Content_Opaque:
			parts = append(parts, anthropic.NewRedactedThinkingBlock(c.Opaque))
		case *v1.Content_Image:
			switch src := c.Image.Source.(type) {
			case *v1.Image_Url:
				parts = append(parts, anthropic.NewImageBlock(
					anthropic.URLImageSourceParam{
						URL: src.Url,
					},
				))
			case *v1.Image_Data:
				parts = append(parts, anthropic.NewImageBlockBase64(
					c.Image.MimeType,
					base64.StdEncoding.EncodeToString(src.Data),
				))
			case *v1.Image_Base64:
				parts = append(parts, anthropic.NewImageBlockBase64(
					c.Image.MimeType,
					src.Base64,
				))
			}
		case *v1.Content_ToolUse:
			textualInput := c.ToolUse.GetTextualInput()

			var input any
			err := json.Unmarshal([]byte(textualInput), &input)
			if err != nil {
				// Fallback to string
				input = textualInput
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
				case *v1.ToolResult_Output_Image:
					switch src := o.Image.Source.(type) {
					case *v1.Image_Url:
						outputs = append(outputs, anthropic.ToolResultBlockParamContentUnion{
							OfImage: &anthropic.ImageBlockParam{
								Source: anthropic.ImageBlockParamSourceUnion{
									OfURL: &anthropic.URLImageSourceParam{
										URL: src.Url,
									},
								},
							},
						})
					case *v1.Image_Data:
						outputs = append(outputs, anthropic.ToolResultBlockParamContentUnion{
							OfImage: &anthropic.ImageBlockParam{
								Source: anthropic.ImageBlockParamSourceUnion{
									OfBase64: &anthropic.Base64ImageSourceParam{
										Data:      base64.StdEncoding.EncodeToString(src.Data),
										MediaType: anthropic.Base64ImageSourceMediaType(o.Image.MimeType),
									},
								},
							},
						})
					case *v1.Image_Base64:
						outputs = append(outputs, anthropic.ToolResultBlockParamContentUnion{
							OfImage: &anthropic.ImageBlockParam{
								Source: anthropic.ImageBlockParamSourceUnion{
									OfBase64: &anthropic.Base64ImageSourceParam{
										Data:      src.Base64,
										MediaType: anthropic.Base64ImageSourceMediaType(o.Image.MimeType),
									},
								},
							},
						})
					}
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

func convertSchemaToAnthropicToolInputSchema(params *structpb.Struct) (schema anthropic.ToolInputSchemaParam) {
	if params == nil {
		return
	}
	schema.Type = "object" // OpenRouter validates this field which will be omitted if not set explicitly

	fields := params.AsMap()
	if properties, ok := fields["properties"]; ok {
		schema.Properties = properties
	} else {
		schema.Properties = map[string]any{}
	}
	if required, ok := util.StringSliceFromAny(fields["required"]); ok {
		schema.Required = required
	}

	extraFields := make(map[string]any)
	for key, value := range fields {
		switch key {
		case "properties", "required", "type":
		default:
			extraFields[key] = value
		}
	}
	if len(extraFields) > 0 {
		schema.ExtraFields = extraFields
	}
	return
}

func convertMessageFromAnthropic(msg *anthropic.Message) *v1.Message {
	message := &v1.Message{
		Id:   msg.ID,
		Role: v1.Role_MODEL,
	}

	for _, content := range msg.Content {
		switch content.Type {
		case "thinking":
			message.Contents = append(message.Contents, &v1.Content{
				Signature: content.Signature,
				Phase:     v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content:   v1.NewTextContent(content.Thinking),
			})
		case "redacted_thinking":
			message.Contents = append(message.Contents, &v1.Content{
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Opaque{Opaque: content.Data},
			})
		case "text":
			message.Contents = append(message.Contents, &v1.Content{
				Content: v1.NewTextContent(content.Text),
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

func (c *anthropicChatStreamClient) convertStreamEventFromAnthropic(event *anthropic.MessageStreamEventUnion) []*entity.ChatEvent {
	switch event.Type {
	case "message_start":
		c.messageID = event.Message.ID
		c.model = event.Message.Model
		chatEvent := c.newChatEvent(v1.NewMessageStartEvent(c.messageID, c.model))
		// Anthropic reports input and cache token counts only in message_start;
		// message_delta carries only output tokens. Capture them here so the
		// cumulative usage snapshot retains the input accounting.
		if statistics := convertStatisticsFromAnthropic(&event.Message.Usage); statistics != nil {
			chatEvent.Usage = statistics.Usage
		}
		return []*entity.ChatEvent{chatEvent}

	case "content_block_start":
		index := uint32(event.Index)
		switch event.ContentBlock.Type {
		case "text":
			events := []*entity.ChatEvent{c.newChatEvent(v1.NewContentStartTextEvent(index, v1.ContentPhase_CONTENT_PHASE_NORMAL))}
			if event.ContentBlock.Text != "" {
				events = append(events, c.newChatEvent(v1.NewContentDeltaTextEvent(index, event.ContentBlock.Text)))
			}
			return events
		case "thinking":
			events := []*entity.ChatEvent{c.newChatEvent(v1.NewContentStartTextEvent(index, v1.ContentPhase_CONTENT_PHASE_REASONING))}
			if event.ContentBlock.Thinking != "" {
				events = append(events, c.newChatEvent(v1.NewContentDeltaTextEvent(index, event.ContentBlock.Thinking)))
			}
			if event.ContentBlock.Signature != "" {
				events = append(events, c.newChatEvent(v1.NewContentDeltaSignatureEvent(index, event.ContentBlock.Signature)))
			}
			return events
		case "redacted_thinking":
			if c.pendingSnapshotStops == nil {
				c.pendingSnapshotStops = map[uint32]struct{}{}
			}
			c.pendingSnapshotStops[index] = struct{}{}
			return []*entity.ChatEvent{c.newChatEvent(v1.NewContentSnapshotEvent(&v1.Content{
				Index:   &index,
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Opaque{Opaque: event.ContentBlock.Data},
			}))}
		case "tool_use":
			return []*entity.ChatEvent{c.newChatEvent(v1.NewContentStartToolUseEvent(index, event.ContentBlock.ID, event.ContentBlock.Name))}
		}
		return nil

	case "content_block_delta":
		index := uint32(event.Index)
		switch event.Delta.Type {
		case "thinking_delta":
			return []*entity.ChatEvent{c.newChatEvent(v1.NewContentDeltaTextEvent(index, event.Delta.Thinking))}
		case "signature_delta":
			return []*entity.ChatEvent{c.newChatEvent(v1.NewContentDeltaSignatureEvent(index, event.Delta.Signature))}
		case "text_delta":
			return []*entity.ChatEvent{c.newChatEvent(v1.NewContentDeltaTextEvent(index, event.Delta.Text))}
		case "input_json_delta":
			return []*entity.ChatEvent{c.newChatEvent(v1.NewContentDeltaToolInputTextEvent(index, event.Delta.PartialJSON))}
		}
		return nil

	case "content_block_stop":
		index := uint32(event.Index)
		if _, ok := c.pendingSnapshotStops[index]; ok {
			delete(c.pendingSnapshotStops, index)
			return nil
		}
		return []*entity.ChatEvent{c.newChatEvent(v1.NewContentStopEvent(index))}

	case "message_delta":
		hasUsage := event.Usage.InputTokens != 0 || event.Usage.OutputTokens != 0 ||
			event.Usage.CacheReadInputTokens != 0 || event.Usage.CacheCreationInputTokens != 0
		if !hasUsage && event.Delta.StopReason == "" {
			return nil
		}

		var chatEvent *entity.ChatEvent
		if event.Delta.StopReason != "" {
			chatEvent = c.newChatEvent(v1.NewMessageStopEvent(convertStatusFromAnthropic(event.Delta.StopReason)))
		} else {
			chatEvent = c.newChatEvent(nil)
		}
		if hasUsage {
			cachedInput := uint32(max(event.Usage.CacheReadInputTokens+event.Usage.CacheCreationInputTokens, 0))
			chatEvent.Usage = &v1.Usage{
				InputTokens:       uint32(max(event.Usage.InputTokens, 0)) + cachedInput,
				OutputTokens:      uint32(max(event.Usage.OutputTokens, 0)),
				CachedInputTokens: cachedInput,
			}
		}
		return []*entity.ChatEvent{chatEvent}

	default:
		return nil
	}
}

func (c *anthropicChatStreamClient) newChatEvent(event v1.ChatEventPayload) *entity.ChatEvent {
	return v1.NewChatEvent(c.req.GetId(), event)
}

func convertStatusFromAnthropic(stopReason anthropic.StopReason) v1.ChatStatus {
	switch stopReason {
	case anthropic.StopReasonToolUse:
		return v1.ChatStatus_CHAT_PENDING_TOOL_USE
	case anthropic.StopReasonEndTurn, anthropic.StopReasonStopSequence:
		return v1.ChatStatus_CHAT_COMPLETED
	case anthropic.StopReasonMaxTokens:
		return v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT
	case anthropic.StopReasonRefusal:
		return v1.ChatStatus_CHAT_REFUSED
	default:
		return v1.ChatStatus_CHAT_IN_PROGRESS
	}
}

func convertStatisticsFromAnthropic(usage *anthropic.Usage) *v1.Statistics {
	if usage == nil {
		return nil
	}

	if usage.InputTokens == 0 && usage.OutputTokens == 0 &&
		usage.CacheCreationInputTokens == 0 && usage.CacheReadInputTokens == 0 {
		return nil
	}

	cachedInput := uint32(max(usage.CacheReadInputTokens+usage.CacheCreationInputTokens, 0))
	return &v1.Statistics{
		Usage: &v1.Usage{
			InputTokens:       uint32(max(usage.InputTokens, 0)) + cachedInput,
			OutputTokens:      uint32(max(usage.OutputTokens, 0)),
			CachedInputTokens: cachedInput,
		},
	}
}

func convertReasoningEffortToAnthropic(effort v1.ReasoningEffort) anthropic.OutputConfigEffort {
	switch effort {
	case v1.ReasoningEffort_REASONING_EFFORT_MINIMAL, v1.ReasoningEffort_REASONING_EFFORT_LOW:
		return anthropic.OutputConfigEffortLow
	case v1.ReasoningEffort_REASONING_EFFORT_MEDIUM:
		return anthropic.OutputConfigEffortMedium
	case v1.ReasoningEffort_REASONING_EFFORT_HIGH:
		return anthropic.OutputConfigEffortHigh
	case v1.ReasoningEffort_REASONING_EFFORT_EXTRA_HIGH:
		return anthropic.OutputConfigEffortXhigh
	case v1.ReasoningEffort_REASONING_EFFORT_MAX:
		return anthropic.OutputConfigEffortMax
	default:
		return ""
	}
}
