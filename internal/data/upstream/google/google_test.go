package google

import (
	"context"
	"errors"
	"fmt"
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

const mockGenerateContentResp = `{
    "candidates": [
        {
            "content": {
                "parts": [
                    {
                        "text": "Right, okay, let's keep it professional, friendly, and helpful. As an AI assistant, my primary function is to be of service.",
                        "thought": true
                    },
                    {
                        "text": "Hello there! How can I help you today?"
                    }
                ],
                "role": "model"
            },
            "finishReason": "STOP",
            "index": 0
        }
    ],
    "usageMetadata": {
        "promptTokenCount": 10,
        "candidatesTokenCount": 8,
        "totalTokenCount": 47,
        "promptTokensDetails": [
            {
                "modality": "TEXT",
                "tokenCount": 10
            }
        ],
        "thoughtsTokenCount": 27
    },
    "modelVersion": "gemini-2.5-flash",
    "responseId": "BmcPafuSOIej2roP08fp6AI"
}`

const mockGenerateContentStreamResp = `data: {"candidates": [{"content": {"parts": [{"text": "**Considering First Interactions**\n\nI've been examining how I respond to simple greetings like \"hi.\" My primary focus is acknowledging the greeting as a social foundation. Next, I consider the formality and tone of the initial message to ensure an appropriate response. It's a fundamental part of the interaction.\n\n\n","thought": true}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 2,"totalTokenCount": 73,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 2}],"thoughtsTokenCount": 71},"modelVersion": "gemini-2.5-flash","responseId": "jHIPaZKcEoOR0-kPxMfy0Ag"}

data: {"candidates": [{"content": {"parts": [{"text": "**Developing Optimal Response Strategies**\n\nI'm refining my response to \"hi\" to maximize usefulness. My focus now is on creating the perfect blend of acknowledgment and assistance. I'm prioritizing directness and helpfulness. The response \"Hi there! How can I help you today?\" is shaping up as the most effective starting point. This ensures a welcoming tone while immediately offering support.\n\n\n","thought": true}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 2,"totalTokenCount": 315,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 2}],"thoughtsTokenCount": 313},"modelVersion": "gemini-2.5-flash","responseId": "jHIPaZKcEoOR0-kPxMfy0Ag"}

data: {"candidates": [{"content": {"parts": [{"text": "**Formulating Ideal Responses**\n\nI'm solidifying my approach to \"hi\" and similar greetings. I now focus on a consistent flow, acknowledging the greeting and then immediately offering assistance. The process incorporates mirroring the tone, ensuring an open-ended invitation.  \"Hi there! How can I help you today?\" efficiently fulfills these conditions, combining a friendly greeting with an immediate offer of help.\n\n\n","thought": true}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 2,"totalTokenCount": 346,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 2}],"thoughtsTokenCount": 344},"modelVersion": "gemini-2.5-flash","responseId": "jHIPaZKcEoOR0-kPxMfy0Ag"}

data: {"candidates": [{"content": {"parts": [{"text": "Hi there! How can I help you today?"}],"role": "model"},"finishReason": "STOP","index": 0}],"usageMetadata": {"promptTokenCount": 2,"candidatesTokenCount": 7,"totalTokenCount": 353,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 2}],"thoughtsTokenCount": 344},"modelVersion": "gemini-2.5-flash","responseId": "jHIPaZKcEoOR0-kPxMfy0Ag"}
`

const mockGenerateContentWithToolCall = `{
    "candidates": [
        {
            "content": {
                "parts": [
                    {
                        "text": "**Okay, here's what I'm thinking:**\n\nAlright, so the user wants to know the weather in Shanghai. Easy enough. I know I have a tool specifically for that, a get_weather function. Looks like I just need to feed it a city name. So, I should call that tool, and I'll use city='shanghai' as the input. That should give me the weather data they're looking for. Seems straightforward!\n",
                        "thought": true
                    },
                    {
                        "functionCall": {
                            "name": "get_weather",
                            "args": {
                                "city": "shanghai"
                            }
                        },
                        "thoughtSignature": "Cu0BAdHtim/1tmi8ZRpfrEgl+Gi++Kc324ShOTqoDpHiN9X82Sp1pgvhLoKcAEBTzdPdWviBCNREjUqhFSfCaIWqPq9Mum5g9k8gj76Cz5Tzxjf3f7TEypscN3r/EnpleAQQNp105mMKqhi3hwpSFIpn0T2zgT142ow+vzgfPQeA6crel3/yi/3FGyRsQL2K/GuyqDrKBhJei0pXAP/rLMlH5FOJcGuxNrJlH2dZhXIcCYVnABxZg0Qchlxr0kZ6fe7oLOu7hVGbl1vxq12alaTsuetmQil738x1Tmwd4g7tFvSGnZZG37quTBrpXzNa"
                    }
                ],
                "role": "model"
            },
            "finishReason": "STOP",
            "index": 0,
            "finishMessage": "Model generated function call(s)."
        }
    ],
    "usageMetadata": {
        "promptTokenCount": 50,
        "candidatesTokenCount": 17,
        "totalTokenCount": 119,
        "promptTokensDetails": [
            {
                "modality": "TEXT",
                "tokenCount": 50
            }
        ],
        "thoughtsTokenCount": 52
    },
    "modelVersion": "gemini-2.5-flash",
    "responseId": "J3YPafGkDK3H1e8Pv57FgAQ"
}`

const mockGenerateContentStreamWithToolCall = `data: {"candidates": [{"content": {"parts": [{"text": "**Pinpointing Location Details**\n\nI've zeroed in on the initial challenge: understanding \"capital of us\" requires pinpointing Washington D.C. as the target location. Now, I can confidently deploy the get_weather tool, feeding it \"Washington D.C.\" for precise weather data.\n\n\n","thought": true}],"role": "model"},"index": 0}],"usageMetadata": {"promptTokenCount": 52,"totalTokenCount": 108,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 52}],"thoughtsTokenCount": 56},"modelVersion": "gemini-2.5-flash","responseId": "CnIPabyjF53Ivr0P4f-e2QU"}

data: {"candidates": [{"content": {"parts": [{"functionCall": {"name": "get_weather","args": {"city": "Washington D.C."}},"thoughtSignature": "CiQB0e2Kb6TkudQsDxtD31FVR08uogP/Bg0v09ujrGIhrx9VQ1IKYgHR7YpvNGHEHX8O/GkY5kxUtxywioE7wlsDXFy8wzq/ang5CWJp1w6DdCRk/z56XrimEJIFMap6AfwCQoEgQnyxS+CxYubeq8RGdTS9X1+tVDCKMR1HcIwzMONal/IcIwt6CqoBAdHtim820E3x/2hGPtduP5IxMdnBAn0srutyqsG/eA3VFhP4qSRUXBSrg1kDZuUtpKsg87IjddAXh4Yz7YVh1KahPnfLO+0h7eMiBYqw/0B61AUTQfA3zbF3byCgfvpSJm/MEhUevoIY2cVTlFv2G57/D1t0ssJllAJ1rZrk1zv1JF5Pq5RBZrB+z8Ox4Uc9HeW7nPAlFzGeSxGzUr3K+kvrSSDNGx8AuZM="}],"role": "model"},"finishReason": "STOP","index": 0,"finishMessage": "Model generated function call(s)."}],"usageMetadata": {"promptTokenCount": 52,"candidatesTokenCount": 17,"totalTokenCount": 125,"promptTokensDetails": [{"modality": "TEXT","tokenCount": 52}],"thoughtsTokenCount": 56},"modelVersion": "gemini-2.5-flash","responseId": "CnIPabyjF53Ivr0P4f-e2QU"}
`

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
			Model: "gemini-2.5-flash",
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
				So(req.URL.Path, ShouldContainSubstring, "/gemini-2.5-flash:generateContent")
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
					Body: io.NopCloser(strings.NewReader(mockGenerateContentResp)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-req-id")
				So(resp.Model, ShouldEqual, "gemini-2.5-flash")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Id, ShouldNotBeEmpty)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(resp.Message.Contents, ShouldHaveLength, 2)
				So(resp.Message.Contents[0].Reasoning, ShouldBeTrue)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "Right, okay, let's keep it professional, friendly, and helpful. As an AI assistant, my primary function is to be of service.")
				So(resp.Message.Contents[1].GetText(), ShouldEqual, "Hello there! How can I help you today?")
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.InputTokens, ShouldEqual, 10)
				So(resp.Statistics.Usage.OutputTokens, ShouldEqual, 35)
				So(resp.Statistics.Usage.CachedInputTokens, ShouldEqual, 0)
			})
		})
	})
}

