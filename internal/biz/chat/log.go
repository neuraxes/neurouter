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
	"strings"
	"text/template"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

const chatPrettyPrintTmpl = `
{{- if .Request}}
<chat_request id="{{.Request.Id}}" model="{{.Request.Model}}">
{{- if .Request.Config}}
  <generation_config>
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
  </generation_config>
{{- end}}
  <messages len={{len .Request.Messages}}>
{{- range $i, $msg := .Request.Messages}}
    <message index={{$i}} id="{{$msg.Id}}" role="{{$msg.Role}}" name="{{$msg.Name}}">
      <contents len={{len $msg.Contents}}>
{{- range $j, $content := $msg.Contents}}
        <content index={{$j}} reasoning="{{ $content.Reasoning }}">
{{formatContent $content}}
        </content>
{{- end}}
      </contents>
    </message>
{{- end}}
  </messages>
{{- if .Request.Tools}}
  <tool_declarations len={{len .Request.Tools}}>
{{- range $i, $tool := .Request.Tools}}
    <tool_declaration index={{$i}}>
{{- if $tool.GetFunction}}
  • Function: {{$tool.GetFunction.Name}}
  • Description: {{$tool.GetFunction.Description}}
  • Parameters: {{formatSchema $tool.GetFunction.Parameters}}
{{- end}}
    </tool_declaration>
{{- end}}
  </tool_declarations>
{{- end}}
{{- end}}
</chat_request>
{{- if .Response}}
<chat_response id="{{.Response.Id}}" model="{{.Response.Model}}">
  <message id="{{.Response.Message.Id}}" role="{{.Response.Message.Role}}">
    <contents len={{len .Response.Message.Contents}}>
{{- range $i, $content := .Response.Message.Contents}}
      <content index={{$i}} reasoning="{{ $content.Reasoning }}">
{{formatContent $content}}
      </content>
{{- end}}
    </contents>
  </message>
{{- if .Response.Statistics}}
  <statistics>
{{- if .Response.Statistics.Usage}}
   • Input Tokens: {{.Response.Statistics.Usage.InputTokens}}
   • Output Tokens: {{.Response.Statistics.Usage.OutputTokens}}
   • Cached Input Tokens: {{.Response.Statistics.Usage.CachedInputTokens}}
{{- end}}
  </statistics>
{{- end}}
</chat_response>
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
	sb := strings.Builder{}

	switch c := content.Content.(type) {
	case *v1.Content_Text:
		sb.WriteString("<content_text>\n")
		sb.WriteString(c.Text)
		sb.WriteString("\n</content_text>")
	case *v1.Content_Image:
		switch src := c.Image.Source.(type) {
		case *v1.Image_Url:
			sb.WriteString("<content_image_url>")
			sb.WriteString(src.Url)
			sb.WriteString("</content_image_url>")
		case *v1.Image_Data:
			sb.WriteString("<content_image_data>")
			sb.WriteString(fmt.Sprintf("%d bytes", len(src.Data)))
			sb.WriteString("</content_image_data>")
		}
	case *v1.Content_ToolUse:
		sb.WriteString(fmt.Sprintf(`<content_tool_use id="%s" name="%s">`, c.ToolUse.Id, c.ToolUse.Name))
		sb.WriteString("\n")
		sb.WriteString(c.ToolUse.GetTextualInput())
		sb.WriteString("\n</content_tool_use>")
	case *v1.Content_ToolResult:
		sb.WriteString(fmt.Sprintf(`<content_tool_result id="%s">`, c.ToolResult.Id))
		sb.WriteString("\n")
		sb.WriteString(c.ToolResult.GetTextualOutput())
		sb.WriteString("\n</content_tool_result>")
	}

	return sb.String()
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
