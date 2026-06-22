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
	"encoding/base64"
	"fmt"
	"strings"
)

// ParseImageDataURL parses a data URL (data:<mime>;base64,<data>) without decoding the payload.
func ParseImageDataURL(url string) (encoded string, mimeType string, err error) {
	if !strings.HasPrefix(url, "data:") {
		return "", "", fmt.Errorf("image data URL must start with data:")
	}

	meta, encoded, ok := strings.Cut(url[len("data:"):], ",")
	if !ok {
		return "", "", fmt.Errorf("image data URL missing payload separator")
	}
	if !strings.HasSuffix(meta, ";base64") {
		return "", "", fmt.Errorf("image data URL must use base64 encoding")
	}

	mimeType, _, _ = strings.Cut(meta, ";")
	return encoded, mimeType, nil
}

// EncodeImageDataToURL encodes binary data and a MIME type into a data URL
// (data:<mime>;base64,<data>).
func EncodeImageDataToURL(data []byte, mimeType string) string {
	return EncodeImageBase64ToURL(base64.StdEncoding.EncodeToString(data), mimeType)
}

// EncodeImageBase64ToURL encodes base64 image data and a MIME type into a data
// URL (data:<mime>;base64,<data>) without decoding and re-encoding the payload.
func EncodeImageBase64ToURL(encoded string, mimeType string) string {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// InferImageMimeType infers the MIME type from image bytes.
func InferImageMimeType(data []byte) string {
	if len(data) >= 8 {
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 && data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
			return "image/png"
		}
	}
	if len(data) >= 3 {
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
		if data[0] == 'G' && data[1] == 'I' && data[2] == 'F' {
			return "image/gif"
		}
	}
	if len(data) >= 12 {
		if data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P' {
			return "image/webp"
		}
	}
	if len(data) >= 2 {
		if data[0] == 'B' && data[1] == 'M' {
			return "image/bmp"
		}
	}
	return "application/octet-stream"
}
