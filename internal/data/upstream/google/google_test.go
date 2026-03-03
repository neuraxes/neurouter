package google

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

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/biz/entity"
	"github.com/neuraxes/neurouter/internal/conf"
)

type mockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return nil, errors.New("RoundTripFunc is not set")
}

func TestNewGoogleUpstream(t *testing.T) {
	Convey("Given a configuration and logger", t, func() {
		config := &conf.GoogleConfig{
			ApiKey: "test-api-key",
		}

		Convey("When newGoogleUpstream is called", func() {
			repo, err := newGoogleUpstream(config, log.DefaultLogger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
			})
		})

		Convey("When newGoogleUpstreamWithClient is called with a custom HTTP client", func() {
			mockClient := &http.Client{Transport: &mockRoundTripper{}}
			repo, err := newGoogleUpstreamWithClient(config, mockClient, log.DefaultLogger)

			Convey("Then it should return a new upstream and no error", func() {
				So(err, ShouldBeNil)
				So(repo, ShouldNotBeNil)
				u, ok := repo.(*upstream)
				So(ok, ShouldBeTrue)
				So(u.config.ApiKey, ShouldEqual, "test-api-key")
				So(u.client, ShouldNotBeNil)
			})
		})
	})
}

func TestChat(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper", t, func() {
		mockRoundTripper := &mockRoundTripper{}
		repo, err := newGoogleUpstreamWithClient(
			&conf.GoogleConfig{
				ApiKey: "test-api-key",
			},
			&http.Client{Transport: mockRoundTripper},
			log.DefaultLogger,
		)
		So(err, ShouldBeNil)

		Convey("When Chat is called and the request is successful", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				So(req.URL.Path, ShouldContainSubstring, "/gemini-2.5-flash:generateContent")
				So(req.Header.Get("X-Goog-Api-Key"), ShouldEqual, "test-api-key")

				body, err := io.ReadAll(req.Body)
				So(err, ShouldBeNil)

				var reqMap map[string]any
				err = json.Unmarshal(body, &reqMap)
				So(err, ShouldBeNil)

				var expectedMap map[string]any
				err = json.Unmarshal([]byte(mockGenerateContentRequestBody), &expectedMap)
				So(err, ShouldBeNil)

				So(reqMap, ShouldResemble, expectedMap)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockGenerateContentResponseBody)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(proto.Equal(resp, mockChatResp), ShouldBeTrue)
			})
		})

		Convey("When the API call fails with a network error", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			_, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})

		Convey("When the API returns empty candidates", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"candidates": [], "modelVersion": "gemini-2.5-flash"}`)),
				}, nil
			}

			_, err := repo.Chat(context.Background(), mockChatReq)

			Convey("Then it should return an error about no candidates", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "no candidates")
			})
		})
	})
}

func TestChatStream(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper for streaming", t, func() {
		mockRoundTripper := &mockRoundTripper{}
		repo, err := newGoogleUpstreamWithClient(
			&conf.GoogleConfig{
				ApiKey: "test-api-key",
			},
			&http.Client{Transport: mockRoundTripper},
			log.DefaultLogger,
		)
		So(err, ShouldBeNil)

		Convey("When ChatStream is called and the request is successful", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				So(req.URL.Path, ShouldContainSubstring, "/gemini-2.5-flash:streamGenerateContent")
				So(req.Header.Get("X-Goog-Api-Key"), ShouldEqual, "test-api-key")

				body, err := io.ReadAll(req.Body)
				So(err, ShouldBeNil)

				var reqMap map[string]any
				err = json.Unmarshal(body, &reqMap)
				So(err, ShouldBeNil)

				var expectedMap map[string]any
				err = json.Unmarshal([]byte(mockGenerateContentRequestBody), &expectedMap)
				So(err, ShouldBeNil)

				So(reqMap, ShouldResemble, expectedMap)
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(mockStreamGenerateContentResponseBody)),
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

				So(responses, ShouldHaveLength, len(mockStreamChatResp))

				for i, resp := range responses {
					if !proto.Equal(resp, mockStreamChatResp[i]) {
						fmt.Println("\n", resp.String(), "\n", mockStreamChatResp[i].String())
					}
					So(proto.Equal(resp, mockStreamChatResp[i]), ShouldBeTrue)
				}
			})
		})

		Convey("When the API call fails with a network error", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			seq := repo.ChatStream(context.Background(), mockChatReq)

			Convey("Then the iterator should yield an error", func() {
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

func TestEmbed(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper", t, func() {
		mockRT := &mockRoundTripper{}
		repo, err := newGoogleUpstreamWithClient(
			&conf.GoogleConfig{ApiKey: "test-api-key"},
			&http.Client{Transport: mockRT},
			log.DefaultLogger,
		)
		So(err, ShouldBeNil)

		u, ok := repo.(*upstream)
		So(ok, ShouldBeTrue)

		Convey("When Embed is called and the request is successful", func() {
			mockRT.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				So(req.URL.Path, ShouldContainSubstring, "/text-embedding-004")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(strings.NewReader(`{
						"embeddings": [{"values": [0.1, 0.2, 0.3]}]
					}`)),
				}, nil
			}

			embedReq := &entity.EmbedReq{
				Id:    "embed-1",
				Model: "text-embedding-004",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "hello world"}},
				},
			}

			resp, err := u.Embed(context.Background(), embedReq)

			Convey("Then it should return a valid embedding response", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "embed-1")
				So(resp.Embedding, ShouldHaveLength, 3)
				So(resp.Embedding[0], ShouldAlmostEqual, 0.1, 0.001)
				So(resp.Embedding[1], ShouldAlmostEqual, 0.2, 0.001)
				So(resp.Embedding[2], ShouldAlmostEqual, 0.3, 0.001)
			})
		})

		Convey("When Embed fails with a network error", func() {
			mockRT.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			}

			embedReq := &entity.EmbedReq{
				Id:    "embed-1",
				Model: "text-embedding-004",
				Contents: []*v1.Content{
					{Content: &v1.Content_Text{Text: "hello world"}},
				},
			}

			_, err := u.Embed(context.Background(), embedReq)

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})
		})
	})
}