var mockChatStreamResp = []*entity.ChatResp{
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "jHIPaZKcEoOR0-kPxMfy0Ag",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Reasoning: true, Content: &v1.Content_Text{Text: "**Considering First Interactions**\n\nI've been examining how I respond to simple greetings like \"hi.\" My primary focus is acknowledging the greeting as a social foundation. Next, I consider the formality and tone of the initial message to ensure an appropriate response. It's a fundamental part of the interaction.\n\n\n"}},
			},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "jHIPaZKcEoOR0-kPxMfy0Ag",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Reasoning: true, Content: &v1.Content_Text{Text: "**Developing Optimal Response Strategies**\n\nI'm refining my response to \"hi\" to maximize usefulness. My focus now is on creating the perfect blend of acknowledgment and assistance. I'm prioritizing directness and helpfulness. The response \"Hi there! How can I help you today?\" is shaping up as the most effective starting point. This ensures a welcoming tone while immediately offering support.\n\n\n"}},
			},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "jHIPaZKcEoOR0-kPxMfy0Ag",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Reasoning: true, Content: &v1.Content_Text{Text: "**Formulating Ideal Responses**\n\nI'm solidifying my approach to \"hi\" and similar greetings. I now focus on a consistent flow, acknowledging the greeting and then immediately offering assistance. The process incorporates mirroring the tone, ensuring an open-ended invitation.  \"Hi there! How can I help you today?\" efficiently fulfills these conditions, combining a friendly greeting with an immediate offer of help.\n\n\n"}},
			},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "jHIPaZKcEoOR0-kPxMfy0Ag",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Content: &v1.Content_Text{Text: "Hi there! How can I help you today?"}},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  2,
				OutputTokens: 351,
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
			Model: "gemini-2.5-flash",
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
				So(req.URL.Path, ShouldContainSubstring, "/gemini-2.5-flash:streamGenerateContent")
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
					Body: io.NopCloser(strings.NewReader(mockGenerateContentStreamResp)),
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

				So(responses, ShouldHaveLength, len(mockChatStreamResp))

				for i, resp := range responses {
					if !proto.Equal(resp, mockChatStreamResp[i]) {
						fmt.Println("\n", resp.String(), "\n", mockChatStreamResp[i].String())
					}
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
			Model: "gemini-2.5-flash",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "What is the weather in Shanghai?"}}}},
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
									"city": {
										Type:        v1.Schema_TYPE_STRING,
										Description: "City name",
									},
								},
								Required: []string{"city"},
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
				So(req.URL.Path, ShouldContainSubstring, "/gemini-2.5-flash:generateContent")
				So(gjson.Get(bodyStr, "contents.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.0.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.parts.0.text").String(), ShouldEqual, "What is the weather in Shanghai?")
				So(gjson.Get(bodyStr, "tools.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.name").String(), ShouldEqual, "get_weather")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.description").String(), ShouldEqual, "Get the current weather for a city")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.parameters.required.0").String(), ShouldEqual, "city")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockGenerateContentWithToolCall)),
				}, nil
			}

			resp, err := repo.Chat(context.Background(), req)

			Convey("Then it should return a valid response and no error", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Id, ShouldEqual, "test-req-id")
				So(resp.Model, ShouldEqual, "gemini-2.5-flash")
				So(resp.Message, ShouldNotBeNil)
				So(resp.Message.Id, ShouldNotBeEmpty)
				So(resp.Message.Role, ShouldEqual, v1.Role_MODEL)
				So(resp.Message.Contents, ShouldHaveLength, 2)
				So(resp.Message.Contents[0].Reasoning, ShouldBeTrue)
				So(resp.Message.Contents[0].GetText(), ShouldEqual, "**Okay, here's what I'm thinking:**\n\nAlright, so the user wants to know the weather in Shanghai. Easy enough. I know I have a tool specifically for that, a get_weather function. Looks like I just need to feed it a city name. So, I should call that tool, and I'll use city='shanghai' as the input. That should give me the weather data they're looking for. Seems straightforward!\n")
				So(resp.Message.Contents[1].GetToolUse().GetName(), ShouldEqual, "get_weather")
				So(resp.Message.Contents[1].GetToolUse().GetTextualInput(), ShouldEqual, "{\"city\":\"shanghai\"}")
				So(resp.Statistics, ShouldNotBeNil)
				So(resp.Statistics.Usage.InputTokens, ShouldEqual, 50)
				So(resp.Statistics.Usage.OutputTokens, ShouldEqual, 69)
			})
		})
	})
}

