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
	"encoding/json"

	"github.com/openai/openai-go/v3/responses"
)

type responsesResponse struct {
	ID                string                               `json:"id"`
	Object            string                               `json:"object"`
	CreatedAt         int64                                `json:"created_at"`
	Model             string                               `json:"model"`
	Status            string                               `json:"status"`
	Output            []*responsesOutputItem               `json:"output"`
	Error             *responses.ResponseError             `json:"error"`
	IncompleteDetails *responses.ResponseIncompleteDetails `json:"incomplete_details"`
	Usage             *responses.ResponseUsage             `json:"usage"`
}

type responsesOutputItem struct {
	ID               string                                   `json:"id"`
	Type             string                                   `json:"type"`
	Status           string                                   `json:"status"`
	Role             string                                   `json:"role,omitempty"`
	Phase            string                                   `json:"phase,omitempty"`
	Content          []responses.ResponseOutputText           `json:"content,omitempty"`
	Summary          []responses.ResponseReasoningItemSummary `json:"summary,omitempty"`
	EncryptedContent string                                   `json:"encrypted_content,omitempty"`
	CallID           string                                   `json:"call_id,omitempty"`
	Name             string                                   `json:"name,omitempty"`
	Arguments        *string                                  `json:"arguments,omitempty"`
}

// MarshalJSON keeps the (empty) arrays that Responses clients iterate over.
func (item *responsesOutputItem) MarshalJSON() ([]byte, error) {
	if item == nil {
		return []byte("null"), nil
	}

	switch item.Type {
	case "message":
		return json.Marshal(struct {
			ID      string                         `json:"id"`
			Type    string                         `json:"type"`
			Status  string                         `json:"status"`
			Role    string                         `json:"role,omitempty"`
			Phase   string                         `json:"phase,omitempty"`
			Content []responses.ResponseOutputText `json:"content"`
		}{
			ID:      item.ID,
			Type:    item.Type,
			Status:  item.Status,
			Role:    item.Role,
			Phase:   item.Phase,
			Content: responsesArray(item.Content),
		})
	case "reasoning":
		return json.Marshal(struct {
			ID               string                                   `json:"id"`
			Type             string                                   `json:"type"`
			Status           string                                   `json:"status"`
			Summary          []responses.ResponseReasoningItemSummary `json:"summary"`
			Content          []responses.ResponseOutputText           `json:"content"`
			EncryptedContent string                                   `json:"encrypted_content,omitempty"`
		}{
			ID:               item.ID,
			Type:             item.Type,
			Status:           item.Status,
			Summary:          responsesArray(item.Summary),
			Content:          responsesArray(item.Content),
			EncryptedContent: item.EncryptedContent,
		})
	case "function_call":
		return json.Marshal(struct {
			ID        string  `json:"id"`
			Type      string  `json:"type"`
			Status    string  `json:"status"`
			CallID    string  `json:"call_id,omitempty"`
			Name      string  `json:"name,omitempty"`
			Arguments *string `json:"arguments,omitempty"`
		}{
			ID:        item.ID,
			Type:      item.Type,
			Status:    item.Status,
			CallID:    item.CallID,
			Name:      item.Name,
			Arguments: item.Arguments,
		})
	default:
		return json.Marshal(struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Status string `json:"status"`
		}{
			ID:     item.ID,
			Type:   item.Type,
			Status: item.Status,
		})
	}
}

func responsesArray[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
