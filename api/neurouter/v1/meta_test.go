package v1

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