var mockChatStreamRespWithToolCall = []*entity.ChatResp{
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "CnIPabyjF53Ivr0P4f-e2QU",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{Reasoning: true, Content: &v1.Content_Text{Text: "**Pinpointing Location Details**\n\nI've zeroed in on the initial challenge: understanding \"capital of us\" requires pinpointing Washington D.C. as the target location. Now, I can confidently deploy the get_weather tool, feeding it \"Washington D.C.\" for precise weather data.\n\n\n"}},
			},
		},
	},
	{
		Id:    "test-stream-req-id",
		Model: "gemini-2.5-flash",
		Message: &v1.Message{
			Id:   "CnIPabyjF53Ivr0P4f-e2QU",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Metadata: map[string]string{
						"thoughtSignature": "CiQB0e2Kb6TkudQsDxtD31FVR08uogP/Bg0v09ujrGIhrx9VQ1IKYgHR7YpvNGHEHX8O/GkY5kxUtxywioE7wlsDXFy8wzq/ang5CWJp1w6DdCRk/z56XrimEJIFMap6AfwCQoEgQnyxS+CxYubeq8RGdTS9X1+tVDCKMR1HcIwzMONal/IcIwt6CqoBAdHtim820E3x/2hGPtduP5IxMdnBAn0srutyqsG/eA3VFhP4qSRUXBSrg1kDZuUtpKsg87IjddAXh4Yz7YVh1KahPnfLO+0h7eMiBYqw/0B61AUTQfA3zbF3byCgfvpSJm/MEhUevoIY2cVTlFv2G57/D1t0ssJllAJ1rZrk1zv1JF5Pq5RBZrB+z8Ox4Uc9HeW7nPAlFzGeSxGzUr3K+kvrSSDNGx8AuZM=",
					},
					Content: &v1.Content_ToolUse{
						ToolUse: &v1.ToolUse{
							Id:   "get_weather",
							Name: "get_weather",
							Inputs: []*v1.ToolUse_Input{
								{
									Input: &v1.ToolUse_Input_Text{Text: "{\"city\":\"Washington D.C.\"}"},
								},
							},
						},
					},
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				InputTokens:  52,
				OutputTokens: 73,
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
			Model: "gemini-2.5-flash",
			Messages: []*v1.Message{
				{Role: v1.Role_USER, Contents: []*v1.Content{{Content: &v1.Content_Text{Text: "What is the weather in the capital of us"}}}},
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
									"city": {
										Type:        v1.Schema_TYPE_STRING,
										Description: "City name",
									},
								},
								Required: []string{"city"},
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
				So(req.URL.Path, ShouldContainSubstring, "/gemini-2.5-flash:streamGenerateContent")
				So(gjson.Get(bodyStr, "contents.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.role").String(), ShouldEqual, "user")
				So(gjson.Get(bodyStr, "contents.0.parts.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "contents.0.parts.0.text").String(), ShouldEqual, "What is the weather in the capital of us")
				So(gjson.Get(bodyStr, "tools.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.#").Int(), ShouldEqual, 1)
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.name").String(), ShouldEqual, "get_weather")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.description").String(), ShouldEqual, "Get the current weather for a city")
				So(gjson.Get(bodyStr, "tools.0.functionDeclarations.0.parameters.required.0").String(), ShouldEqual, "city")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(mockGenerateContentStreamWithToolCall)),
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

				So(responses, ShouldHaveLength, len(mockChatStreamRespWithToolCall))

				for i, resp := range responses {
					So(proto.Equal(resp, mockChatStreamRespWithToolCall[i]), ShouldBeTrue)
				}
			})
		})
	})
}
