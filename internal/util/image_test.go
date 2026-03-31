// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

func TestDecodeImageDataFromUrl(t *testing.T) {
	Convey("DecodeImageDataFromUrl", t, func() {
		Convey("should decode a valid base64 data URL", func() {
			url := "data:image/png;base64,iVBORw0KGgo="
			data, mimeType := DecodeImageDataFromUrl(url)
			So(data, ShouldNotBeNil)
			So(mimeType, ShouldEqual, "image/png")
		})

		Convey("should decode a data URL with jpeg mime type", func() {
			url := "data:image/jpeg;base64,/9j/4AAQ"
			data, mimeType := DecodeImageDataFromUrl(url)
			So(data, ShouldNotBeNil)
			So(mimeType, ShouldEqual, "image/jpeg")
		})

		Convey("should return nil for a regular HTTP URL", func() {
			data, mimeType := DecodeImageDataFromUrl("https://example.com/image.png")
			So(data, ShouldBeNil)
			So(mimeType, ShouldBeEmpty)
		})

		Convey("should return nil for an empty string", func() {
			data, mimeType := DecodeImageDataFromUrl("")
			So(data, ShouldBeNil)
			So(mimeType, ShouldBeEmpty)
		})

		Convey("should return nil for data URL without base64 encoding", func() {
			data, mimeType := DecodeImageDataFromUrl("data:text/plain,hello")
			So(data, ShouldBeNil)
			So(mimeType, ShouldBeEmpty)
		})

		Convey("should return nil for invalid base64 data", func() {
			data, mimeType := DecodeImageDataFromUrl("data:image/png;base64,not-valid-base64!!!")
			So(data, ShouldBeNil)
			So(mimeType, ShouldBeEmpty)
		})

		Convey("should handle data URL with empty mime type", func() {
			url := "data:;base64,aGVsbG8="
			data, mimeType := DecodeImageDataFromUrl(url)
			So(data, ShouldNotBeNil)
			So(string(data), ShouldEqual, "hello")
			So(mimeType, ShouldBeEmpty)
		})
	})
}

func TestEncodeImageDataToUrl(t *testing.T) {
	Convey("EncodeImageDataToUrl", t, func() {
		Convey("should encode data with mime type", func() {
			result := EncodeImageDataToUrl([]byte("hello"), "image/png")
			So(result, ShouldEqual, "data:image/png;base64,aGVsbG8=")
		})

		Convey("should use application/octet-stream for empty mime type", func() {
			result := EncodeImageDataToUrl([]byte("hello"), "")
			So(result, ShouldEqual, "data:application/octet-stream;base64,aGVsbG8=")
		})
	})
}
