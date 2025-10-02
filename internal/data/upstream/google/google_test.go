package google

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"
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

const mockChatResp = `[
    {
        "candidates": [
            {
                "content": {
                    "parts": [
                        {
                            "text": "Hello"
                        }
                    ],
                    "role": "model"
                }
            }
        ],
        "usageMetadata": {
            "promptTokenCount": 2,
            "totalTokenCount": 2,
            "promptTokensDetails": [
                {
                    "modality": 1,
                    "tokenCount": 2
                }
            ]
        },
        "modelVersion": "gemini-2.0-flash",
        "responseId": "bIreaOrWOZLEnvgPo8mZ8A4"
    },
    {
        "candidates": [
            {
                "content": {
                    "parts": [
                        {
                            "text": " there! How"
                        }
                    ],
                    "role": "model"
                }
            }
        ],
        "usageMetadata": {
            "promptTokenCount": 2,
            "totalTokenCount": 2,
            "promptTokensDetails": [
                {
                    "modality": 1,
                    "tokenCount": 2
                }
            ]
        },
        "modelVersion": "gemini-2.0-flash",
        "responseId": "bIreaOrWOZLEnvgPo8mZ8A4"
    },
    {
        "candidates": [
            {
                "content": {
                    "parts": [
                        {
                            "text": " can I help you today?\n"
                        }
                    ],
                    "role": "model"
                },
                "finishReason": 1
            }
        ],
        "usageMetadata": {
            "promptTokenCount": 1,
            "candidatesTokenCount": 11,
            "totalTokenCount": 12,
            "promptTokensDetails": [
                {
                    "modality": 1,
                    "tokenCount": 1
                }
            ],
            "candidatesTokensDetails": [
                {
                    "modality": 1,
                    "tokenCount": 11
                }
            ]
        },
        "modelVersion": "gemini-2.0-flash",
        "responseId": "bIreaOrWOZLEnvgPo8mZ8A4"
    }
]`

const mockChatRespWithToolCall = `[
    {
        "candidates": [
            {
                "content": {
                    "parts": [
                        {
                            "functionCall": {
                                "name": "get_weather",
                                "args": {
                                    "location": "Tokyo"
                                }
                            }
                        }
                    ],
                    "role": "model"
                },
                "finishReason": 1
            }
        ],
        "usageMetadata": {
            "promptTokenCount": 23,
            "candidatesTokenCount": 5,
            "totalTokenCount": 28,
            "promptTokensDetails": [
                {
                    "modality": 1,
                    "tokenCount": 23
                }
            ],
            "candidatesTokensDetails": [
                {
                    "modality": 1,
                    "tokenCount": 5
                }
            ]
        },
        "modelVersion": "gemini-2.0-flash",
        "responseId": "YqDeaJ6dGbqN2PgPnp-G6Aw"
    }
]`

func TestNewGoogleUpstream(t *testing.T) {
	// TODO
}

func TestChat(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper", t, func() {
		config := &conf.GoogleConfig{
			ApiKey: "test-key",
		}
		mockRoundTripper := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockRoundTripper}
		repo, err := newGoogleUpstreamWithClient(config, httpClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-req-id",
			Model: "gemini-2.0-flash",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "You are a helpful assistant."}}}},
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello!"}}}},
			},
		}

		Convey("When Chat is called and the request is successful", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				So(err, ShouldBeNil)
				bodyStr := string(body)
				So(gjson.Get(bodyStr, "model").String(), ShouldEqual, "models/gemini-2.0-flash")
				So(gjson.Get(bodyStr, "contents.#").Int(), ShouldEqual, 2)
				So(gjson.Get(bodyStr, "contents.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.0.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.parts.0.text").String(), ShouldEqual, "You are a helpful assistant.")
				So(gjson.Get(bodyStr, "contents.1.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.1.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.1.parts.0.text").String(), ShouldEqual, "Hello!")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatResp)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-req-id")
				So(resp.Model, ShouldEqual, "gemini-2.0-flash")
				So(resp.Message, ShouldNotBeNil)
				So(len(resp.Message.Id), ShouldEqual, 36)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Hello there! How can I help you today?\n")
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 2)
				// TODO: Completion token count is currently broken in google's SDK, add when available
			})
		})
	})
}

var mockChatStreamResp = []*entity.ChatResp{
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.0-flash",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: "Hello"}},
			},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.0-flash",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: " there! How"}},
			},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.0-flash",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: " can I help you today?\n"}},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     1,
				CompletionTokens: 11,
			},
		},
	},
}

func TestChatStream(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper for streaming", t, func() {
		config := &conf.GoogleConfig{
			ApiKey: "test-key",
		}
		mockRoundTripper := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockRoundTripper}
		repo, err := newGoogleUpstreamWithClient(config, httpClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-stream-req-id",
			Model: "gemini-2.0-flash",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "You are a helpful assistant."}}}},
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "Hello!"}}}},
			},
		}

		Convey("When ChatStream is called and the request is successful", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				So(err, ShouldBeNil)
				bodyStr := string(body)
				So(gjson.Get(bodyStr, "model").String(), ShouldEqual, "models/gemini-2.0-flash")
				So(gjson.Get(bodyStr, "contents.#").Int(), ShouldEqual, 2)
				So(gjson.Get(bodyStr, "contents.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.0.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.parts.0.text").String(), ShouldEqual, "You are a helpful assistant.")
				So(gjson.Get(bodyStr, "contents.1.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.1.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.1.parts.0.text").String(), ShouldEqual, "Hello!")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatResp)),
				}, nil
			}

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var messageID string
				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)

					messageID = resp.Message.Id
					responses = append(responses, resp)
				}

				So(responses, ShouldHaveLength, len(mockChatStreamResp))

				for _, mockResp := range mockChatStreamResp {
					mockResp.Message.Id = messageID
				}

				for i, resp := range responses {
					So(proto.Equal(resp, mockChatStreamResp[i]), ShouldBeTrue)
				}
			})
		})
	})
}

