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

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

// convertFunctionParamTypeToGoogle maps string type to [genai.Type]
func convertFunctionParamTypeToGoogle(t string) genai.Type {
	switch t {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeUnspecified
	}
}

// convertFunctionParametersToGoogle converts [v1.Tool_Function_Parameters] to [*genai.Schema]
func convertFunctionParametersToGoogle(params *v1.Tool_Function_Parameters) *genai.Schema {
	if params == nil {
		return nil
	}
	schema := &genai.Schema{
		Type:       convertFunctionParamTypeToGoogle(params.Type),
		Properties: map[string]*genai.Schema{},
		Required:   params.Required,
	}
	for k, v := range params.Properties {
		schema.Properties[k] = &genai.Schema{
			Type:        convertFunctionParamTypeToGoogle(v.Type),
			Description: v.Description,
		}
	}
	return schema
}

// convertToolsToGoogle converts a slice of [v1.Tool] to a slice of [genai.Tool]
func convertToolsToGoogle(tools []*v1.Tool) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}
	var functionDecls []*genai.FunctionDeclaration
	for _, tool := range tools {
		switch t := tool.Tool.(type) {
		case *v1.Tool_Function_:
			fn := t.Function
			functionDecls = append(functionDecls, &genai.FunctionDeclaration{
				Name:        fn.GetName(),
				Description: fn.GetDescription(),
				Parameters:  convertFunctionParametersToGoogle(fn.GetParameters()),
			})
		}
	}
	return []*genai.Tool{{FunctionDeclarations: functionDecls}}
}

// inferImageType infers the image type from the byte data
func inferImageType(data []byte) string {
	if len(data) >= 8 {
		// PNG: 89 50 4E 47 0D 0A 1A 0A
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 && data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
			return "png"
		}
	}
	if len(data) >= 3 {
		// JPEG: FF D8 FF
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "jpeg"
		}
		// GIF: GIF87a or GIF89a
		if data[0] == 'G' && data[1] == 'I' && data[2] == 'F' {
			return "gif"
		}
	}
	if len(data) >= 12 {
		// WEBP: RIFF....WEBP
		if data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P' {
			return "webp"
		}
	}
	if len(data) >= 2 {
		// BMP: BM
		if data[0] == 'B' && data[1] == 'M' {
			return "bmp"
		}
	}
	return "unknown"
}

// convertContentToGoogle converts [v1.Content] to [genai.Part]
func convertContentToGoogle(content *v1.Content) genai.Part {
	switch v := content.Content.(type) {
	case *v1.Content_Text:
		return genai.Text(v.Text)
	case *v1.Content_Image:
		switch source := v.Image.Source.(type) {
		case *v1.Image_Data:
			return genai.ImageData(inferImageType(source.Data), source.Data)
		default:
			return nil
		}
	case *v1.Content_ToolCall:
		switch t := v.ToolCall.Tool.(type) {
		case *v1.ToolCall_Function:
			var args map[string]any
			if t.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(t.Function.Arguments), &args); err != nil {
					return nil
				}
			}
			return genai.FunctionCall{
				Name: t.Function.Name,
				Args: args,
			}
		default:
			return nil
		}
	default:
		return nil
	}
}

// convertMessageToGoogle converts [v1.Message] to [genai.Content]
func convertMessageToGoogle(msg *v1.Message) *genai.Content {
	var parts []genai.Part
	if msg.Role == v1.Role_TOOL {
		parts = append(parts, genai.FunctionResponse{
			Name: msg.ToolCallId,
			Response: map[string]any{
				"content": msg.Contents[0].GetText(), // TODO: Handle multiple contents or different types
			},
		})
	} else {
		for _, content := range msg.Contents {
			if part := convertContentToGoogle(content); part != nil {
				parts = append(parts, part)
			}
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
	case v1.Role_TOOL:
		role = "user" // Google AI doesn't support tool role, use user role instead
	}

	return &genai.Content{
		Parts: parts,
		Role:  role,
	}
}

// convertMessageFromGoogle converts a [genai.Content] to a [v1.Message]
func convertMessageFromGoogle(content *genai.Content) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	for _, part := range content.Parts {
		switch part := part.(type) {
		case genai.Text:
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: string(part),
				},
			})
		case genai.FunctionCall:
			args, err := json.Marshal(part.Args)
			if err != nil {
				continue // Skip if arguments cannot be marshaled
			}
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_ToolCall{
					ToolCall: &v1.ToolCall{
						Id: part.Name,
						Tool: &v1.ToolCall_Function{
							Function: &v1.ToolCall_FunctionCall{
								Name:      part.Name,
								Arguments: string(args),
							},
						},
					},
				},
			})
		}
	}

	return message
}
