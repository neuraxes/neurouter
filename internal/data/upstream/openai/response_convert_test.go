package openai

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/openai/openai-go/v3/responses"
	. "github.com/smartystreets/goconvey/convey"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

func newTestUpstream() *upstream {
	return &upstream{
		config: &conf.OpenAIConfig{},
		log:    log.NewHelper(log.DefaultLogger),
	}
}

func TestConvertStatisticsFromOpenAIResponse(t *testing.T) {
	Convey("Test convertStatisticsFromOpenAIResponse", t, func() {
		Convey("with nil usage", func() {
			result := convertStatisticsFromOpenAIResponse(nil)
			So(result, ShouldBeNil)
		})

		Convey("with valid usage statistics", func() {
			usage := &responses.ResponseUsage{
				InputTokens:  10,
				OutputTokens: 20,
				InputTokensDetails: responses.ResponseUsageInputTokensDetails{
					CachedTokens: 5,
				},
				OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{
					ReasoningTokens: 12,
				},
			}

			result := convertStatisticsFromOpenAIResponse(usage)

			So(result, ShouldNotBeNil)
			So(result.Usage, ShouldNotBeNil)
			So(result.Usage.InputTokens, ShouldEqual, 10)
			So(result.Usage.OutputTokens, ShouldEqual, 20)
			So(result.Usage.CachedInputTokens, ShouldEqual, 5)
			So(result.Usage.ReasoningTokens, ShouldEqual, 12)
		})

		Convey("with reasoning tokens only", func() {
			usage := &responses.ResponseUsage{
				OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{
					ReasoningTokens: 3,
				},
			}

			result := convertStatisticsFromOpenAIResponse(usage)

			So(result, ShouldNotBeNil)
			So(result.Usage, ShouldNotBeNil)
			So(result.Usage.ReasoningTokens, ShouldEqual, 3)
		})

		Convey("with all zero token counts", func() {
			usage := &responses.ResponseUsage{}

			result := convertStatisticsFromOpenAIResponse(usage)

			So(result, ShouldBeNil)
		})
	})
}

// TestConvertResponseFromOpenAIResponseRefusal exercises the refusal path of
// the non-streaming Responses API ingest converter.
func TestConvertResponseFromOpenAIResponseRefusal(t *testing.T) {
	Convey("Given a non-streaming Responses API result with a refusal", t, func() {
		r := newTestUpstream()
		raw := `{
			"id": "resp_refusal_1",
			"model": "gpt-4o",
			"object": "response",
			"status": "completed",
			"output": [{
				"type": "message",
				"id": "msg_refused",
				"role": "assistant",
				"content": [
					{"type": "refusal", "refusal": "I can't help with that."}
				]
			}]
		}`
		var openAIResp responses.Response
		So(openAIResp.UnmarshalJSON([]byte(raw)), ShouldBeNil)

		resp := r.convertResponseFromOpenAIResponse(&openAIResp)

		Convey("Then status flips to CHAT_REFUSED and metadata flags the refusal", func() {
			So(resp.Status, ShouldEqual, v1.ChatStatus_CHAT_REFUSED)
			So(resp.Message, ShouldNotBeNil)
			So(resp.Message.Contents, ShouldHaveLength, 1)
			c := resp.Message.Contents[0]
			So(c.GetText(), ShouldEqual, "I can't help with that.")
			So(c.Meta("refusal"), ShouldEqual, "true")
			So(c.Id, ShouldEqual, "msg_refused")
		})
	})
}

// TestConvertResponseFromOpenAIResponseReasoningMerge verifies that encrypted
// content, multiple summaries, and reasoning text from a single reasoning
// output item are merged into the v1.Content layout the rest of the system
// expects (one Content per summary entry, with reasoning text occupying the
// last empty slot).
func TestConvertResponseFromOpenAIResponseReasoningMerge(t *testing.T) {
	Convey("Given a Responses API result with rich reasoning", t, func() {
		r := newTestUpstream()
		raw := `{
			"id": "resp_reasoning_1",
			"model": "o3",
			"object": "response",
			"status": "completed",
			"output": [{
				"type": "reasoning",
				"id": "rs_1",
				"encrypted_content": "opaque-bytes",
				"summary": [
					{"type": "summary_text", "text": "first"},
					{"type": "summary_text", "text": "second"}
				],
				"content": [
					{"type": "reasoning_text", "text": "internal trace"}
				]
			}]
		}`
		var openAIResp responses.Response
		So(openAIResp.UnmarshalJSON([]byte(raw)), ShouldBeNil)

		resp := r.convertResponseFromOpenAIResponse(&openAIResp)

		Convey("Then layout matches the documented merge protocol", func() {
			So(resp.Message, ShouldNotBeNil)
			So(resp.Message.Contents, ShouldHaveLength, 2)

			first := resp.Message.Contents[0]
			So(first.Reasoning, ShouldBeTrue)
			So(first.Meta("encrypted"), ShouldEqual, "opaque-bytes")
			So(first.Meta("summary"), ShouldEqual, "first")
			So(first.Meta("summary_index"), ShouldEqual, "0")
			So(first.GetText(), ShouldEqual, "")

			second := resp.Message.Contents[1]
			So(second.Reasoning, ShouldBeTrue)
			So(second.Meta("encrypted"), ShouldBeEmpty)
			So(second.Meta("summary"), ShouldEqual, "second")
			So(second.Meta("summary_index"), ShouldEqual, "1")
			So(second.GetText(), ShouldEqual, "internal trace")
		})
	})
}

