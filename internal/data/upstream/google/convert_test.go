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

	"github.com/google/generative-ai-go/genai"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestConvertFunctionParamTypeToGoogle(t *testing.T) {
	Convey("convertFunctionParamTypeToGoogle should map types correctly", t, func() {
		So(convertFunctionParamTypeToGoogle("string"), ShouldEqual, genai.TypeString)
		So(convertFunctionParamTypeToGoogle("number"), ShouldEqual, genai.TypeNumber)
		So(convertFunctionParamTypeToGoogle("integer"), ShouldEqual, genai.TypeInteger)
		So(convertFunctionParamTypeToGoogle("boolean"), ShouldEqual, genai.TypeBoolean)
		So(convertFunctionParamTypeToGoogle("array"), ShouldEqual, genai.TypeArray)
		So(convertFunctionParamTypeToGoogle("object"), ShouldEqual, genai.TypeObject)
		So(convertFunctionParamTypeToGoogle("unknown"), ShouldEqual, genai.TypeUnspecified)
	})
}

func TestConvertFunctionParametersToGoogle(t *testing.T) {
	Convey("convertFunctionParametersToGoogle should convert parameters", t, func() {
		params := &v1.Tool_Function_Parameters{
			Type:     "object",
			Required: []string{"foo"},
			Properties: map[string]*v1.Tool_Function_Parameters_Property{
				"foo": {Type: "string", Description: "desc"},
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
						Parameters: &v1.Tool_Function_Parameters{
							Type: "object",
						},
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
		So(inferImageType(data), ShouldEqual, "png")
	})
	Convey("inferImageType should detect jpeg", t, func() {
		data := []byte{0xFF, 0xD8, 0xFF}
		So(inferImageType(data), ShouldEqual, "jpeg")
	})
	Convey("inferImageType should detect gif", t, func() {
		data := []byte{'G', 'I', 'F'}
		So(inferImageType(data), ShouldEqual, "gif")
	})
	Convey("inferImageType should detect webp", t, func() {
		data := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}
		So(inferImageType(data), ShouldEqual, "webp")
	})
	Convey("inferImageType should detect bmp", t, func() {
		data := []byte{'B', 'M'}
		So(inferImageType(data), ShouldEqual, "bmp")
	})
	Convey("inferImageType should return unknown", t, func() {
		data := []byte{0x00, 0x01}
		So(inferImageType(data), ShouldEqual, "unknown")
	})
}

func TestConvertContentToGoogle(t *testing.T) {
	Convey("convertContentToGoogle should convert text", t, func() {
		content := &v1.Content{
			Content: &v1.Content_Text{Text: "hello"},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldResemble, genai.Text("hello"))
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
		So(part.(genai.Blob).MIMEType, ShouldEqual, "image/png")
	})

	Convey("convertContentToGoogle should convert tool call", t, func() {
		args := map[string]any{"foo": "bar"}
		argsJSON, _ := json.Marshal(args)
		content := &v1.Content{
			Content: &v1.Content_ToolCall{
				ToolCall: &v1.ToolCall{
					Tool: &v1.ToolCall_Function{
						Function: &v1.ToolCall_FunctionCall{
							Name:      "fn",
							Arguments: string(argsJSON),
						},
					},
				},
			},
		}
		part := convertContentToGoogle(content)
		So(part, ShouldHaveSameTypeAs, genai.FunctionCall{})
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

	Convey("convertContentToGoogle should return nil for tool call with invalid JSON", t, func() {
		content := &v1.Content{
			Content: &v1.Content_ToolCall{
				ToolCall: &v1.ToolCall{
					Tool: &v1.ToolCall_Function{
						Function: &v1.ToolCall_FunctionCall{
							Name:      "fn",
							Arguments: "{invalid json}",
						},
					},
				},
			},
		}
		So(convertContentToGoogle(content), ShouldBeNil)
	})

	Convey("convertContentToGoogle should return nil for unknown tool call type", t, func() {
		content := &v1.Content{
			Content: &v1.Content_ToolCall{
				ToolCall: &v1.ToolCall{
					Tool: nil, // triggers default case
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
		So(result.Parts[0].(genai.Text), ShouldEqual, genai.Text("hi"))
	})

	Convey("convertMessageToGoogle should convert tool message", t, func() {
		msg := &v1.Message{
			Role:       v1.Role_TOOL,
			ToolCallId: "tool1",
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: "result"}},
			},
		}
		result := convertMessageToGoogle(msg)
		So(result, ShouldNotBeNil)
		So(result.Role, ShouldEqual, "user")
		So(result.Parts, ShouldHaveLength, 1)
		So(result.Parts[0].(genai.FunctionResponse).Name, ShouldEqual, "tool1")
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
		So(result.Parts[0].(genai.Text), ShouldEqual, genai.Text("sys"))
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
		So(result.Parts[0].(genai.Text), ShouldEqual, genai.Text("model"))
	})
}

func TestConvertMessageFromGoogle(t *testing.T) {
	Convey("convertMessageFromGoogle should convert text", t, func() {
		content := &genai.Content{
			Parts: []genai.Part{genai.Text("hello")},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Role, ShouldEqual, v1.Role_MODEL)
		So(msg.Contents, ShouldHaveLength, 1)
		So(msg.Contents[0].GetText(), ShouldEqual, "hello")
	})

	Convey("convertMessageFromGoogle should convert function call", t, func() {
		part := genai.FunctionCall{
			Name: "fn",
			Args: map[string]any{"foo": "bar"},
		}
		content := &genai.Content{
			Parts: []genai.Part{part},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Contents, ShouldHaveLength, 1)
		So(msg.Contents[0].GetToolCall().GetTool().(*v1.ToolCall_Function).Function.Name, ShouldEqual, "fn")
		So(msg.Contents[0].GetToolCall().GetTool().(*v1.ToolCall_Function).Function.Arguments, ShouldEqual, `{"foo":"bar"}`)
	})

	Convey("convertMessageFromGoogle should skip function call with marshal error", t, func() {
		// Use a value that cannot be marshaled to JSON (e.g., a channel)
		part := genai.FunctionCall{
			Name: "fn",
			Args: map[string]any{"bad": make(chan int)},
		}
		content := &genai.Content{
			Parts: []genai.Part{part},
			Role:  "model",
		}
		msg := convertMessageFromGoogle(content)
		So(msg, ShouldNotBeNil)
		So(msg.Contents, ShouldBeEmpty)
	})
}
