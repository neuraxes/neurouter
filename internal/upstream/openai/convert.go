package openai

import (
	"strings"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz"
)

// convertMessageToOpenAI converts an internal message to a message that can be sent to the OpenAI API.
func (r *ChatRepo) convertMessageToOpenAI(message *v1.Message) openai.ChatCompletionMessageParamUnion {
	plainText := ""
	isPlainText := true

	{
		var sb strings.Builder
		for _, content := range message.Contents {
			if textContent, ok := content.GetContent().(*v1.Content_Text); ok {
				sb.WriteString(textContent.Text)
			} else {
				isPlainText = false
			}
		}
		plainText = sb.String()
	}

	switch message.Role {
	case v1.Role_SYSTEM:
		m := openai.ChatCompletionSystemMessageParam{
			Role: openai.F(openai.ChatCompletionSystemMessageParamRoleSystem),
		}
		if message.Name != "" {
			m.Name = openai.F(message.Name)
		}
		var parts []openai.ChatCompletionContentPartTextParam
		if isPlainText && (r.config.PreferStringContentForSystem || r.config.PreferSinglePartContent) {
			parts = append(parts, openai.TextPart(plainText))
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				default:
					r.log.Errorf("unsupported content for system: %v", c)
				}
			}
		}
		m.Content = openai.F(parts)
		return m
	case v1.Role_USER:
		var parts []openai.ChatCompletionContentPartUnionParam
		if isPlainText && (r.config.PreferStringContentForUser || r.config.PreferSinglePartContent) {
			parts = append(parts, openai.TextPart(plainText))
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				case *v1.Content_Image_:
					parts = append(parts, openai.ImagePart(c.Image.Url))
				default:
					r.log.Errorf("unsupported content for user: %v", c)
				}
			}
		}
		m := openai.UserMessageParts(parts...)
		if message.Name != "" {
			m.Name = openai.F(message.Name)
		}
		return m
	case v1.Role_MODEL:
		m := openai.ChatCompletionAssistantMessageParam{
			Role: openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
		}

		if message.Name != "" {
			m.Name = openai.F(message.Name)
		}

		if message.Contents != nil {
			var parts []openai.ChatCompletionAssistantMessageParamContentUnion
			if isPlainText && (r.config.PreferStringContentForAssistant || r.config.PreferSinglePartContent) {
				parts = append(parts, openai.TextPart(plainText))
			} else {
				for _, content := range message.Contents {
					switch c := content.GetContent().(type) {
					case *v1.Content_Text:
						parts = append(parts, openai.TextPart(c.Text))
					default:
						r.log.Errorf("unsupported content for assistant: %v", c)
					}
				}
			}
			m.Content = openai.F(parts)
		}

		if message.ToolCalls != nil {
			var toolCalls []openai.ChatCompletionMessageToolCallParam
			for _, toolCall := range message.ToolCalls {
				switch t := toolCall.Tool.(type) {
				case *v1.ToolCall_Function:
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID:   openai.F(toolCall.Id),
						Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
						Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      openai.F(t.Function.Name),
							Arguments: openai.F(t.Function.Arguments),
						}),
					})
				default:
					r.log.Errorf("unsupported tool call: %v", t)
				}
			}
			m.ToolCalls = openai.F(toolCalls)
		}

		return m
	case v1.Role_TOOL:
		m := openai.ChatCompletionToolMessageParam{
			Role:       openai.F(openai.ChatCompletionToolMessageParamRoleTool),
			ToolCallID: openai.F(message.ToolCallId),
		}
		var parts []openai.ChatCompletionContentPartTextParam
		if isPlainText && (r.config.PreferStringContentForTool || r.config.PreferSinglePartContent) {
			parts = append(parts, openai.TextPart(plainText))
		} else {
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				default:
					r.log.Errorf("unsupported content for system: %v", c)
				}
			}
		}
		m.Content = openai.F(parts)
		return m
	default:
		r.log.Errorf("unsupported role: %v", message.Role)
		return nil
	}
}

// convertRequestToOpenAI converts an internal request to a request that can be sent to the OpenAI API.
func (r *ChatRepo) convertRequestToOpenAI(req *biz.ChatReq) openai.ChatCompletionNewParams {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, message := range req.Messages {
		m := r.convertMessageToOpenAI(message)
		if m != nil {
			messages = append(messages, m)
		}
	}

	openAIReq := openai.ChatCompletionNewParams{
		Model:    openai.F(req.Model),
		Messages: openai.F(messages),
	}

	if c := req.Config; c != nil {
		if c.MaxTokens != 0 {
			openAIReq.MaxCompletionTokens = openai.F(c.MaxTokens)
		}
		openAIReq.Temperature = openai.F(float64(c.Temperature))
		if c.TopP != 0 {
			openAIReq.TopP = openai.F(float64(c.TopP))
		}
		openAIReq.FrequencyPenalty = openai.F(float64(c.FrequencyPenalty))
		openAIReq.PresencePenalty = openai.F(float64(c.PresencePenalty))
		if c.GetPresetGrammar() == "json_object" {
			openAIReq.ResponseFormat = openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
				openai.ResponseFormatJSONObjectParam{
					Type: openai.F(openai.ResponseFormatJSONObjectTypeJSONObject),
				},
			)
		}
	}

	if req.Tools != nil {
		var tools []openai.ChatCompletionToolParam
		for _, tool := range req.Tools {
			switch t := tool.Tool.(type) {
			case *v1.Tool_Function_:
				tools = append(tools, openai.ChatCompletionToolParam{
					Type: openai.F(openai.ChatCompletionToolTypeFunction),
					Function: openai.F(shared.FunctionDefinitionParam{
						Name:        openai.F(t.Function.Name),
						Description: openai.F(t.Function.Description),
						Parameters:  openai.F(toolFunctionParametersToOpenAI(t.Function.Parameters)),
					}),
				})
			default:
				r.log.Errorf("unsupported tool: %v", t)
			}
		}
		openAIReq.Tools = openai.F(tools)
	}

	return openAIReq
}

func toolFunctionParametersToOpenAI(parameters *v1.Tool_Function_Parameters) openai.FunctionParameters {
	return map[string]any{
		"type":       parameters.Type,
		"properties": parameters.Properties,
		"required":   parameters.Required,
	}
}

func (r *ChatRepo) convertMessageFromOpenAI(openAIMessage *openai.ChatCompletionMessage) *v1.Message {
	id, err := uuid.NewUUID()
	if err != nil {
		r.log.Fatalf("failed to generate UUID: %v", err)
	}

	message := &v1.Message{
		Id:   id.String(),
		Role: v1.Role_MODEL,
	}

	if openAIMessage.Content != "" {
		message.Contents = []*v1.Content{
			{
				Content: &v1.Content_Text{
					// The result may contain a leading space, so we need to trim it
					Text: strings.TrimSpace(openAIMessage.Content),
				},
			},
		}
	}

	if openAIMessage.ToolCalls != nil {
		var toolCalls []*v1.ToolCall
		for _, toolCall := range openAIMessage.ToolCalls {
			switch toolCall.Type {
			case openai.ChatCompletionMessageToolCallTypeFunction:
				toolCalls = append(toolCalls, &v1.ToolCall{
					Id: toolCall.ID,
					Tool: &v1.ToolCall_Function{
						Function: &v1.ToolCall_FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					},
				})
			default:
				r.log.Errorf("unsupported tool call: %v", toolCall)
			}
		}
		message.ToolCalls = toolCalls
	}

	return message
}
