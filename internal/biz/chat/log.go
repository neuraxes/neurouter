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

package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

const chatPrettyPrintTmpl = `
{{- if .Request}}
--- CHAT REQUEST ---
ID: {{.Request.Id}}
Model: {{.Request.Model}}
{{- if .Request.Config}}
Configuration:
{{- if .Request.Config.MaxTokens}}
 • Max Tokens: {{.Request.Config.MaxTokens}}
{{- end}}
{{- if .Request.Config.Temperature}}
 • Temperature: {{printf "%.2f" .Request.Config.GetTemperature}}
{{- end}}
{{- if .Request.Config.TopP}}
 • Top P: {{printf "%.2f" .Request.Config.GetTopP}}
{{- end}}
{{- if .Request.Config.TopK}}
 • Top K: {{.Request.Config.GetTopK}}
{{- end}}
{{- if .Request.Config.FrequencyPenalty}}
 • Frequency Penalty: {{printf "%.2f" .Request.Config.GetFrequencyPenalty}}
{{- end}}
{{- if .Request.Config.PresencePenalty}}
 • Presence Penalty: {{printf "%.2f" .Request.Config.GetPresencePenalty}}
{{- end}}
{{- if .Request.Config.GetPresetTemplate}}
 • Preset Template: {{.Request.Config.GetPresetTemplate}}
{{- end}}
{{- if .Request.Config.GetPresetGrammar}}
 • Preset Grammar: {{.Request.Config.GetPresetGrammar}}
{{- end}}
{{- if .Request.Config.GetGbnfGrammar}}
 • GBNF Grammar: {{.Request.Config.GetGbnfGrammar}}
{{- end}}
{{- if .Request.Config.GetJsonSchema}}
 • JSON Schema: {{.Request.Config.GetJsonSchema}}
{{- end}}
{{- end}}
Messages ({{len .Request.Messages}}):
{{- range $i, $msg := .Request.Messages}}
Message {{$i}}:
 • ID: {{$msg.Id}}
 • Role: {{$msg.Role}}
 • Name: {{$msg.Name}}
 • ToolCallID: {{$msg.ToolCallId}}
 • Contents ({{len $msg.Contents}}):
{{- range $j, $content := $msg.Contents}}
 • Content {{$j}}:
{{formatContent $content}}
{{- end}}
{{- end}}
{{- if .Request.Tools}}
Tools ({{len .Request.Tools}}):
{{- range $i, $tool := .Request.Tools}}
 • Tool {{$i}}:
{{- if $tool.GetFunction}}
  • Function: {{$tool.GetFunction.Name}}
  • Description: {{$tool.GetFunction.Description}}
  • Parameters: {{formatSchema $tool.GetFunction.Parameters}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}

{{- if .Response}}
--- CHAT RESPONSE ---
ID: {{.Response.Id}}
Model: {{.Response.Model}}
Message:
 • ID: {{.Response.Message.Id}}
 • Role: {{.Response.Message.Role}}
 • Contents ({{len .Response.Message.Contents}}):
{{- range $i, $content := .Response.Message.Contents}}
 • Content {{$i}}:
{{formatContent $content}}
{{- end}}
{{- if .Response.Statistics}}
{{- if .Response.Statistics.Usage}}
Statistics:
 • Prompt Tokens: {{.Response.Statistics.Usage.PromptTokens}}
 • Completion Tokens: {{.Response.Statistics.Usage.CompletionTokens}}
 • Cached Prompt Tokens: {{.Response.Statistics.Usage.CachedPromptTokens}}
{{- end}}
{{- end}}
{{- end}}
`

var (
	chatPrettyPrintCompiledTmpl *template.Template
)

func init() {
	chatPrettyPrintCompiledTmpl = template.Must(
		template.New("chat").Funcs(template.FuncMap{
			"formatContent": formatContent,
			"formatSchema":  formatSchema,
		}).Parse(chatPrettyPrintTmpl),
	)
}

func formatContent(content *v1.Content) string {
	switch c := content.Content.(type) {
	case *v1.Content_Text:
		return "<TEXT>\n" + c.Text + "\n</TEXT>"
	case *v1.Content_Image:
		switch src := c.Image.Source.(type) {
		case *v1.Image_Url:
			return "<IMAGE_URL>" + src.Url + "</IMAGE_URL>"
		case *v1.Image_Data:
			return "<IMAGE_DATA>" + fmt.Sprintf("%d bytes", len(src.Data)) + "</IMAGE_DATA>"
		}
	case *v1.Content_Thinking:
		return "<THINKING>\n" + c.Thinking + "\n</THINKING>"
	case *v1.Content_FunctionCall:
		return fmt.Sprintf(
			`<FN_CALL>
ID: %s
Name: %s
Args: %s
</FN_CALL>`,
			c.FunctionCall.Id,
			c.FunctionCall.Name,
			c.FunctionCall.Arguments,
		)
	}

	return "<UNKNOWN>"
}

func formatSchema(schema *v1.Schema) string {
	j, _ := json.MarshalIndent(schema, "    ", "  ")
	return string(j)
}

func (uc *chatUseCase) printChat(req *v1.ChatReq, resp *v1.ChatResp) error {
	data := struct {
		Request  *v1.ChatReq
		Response *v1.ChatResp
	}{
		Request:  req,
		Response: resp,
	}

	var buf bytes.Buffer
	if err := chatPrettyPrintCompiledTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render chat pretty print template: %w", err)
	}

	uc.log.Debug(buf.String())
	return nil
}
