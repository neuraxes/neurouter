package anthropic

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

// mockHTTPClient is a mock implementation of option.HTTPClient for testing.
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, errors.New("DoFunc is not set")
}

const mockChatCompletionResp = `{
    "id": "14bfe852-1643-486f-a6fb-1e64b253986a",
    "type": "message",
    "role": "assistant",
    "model": "claude-3-7-sonnet-latest",
    "content": [
        {
            "type": "thinking",
            "thinking": "Hmm, the user just said \"Hello!\" - a simple and friendly greeting. No complex queries or specific requests here. \n\nSince it's a casual opening, a warm and welcoming response would be appropriate. Can mirror their cheerful tone while keeping it concise. \n\nThe response should acknowledge the greeting, express enthusiasm about helping, and leave the conversation open-ended for them to continue. No need for lengthy explanations or assumptions about their needs at this stage.",
            "signature": "14bfe852-1643-486f-a6fb-1e64b253986a"
        },
        {
            "type": "text",
            "text": "Hello! 😊 How are you doing today? Is there anything I can help you with or would you like to chat about something in particular?"
        }
    ],
    "stop_reason": "end_turn",
    "usage": {
        "input_tokens": 6,
        "cache_creation_input_tokens": 0,
        "cache_read_input_tokens": 1,
        "output_tokens": 120,
        "service_tier": "standard"
    }
}`

const mockChatCompletionStreamResp = `event: message_start
data: {"type":"message_start","message":{"id":"887bb5af-ed3d-4980-8dc8-2a457f57ba39","type":"message","role":"assistant","model":"claude-3-7-sonnet-latest","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":6,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":0,"service_tier":"standard"}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":""}}

event: ping
data: {"type":"ping"}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"This is a simple greeting, so no complex analysis is needed."}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":" \n\n"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":" I should respond in a warm and welcoming manner to match their energy"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"."}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"887bb5af-ed3d-4980-8dc8-2a457f57ba39"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Hello there! 👋 It"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"'s"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":" great to see"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":" you"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"!"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":" 😊"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":167}}

event: message_stop
data: {"type":"message_stop"}

`

func TestNewAnthropicUpstream(t *testing.T) {
	Convey("Given a configuration and logger", t, func() {
		config := &conf.AnthropicConfig{
			BaseUrl: "https://api.anthropic.com/",
			ApiKey:  "test-key",
		}

		Convey("When newAnthropicUpstream is called", func() {
			repo, err := newAnthropicUpstream(config, log.DefaultLogger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				upstream, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(upstream.config.BaseUrl, ShouldEqual, "https://api.anthropic.com/")
				So(upstream.config.ApiKey, ShouldEqual, "test-key")
				So(upstream.client, ShouldNotBeNil)
			})
		})

		Convey("When newAnthropicUpstreamWithClient is called", func() {
			mockClient := &mockHTTPClient{}
			repo, err := newAnthropicUpstreamWithClient(config, mockClient, log.DefaultLogger)

			Convey("Then it should return a new upstream with the custom client", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				upstream, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(upstream.config.BaseUrl, ShouldEqual, "https://api.anthropic.com/")
				So(upstream.config.ApiKey, ShouldEqual, "test-key")
				So(upstream.client, ShouldNotBeNil)
			})
		})
	})
}

func TestChat(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client", t, func() {
		config := &conf.AnthropicConfig{
			BaseUrl: "https://api.anthropic.com/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newAnthropicUpstreamWithClient(config, mockClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-req-id",
			Model: "claude-3-opus-20240229",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello"}}}},
			},
		}

		Convey("When Chat is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.anthropic.com/v1/messages")
				So(httpReq.Header.Get("x-api-key"), ShouldEqual, "test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionResp)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-req-id")
				So(resp.Model, ShouldEqual, "claude-3-opus-20240229")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Id, ShouldHaveLength, 36)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(resp.Message.Contents, ShouldHaveLength, 2)
				So(resp.Message.Contents[0].GetThinking(), ShouldEqual, "Hmm, the user just said \"Hello!\" - a simple and friendly greeting. No complex queries or specific requests here. \n\nSince it's a casual opening, a warm and welcoming response would be appropriate. Can mirror their cheerful tone while keeping it concise. \n\nThe response should acknowledge the greeting, express enthusiasm about helping, and leave the conversation open-ended for them to continue. No need for lengthy explanations or assumptions about their needs at this stage.")
				So(resp.Message.Contents[1].GetText(), ShouldEqual, "Hello! 😊 How are you doing today? Is there anything I can help you with or would you like to chat about something in particular?")
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 6)
				So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 120)
				So(resp.Statistics.Usage.CachedPromptTokens, ShouldEqual, 1)
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			_, err := repo.Chat(context.Background(), req)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}

var mockChatStreamResp = []*entity.ChatResp{
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Thinking{Thinking: "This is a simple greeting, so no complex analysis is needed."},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Thinking{Thinking: " \n\n"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Thinking{Thinking: " I should respond in a warm and welcoming manner to match their energy"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Thinking{Thinking: "."},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "Hello there! 👋 It"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "'s"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " great to see"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " you"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: "!"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Message: &v1.Message{
			Id:   "887bb5af-ed3d-4980-8dc8-2a457f57ba39",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{{
				Content: &v1.Content_Text{Text: " 😊"},
			}},
		},
	},
	{
		Model: "claude-3-7-sonnet-latest",
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     6,
				CompletionTokens: 167,
			},
		},
	},
}

func TestChatStream(t *testing.T) {
	Convey("Given a upstream with a mock HTTP client for streaming", t, func() {
		config := &conf.AnthropicConfig{
			BaseUrl: "https://api.anthropic.com/",
			ApiKey:  "test-key",
		}
		mockClient := &mockHTTPClient{}
		repo, err := newAnthropicUpstreamWithClient(config, mockClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-stream-req-id",
			Model: "claude-3-opus-20240229",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello"}}}},
			},
		}

		Convey("When ChatStream is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.anthropic.com/v1/messages")
				So(httpReq.Header.Get("x-api-key"), ShouldEqual, "test-key")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatCompletionStreamResp)),
				}, nil
			}

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)
					responses = append(responses, resp)
				}

				So(len(responses), ShouldEqual, len(mockChatStreamResp))

				for i, mockResp := range responses {
					So(proto.Equal(mockResp.Message, mockChatStreamResp[i].Message), ShouldBeTrue)
				}
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return an error on Recv", func() {
				So(err, ShouldBeNil)
				resp, err := streamClient.Recv()
				So(resp, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}
