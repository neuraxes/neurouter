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

func TestParseImageDataURL(t *testing.T) {
	Convey("ParseImageDataURL", t, func() {
		Convey("should parse a valid base64 data URL without decoding", func() {
			encoded, mimeType, err := ParseImageDataURL("data:image/png;base64,aGVsbG8=")
			So(err, ShouldBeNil)
			So(encoded, ShouldEqual, "aGVsbG8=")
			So(mimeType, ShouldEqual, "image/png")
		})

		Convey("should not validate base64 data", func() {
			encoded, mimeType, err := ParseImageDataURL("data:image/png;base64,not-valid-base64!!!")
			So(err, ShouldBeNil)
			So(encoded, ShouldEqual, "not-valid-base64!!!")
			So(mimeType, ShouldEqual, "image/png")
		})

		Convey("should return error for a non-data URL", func() {
			encoded, mimeType, err := ParseImageDataURL("https://example.com/image.png")
			So(err, ShouldNotBeNil)
			So(encoded, ShouldBeEmpty)
			So(mimeType, ShouldBeEmpty)
		})

		Convey("should return error for data URL without payload separator", func() {
			encoded, mimeType, err := ParseImageDataURL("data:image/png;base64")
			So(err, ShouldNotBeNil)
			So(encoded, ShouldBeEmpty)
			So(mimeType, ShouldBeEmpty)
		})

		Convey("should return error for data URL without base64 encoding", func() {
			encoded, mimeType, err := ParseImageDataURL("data:text/plain,hello")
			So(err, ShouldNotBeNil)
			So(encoded, ShouldBeEmpty)
			So(mimeType, ShouldBeEmpty)
		})
	})
}

func TestEncodeImageDataToUrl(t *testing.T) {
	Convey("EncodeImageDataToUrl", t, func() {
		Convey("should encode data with mime type", func() {
			result := EncodeImageDataToURL([]byte("hello"), "image/png")
			So(result, ShouldEqual, "data:image/png;base64,aGVsbG8=")
		})

		Convey("should use application/octet-stream for empty mime type", func() {
			result := EncodeImageDataToURL([]byte("hello"), "")
			So(result, ShouldEqual, "data:application/octet-stream;base64,aGVsbG8=")
		})
	})
}

func TestEncodeImageBase64ToURL(t *testing.T) {
	Convey("EncodeImageBase64ToURL", t, func() {
		Convey("should reuse base64 data with mime type", func() {
			result := EncodeImageBase64ToURL("aGVsbG8=", "image/png")
			So(result, ShouldEqual, "data:image/png;base64,aGVsbG8=")
		})

		Convey("should use application/octet-stream for empty mime type", func() {
			result := EncodeImageBase64ToURL("aGVsbG8=", "")
			So(result, ShouldEqual, "data:application/octet-stream;base64,aGVsbG8=")
		})
	})
}

func TestInferImageMimeType(t *testing.T) {
	Convey("InferImageMimeType", t, func() {
		Convey("should detect PNG", func() {
			data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
			So(InferImageMimeType(data), ShouldEqual, "image/png")
		})

		Convey("should detect JPEG", func() {
			data := []byte{0xFF, 0xD8, 0xFF}
			So(InferImageMimeType(data), ShouldEqual, "image/jpeg")
		})

		Convey("should detect GIF", func() {
			data := []byte{'G', 'I', 'F'}
			So(InferImageMimeType(data), ShouldEqual, "image/gif")
		})

		Convey("should detect WEBP", func() {
			data := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}
			So(InferImageMimeType(data), ShouldEqual, "image/webp")
		})

		Convey("should detect BMP", func() {
			data := []byte{'B', 'M'}
			So(InferImageMimeType(data), ShouldEqual, "image/bmp")
		})

		Convey("should return default for unknown data", func() {
			data := []byte{0x00, 0x01}
			So(InferImageMimeType(data), ShouldEqual, "application/octet-stream")
		})
	})
}