// TestConvertStreamEventRefusalDelta makes sure the streaming converter tags
// refusal deltas with the refusal=true metadata so downstream consumers (e.g.
// the Responses server) can branch on it without having to wait for the
// terminal status chunk.
func TestConvertStreamEventRefusalDelta(t *testing.T) {
	Convey("Given a refusal delta SSE event", t, func() {
		c := &openAIResponseStreamClient{
			req:       &entity.ChatReq{Id: "req_1"},
			respModel: "gpt-4o",
			messageID: "msg_stream_1",
		}
		raw := `{
			"type": "response.refusal.delta",
			"item_id": "msg_refused_stream",
			"output_index": 0,
			"delta": "I cannot ",
			"sequence_number": 5
		}`
		var event responses.ResponseStreamEventUnion
		So(event.UnmarshalJSON([]byte(raw)), ShouldBeNil)

		resp := c.convertStreamEventFromOpenAIResponse(&event)

		Convey("Then the chunk carries the delta text and refusal metadata", func() {
			So(c.hasRefused, ShouldBeTrue)
			So(resp, ShouldNotBeNil)
			So(resp.Message, ShouldNotBeNil)
			So(resp.Message.Contents, ShouldHaveLength, 1)
			content := resp.Message.Contents[0]
			So(content.Id, ShouldEqual, "msg_refused_stream")
			So(content.GetText(), ShouldEqual, "I cannot ")
			So(content.Meta("refusal"), ShouldEqual, "true")
		})
	})
}

// TestConvertStreamEventReasoningSummary covers the summary delta path -
// summary_index has to be propagated as metadata so the server can group
// multiple deltas into a single part on the wire.
func TestConvertStreamEventReasoningSummary(t *testing.T) {
	Convey("Given a reasoning_summary_text.delta event", t, func() {
		c := &openAIResponseStreamClient{
			req:       &entity.ChatReq{Id: "req_2"},
			respModel: "o3",
			messageID: "msg_stream_2",
		}
		raw := `{
			"type": "response.reasoning_summary_text.delta",
			"item_id": "rs_stream_1",
			"output_index": 0,
			"summary_index": 1,
			"delta": "thinking deeper",
			"sequence_number": 7
		}`
		var event responses.ResponseStreamEventUnion
		So(event.UnmarshalJSON([]byte(raw)), ShouldBeNil)

		resp := c.convertStreamEventFromOpenAIResponse(&event)

		Convey("Then the chunk carries summary text plus indexed metadata", func() {
			So(resp, ShouldNotBeNil)
			So(resp.Message.Contents, ShouldHaveLength, 1)
			content := resp.Message.Contents[0]
			So(content.Reasoning, ShouldBeTrue)
			So(content.Id, ShouldEqual, "rs_stream_1")
			So(content.Meta("summary"), ShouldEqual, "thinking deeper")
			So(content.Meta("summary_index"), ShouldEqual, "1")
			So(content.GetText(), ShouldEqual, "")
		})
	})
}

// TestConvertStreamEventReasoningEncryptedDone covers the case where the
// encrypted reasoning blob arrives with the output_item.done event - we need
// to expose it as an `encrypted` metadata blob on a synthesised content so the
// server can echo it back in the next request.
func TestConvertStreamEventReasoningEncryptedDone(t *testing.T) {
	Convey("Given an output_item.done for a reasoning item with encrypted blob", t, func() {
		c := &openAIResponseStreamClient{
			req:       &entity.ChatReq{Id: "req_3"},
			respModel: "o3",
			messageID: "msg_stream_3",
		}
		raw := `{
			"type": "response.output_item.done",
			"output_index": 0,
			"sequence_number": 9,
			"item": {
				"type": "reasoning",
				"id": "rs_done_1",
				"encrypted_content": "opaque-stream"
			}
		}`
		var event responses.ResponseStreamEventUnion
		So(event.UnmarshalJSON([]byte(raw)), ShouldBeNil)

		resp := c.convertStreamEventFromOpenAIResponse(&event)

		Convey("Then encrypted content is surfaced as metadata on a reasoning chunk", func() {
			So(resp, ShouldNotBeNil)
			So(resp.Message.Contents, ShouldHaveLength, 1)
			content := resp.Message.Contents[0]
			So(content.Reasoning, ShouldBeTrue)
			So(content.Id, ShouldEqual, "rs_done_1")
			So(content.Meta("encrypted"), ShouldEqual, "opaque-stream")
		})
	})
}

