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
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// preferStringContent is a middleware that modifies the request to prefer string content over structured content.
func (r *ChatRepo) preferStringContent(req *http.Request, next option.MiddlewareNext) (resp *http.Response, err error) {
	if !strings.HasSuffix(req.URL.Path, "/chat/completions") {
		return next(req)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}
	err = req.Body.Close()
	if err != nil {
		return
	}

	for i, msg := range gjson.GetBytes(body, "messages").Array() {
		role := msg.Get("role").String()
		if (role == string(openai.ChatCompletionSystemMessageParamRoleSystem) && !r.config.PreferStringContentForSystem) ||
			(role == string(openai.ChatCompletionMessageParamRoleUser) && !r.config.PreferStringContentForUser) ||
			(role == string(openai.ChatCompletionMessageParamRoleAssistant) && !r.config.PreferStringContentForAssistant) ||
			(role == string(openai.ChatCompletionMessageParamRoleTool) && !r.config.PreferStringContentForTool) {
			continue
		}
		content := msg.Get("content")
		if content.Exists() && content.IsArray() {
			parts := content.Array()
			if len(parts) == 1 && parts[0].Get("type").String() == "text" {
				body, err = sjson.SetBytes(body, "messages."+strconv.Itoa(i)+".content", parts[0].Get("text").String())
				if err != nil {
					return
				}
			}
		}
	}

	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	return next(req)
}
