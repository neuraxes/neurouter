package chat

import (
	"reflect"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

type ChatRespAccumulator struct {
	resp *v1.ChatResp
}

func NewChatRespAccumulator() *ChatRespAccumulator {
	return &ChatRespAccumulator{
		resp: &v1.ChatResp{},
	}
}

func (a *ChatRespAccumulator) Resp() *v1.ChatResp {
	return a.resp
}

func (a *ChatRespAccumulator) Accumulate(resp *v1.ChatResp) {
	if resp == nil {
		return
	}

	a.resp.Id = resp.Id
	a.resp.Model = resp.Model

	// Accumulate message content
	a.accumulateMessage(resp.Message)

	// Accumulate statistics
	a.accumulateStatistics(resp.Statistics)
}

func (a *ChatRespAccumulator) lastContent() *v1.Content {
	contents := a.resp.GetMessage().GetContents()
	contentLen := len(contents)
	if contentLen == 0 {
		return nil
	}
	return contents[contentLen-1]
}

func (a *ChatRespAccumulator) accumulateMessage(message *v1.Message) {
	if message == nil {
		return
	}

	if a.resp.Message == nil {
		a.resp.Message = &v1.Message{}
	}

	a.resp.Message.Id = message.Id
	a.resp.Message.Role = message.Role
	a.resp.Message.Name += message.Name

	// Accumulate contents
	for _, content := range message.Contents {
		lastContent := a.lastContent()
		if lastContent == nil || reflect.TypeOf(lastContent.Content) != reflect.TypeOf(content.Content) || lastContent.Reasoning != content.Reasoning {
			a.resp.Message.Contents = append(a.resp.Message.Contents, content)
			continue
		}

		switch c := content.Content.(type) {
		case *v1.Content_Text:
			lastText := lastContent.Content.(*v1.Content_Text)
			lastText.Text += c.Text

		case *v1.Content_ToolUse:
			if c.ToolUse.Id != "" {
				// A new function call, append as new content
				a.resp.Message.Contents = append(a.resp.Message.Contents, content)
				continue
			}
			lastFunctionCall := lastContent.Content.(*v1.Content_ToolUse)
			lastFunctionCall.ToolUse.Id += c.ToolUse.Id
			lastFunctionCall.ToolUse.Name += c.ToolUse.Name
			if len(c.ToolUse.Inputs) > 0 {
				if len(lastFunctionCall.ToolUse.Inputs) > 0 {
					lastFunctionCall.ToolUse.Inputs[0].Input.(*v1.ToolUse_Input_Text).Text += c.ToolUse.GetTextualInput()
				} else {
					lastFunctionCall.ToolUse.Inputs = append(lastFunctionCall.ToolUse.Inputs, c.ToolUse.Inputs...)
				}
			}
		}
	}
}

func (a *ChatRespAccumulator) accumulateStatistics(statistics *v1.Statistics) {
	if statistics == nil {
		return
	}

	if a.resp.Statistics == nil {
		a.resp.Statistics = &v1.Statistics{}
	}

	if statistics.Usage != nil {
		if a.resp.Statistics.Usage == nil {
			a.resp.Statistics.Usage = &v1.Statistics_Usage{}
		}

		a.resp.Statistics.Usage.InputTokens = statistics.Usage.InputTokens
		a.resp.Statistics.Usage.OutputTokens += statistics.Usage.OutputTokens
	}
}
