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
