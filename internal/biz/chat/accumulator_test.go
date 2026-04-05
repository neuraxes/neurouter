package chat

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestChatRespAccumulator(t *testing.T) {
	Convey("Test ChatRespAccumulator", t, func() {
		Convey("NewChatRespAccumulator returns empty resp", func() {
			acc := NewChatRespAccumulator()
			resp := acc.Resp()
			So(resp, ShouldNotBeNil)
			So(proto.Equal(resp, &v1.ChatResp{}), ShouldBeTrue)
		})

		Convey("Accumulate nil resp is no-op", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(nil)
			resp := acc.Resp()
			So(proto.Equal(resp, &v1.ChatResp{}), ShouldBeTrue)
		})

		Convey("Accumulate overwrites Id, Model, and Status", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Id:     "id-1",
				Model:  "model-a",
				Status: v1.ChatStatus_CHAT_IN_PROGRESS,
			})
			acc.Accumulate(&v1.ChatResp{
				Id:     "id-2",
				Model:  "model-b",
				Status: v1.ChatStatus_CHAT_COMPLETED,
			})
			resp := acc.Resp()
			So(resp.Id, ShouldEqual, "id-2")
			So(resp.Model, ShouldEqual, "model-b")
			So(resp.Status, ShouldEqual, v1.ChatStatus_CHAT_COMPLETED)
		})

		Convey("Accumulate text content", func() {
			Convey("Single text chunk", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Id:   "msg-1",
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Hello, World!"}},
						},
					},
				})
				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello, World!")
			})

			Convey("Multiple text chunks merge into one", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Id:   "msg-1",
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Hello"}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Id:   "msg-1",
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: ", World!"}},
						},
					},
				})
				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello, World!")
			})

			Convey("Text with different reasoning flags are separate", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Reasoning: true, Content: &v1.Content_Text{Text: "thinking"}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Reasoning: true, Content: &v1.Content_Text{Text: " more"}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "answer"}},
						},
					},
				})
				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 2)
				So(resp.Message.Contents[0].Reasoning, ShouldBeTrue)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "thinking more")
				So(resp.Message.Contents[1].Reasoning, ShouldBeFalse)
				So(resp.Message.Contents[1].GetText(), ShouldEqual, "answer")
			})

			Convey("Text with same index merges", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Index: new(uint32(0)), Content: &v1.Content_Text{Text: "part-1"}},
						},
					},
				})

				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Index: new(uint32(0)), Content: &v1.Content_Text{Text: "-part-2"}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "part-1-part-2")
			})

			Convey("Text with different indices does not merge", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Index: new(uint32(0)), Content: &v1.Content_Text{Text: "first"}},
						},
					},
				})

				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Index: new(uint32(1)), Content: &v1.Content_Text{Text: "second"}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 2)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "first")
				So(resp.Message.Contents[1].GetText(), ShouldEqual, "second")
			})

			Convey("Reasoning signature metadata is accumulated", func() {
				acc := NewChatRespAccumulator()
				idx := uint32(0)
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{
								Index:     &idx,
								Reasoning: true,
								Content:   &v1.Content_Text{Text: "think-"},
							},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{
								Index:     &idx,
								Reasoning: true,
								Metadata: map[string]string{
									"signature": "sig-",
								},
								Content: &v1.Content_Text{Text: "ing"},
							},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{
								Index:     &idx,
								Reasoning: true,
								Metadata: map[string]string{
									"signature": "nature",
								},
								Content: &v1.Content_Text{Text: "-done"},
							},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "think-ing-done")
				So(resp.Message.Contents[0].Metadata["signature"], ShouldEqual, "sig-nature")
			})

			Convey("Text metadata key is initialized when missing", func() {
				acc := NewChatRespAccumulator()
				idx := uint32(1)
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{
								Index:   &idx,
								Content: &v1.Content_Text{Text: "hello"},
							},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{
								Index: &idx,
								Metadata: map[string]string{
									"trace_id": "abc",
								},
								Content: &v1.Content_Text{Text: " world"},
							},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "hello world")
				So(resp.Message.Contents[0].Metadata["trace_id"], ShouldEqual, "abc")
			})
		})

		Convey("Accumulate tool use content", func() {
			Convey("Single tool use with streamed input", func() {
				acc := NewChatRespAccumulator()
				// First chunk: tool use header with id and name
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "get_weather",
							}}},
						},
					},
				})
				// Second chunk: input data
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Inputs: []*v1.ToolUse_Input{
									{Input: &v1.ToolUse_Input_Text{Text: `{"loc`}},
								},
							}}},
						},
					},
				})
				// Third chunk: more input data
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Inputs: []*v1.ToolUse_Input{
									{Input: &v1.ToolUse_Input_Text{Text: `ation":"NYC"}`}},
								},
							}}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)

				toolUse := resp.Message.Contents[0].GetToolUse()
				So(toolUse, ShouldNotBeNil)
				So(toolUse.Id, ShouldEqual, "call-1")
				So(toolUse.Name, ShouldEqual, "get_weather")
				So(toolUse.GetTextualInput(), ShouldEqual, `{"location":"NYC"}`)
			})

			Convey("Multiple tool uses with different ids", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "get_weather",
								Inputs: []*v1.ToolUse_Input{
									{Input: &v1.ToolUse_Input_Text{Text: `{"city":"NYC"}`}},
								},
							}}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-2",
								Name: "get_time",
								Inputs: []*v1.ToolUse_Input{
									{Input: &v1.ToolUse_Input_Text{Text: `{"tz":"EST"}`}},
								},
							}}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 2)

				So(resp.Message.Contents[0].GetToolUse().Id, ShouldEqual, "call-1")
				So(resp.Message.Contents[0].GetToolUse().Name, ShouldEqual, "get_weather")
				So(resp.Message.Contents[0].GetToolUse().GetTextualInput(), ShouldEqual, `{"city":"NYC"}`)

				So(resp.Message.Contents[1].GetToolUse().Id, ShouldEqual, "call-2")
				So(resp.Message.Contents[1].GetToolUse().Name, ShouldEqual, "get_time")
				So(resp.Message.Contents[1].GetToolUse().GetTextualInput(), ShouldEqual, `{"tz":"EST"}`)
			})

			Convey("Streamed tool use name", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "get_",
							}}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Name: "weather",
							}}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetToolUse().Name, ShouldEqual, "get_weather")
			})
		})

		Convey("Accumulate mixed content types", func() {
			Convey("Text followed by tool use", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Let me check "}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "the weather."}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "get_weather",
							}}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 2)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Let me check the weather.")
				So(resp.Message.Contents[1].GetToolUse().Id, ShouldEqual, "call-1")
			})

			Convey("Reasoning then text then tool use", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Reasoning: true, Content: &v1.Content_Text{Text: "I should check"}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Text{Text: "Checking now"}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
								Id:   "call-1",
								Name: "search",
							}}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 3)
				So(resp.Message.Contents[0].Reasoning, ShouldBeTrue)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "I should check")
				So(resp.Message.Contents[1].Reasoning, ShouldBeFalse)
				So(resp.Message.Contents[1].GetText(), ShouldEqual, "Checking now")
				So(resp.Message.Contents[2].GetToolUse().Id, ShouldEqual, "call-1")
				So(resp.Message.Contents[2].GetToolUse().Name, ShouldEqual, "search")
			})
		})

		Convey("Accumulate image content", func() {
			Convey("Consecutive images are not merged", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Image{Image: &v1.Image{
								MimeType: "image/png",
								Source:   &v1.Image_Url{Url: "http://example.com/1.png"},
							}}},
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Message: &v1.Message{
						Role: v1.Role_MODEL,
						Contents: []*v1.Content{
							{Content: &v1.Content_Image{Image: &v1.Image{
								MimeType: "image/jpeg",
								Source:   &v1.Image_Url{Url: "http://example.com/2.jpg"},
							}}},
						},
					},
				})

				resp := acc.Resp()
				So(len(resp.Message.Contents), ShouldEqual, 2)
				So(resp.Message.Contents[0].GetImage().GetUrl(), ShouldEqual, "http://example.com/1.png")
				So(resp.Message.Contents[1].GetImage().GetUrl(), ShouldEqual, "http://example.com/2.jpg")
			})
		})

		Convey("Accumulate statistics", func() {
			Convey("Single statistics chunk", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Statistics: &v1.Statistics{
						Usage: &v1.Statistics_Usage{
							InputTokens:       100,
							OutputTokens:      50,
							CachedInputTokens: 20,
							ReasoningTokens:   15,
						},
					},
				})
				resp := acc.Resp()
				So(resp.Statistics.Usage.InputTokens, ShouldEqual, 100)
				So(resp.Statistics.Usage.OutputTokens, ShouldEqual, 50)
				So(resp.Statistics.Usage.CachedInputTokens, ShouldEqual, 20)
				So(resp.Statistics.Usage.ReasoningTokens, ShouldEqual, 15)
			})

			Convey("Tokens update by non-zero fields across chunks", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Statistics: &v1.Statistics{
						Usage: &v1.Statistics_Usage{
							InputTokens:       100,
							OutputTokens:      200,
							CachedInputTokens: 300,
							ReasoningTokens:   50,
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Statistics: &v1.Statistics{
						Usage: &v1.Statistics_Usage{
							InputTokens:       0,
							OutputTokens:      300,
							CachedInputTokens: 0,
							ReasoningTokens:   70,
						},
					},
				})
				acc.Accumulate(&v1.ChatResp{
					Statistics: &v1.Statistics{
						Usage: &v1.Statistics_Usage{
							InputTokens:       200,
							OutputTokens:      0,
							CachedInputTokens: 400,
							ReasoningTokens:   0,
						},
					},
				})

				resp := acc.Resp()
				So(resp.Statistics.Usage.InputTokens, ShouldEqual, 200)
				So(resp.Statistics.Usage.OutputTokens, ShouldEqual, 300)
				So(resp.Statistics.Usage.CachedInputTokens, ShouldEqual, 400)
				So(resp.Statistics.Usage.ReasoningTokens, ShouldEqual, 70)
			})

			Convey("Nil usage is skipped", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{
					Statistics: &v1.Statistics{},
				})
				resp := acc.Resp()
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage, ShouldBeNil)
			})

			Convey("Nil statistics is skipped", func() {
				acc := NewChatRespAccumulator()
				acc.Accumulate(&v1.ChatResp{})
				resp := acc.Resp()
				So(resp.Statistics, ShouldBeNil)
			})
		})

		Convey("Accumulate nil message is skipped", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Id:    "id-1",
				Model: "model-a",
			})
			resp := acc.Resp()
			So(resp.Id, ShouldEqual, "id-1")
			So(resp.Message, ShouldBeNil)
		})

		Convey("Accumulate message with no contents", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Id:   "msg-1",
					Role: v1.Role_MODEL,
				},
			})
			resp := acc.Resp()
			So(resp.Message, ShouldNotBeNil)
			So(resp.Message.Id, ShouldEqual, "msg-1")
			So(len(resp.Message.Contents), ShouldEqual, 0)
		})

		Convey("Accumulate message with multiple contents in a single chunk", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_Text{Text: "Hello"}},
						{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
							Id:   "call-1",
							Name: "test",
						}}},
					},
				},
			})

			resp := acc.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 2)
			So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello")
			So(resp.Message.Contents[1].GetToolUse().Id, ShouldEqual, "call-1")
		})

		Convey("Message name is appended", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Name: "assist",
				},
			})
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Name: "ant",
				},
			})
			resp := acc.Resp()
			So(resp.Message.Name, ShouldEqual, "assistant")
		})

		Convey("Message id and role use latest value", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Id:   "msg-1",
					Role: v1.Role_USER,
				},
			})
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Id:   "msg-2",
					Role: v1.Role_MODEL,
				},
			})
			resp := acc.Resp()
			So(resp.Message.Id, ShouldEqual, "msg-2")
			So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
		})

		Convey("Tool use with no inputs then inputs arrive", func() {
			acc := NewChatRespAccumulator()
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
							Id:   "call-1",
							Name: "func",
						}}},
					},
				},
			})
			acc.Accumulate(&v1.ChatResp{
				Message: &v1.Message{
					Role: v1.Role_MODEL,
					Contents: []*v1.Content{
						{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
							Inputs: []*v1.ToolUse_Input{
								{Input: &v1.ToolUse_Input_Text{Text: `{"key":"val"}`}},
							},
						}}},
					},
				},
			})

			resp := acc.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 1)
			toolUse := resp.Message.Contents[0].GetToolUse()
			So(toolUse.Id, ShouldEqual, "call-1")
			So(toolUse.Name, ShouldEqual, "func")
			So(len(toolUse.Inputs), ShouldEqual, 1)
			So(toolUse.GetTextualInput(), ShouldEqual, `{"key":"val"}`)
		})
	})
}
