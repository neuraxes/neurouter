package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"

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

		Convey("When Chat is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.anthropic.com/v1/messages")
				So(httpReq.Header.Get("x-api-key"), ShouldEqual, "test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				body, err := io.ReadAll(httpReq.Body)
				So(err, ShouldBeNil)

				var reqMap map[string]any
				err = json.Unmarshal(body, &reqMap)
				So(err, ShouldBeNil)

				var expectedMap map[string]any
				err = json.Unmarshal([]byte(mockMessagesRequetBody), &expectedMap)
				So(err, ShouldBeNil)

				So(reqMap, ShouldResemble, expectedMap)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockMessagesResponseBody)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				fmt.Println()
				fmt.Println(resp)
				fmt.Println(mockChatResp)
				So(proto.Equal(resp, mockChatResp), ShouldBeTrue)
			})
		})

		Convey("When the API call fails", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			_, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
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

		Convey("When ChatStream is called and the request is successful", func() {
			mockClient.DoFunc = func(httpReq *http.Request) (*http.Response, error) {
				So(httpReq.Method, ShouldEqual, http.MethodPost)
				So(httpReq.URL.String(), ShouldEqual, "https://api.anthropic.com/v1/messages")
				So(httpReq.Header.Get("x-api-key"), ShouldEqual, "test-key")
				So(httpReq.Header.Get("Content-Type"), ShouldEqual, "application/json")

				body, err := io.ReadAll(httpReq.Body)
				So(err, ShouldBeNil)

				var reqMap map[string]any
				err = json.Unmarshal(body, &reqMap)
				So(err, ShouldBeNil)

				var expectedMap map[string]any
				err = json.Unmarshal([]byte(mockMessagesStreamRequetBody), &expectedMap)
				So(err, ShouldBeNil)

				So(reqMap, ShouldResemble, expectedMap)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(mockMessagesStreamResponseBody)),
				}, nil
			}

			seq := repo.ChatStream(context.Background(), mockChatReq)

			Convey("Then it should return a sequence and no error", func() {
				So(seq, ShouldNotBeNil)

				var responses []*entity.ChatResp
				for resp, err := range seq {
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

			seq := repo.ChatStream(context.Background(), mockChatReq)

			Convey("Then it should return an error in the iterator", func() {
				So(seq, ShouldNotBeNil)

				for resp, err := range seq {
					So(resp, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "network error")
				}
			})
		})
	})
}
