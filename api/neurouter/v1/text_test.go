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

package v1

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToolUse_GetTextualInput(t *testing.T) {
	Convey("GetTextualInput", t, func() {
		Convey("should concatenate all textual inputs in order", func() {
			toolUse := &ToolUse{
				Inputs: []*ToolUse_Input{
					{Input: &ToolUse_Input_Text{Text: "hello"}},
					{},
					{Input: &ToolUse_Input_Text{Text: " world"}},
				},
			}

			So(toolUse.GetTextualInput(), ShouldEqual, "hello world")
		})

		Convey("should return empty string for nil receiver", func() {
			var toolUse *ToolUse
			So(toolUse.GetTextualInput(), ShouldEqual, "")
		})

		Convey("should return empty string when there are no textual inputs", func() {
			toolUse := &ToolUse{}
			So(toolUse.GetTextualInput(), ShouldEqual, "")
		})
	})
}

func TestToolResult_GetTextualOutput(t *testing.T) {
	Convey("GetTextualOutput", t, func() {
		Convey("should concatenate all textual outputs in order", func() {
			toolResult := &ToolResult{
				Outputs: []*ToolResult_Output{
					{Output: &ToolResult_Output_Text{Text: "foo"}},
					{Output: &ToolResult_Output_Image{Image: &Image{MimeType: "image/png"}}},
					{Output: &ToolResult_Output_Text{Text: "bar"}},
				},
			}

			So(toolResult.GetTextualOutput(), ShouldEqual, "foobar")
		})

		Convey("should return empty string for nil receiver", func() {
			var toolResult *ToolResult
			So(toolResult.GetTextualOutput(), ShouldEqual, "")
		})

		Convey("should return empty string when there are no textual outputs", func() {
			toolResult := &ToolResult{}
			So(toolResult.GetTextualOutput(), ShouldEqual, "")
		})
	})
}
