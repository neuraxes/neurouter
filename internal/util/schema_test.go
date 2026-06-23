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

package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testSchema struct {
	Type       string                `json:"type"`
	Properties map[string]testSchema `json:"properties,omitempty"`
	Required   []string              `json:"required,omitempty"`
}

func TestStructFromAny(t *testing.T) {
	Convey("StructFromAny", t, func() {
		Convey("should convert a JSON-marshaled struct", func() {
			schema, err := StructFromAny(testSchema{
				Type: "object",
				Properties: map[string]testSchema{
					"city": {Type: "string"},
				},
				Required: []string{"city"},
			})

			So(err, ShouldBeNil)
			So(schema.AsMap()["type"], ShouldEqual, "object")
			So(schema.AsMap()["properties"], ShouldContainKey, "city")
			So(schema.AsMap()["required"], ShouldResemble, []any{"city"})
		})

		Convey("should reject top-level non-object values", func() {
			schema, err := StructFromAny([]string{"city"})

			So(err, ShouldNotBeNil)
			So(schema, ShouldBeNil)
		})
	})
}

func TestStructFromMap(t *testing.T) {
	Convey("StructFromMap", t, func() {
		Convey("should return nil for nil input", func() {
			schema, err := StructFromMap(nil)

			So(err, ShouldBeNil)
			So(schema, ShouldBeNil)
		})

		Convey("should convert a schema map", func() {
			schema, err := StructFromMap(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"city": map[string]any{"type": "string"},
				},
			})

			So(err, ShouldBeNil)
			So(schema.AsMap()["type"], ShouldEqual, "object")
			So(schema.AsMap()["properties"], ShouldContainKey, "city")
		})
	})
}
