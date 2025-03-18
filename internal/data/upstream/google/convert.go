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

package google

import (
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func convertContentToGoogle(content *v1.Content) genai.Part {
	switch v := content.Content.(type) {
	case *v1.Content_Text:
		return genai.Text(v.Text)
	case *v1.Content_Image_:
		//   TODO: Handle image content when supported
		return nil
	default:
		return nil
	}
}

func convertMessageToGoogle(msg *v1.Message) *genai.Content {
	var parts []genai.Part
	for _, content := range msg.Contents {
		if part := convertContentToGoogle(content); part != nil {
			parts = append(parts, part)
		}
	}

	role := ""
	switch msg.Role {
	case v1.Role_SYSTEM:
		role = "user" // Google AI doesn't support system role, use user role instead
	case v1.Role_USER:
		role = "user"
	case v1.Role_MODEL:
		role = "model"
	case v1.Role_TOOL:
		role = "user" // Google AI doesn't support tool role, use user role instead
	}

	return &genai.Content{
		Parts: parts,
		Role:  role,
	}
}

func convertMessageFromGoogle(content *genai.Content) *v1.Message {
	message := &v1.Message{
		Id:   uuid.NewString(),
		Role: v1.Role_MODEL,
	}

	for _, part := range content.Parts {
		if text, ok := part.(genai.Text); ok {
			message.Contents = append(message.Contents, &v1.Content{
				Content: &v1.Content_Text{
					Text: string(text),
				},
			})
		}
	}

	return message
}
