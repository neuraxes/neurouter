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

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
)

func TestRemoveReasoningContents(t *testing.T) {
	Convey("Given opaque reasoning state and plain reasoning items in a request", t, func() {
		reasoning := func(id, text string) *v1.Content {
			return &v1.Content{Id: id, Phase: v1.ContentPhase_CONTENT_PHASE_REASONING, Content: v1.NewTextContent(text)}
		}
		opaqueReasoning := func(id string) *v1.Content {
			return &v1.Content{Id: id, Phase: v1.ContentPhase_CONTENT_PHASE_REASONING, Content: &v1.Content_Opaque{Opaque: "invalid"}}
		}
		plain := reasoning("rs-plain", "Remove this summary")
		anonymousPlain := reasoning("", "Remove this anonymous summary")
		text := &v1.Content{Content: v1.NewTextContent("Keep the answer")}
		tool := &v1.Content{Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{Id: "call-1", Name: "get_weather"}}}
		req := &entity.ChatRequest{Model: "gpt-5", Messages: []*v1.Message{
			nil,
			{
				Role: v1.Role_ROLE_MODEL,
				Contents: []*v1.Content{
					nil,
					reasoning("rs-bad", "Remove this summary"),
					plain,
					opaqueReasoning("rs-bad"),
					text,
					reasoning("", "Remove this anonymous summary"),
					opaqueReasoning(""),
					text,
					anonymousPlain,
					tool,
				},
			},
			{
				Role:     v1.Role_ROLE_MODEL,
				Contents: []*v1.Content{reasoning("rs-bad", "Remove another message's summary")},
			},
		}}
		expected := proto.Clone(req).(*entity.ChatRequest)
		expected.Messages[1].Contents = []*v1.Content{nil, text, text, tool}
		expected.Messages[2].Contents = nil

		Convey("When all reasoning contents are removed from the proto request", func() {
			So(removeReasoningContents(req), ShouldBeTrue)
			So(proto.Equal(req, expected), ShouldBeTrue)
			So(removeReasoningContents(req), ShouldBeFalse)
		})
	})
}
