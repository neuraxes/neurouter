package openai

import (
	"testing"

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
