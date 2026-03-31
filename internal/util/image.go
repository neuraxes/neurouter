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

// DecodeImageDataFromUrl parses a data URL (data:<mime>;base64,<data>) and
// returns the decoded bytes and MIME type.
func DecodeImageDataFromUrl(url string) (data []byte, mimeType string) {
	if !strings.HasPrefix(url, "data:") {
		return
	}

	meta, encoded, ok := strings.Cut(url[len("data:"):], ",")
	if !ok || !strings.HasSuffix(meta, ";base64") {
		return
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		data = nil
		return
	}

	mimeType, _, _ = strings.Cut(meta, ";")
	return data, mimeType
}

// EncodeImageDataToUrl encodes binary data and a MIME type into a data URL
// (data:<mime>;base64,<data>).
func EncodeImageDataToUrl(data []byte, mimeType string) string {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}
