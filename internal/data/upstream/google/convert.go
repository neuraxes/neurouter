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

package google

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/genai"
	"k8s.io/utils/ptr"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func convertGenerationConfigToGoogle(config *v1.GenerationConfig, googleConfig *genai.GenerateContentConfig) {
	if config == nil || googleConfig == nil {
		return
	}
	if config.MaxTokens != nil {
		googleConfig.MaxOutputTokens = int32(*config.MaxTokens)
	}
	if config.Temperature != nil {
		googleConfig.Temperature = config.Temperature
	}
	if config.TopP != nil {
		googleConfig.TopP = config.TopP
	}
	if config.TopK != nil {
		googleConfig.TopK = ptr.To(float32(*config.TopK))
	}
}

func convertFunctionParametersToGoogle(params *v1.Schema) *genai.Schema {
	if params == nil {
		return nil
	}
	schema := &genai.Schema{
		Type:        genai.Type(strings.TrimPrefix(v1.Schema_Type_name[int32(params.Type)], "TYPE_")),
		Description: params.Description,
		Required:    params.Required,
		Enum:        params.Enum,
	}
	switch params.Type {
	case v1.Schema_TYPE_ARRAY:
		schema.Items = convertFunctionParametersToGoogle(params.Items)
	case v1.Schema_TYPE_OBJECT:
		schema.Properties = make(map[string]*genai.Schema)
		for key, prop := range params.Properties {
			schema.Properties[key] = convertFunctionParametersToGoogle(prop)
		}
	}
	return schema
}

func convertToolsToGoogle(tools []*v1.Tool) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}
	var functionDecls []*genai.FunctionDeclaration
	for _, tool := range tools {
		switch t := tool.Tool.(type) {
		case *v1.Tool_Function_:
			functionDecls = append(functionDecls, &genai.FunctionDeclaration{
				Name:        t.Function.GetName(),
				Description: t.Function.GetDescription(),
				Parameters:  convertFunctionParametersToGoogle(t.Function.GetParameters()),
			})
		}
	}
	return []*genai.Tool{{FunctionDeclarations: functionDecls}}
}

// inferImageType infers the MIME type from the byte data
func inferImageType(data []byte) string {
	if len(data) >= 8 {
		// PNG: 89 50 4E 47 0D 0A 1A 0A
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 && data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
			return "image/png"
		}
	}
	if len(data) >= 3 {
		// JPEG: FF D8 FF
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
		// GIF: GIF87a or GIF89a
		if data[0] == 'G' && data[1] == 'I' && data[2] == 'F' {
			return "image/gif"
		}
	}
	if len(data) >= 12 {
		// WEBP: RIFF....WEBP
		if data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P' {
			return "image/webp"
		}
	}
	if len(data) >= 2 {
		// BMP: BM
		if data[0] == 'B' && data[1] == 'M' {
			return "image/bmp"
		}
	}
	return "application/octet-stream"
}

func convertContentToGoogle(content *v1.Content) *genai.Part {
	switch c := content.Content.(type) {
	case *v1.Content_Text:
		return genai.NewPartFromText(c.Text)
	case *v1.Content_Image:
		mimeType := c.Image.MimeType
		switch source := c.Image.Source.(type) {
		case *v1.Image_Url:
			return genai.NewPartFromURI(source.Url, mimeType)
		case *v1.Image_Data:
			if mimeType == "" {
				mimeType = inferImageType(source.Data)
			}
			return genai.NewPartFromBytes(source.Data, mimeType)
		default:
			return nil
		}
	case *v1.Content_ToolUse:
		textualInput := c.ToolUse.GetTextualInput()

		var args map[string]any
		if textualInput != "" {
			if err := json.Unmarshal([]byte(textualInput), &args); err != nil {
				args = map[string]any{
					"args": textualInput,
				}
			}
		}

		return genai.NewPartFromFunctionCall(c.ToolUse.Name, args)
	case *v1.Content_ToolResult:
		textualOutput := c.ToolResult.GetTextualOutput()

		var output map[string]any
		if textualOutput != "" {
			if err := json.Unmarshal([]byte(textualOutput), &output); err != nil {
				output = map[string]any{
					"result": textualOutput,
				}
			}
		}

		return genai.NewPartFromFunctionResponse(c.ToolResult.Id, output)
	default:
		return nil
	}
}

func convertMessageToGoogle(msg *v1.Message) *genai.Content {
	var parts []*genai.Part
	for _, content := range msg.Contents {
		if part := convertContentToGoogle(content); part != nil {
			parts = append(parts, part)
		}
	}

	role := ""
	switch msg.Role {
	case v1.Role_SYSTEM:
		role = "user" // Google AI doesn't support system role, use user role instead
	case v1.Role_USER:
		role = "user"
	case v1.Role_MODEL:
		role = "model"
	}

	return &genai.Content{
		Parts: parts,
		Role:  role,
	}
}

func convertMessageFromGoogle(content *genai.Content) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	for _, part := range content.Parts {
		if part.Thought {
			if part.Text != "" {
				message.Contents = append(message.Contents, &v1.Content{
					Content: &v1.Content_Reasoning{
						Reasoning: part.Text,
					},
				})
				continue
			}
		}
		if part.Text != "" {
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: part.Text,
				},
			})
			continue
		}
		if part.FunctionCall != nil {
			args, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				continue
			}
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   part.FunctionCall.Name,
						Name: part.FunctionCall.Name,
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
		}
	}

	return message
}

func convertStatisticsFromGoogle(usage *genai.GenerateContentResponseUsageMetadata) *v1.Statistics {
	if usage == nil {
		return nil
	}

	return &v1.Statistics{
		Usage: &v1.Statistics_Usage{
			InputTokens:       uint32(usage.PromptTokenCount),
			OutputTokens:      uint32(usage.CandidatesTokenCount + usage.ThoughtsTokenCount),
			CachedInputTokens: uint32(usage.CachedContentTokenCount),
		},
	}
}
