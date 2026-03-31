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

package openai

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/util"
)

func convertConfigToOpenAIResponse(config *v1.GenerationConfig, req *responses.ResponseNewParams) {
	if config == nil {
		return
	}
	if config.MaxTokens != nil {
		req.MaxOutputTokens = openai.Opt(*config.MaxTokens)
	}
	if config.Temperature != nil {
		req.Temperature = openai.Opt(float64(*config.Temperature))
	}
	if config.TopP != nil {
		req.TopP = openai.Opt(float64(*config.TopP))
	}

	switch g := config.Grammar.(type) {
	case *v1.GenerationConfig_PresetGrammar:
		if g.PresetGrammar == "json_object" {
			req.Text = responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
				},
			}
		}
	case *v1.GenerationConfig_JsonSchema:
		var schema map[string]any
		if err := json.Unmarshal([]byte(g.JsonSchema), &schema); err == nil {
			req.Text = responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
						Name:   "custom_schema",
						Schema: schema,
					},
				},
			}
		}
	case *v1.GenerationConfig_Schema:
		schemaMap := convertSchemaToMap(g.Schema)
		req.Text = responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "custom_schema",
					Schema: schemaMap,
				},
			},
		}
	}
}

func convertImageToOpenAIResponseInputURL(image *v1.Image) string {
	switch image.Source.(type) {
	case *v1.Image_Url:
		return image.GetUrl()
	case *v1.Image_Data:
		return util.EncodeImageDataToUrl(image.GetData(), image.GetMimeType())
	}
	return ""
}

func (r *upstream) convertToolResultToFunctionCallOutput(toolResult *v1.ToolResult) responses.ResponseInputItemUnionParam {
	var outputItems responses.ResponseFunctionCallOutputItemListParam
	hasNonText := false

	for _, output := range toolResult.GetOutputs() {
		switch o := output.GetOutput().(type) {
		case *v1.ToolResult_Output_Text:
			outputItems = append(outputItems, responses.ResponseFunctionCallOutputItemUnionParam{
				OfInputText: &responses.ResponseInputTextContentParam{Text: o.Text},
			})
		case *v1.ToolResult_Output_Image:
			hasNonText = true
			outputItems = append(outputItems, responses.ResponseFunctionCallOutputItemUnionParam{
				OfInputImage: &responses.ResponseInputImageContentParam{
					ImageURL: openai.Opt(convertImageToOpenAIResponseInputURL(o.Image)),
					Detail:   responses.ResponseInputImageContentDetailAuto,
				},
			})
		}
	}

	if !hasNonText {
		if r.config.PreferStringContentForTool {
			return responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: toolResult.Id,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfString: openai.Opt(toolResult.GetTextualOutput()),
					},
				},
			}
		}
		if r.config.PreferSinglePartContent && len(outputItems) > 1 {
			return responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: toolResult.Id,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfResponseFunctionCallOutputItemArray: responses.ResponseFunctionCallOutputItemListParam{
							{OfInputText: &responses.ResponseInputTextContentParam{Text: toolResult.GetTextualOutput()}},
						},
					},
				},
			}
		}
	}

	return responses.ResponseInputItemUnionParam{
		OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
			CallID: toolResult.Id,
			Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
				OfResponseFunctionCallOutputItemArray: outputItems,
			},
		},
	}
}

