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

import "strings"

// NewTextContent creates text content from a plain string.
func NewTextContent(text string) *Content_Text {
	return &Content_Text{Text: &Text{Text: text}}
}

// GetTextualInput collects all textual inputs from a ToolUse.
func (x *ToolUse) GetTextualInput() string {
	var sb strings.Builder
	for _, part := range x.GetInputs() {
		switch input := part.Input.(type) {
		case *ToolUse_Input_Text:
			sb.WriteString(input.Text)
		}
	}
	return sb.String()
}

// GetTextualOutput collects all textual outputs from a ToolResult.
func (x *ToolResult) GetTextualOutput() string {
	var sb strings.Builder
	for _, part := range x.GetOutputs() {
		switch output := part.Output.(type) {
		case *ToolResult_Output_Text:
			sb.WriteString(output.Text)
		}
	}
	return sb.String()
}
