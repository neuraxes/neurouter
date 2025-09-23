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
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaType_MarshalJSON(t *testing.T) {
	Convey("MarshalJSON", t, func() {
		Convey("should marshal valid schema types to lowercase strings", func() {
			testCases := []struct {
				schemaType Schema_Type
				expected   string
			}{
				{Schema_TYPE_UNSPECIFIED, `"unspecified"`},
				{Schema_TYPE_STRING, `"string"`},
				{Schema_TYPE_NUMBER, `"number"`},
				{Schema_TYPE_INTEGER, `"integer"`},
				{Schema_TYPE_BOOLEAN, `"boolean"`},
				{Schema_TYPE_OBJECT, `"object"`},
				{Schema_TYPE_ARRAY, `"array"`},
			}

			for _, tc := range testCases {
				data, err := tc.schemaType.MarshalJSON()
				So(err, ShouldBeNil)
				So(string(data), ShouldEqual, tc.expected)
			}
		})

		Convey("should return error for invalid schema type", func() {
			invalidType := Schema_Type(999) // Invalid value
			_, err := invalidType.MarshalJSON()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid schema type")
		})
	})
}

func TestSchemaType_UnmarshalJSON(t *testing.T) {
	Convey("UnmarshalJSON", t, func() {
		Convey("should unmarshal valid lowercase strings to schema types", func() {
			testCases := []struct {
				jsonStr  string
				expected Schema_Type
			}{
				{`"unspecified"`, Schema_TYPE_UNSPECIFIED},
				{`"string"`, Schema_TYPE_STRING},
				{`"number"`, Schema_TYPE_NUMBER},
				{`"integer"`, Schema_TYPE_INTEGER},
				{`"boolean"`, Schema_TYPE_BOOLEAN},
				{`"object"`, Schema_TYPE_OBJECT},
				{`"array"`, Schema_TYPE_ARRAY},
			}

			for _, tc := range testCases {
				var st Schema_Type
				err := json.Unmarshal([]byte(tc.jsonStr), &st)
				So(err, ShouldBeNil)
				So(st, ShouldEqual, tc.expected)
			}
		})

		Convey("should return error for invalid string", func() {
			var st Schema_Type
			err := json.Unmarshal([]byte(`"invalid"`), &st)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid schema type")
		})

		Convey("should return error for non-string input", func() {
			var st Schema_Type
			err := json.Unmarshal([]byte(`123`), &st)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "schema type should be a string")
		})
	})
}