func (r *upstream) convertMessageToOpenAIResponseInput(message *v1.Message) []responses.ResponseInputItemUnionParam {
	plainText := ""
	isPlainText := true

	{
		var sb strings.Builder
		for _, content := range message.Contents {
			switch c := content.Content.(type) {
			case *v1.Content_Text:
				sb.WriteString(c.Text)
			default:
				isPlainText = false
			}
		}
		plainText = sb.String()
	}

	switch message.Role {
	case v1.Role_SYSTEM:
		if isPlainText && r.config.PreferStringContentForSystem {
			return []responses.ResponseInputItemUnionParam{{
				OfMessage: &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(plainText)},
				},
			}}
		} else if isPlainText && r.config.PreferSinglePartContent {
			return []responses.ResponseInputItemUnionParam{{
				OfMessage: &responses.EasyInputMessageParam{
					Role: responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{
						OfInputItemContentList: responses.ResponseInputMessageContentListParam{
							{OfInputText: &responses.ResponseInputTextParam{Text: plainText}},
						},
					},
				},
			}}
		} else {
			var contents responses.ResponseInputMessageContentListParam

			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					contents = append(contents, responses.ResponseInputContentUnionParam{
						OfInputText: &responses.ResponseInputTextParam{Text: c.Text},
					})
				case *v1.Content_Image:
					contents = append(contents, responses.ResponseInputContentUnionParam{
						OfInputImage: &responses.ResponseInputImageParam{
							Detail:   responses.ResponseInputImageDetailAuto,
							ImageURL: openai.Opt(convertImageToOpenAIResponseInputURL(c.Image)),
						},
					})
				default:
					r.log.Errorf("unsupported content for system: %v", c)
				}
			}

			return []responses.ResponseInputItemUnionParam{{
				OfMessage: &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{OfInputItemContentList: contents},
				},
			}}
		}
	case v1.Role_USER:
		var result []responses.ResponseInputItemUnionParam
		var userContents []*v1.Content

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_ToolResult:
				result = append(result, r.convertToolResultToFunctionCallOutput(c.ToolResult))
			default:
				userContents = append(userContents, content)
			}
		}

		if len(result) == 0 || len(userContents) > 0 {
			msg := &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
			}

			if isPlainText && r.config.PreferStringContentForUser {
				msg.Content = responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(plainText)}
			} else if isPlainText && r.config.PreferSinglePartContent {
				msg.Content = responses.EasyInputMessageContentUnionParam{
					OfInputItemContentList: responses.ResponseInputMessageContentListParam{{
						OfInputText: &responses.ResponseInputTextParam{Text: plainText},
					}},
				}
			} else {
				var contents responses.ResponseInputMessageContentListParam

				for _, content := range userContents {
					switch c := content.GetContent().(type) {
					case *v1.Content_Text:
						contents = append(contents, responses.ResponseInputContentUnionParam{
							OfInputText: &responses.ResponseInputTextParam{Text: c.Text},
						})
					case *v1.Content_Image:
						contents = append(contents, responses.ResponseInputContentUnionParam{
							OfInputImage: &responses.ResponseInputImageParam{
								Detail:   responses.ResponseInputImageDetailAuto,
								ImageURL: openai.Opt(convertImageToOpenAIResponseInputURL(c.Image)),
							},
						})
					default:
						r.log.Errorf("unsupported content for user: %v", c)
					}
				}

				msg.Content = responses.EasyInputMessageContentUnionParam{
					OfInputItemContentList: contents,
				}
			}

			result = append(result, responses.ResponseInputItemUnionParam{OfMessage: msg})
		}

		return result
	case v1.Role_MODEL:
		var result []responses.ResponseInputItemUnionParam
		var reasoning *responses.ResponseReasoningItemParam

		flushReasoning := func() {
			if reasoning != nil {
				result = append(result, responses.ResponseInputItemUnionParam{OfReasoning: reasoning})
				reasoning = nil
			}
		}

		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				if content.Reasoning {
					id := content.Meta("id")
					if reasoning == nil || reasoning.ID != id {
						flushReasoning()
						reasoning = &responses.ResponseReasoningItemParam{ID: id}
					}

					if enc := content.Meta("encrypted_content"); enc != "" {
						reasoning.EncryptedContent = openai.Opt(enc)
					}

					switch content.Meta("kind") {
					case "summary":
						reasoning.Summary = append(reasoning.Summary,
							responses.ResponseReasoningItemSummaryParam{Text: c.Text})
					case "content":
						reasoning.Content = append(reasoning.Content,
							responses.ResponseReasoningItemContentParam{Text: c.Text})
					}
					continue
				}
				flushReasoning()
				msg := &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleAssistant,
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.Opt(c.Text)},
				}
				if phase := message.Meta("phase"); phase != "" {
					msg.Phase = responses.EasyInputMessagePhase(phase)
				}
				result = append(result, responses.ResponseInputItemUnionParam{OfMessage: msg})
			case *v1.Content_ToolUse:
				flushReasoning()
				result = append(result, responses.ResponseInputItemUnionParam{
					OfFunctionCall: &responses.ResponseFunctionToolCallParam{
						CallID:    c.ToolUse.Id,
						Name:      c.ToolUse.Name,
						Arguments: c.ToolUse.GetTextualInput(),
					},
				})
			default:
				r.log.Errorf("unsupported content for assistant: %v", c)
			}
		}
		flushReasoning()
		return result
	default:
		r.log.Errorf("invalid role: %v", message.Role)
		return nil
	}
}

func (r *upstream) convertRequestToOpenAIResponse(req *entity.ChatReq) responses.ResponseNewParams {
	openAIReq := responses.ResponseNewParams{
		Model: responses.ResponsesModel(req.Model),
		Store: openai.Opt(false),
	}

	if req.Config != nil {
		convertConfigToOpenAIResponse(req.Config, &openAIReq)
	}

	for _, message := range req.Messages {
		items := r.convertMessageToOpenAIResponseInput(message)
		openAIReq.Input.OfInputItemList = append(openAIReq.Input.OfInputItemList, items...)
	}

	if req.Tools != nil {
		var tools []responses.ToolUnionParam
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				ft := &responses.FunctionToolParam{
					Name:       t.Function.Name,
					Parameters: convertSchemaToMap(t.Function.Parameters),
				}
				if t.Function.Description != "" {
					ft.Description = openai.Opt(t.Function.Description)
				}
				tools = append(tools, responses.ToolUnionParam{OfFunction: ft})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		openAIReq.Tools = tools
	}

	return openAIReq
}

func convertStatusFromOpenAIResponse(status responses.ResponseStatus) v1.ChatStatus {
	switch status {
	case responses.ResponseStatusCompleted:
		return v1.ChatStatus_CHAT_COMPLETED
	case responses.ResponseStatusFailed:
		return v1.ChatStatus_CHAT_REFUSED
	case responses.ResponseStatusCancelled:
		return v1.ChatStatus_CHAT_CANCELLED
	case responses.ResponseStatusIncomplete:
		return v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT
	default:
		return v1.ChatStatus_CHAT_IN_PROGRESS
	}
}

