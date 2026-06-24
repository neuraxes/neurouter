package openai

import (
	"testing"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/openai/openai-go/v3/responses"
	. "github.com/smartystreets/goconvey/convey"
)

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

func TestConvertStreamEventFromOpenAIResponse(t *testing.T) {
	Convey("Test convertStreamEventFromOpenAIResponse", t, func() {
		client := &openAIResponseStreamClient{
			req: &entity.ChatReq{Id: "chat-123"},
		}

		Convey("it should keep the output item phase on text deltas", func() {
			added := &responses.ResponseStreamEventUnion{
				Type:        "response.output_item.added",
				OutputIndex: 0,
				Item: responses.ResponseOutputItemUnion{
					ID:    "msg-123",
					Type:  "message",
					Phase: responses.ResponseOutputMessagePhaseFinalAnswer,
				},
			}
			resp := client.convertStreamEventFromOpenAIResponse(added)
			So(resp, ShouldNotBeNil)
			So(resp.Message.Contents[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_OUTCOME)

			delta := &responses.ResponseStreamEventUnion{
				Type:        "response.output_text.delta",
				OutputIndex: 0,
				Delta:       "final answer",
			}
			resp = client.convertStreamEventFromOpenAIResponse(delta)
			So(resp, ShouldNotBeNil)
			So(resp.Message.Contents[0].GetPhase(), ShouldEqual, v1.ContentPhase_CONTENT_PHASE_OUTCOME)
			So(resp.Message.Contents[0].GetText().GetText(), ShouldEqual, "final answer")
		})
	})
}
