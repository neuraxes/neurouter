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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/genai"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestConvertFunctionParametersToGoogle(t *testing.T) {
	Convey("convertFunctionParametersToGoogle should convert parameters", t, func() {
		params := &v1.Schema{
			Type:     v1.Schema_TYPE_OBJECT,
			Required: []string{"foo"},
			Properties: map[string]*v1.Schema{
				"foo": {Type: v1.Schema_TYPE_STRING, Description: "desc"},
			},
		}
		schema := convertFunctionParametersToGoogle(params)
		So(schema, ShouldNotBeNil)
		So(schema.Type, ShouldEqual, genai.TypeObject)
		So(schema.Required, ShouldResemble, []string{"foo"})
		So(schema.Properties["foo"].Type, ShouldEqual, genai.TypeString)
		So(schema.Properties["foo"].Description, ShouldEqual, "desc")
	})

	Convey("convertFunctionParametersToGoogle should return nil for nil input", t, func() {
		So(convertFunctionParametersToGoogle(nil), ShouldBeNil)
	})
}

func TestConvertToolsToGoogle(t *testing.T) {
	Convey("convertToolsToGoogle should convert tools", t, func() {
		tools := []*v1.Tool{
			{
				Tool: &v1.Tool_Function_{
					Function: &v1.Tool_Function{
						Name:        "fn",
						Description: "desc",
						Parameters:  &v1.Schema{Type: v1.Schema_TYPE_OBJECT},
					},
				},
			},
		}
		result := convertToolsToGoogle(tools)
		So(result, ShouldHaveLength, 1)
		So(result[0].FunctionDeclarations, ShouldHaveLength, 1)
		So(result[0].FunctionDeclarations[0].Name, ShouldEqual, "fn")
		So(result[0].FunctionDeclarations[0].Description, ShouldEqual, "desc")
		So(result[0].FunctionDeclarations[0].Parameters.Type, ShouldEqual, genai.TypeObject)
	})

	Convey("convertToolsToGoogle should return nil for empty input", t, func() {
		So(convertToolsToGoogle(nil), ShouldBeNil)
		So(convertToolsToGoogle([]*v1.Tool{}), ShouldBeNil)
	})
}

func TestInferImageType(t *testing.T) {
	Convey("inferImageType should detect png", t, func() {
		data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		So(inferImageType(data), ShouldEqual, "image/png")
	})
	Convey("inferImageType should detect jpeg", t, func() {
		data := []byte{0xFF, 0xD8, 0xFF}
		So(inferImageType(data), ShouldEqual, "image/jpeg")
	})
	Convey("inferImageType should detect gif", t, func() {
		data := []byte{'G', 'I', 'F'}
		So(inferImageType(data), ShouldEqual, "image/gif")
	})
	Convey("inferImageType should detect webp", t, func() {
		data := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}
		So(inferImageType(data), ShouldEqual, "image/webp")
	})
	Convey("inferImageType should detect bmp", t, func() {
		data := []byte{'B', 'M'}
		So(inferImageType(data), ShouldEqual, "image/bmp")
	})
	Convey("inferImageType should return default for unknown", t, func() {
		data := []byte{0x00, 0x01}
		So(inferImageType(data), ShouldEqual, "application/octet-stream")
	})
}

