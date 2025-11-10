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

			streamClient, err := repo.ChatStream(context.Background(), mockChatReq)

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

				So(responses, ShouldHaveLength, len(mockStreamChatResp))

				for i, resp := range responses {
					if !proto.Equal(resp, mockStreamChatResp[i]) {
						fmt.Println("\n", resp.String(), "\n", mockStreamChatResp[i].String())
					}
					So(proto.Equal(resp, mockStreamChatResp[i]), ShouldBeTrue)
				}
			})
		})
	})
}
