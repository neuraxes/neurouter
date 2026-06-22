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
	"testing"

	"github.com/openai/openai-go/v3/shared"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func TestConvertEffortToOpenAI(t *testing.T) {
	Convey("Test convertEffortToOpenAI", t, func() {
		cases := []struct {
			effort   v1.ReasoningEffort
			expected shared.ReasoningEffort
		}{
			{v1.ReasoningEffort_REASONING_EFFORT_NONE, shared.ReasoningEffortNone},
			{v1.ReasoningEffort_REASONING_EFFORT_MINIMAL, shared.ReasoningEffortMinimal},
			{v1.ReasoningEffort_REASONING_EFFORT_LOW, shared.ReasoningEffortLow},
			{v1.ReasoningEffort_REASONING_EFFORT_MEDIUM, shared.ReasoningEffortMedium},
			{v1.ReasoningEffort_REASONING_EFFORT_HIGH, shared.ReasoningEffortHigh},
			{v1.ReasoningEffort_REASONING_EFFORT_EXTRA_HIGH, shared.ReasoningEffortXhigh},
			{v1.ReasoningEffort_REASONING_EFFORT_MAX, shared.ReasoningEffortXhigh},
			{v1.ReasoningEffort_REASONING_EFFORT_UNSPECIFIED, shared.ReasoningEffort("")},
		}

		for _, c := range cases {
			So(convertEffortToOpenAI(c.effort), ShouldEqual, c.expected)
		}
	})
}

func TestConvertImageToOpenAIURL(t *testing.T) {
	Convey("Test convertImageToOpenAIURL", t, func() {
		Convey("with nil image", func() {
			So(convertImageToOpenAIURL(nil), ShouldEqual, "")
		})

		Convey("with URL source", func() {
			image := &v1.Image{Source: &v1.Image_Url{Url: "https://example.com/image.png"}}
			So(convertImageToOpenAIURL(image), ShouldEqual, "https://example.com/image.png")
		})

		Convey("with data source", func() {
			image := &v1.Image{MimeType: "image/png", Source: &v1.Image_Data{Data: []byte("hello")}}
			So(convertImageToOpenAIURL(image), ShouldEqual, "data:image/png;base64,aGVsbG8=")
		})

		Convey("with base64 source", func() {
			image := &v1.Image{MimeType: "image/png", Source: &v1.Image_Base64{Base64: "aGVsbG8="}}
			So(convertImageToOpenAIURL(image), ShouldEqual, "data:image/png;base64,aGVsbG8=")
		})
	})
}

func TestConvertSchemaToMap(t *testing.T) {
	Convey("Test convertSchemaToMap", t, func() {
		Convey("with object schema and properties", func() {
			schema := &v1.Schema{
				Type: v1.Schema_TYPE_OBJECT,
				Properties: map[string]*v1.Schema{
					"name": {Type: v1.Schema_TYPE_STRING},
				},
				Required: []string{"name"},
			}

			params := convertSchemaToMap(schema)
			So(params["type"], ShouldEqual, "object")
			So(params["required"], ShouldResemble, []any{"name"})
			props := params["properties"].(map[string]any)
			name := props["name"].(map[string]any)
			So(name["type"], ShouldEqual, "string")
		})

		Convey("with object schema lacking properties", func() {
			schema := &v1.Schema{Type: v1.Schema_TYPE_OBJECT}

			params := convertSchemaToMap(schema)
			So(params["properties"], ShouldResemble, map[string]any{})
		})
	})
}
