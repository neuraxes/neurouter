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

package anthropic

import (
	"github.com/anthropics/anthropic-sdk-go"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
	"git.xdea.xyz/Turing/neurouter/internal/biz/entity"
)

// convertSystemToAnthropic converts system messages to a format that can be sent to the Anthropic API.
func (r *ChatRepo) convertSystemToAnthropic(messages []*v1.Message) []anthropic.TextBlockParam {
	var parts []anthropic.TextBlockParam
	for _, message := range messages {
		if message.Role != v1.Role_SYSTEM {
			continue
		}
		for _, content := range message.Contents {
			switch c := content.GetContent().(type) {
			case *v1.Content_Text:
				parts = append(parts, anthropic.NewTextBlock(c.Text))
			}
		}
	}
	return parts
}

// convertMessageToAnthropic converts an internal message to a message that can be sent to the Anthropic API.
func (r *ChatRepo) convertMessageToAnthropic(message *v1.Message) anthropic.MessageParam {
	var parts []anthropic.ContentBlockParamUnion
	for _, content := range message.Contents {
		switch c := content.GetContent().(type) {
		case *v1.Content_Text:
			parts = append(parts, anthropic.NewTextBlock(c.Text))
		case *v1.Content_Image_:
			// TODO: Implement image support
		}
	}
	if message.Role == v1.Role_USER || message.Role == v1.Role_SYSTEM {
		return anthropic.NewUserMessage(parts...)
	} else {
		return anthropic.NewAssistantMessage(anthropic.NewTextBlock(message.Contents[0].GetText()))
	}
}

// convertRequestToAnthropic converts an internal request to a request that can be sent to the Anthropic API.
func (r *ChatRepo) convertRequestToAnthropic(req *entity.ChatReq) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model: anthropic.F(req.Model),
	}

	if !r.config.MergeSystem {
		params.System = anthropic.F(r.convertSystemToAnthropic(req.Messages))
	}

	var messages []anthropic.MessageParam
	for _, message := range req.Messages {
		if !r.config.MergeSystem && message.Role == v1.Role_SYSTEM {
			continue
		}
		messages = append(messages, r.convertMessageToAnthropic(message))
	}
	params.Messages = anthropic.F(messages)

	return params
}
