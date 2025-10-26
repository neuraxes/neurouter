package v1

import "strings"

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
