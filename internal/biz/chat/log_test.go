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

package chat

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/utils/ptr"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestPrintChat(t *testing.T) {
	Convey("Test PrintChat", t, func() {
		uc := &chatUseCase{log: log.NewHelper(log.DefaultLogger)}
		req := &v1.ChatReq{
			Id:    "req-id",
			Model: "gemini-2.5-pro",
			Config: &v1.GenerationConfig{
				MaxTokens:        ptr.To[int64](128),
				Temperature:      ptr.To[float32](0.7),
				TopP:             ptr.To[float32](0.1),
				TopK:             ptr.To[int64](40),
				FrequencyPenalty: ptr.To[float32](0.1),
				PresencePenalty:  ptr.To[float32](0.2),
				Template: &v1.GenerationConfig_PresetTemplate{
					PresetTemplate: "example_tmpl",
				},
				Grammar: &v1.GenerationConfig_PresetGrammar{
					PresetGrammar: "json_object",
				},
			},
			Messages: []*v1.Message{
				{
					Id:   "msg-1",
					Role: v1.Role_USER,
					Contents: []*v1.Content{
						{
							Reasoning: true,
							Content: &v1.Content_Text{
								Text: "Reasoning...",
							},
						},
						{
							Content: &v1.Content_Text{
								Text: "Hello,\nWorld!",
							},
						},
						{
							Content: &v1.Content_Image{
								Image: &v1.Image{
									Source: &v1.Image_Url{
										Url: "http://example.com",
									},
								},
							},
						},
						{
							Content: &v1.Content_Image{
								Image: &v1.Image{
									Source: &v1.Image_Data{
										Data: []byte("image data"),
									},
								},
							},
						},
					},
				},
			},
			Tools: []*v1.Tool{
				{
					Tool: &v1.Tool_Function_{
						Function: &v1.Tool_Function{
							Name:        "get_weather",
							Description: "Get the weather",
							Parameters: &v1.Schema{
								Properties: map[string]*v1.Schema{
									"location": &v1.Schema{
										Type: v1.Schema_TYPE_STRING,
									},
								},
							},
						},
					},
				},
			},
		}
		resp := &v1.ChatResp{
			Id:    "resp-id",
			Model: "gemini-2.5-pro",
			Message: &v1.Message{
				Id:   "msg-2",
				Role: v1.Role_MODEL,
				Contents: []*v1.Content{
					{
						Content: &v1.Content_Text{
							Text: "Hello,\nWorld!",
						},
					},
					{
						Content: &v1.Content_ToolUse{
							ToolUse: &v1.ToolUse{
								Id:   "fn-call-1",
								Name: "get_weather",
								Inputs: []*v1.ToolUse_Input{
									{
										Index: ptr.To[uint32](0),
										Input: &v1.ToolUse_Input_Text{
											Text: `{"location": "San Francisco"}`,
										},
									},
								},
							},
						},
					},
				},
			},
			Statistics: &v1.Statistics{
				Usage: &v1.Statistics_Usage{
					InputTokens: 10,
				},
			},
		}
		err := uc.printChat(req, resp)
		So(err, ShouldBeNil)
	})
}
