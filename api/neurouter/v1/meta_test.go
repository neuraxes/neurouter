package v1

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestContent_Meta(t *testing.T) {
	Convey("Content Meta", t, func() {
		Convey("should return metadata value when key exists", func() {
			content := &Content{Metadata: map[string]string{"trace_id": "abc123"}}
			So(content.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should return empty string when key does not exist", func() {
			content := &Content{Metadata: map[string]string{"trace_id": "abc123"}}
			So(content.Meta("request_id"), ShouldEqual, "")
		})

		Convey("should return empty string for nil receiver", func() {
			var content *Content
			So(content.Meta("trace_id"), ShouldEqual, "")
		})
	})
}

func TestContent_SetMeta(t *testing.T) {
	Convey("Content SetMeta", t, func() {
		Convey("should initialize metadata map when nil", func() {
			content := &Content{}
			content.SetMeta("trace_id", "abc123")

			So(content.Metadata, ShouldNotBeNil)
			So(content.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should overwrite existing metadata value", func() {
			content := &Content{Metadata: map[string]string{"trace_id": "old"}}
			content.SetMeta("trace_id", "new")

			So(content.Meta("trace_id"), ShouldEqual, "new")
		})

		Convey("should do nothing for nil receiver", func() {
			var content *Content
			content.SetMeta("trace_id", "abc123")
		})
	})
}

func TestMessage_Meta(t *testing.T) {
	Convey("Message Meta", t, func() {
		Convey("should return metadata value when key exists", func() {
			message := &Message{Metadata: map[string]string{"trace_id": "abc123"}}
			So(message.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should return empty string when key does not exist", func() {
			message := &Message{Metadata: map[string]string{"trace_id": "abc123"}}
			So(message.Meta("request_id"), ShouldEqual, "")
		})

		Convey("should return empty string for nil receiver", func() {
			var message *Message
			So(message.Meta("trace_id"), ShouldEqual, "")
		})
	})
}

func TestMessage_SetMeta(t *testing.T) {
	Convey("Message SetMeta", t, func() {
		Convey("should initialize metadata map when nil", func() {
			message := &Message{}
			message.SetMeta("trace_id", "abc123")

			So(message.Metadata, ShouldNotBeNil)
			So(message.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should overwrite existing metadata value", func() {
			message := &Message{Metadata: map[string]string{"trace_id": "old"}}
			message.SetMeta("trace_id", "new")

			So(message.Meta("trace_id"), ShouldEqual, "new")
		})

		Convey("should do nothing for nil receiver", func() {
			var message *Message
			message.SetMeta("trace_id", "abc123")
		})
	})
}

func TestChatReq_Meta(t *testing.T) {
	Convey("ChatReq Meta", t, func() {
		Convey("should return metadata value when key exists", func() {
			req := &ChatReq{Metadata: map[string]string{"trace_id": "abc123"}}
			So(req.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should return empty string when key does not exist", func() {
			req := &ChatReq{Metadata: map[string]string{"trace_id": "abc123"}}
			So(req.Meta("request_id"), ShouldEqual, "")
		})

		Convey("should return empty string for nil receiver", func() {
			var req *ChatReq
			So(req.Meta("trace_id"), ShouldEqual, "")
		})
	})
}

func TestChatReq_SetMeta(t *testing.T) {
	Convey("ChatReq SetMeta", t, func() {
		Convey("should initialize metadata map when nil", func() {
			req := &ChatReq{}
			req.SetMeta("trace_id", "abc123")

			So(req.Metadata, ShouldNotBeNil)
			So(req.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should overwrite existing metadata value", func() {
			req := &ChatReq{Metadata: map[string]string{"trace_id": "old"}}
			req.SetMeta("trace_id", "new")

			So(req.Meta("trace_id"), ShouldEqual, "new")
		})

		Convey("should do nothing for nil receiver", func() {
			var req *ChatReq
			req.SetMeta("trace_id", "abc123")
		})
	})
}