// TestConvertMessageToOpenAIResponseInputRefusal locks in the outgoing
// converter's behaviour: a model message tagged with refusal=true must
// round-trip back as `OfRefusal` so subsequent prompts retain the original
// refusal semantic instead of leaking it as plain assistant text.
func TestConvertMessageToOpenAIResponseInputRefusal(t *testing.T) {
	Convey("Given a model message carrying a refusal-flagged content", t, func() {
		r := newTestUpstream()
		refusal := &v1.Content{
			Id:      "msg_history_refusal",
			Content: &v1.Content_Text{Text: "I cannot help."},
		}
		refusal.SetMeta("refusal", "true")
		msg := &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{refusal},
		}

		items := r.convertMessageToOpenAIResponseInput(msg)

		Convey("Then the produced item is an output_message containing OfRefusal", func() {
			So(items, ShouldHaveLength, 1)
			outMsg := items[0].OfOutputMessage
			So(outMsg, ShouldNotBeNil)
			So(outMsg.ID, ShouldEqual, "msg_history_refusal")
			So(outMsg.Content, ShouldHaveLength, 1)
			So(outMsg.Content[0].OfRefusal, ShouldNotBeNil)
			So(outMsg.Content[0].OfRefusal.Refusal, ShouldEqual, "I cannot help.")
			So(outMsg.Content[0].OfOutputText, ShouldBeNil)
		})
	})
}

// TestConvertMessageToOpenAIResponseInputReasoning verifies that a model
// message carrying reasoning content (with encrypted blob and summary
// metadata) is rebuilt into a single ResponseReasoningItemParam, mirroring
// what the upstream originally produced.
func TestConvertMessageToOpenAIResponseInputReasoning(t *testing.T) {
	Convey("Given a model message with reasoning chunks", t, func() {
		r := newTestUpstream()
		first := &v1.Content{
			Id:        "rs_history_1",
			Reasoning: true,
			Content:   &v1.Content_Text{},
		}
		first.SetMeta("encrypted", "opaque")
		first.SetMeta("summary", "first sum")
		first.SetMeta("summary_index", "0")
		second := &v1.Content{
			Id:        "rs_history_1",
			Reasoning: true,
			Content:   &v1.Content_Text{Text: "long thought"},
		}
		second.SetMeta("summary", "second sum")
		second.SetMeta("summary_index", "1")
		msg := &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{first, second},
		}

		items := r.convertMessageToOpenAIResponseInput(msg)

		Convey("Then the chunks are consolidated into a single reasoning item", func() {
			So(items, ShouldHaveLength, 1)
			rs := items[0].OfReasoning
			So(rs, ShouldNotBeNil)
			So(rs.ID, ShouldEqual, "rs_history_1")
			So(rs.EncryptedContent.Value, ShouldEqual, "opaque")
			So(rs.Summary, ShouldHaveLength, 2)
			So(rs.Summary[0].Text, ShouldEqual, "first sum")
			So(rs.Summary[1].Text, ShouldEqual, "second sum")
			So(rs.Content, ShouldHaveLength, 1)
			So(rs.Content[0].Text, ShouldEqual, "long thought")
		})
	})
}

// TestConvertMessageToOpenAIResponseInputMixedTurn checks ordering when a
// single MODEL message contains reasoning, function call, and final text in
// that order - each one should round-trip into its dedicated item.
func TestConvertMessageToOpenAIResponseInputMixedTurn(t *testing.T) {
	Convey("Given a MODEL message with reasoning + tool use + plain text", t, func() {
		r := newTestUpstream()
		reasoning := &v1.Content{
			Id:        "rs_mix",
			Reasoning: true,
			Content:   &v1.Content_Text{Text: "deep think"},
		}
		toolUse := &v1.Content{
			Id: "fc_mix",
			Content: &v1.Content_ToolUse{ToolUse: &v1.ToolUse{
				Id:   "call_mix",
				Name: "lookup",
				Inputs: []*v1.ToolUse_Input{{
					Input: &v1.ToolUse_Input_Text{Text: `{"q":"x"}`},
				}},
			}},
		}
		text := &v1.Content{
			Id:      "msg_mix",
			Content: &v1.Content_Text{Text: "final answer"},
		}
		msg := &v1.Message{
			Role:     v1.Role_MODEL,
			Contents: []*v1.Content{reasoning, toolUse, text},
		}

		items := r.convertMessageToOpenAIResponseInput(msg)

		Convey("Then we get reasoning, function_call, output_message in order", func() {
			So(items, ShouldHaveLength, 3)
			So(items[0].OfReasoning, ShouldNotBeNil)
			So(items[0].OfReasoning.ID, ShouldEqual, "rs_mix")
			So(items[1].OfFunctionCall, ShouldNotBeNil)
			So(items[1].OfFunctionCall.CallID, ShouldEqual, "call_mix")
			So(items[1].OfFunctionCall.ID.Value, ShouldEqual, "fc_mix")
			So(items[2].OfOutputMessage, ShouldNotBeNil)
			So(items[2].OfOutputMessage.ID, ShouldEqual, "msg_mix")
			So(items[2].OfOutputMessage.Content[0].OfOutputText.Text, ShouldEqual, "final answer")
		})
	})
}