func TestChatWithToolCalls(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper", t, func() {
		config := &conf.GoogleConfig{
			ApiKey: "test-key",
		}
		mockRoundTripper := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockRoundTripper}
		repo, err := newGoogleUpstreamWithClient(config, httpClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-req-id",
			Model: "gemini-2.0-flash",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "What is the weather in Tokyo?"}}}},
			},
			Tools: []*v1.Tool{
				{
					Tool: &v1.Tool_Function_{
						Function: &v1.Tool_Function{
							Name:        "get_weather",
							Description: "Get the current weather for a city",
							Parameters: &v1.Schema{
								Type: v1.Schema_TYPE_OBJECT,
								Properties: map[string]*v1.Schema{
									"location": {
										Type:        v1.Schema_TYPE_STRING,
										Description: "City name",
									},
								},
								Required: []string{"location"},
							},
						},
					},
				},
			},
		}

		Convey("When Chat is called and the request is successful", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				So(err, ShouldBeNil)
				bodyStr := string(body)
				So(gjson.Get(bodyStr, "model").String(), ShouldEqual, "models/gemini-2.0-flash")
				So(gjson.Get(bodyStr, "contents.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.0.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.parts.0.text").String(), ShouldEqual, "What is the weather in Tokyo?")
				So(gjson.Get(bodyStr, "tools.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.name").String(), ShouldEqual, "get_weather")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.description").String(), ShouldEqual, "Get the current weather for a city")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.parameters.required.0").String(), ShouldEqual, "location")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatRespWithToolCall)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-req-id")
				So(resp.Model, ShouldEqual, "gemini-2.0-flash")
				So(resp.Message, ShouldNotBeNil)
				So(len(resp.Message.Id), ShouldEqual, 36)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(len(resp.Message.Contents), ShouldEqual, 1)
				So(resp.Message.Contents[0].GetFunctionCall().GetName(), ShouldEqual, "get_weather")
				So(resp.Message.Contents[0].GetFunctionCall().GetArguments(), ShouldEqual, "{\"location\":\"Tokyo\"}")
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.PromptTokens, ShouldEqual, 23)
				So(resp.Statistics.Usage.CompletionTokens, ShouldEqual, 5)
			})
		})
	})
}

var mockChatStreamRespWithToolCall = []*entity.ChatResp{
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.0-flash",
		Message: &v1.Message{
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_FunctionCall{
						FunctionCall: &v1.FunctionCall{
							Id:        "get_weather",
							Name:      "get_weather",
							Arguments: "{\"location\":\"Tokyo\"}",
						},
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     23,
				CompletionTokens: 5,
			},
		},
	},
}

func TestChatStreamWithToolCalls(t *testing.T) {
	Convey("Given a upstream with a mock HTTP round tripper for streaming with tool calls", t, func() {
		config := &conf.GoogleConfig{
			ApiKey: "test-key",
		}
		mockRoundTripper := &mockRoundTripper{}
		httpClient := &http.Client{Transport: mockRoundTripper}
		repo, err := newGoogleUpstreamWithClient(config, httpClient, log.DefaultLogger)
		So(err, ShouldBeNil)

		req := &entity.ChatReq{
			Id:    "test-stream-req-id",
			Model: "gemini-2.0-flash",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "What is the weather in Tokyo?"}}}},
			},
			Tools: []*v1.Tool{
				{
					Tool: &v1.Tool_Function_{
						Function: &v1.Tool_Function{
							Name:        "get_weather",
							Description: "Get the current weather for a city",
							Parameters: &v1.Schema{
								Type: v1.Schema_TYPE_OBJECT,
								Properties: map[string]*v1.Schema{
									"location": {
										Type:        v1.Schema_TYPE_STRING,
										Description: "City name",
									},
								},
								Required: []string{"location"},
							},
						},
					},
				},
			},
		}

		Convey("When ChatStream is called and the request is successful", func() {
			mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				So(err, ShouldBeNil)
				bodyStr := string(body)
				So(gjson.Get(bodyStr, "model").String(), ShouldEqual, "models/gemini-2.0-flash")
				So(gjson.Get(bodyStr, "contents.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.0.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.parts.0.text").String(), ShouldEqual, "What is the weather in Tokyo?")
				So(gjson.Get(bodyStr, "tools.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.name").String(), ShouldEqual, "get_weather")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.description").String(), ShouldEqual, "Get the current weather for a city")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.parameters.required.0").String(), ShouldEqual, "location")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockChatRespWithToolCall)),
				}, nil
			}

			streamClient, err := repo.ChatStream(context.Background(), req)

			Convey("Then it should return a stream client and no error", func() {
				So(err, ShouldBeNil)
				So(streamClient, ShouldNotBeNil)
				defer streamClient.Close()

				var messageID string
				var responses []*entity.ChatResp
				for {
					resp, err := streamClient.Recv()
					if err == io.EOF {
						break
					}
					So(err, ShouldBeNil)
					So(resp, ShouldNotBeNil)

					messageID = resp.Message.Id
					responses = append(responses, resp)
				}

				So(responses, ShouldHaveLength, len(mockChatStreamRespWithToolCall))

				for _, mockResp := range mockChatStreamRespWithToolCall {
					mockResp.Message.Id = messageID
				}

				for i, resp := range responses {
					So(proto.Equal(resp, mockChatStreamRespWithToolCall[i]), ShouldBeTrue)
				}
			})
		})
	})
}
