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

func TestChatRequest_Meta(t *testing.T) {
	Convey("ChatRequest Meta", t, func() {
		Convey("should return metadata value when key exists", func() {
			req := &ChatRequest{Metadata: map[string]string{"trace_id": "abc123"}}
			So(req.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should return empty string when key does not exist", func() {
			req := &ChatRequest{Metadata: map[string]string{"trace_id": "abc123"}}
			So(req.Meta("request_id"), ShouldEqual, "")
		})

		Convey("should return empty string for nil receiver", func() {
			var req *ChatRequest
			So(req.Meta("trace_id"), ShouldEqual, "")
		})
	})
}

func TestChatRequest_SetMeta(t *testing.T) {
	Convey("ChatRequest SetMeta", t, func() {
		Convey("should initialize metadata map when nil", func() {
			req := &ChatRequest{}
			req.SetMeta("trace_id", "abc123")

			So(req.Metadata, ShouldNotBeNil)
			So(req.Meta("trace_id"), ShouldEqual, "abc123")
		})

		Convey("should overwrite existing metadata value", func() {
			req := &ChatRequest{Metadata: map[string]string{"trace_id": "old"}}
			req.SetMeta("trace_id", "new")

			So(req.Meta("trace_id"), ShouldEqual, "new")
		})

		Convey("should do nothing for nil receiver", func() {
			var req *ChatRequest
			req.SetMeta("trace_id", "abc123")
		})
	})
}
