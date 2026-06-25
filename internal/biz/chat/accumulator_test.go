package chat

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func newTestChatEvent(event v1.ChatEventPayload) *v1.ChatEvent {
	return v1.NewChatEvent("", event)
}

func reduceTestChatEvent(r *ChatEventReducer, event v1.ChatEventPayload) {
	r.Reduce(newTestChatEvent(event))
}

func newTestChatEventReducer() *ChatEventReducer {
	return NewChatEventReducer(log.NewHelper(log.DefaultLogger))
}

func TestChatEventReducer(t *testing.T) {
	Convey("Test ChatEventReducer", t, func() {
		Convey("NewChatEventReducer returns empty resp", func() {
			r := newTestChatEventReducer()
			resp := r.Resp()
			So(resp, ShouldNotBeNil)
			So(proto.Equal(resp, &v1.ChatResp{}), ShouldBeTrue)
		})

		Convey("Reduce nil event is no-op", func() {
			r := newTestChatEventReducer()
			r.Reduce(nil)
			So(proto.Equal(r.Resp(), &v1.ChatResp{}), ShouldBeTrue)
		})

		Convey("MessageStart sets id, model and message identity", func() {
			r := newTestChatEventReducer()
			start := newTestChatEvent(v1.NewMessageStartEvent("msg-1", "model-a"))
			start.Id = "session-1"
			r.Reduce(start)

			resp := r.Resp()
			So(resp.Id, ShouldEqual, "session-1")
			So(resp.Model, ShouldEqual, "model-a")
			So(resp.Message.Id, ShouldEqual, "msg-1")
			So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
		})

		Convey("MessageStop sets status", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED))
			So(r.Resp().Status, ShouldEqual, v1.ChatStatus_CHAT_COMPLETED)
		})

		Convey("Text content block start and deltas merge", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_NORMAL))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, "Hello"))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, ", World!"))
			reduceTestChatEvent(r, v1.NewContentStopEvent(0))

			resp := r.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 1)
			So(resp.Message.Contents[0].GetText().GetText(), ShouldEqual, "Hello, World!")
			So(resp.Message.Contents[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_NORMAL)
		})

		Convey("Content start metadata is preserved", func() {
			r := newTestChatEventReducer()
			start := v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_NORMAL)
			start.ContentStart.Metadata = map[string]string{"provider_item_id": "item-1"}
			reduceTestChatEvent(r, start)

			resp := r.Resp()
			So(resp.Message.Contents[0].Metadata["provider_item_id"], ShouldEqual, "item-1")
		})

		Convey("Reasoning block with text and signature deltas", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_REASONING))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, "think-"))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, "ing"))
			reduceTestChatEvent(r, v1.NewContentDeltaSignatureEvent(0, "sig-"))
			reduceTestChatEvent(r, v1.NewContentDeltaSignatureEvent(0, "nature"))
			reduceTestChatEvent(r, v1.NewContentStopEvent(0))

			resp := r.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 1)
			So(resp.Message.Contents[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_REASONING)
			So(resp.Message.Contents[0].GetText().GetText(), ShouldEqual, "think-ing")
			So(resp.Message.Contents[0].Signature, ShouldEqual, "sig-nature")
		})

		Convey("Distinct indices produce distinct content blocks", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_REASONING))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, "thinking"))
			reduceTestChatEvent(r, v1.NewContentStopEvent(0))
			reduceTestChatEvent(r, v1.NewContentStartTextEvent(1, v1.ContentPhase_CONTENT_PHASE_NORMAL))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(1, "answer"))
			reduceTestChatEvent(r, v1.NewContentStopEvent(1))

			resp := r.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 2)
			So(resp.Message.Contents[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_REASONING)
			So(resp.Message.Contents[0].GetText().GetText(), ShouldEqual, "thinking")
			So(resp.Message.Contents[1].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_NORMAL)
			So(resp.Message.Contents[1].GetText().GetText(), ShouldEqual, "answer")
		})

		Convey("Tool use block with streamed input", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewContentStartToolUseEvent(0, "call-1", "get_weather"))
			reduceTestChatEvent(r, v1.NewContentDeltaToolInputTextEvent(0, `{"loc`))
			reduceTestChatEvent(r, v1.NewContentDeltaToolInputTextEvent(0, `ation":"NYC"}`))
			reduceTestChatEvent(r, v1.NewContentStopEvent(0))

			resp := r.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 1)
			toolUse := resp.Message.Contents[0].GetToolUse()
			So(toolUse, ShouldNotBeNil)
			So(toolUse.Id, ShouldEqual, "call-1")
			So(toolUse.Name, ShouldEqual, "get_weather")
			So(toolUse.GetTextualInput(), ShouldEqual, `{"location":"NYC"}`)
		})

		Convey("Multiple tool uses with different indices", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewContentStartToolUseEvent(0, "call-1", "get_weather"))
			reduceTestChatEvent(r, v1.NewContentDeltaToolInputTextEvent(0, `{"city":"NYC"}`))
			reduceTestChatEvent(r, v1.NewContentStopEvent(0))
			reduceTestChatEvent(r, v1.NewContentStartToolUseEvent(1, "call-2", "get_time"))
			reduceTestChatEvent(r, v1.NewContentDeltaToolInputTextEvent(1, `{"tz":"EST"}`))
			reduceTestChatEvent(r, v1.NewContentStopEvent(1))

			resp := r.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 2)
			So(resp.Message.Contents[0].GetToolUse().Id, ShouldEqual, "call-1")
			So(resp.Message.Contents[0].GetToolUse().GetTextualInput(), ShouldEqual, `{"city":"NYC"}`)
			So(resp.Message.Contents[1].GetToolUse().Id, ShouldEqual, "call-2")
			So(resp.Message.Contents[1].GetToolUse().GetTextualInput(), ShouldEqual, `{"tz":"EST"}`)
		})

		Convey("Content snapshot is appended verbatim", func() {
			r := newTestChatEventReducer()
			snapshot := &v1.Content{
				Index:   new(uint32(0)),
				Phase:   v1.ContentPhase_CONTENT_PHASE_REASONING,
				Content: &v1.Content_Opaque{Opaque: "encrypted"},
			}
			reduceTestChatEvent(r, v1.NewContentSnapshotEvent(snapshot))

			resp := r.Resp()
			So(len(resp.Message.Contents), ShouldEqual, 1)
			So(resp.Message.Contents[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_REASONING)
			So(resp.Message.Contents[0].GetOpaque(), ShouldEqual, "encrypted")
		})

		Convey("Usage updates by non-zero fields across events", func() {
			r := newTestChatEventReducer()
			r.Reduce(&v1.ChatEvent{Usage: &v1.Usage{InputTokens: 100, OutputTokens: 200, CachedInputTokens: 300, ReasoningTokens: 50}})
			r.Reduce(&v1.ChatEvent{Usage: &v1.Usage{OutputTokens: 300, ReasoningTokens: 70}})
			r.Reduce(&v1.ChatEvent{Usage: &v1.Usage{InputTokens: 200, CachedInputTokens: 400}})

			usage := r.Resp().Statistics.Usage
			So(usage.InputTokens, ShouldEqual, 200)
			So(usage.OutputTokens, ShouldEqual, 300)
			So(usage.CachedInputTokens, ShouldEqual, 400)
			So(usage.ReasoningTokens, ShouldEqual, 70)
		})

		Convey("Delta without a preceding start is ignored", func() {
			r := newTestChatEventReducer()
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, "orphan"))

			resp := r.Resp()
			So(resp.Message, ShouldBeNil)
		})

		Convey("Full turn produces a complete resp", func() {
			r := newTestChatEventReducer()
			start := newTestChatEvent(v1.NewMessageStartEvent("msg-1", "model-a"))
			start.Id = "session-1"
			r.Reduce(start)
			reduceTestChatEvent(r, v1.NewContentStartTextEvent(0, v1.ContentPhase_CONTENT_PHASE_NORMAL))
			reduceTestChatEvent(r, v1.NewContentDeltaTextEvent(0, "Hi"))
			reduceTestChatEvent(r, v1.NewContentStopEvent(0))
			stop := newTestChatEvent(v1.NewMessageStopEvent(v1.ChatStatus_CHAT_COMPLETED))
			stop.Usage = &v1.Usage{InputTokens: 10, OutputTokens: 5}
			r.Reduce(stop)

			resp := r.Resp()
			So(resp.Id, ShouldEqual, "session-1")
			So(resp.Model, ShouldEqual, "model-a")
			So(resp.Status, ShouldEqual, v1.ChatStatus_CHAT_COMPLETED)
			So(resp.Message.Id, ShouldEqual, "msg-1")
			So(resp.Message.Contents[0].GetText().GetText(), ShouldEqual, "Hi")
			So(resp.Statistics.Usage.InputTokens, ShouldEqual, 10)
			So(resp.Statistics.Usage.OutputTokens, ShouldEqual, 5)
		})
	})
}