func TestConvertContentToGoogle(t *testing.T) {
	Convey("convertContentToGoogle should convert text", t, func() {
		content := &v1.Content{
			Content: &v1.Content_Text{Text: "hello"},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.Text, ShouldEqual, "hello")
	})

	Convey("convertContentToGoogle should convert image data", t, func() {
		content := &v1.Content{
			Content: &v1.Content_Image{
				Image: &v1.Image{
					Source: &v1.Image_Data{
						Data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.InlineData, ShouldNotBeNil)
		So(part.InlineData.MIMEType, ShouldEqual, "image/png")
	})

	Convey("convertContentToGoogle should convert image data with explicit mime type", t, func() {
		content := &v1.Content{
			Content: &v1.Content_Image{
				Image: &v1.Image{
					MimeType: "image/jpeg",
					Source: &v1.Image_Data{
						Data: []byte{0xFF, 0xD8, 0xFF},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.InlineData, ShouldNotBeNil)
		So(part.InlineData.MIMEType, ShouldEqual, "image/jpeg")
	})

	Convey("convertContentToGoogle should convert image URL", t, func() {
		content := &v1.Content{
			Content: &v1.Content_Image{
				Image: &v1.Image{
					MimeType: "image/png",
					Source: &v1.Image_Url{
						Url: "https://example.com/image.png",
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.FileData, ShouldNotBeNil)
		So(part.FileData.FileURI, ShouldEqual, "https://example.com/image.png")
		So(part.FileData.MIMEType, ShouldEqual, "image/png")
	})

	Convey("convertContentToGoogle should convert tool call with valid JSON", t, func() {
		args := map[string]any{"foo": "bar"}
		argsJSON, _ := json.Marshal(args)
		content := &v1.Content{
			Content: &v1.Content_ToolUse{
				ToolUse: &v1.ToolUse{
					Name: "fn",
					Inputs: []*v1.ToolUse_Input{
						{
							Input: &v1.ToolUse_Input_Text{
								Text: string(argsJSON),
							},
						},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.FunctionCall, ShouldNotBeNil)
		So(part.FunctionCall.Name, ShouldEqual, "fn")
		So(part.FunctionCall.Args, ShouldResemble, args)
	})

	Convey("convertContentToGoogle should convert tool call with invalid JSON", t, func() {
		content := &v1.Content{
			Content: &v1.Content_ToolUse{
				ToolUse: &v1.ToolUse{
					Name: "fn",
					Inputs: []*v1.ToolUse_Input{
						{
							Input: &v1.ToolUse_Input_Text{
								Text: "invalid json",
							},
						},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.FunctionCall, ShouldNotBeNil)
		So(part.FunctionCall.Name, ShouldEqual, "fn")
		So(part.FunctionCall.Args, ShouldResemble, map[string]any{"args": "invalid json"})
	})

	Convey("convertContentToGoogle should convert tool result with valid JSON", t, func() {
		output := map[string]any{"status": "success"}
		outputJSON, _ := json.Marshal(output)
		content := &v1.Content{
			Content: &v1.Content_ToolResult{
				ToolResult: &v1.ToolResult{
					Id: "tool-id",
					Outputs: []*v1.ToolResult_Output{
						{
							Output: &v1.ToolResult_Output_Text{
								Text: string(outputJSON),
							},
						},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.FunctionResponse, ShouldNotBeNil)
		So(part.FunctionResponse.Name, ShouldEqual, "tool-id")
		So(part.FunctionResponse.Response, ShouldResemble, output)
	})

	Convey("convertContentToGoogle should convert tool result with plain text", t, func() {
		content := &v1.Content{
			Content: &v1.Content_ToolResult{
				ToolResult: &v1.ToolResult{
					Id: "tool-id",
					Outputs: []*v1.ToolResult_Output{
						{
							Output: &v1.ToolResult_Output_Text{
								Text: "result text",
							},
						},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldNotBeNil)
		So(part.FunctionResponse, ShouldNotBeNil)
		So(part.FunctionResponse.Name, ShouldEqual, "tool-id")
		So(part.FunctionResponse.Response, ShouldResemble, map[string]any{"result": "result text"})
	})

	Convey("convertContentToGoogle should return nil for unknown", t, func() {
		content := &v1.Content{}
		So(convertContentToGoogle(content), ShouldBeNil)
	})

	Convey("convertContentToGoogle should return nil for unknown image source", t, func() {
		content := &v1.Content{
			Content: &v1.Content_Image{
				Image: &v1.Image{
					Source: nil, // triggers default case
				},
			},
		}
		So(convertContentToGoogle(content), ShouldBeNil)
	})
}

func TestConvertMessageToGoogle(t *testing.T) {
	Convey("convertMessageToGoogle should convert user message", t, func() {
		msg := &v1.Message{
			Role: v1.Role_USER,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: "hi"}},
			},
		}
		result := convertMessageToGoogle(msg)
		So(result, ShouldNotBeNil)
		So(result.Role, ShouldEqual, "user")
		So(result.Parts, ShouldHaveLength, 1)
		So(result.Parts[0].Text, ShouldEqual, "hi")
	})

	Convey("convertMessageToGoogle should convert user message with tool result", t, func() {
		msg := &v1.Message{
			Role: v1.Role_USER,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_ToolResult{
						ToolResult: &v1.ToolResult{
							Id: "tool1",
							Outputs: []*v1.ToolResult_Output{
								{
									Output: &v1.ToolResult_Output_Text{
										Text: "result2",
									},
								},
							},
						},
					},
				},
			},
		}
		result := convertMessageToGoogle(msg)
		So(result, ShouldNotBeNil)
		So(result.Role, ShouldEqual, "user")
		So(result.Parts, ShouldHaveLength, 1)
		So(result.Parts[0].FunctionResponse, ShouldNotBeNil)
		So(result.Parts[0].FunctionResponse.Name, ShouldEqual, "tool1")
		So(result.Parts[0].FunctionResponse.Response["result"], ShouldEqual, "result2")
	})

	Convey("convertMessageToGoogle should handle system role", t, func() {
		msg := &v1.Message{
			Role: v1.Role_SYSTEM,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: "sys"}},
			},
		}
		result := convertMessageToGoogle(msg)
		So(result, ShouldNotBeNil)
		So(result.Role, ShouldEqual, "user")
		So(result.Parts, ShouldHaveLength, 1)
		So(result.Parts[0].Text, ShouldEqual, "sys")
	})

	Convey("convertMessageToGoogle should handle model role", t, func() {
		msg := &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: "model"}},
			},
		}
		result := convertMessageToGoogle(msg)
		So(result, ShouldNotBeNil)
		So(result.Role, ShouldEqual, "model")
		So(result.Parts, ShouldHaveLength, 1)
		So(result.Parts[0].Text, ShouldEqual, "model")
	})
}

func TestConvertMessageFromGoogle(t *testing.T) {
	Convey("convertMessageFromGoogle should convert text", t, func() {
		content := &genai.Content{
			Parts: []*genai.Part{{Text: "hello"}},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Role, ShouldEqual, v1.Role_MODEL)
		So(msg.Contents, ShouldHaveLength, 1)
		So(msg.Contents[0].GetText(), ShouldEqual, "hello")
	})

	Convey("convertMessageFromGoogle should convert reasoning (thought)", t, func() {
		content := &genai.Content{
			Parts: []*genai.Part{{Text: "thinking...", Thought: true}},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Role, ShouldEqual, v1.Role_MODEL)
		So(msg.Contents, ShouldHaveLength, 1)
		So(msg.Contents[0].GetReasoning(), ShouldEqual, "thinking...")
	})

	Convey("convertMessageFromGoogle should skip thought without text", t, func() {
		content := &genai.Content{
			Parts: []*genai.Part{{Thought: true}},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Contents, ShouldBeEmpty)
	})

	Convey("convertMessageFromGoogle should convert function call", t, func() {
		part := &genai.FunctionCall{
			Name: "fn",
			Args: map[string]any{"foo": "bar"},
		}
		content := &genai.Content{
			Parts: []*genai.Part{{FunctionCall: part}},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Contents, ShouldHaveLength, 1)
		So(msg.Contents[0].GetToolUse().GetId(), ShouldEqual, "fn")
		So(msg.Contents[0].GetToolUse().GetName(), ShouldEqual, "fn")
		So(msg.Contents[0].GetToolUse().GetTextualInput(), ShouldEqual, `{"foo":"bar"}`)
	})

	Convey("convertMessageFromGoogle should skip function call with marshal error", t, func() {
		// Use a value that cannot be marshaled to JSON (e.g., a channel)
		part := &genai.FunctionCall{
			Name: "fn",
			Args: map[string]any{"bad": make(chan int)},
		}
		content := &genai.Content{
			Parts: []*genai.Part{{FunctionCall: part}},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Contents, ShouldBeEmpty)
	})
}

func TestConvertStatisticsFromGoogle(t *testing.T) {
	Convey("convertStatisticsFromGoogle should convert usage metadata", t, func() {
		usage := &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:        100,
			CandidatesTokenCount:    50,
			CachedContentTokenCount: 25,
		}

		stats := convertStatisticsFromGoogle(usage)

		So(stats, ShouldNotBeNil)
		So(stats.Usage, ShouldNotBeNil)
		So(stats.Usage.InputTokens, ShouldEqual, 100)
		So(stats.Usage.OutputTokens, ShouldEqual, 50)
		So(stats.Usage.CachedInputTokens, ShouldEqual, 25)
	})

	Convey("convertStatisticsFromGoogle should return nil for nil input", t, func() {
		stats := convertStatisticsFromGoogle(nil)
		So(stats, ShouldBeNil)
	})

	Convey("convertStatisticsFromGoogle should handle zero values", t, func() {
		usage := &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:        0,
			CandidatesTokenCount:    0,
			CachedContentTokenCount: 0,
		}

		stats := convertStatisticsFromGoogle(usage)

		So(stats, ShouldNotBeNil)
		So(stats.Usage, ShouldNotBeNil)
		So(stats.Usage.InputTokens, ShouldEqual, 0)
		So(stats.Usage.OutputTokens, ShouldEqual, 0)
		So(stats.Usage.CachedInputTokens, ShouldEqual, 0)
	})
}
