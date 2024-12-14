package data

import (
	"context"
	"io"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/shared"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
	"git.xdea.xyz/Turing/router/internal/biz"
	"git.xdea.xyz/Turing/router/internal/conf"
)

type OpenAIChatRepo struct {
	config *conf.OpenAIConfig
	client *openai.Client
	log    *log.Helper
}

func NewOpenAIChatRepoFactory() biz.OpenAIChatRepoFactory {
	return func(config *conf.OpenAIConfig, logger log.Logger) biz.ChatRepo {
		return NewOpenAIChatRepo(config, logger)
	}
}

func NewOpenAIChatRepo(config *conf.OpenAIConfig, logger log.Logger) biz.ChatRepo {
	options := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
	}

	if config.BaseUrl != "" {
		options = append(options, option.WithBaseURL(config.BaseUrl))
	}

	return &OpenAIChatRepo{
		config: config,
		client: openai.NewClient(options...),
		log:    log.NewHelper(logger),
	}
}

// convertMessageToOpenAI converts an internal message to a message that can be sent to the OpenAI API.
func (r *OpenAIChatRepo) convertMessageToOpenAI(message *v1.Message) openai.ChatCompletionMessageParamUnion {
	isPureText := true
	plainText := ""

	{
		var sb strings.Builder
		for _, content := range message.Contents {
			if textContent, ok := content.GetContent().(*v1.Content_Text); ok {
				sb.WriteString(textContent.Text)
			} else {
				isPureText = false
				break
			}
		}
		plainText = sb.String()
	}

	switch message.Role {
	case v1.Role_SYSTEM:
		return openai.SystemMessage(plainText)
	case v1.Role_USER:
		if r.config.MergeContent && isPureText {
			return openai.UserMessage(plainText)
		} else {
			var parts []openai.ChatCompletionContentPartUnionParam
			for _, content := range message.Contents {
				switch c := content.GetContent().(type) {
				case *v1.Content_Text:
					parts = append(parts, openai.TextPart(c.Text))
				case *v1.Content_ImageUrl:
					parts = append(parts, openai.ImagePart(c.ImageUrl))
				}
			}
			return openai.UserMessageParts(parts...)
		}
	case v1.Role_MODEL:
		openAIMessage := openai.ChatCompletionAssistantMessageParam{
			Role: openai.F(openai.ChatCompletionAssistantMessageParamRoleAssistant),
		}

		if plainText != "" {
			openAIMessage.Content = openai.F([]openai.ChatCompletionAssistantMessageParamContentUnion{
				openai.TextPart(plainText),
			})
		}

		if message.ToolCalls != nil {
			var toolCalls []openai.ChatCompletionMessageToolCallParam
			for _, toolCall := range message.ToolCalls {
				function := toolCall.GetFunction()
				if function == nil {
					// Only support function tool
					continue
				}
				toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
					ID:   openai.F(toolCall.Id),
					Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
					Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      openai.F(function.Name),
						Arguments: openai.F(function.Arguments),
					}),
				})
			}
			openAIMessage.ToolCalls = openai.F(toolCalls)
		}

		return openAIMessage
	case v1.Role_TOOL:
		return openai.ToolMessage(message.ToolCallId, plainText)
	}
	return nil
}

// convertRequestToOpenAI converts an internal request to a request that can be sent to the OpenAI API.
func (r *OpenAIChatRepo) convertRequestToOpenAI(req *biz.ChatReq) openai.ChatCompletionNewParams {
	var messages []openai.ChatCompletionMessageParamUnion
	for _, message := range req.Messages {
		messages = append(messages, r.convertMessageToOpenAI(message))
	}
	openAIRequest := openai.ChatCompletionNewParams{
		Model:    openai.F(req.Model),
		Messages: openai.F(messages),
	}
	if req.Tools != nil {
		var tools []openai.ChatCompletionToolParam
		for _, tool := range req.Tools {
			function := tool.GetFunction()
			if function == nil {
				// Only support function tool
				continue
			}
			params, err := convertProtoMessageToJSONMap(function.Parameters)
			if err != nil {
				r.log.Errorf("failed to convert proto message to map: %v", err)
				continue
			}
			tools = append(tools, openai.ChatCompletionToolParam{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(shared.FunctionDefinitionParam{
					Name:        openai.F(function.Name),
					Description: openai.F(function.Description),
					Parameters:  openai.F(shared.FunctionParameters(params)),
				}),
			})
		}
		openAIRequest.Tools = openai.F(tools)
	}
	return openAIRequest
}

func (r *OpenAIChatRepo) convertMessageFromOpenAI(openAIMessage *openai.ChatCompletionMessage) *v1.Message {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("failed to generate UUID: %v", err)
	}

	message := &v1.Message{
		Id:   id.String(),
		Role: v1.Role_MODEL,
	}

	if openAIMessage.Content != "" {
		message.Contents = []*v1.Content{
			{
				Content: &v1.Content_Text{
					// The result contains a leading space, so we need to trim it
					Text: strings.TrimSpace(openAIMessage.Content),
				},
			},
		}
	}

	if openAIMessage.ToolCalls != nil {
		var toolCalls []*v1.ToolCall
		for _, toolCall := range openAIMessage.ToolCalls {
			if toolCall.Type != openai.ChatCompletionMessageToolCallTypeFunction {
				// Only support function tool
				continue
			}
			toolCalls = append(toolCalls, &v1.ToolCall{
				Id: toolCall.ID,
				Tool: &v1.ToolCall_Function_{
					Function: &v1.ToolCall_Function{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				},
			})
		}
		message.ToolCalls = toolCalls
	}

	return message
}

func (r *OpenAIChatRepo) Chat(ctx context.Context, req *biz.ChatReq) (resp *biz.ChatResp, err error) {
	res, err := r.client.Chat.Completions.New(
		ctx,
		r.convertRequestToOpenAI(req),
	)
	if err != nil {
		return
	}

	resp = &biz.ChatResp{
		Id:      req.Id,
		Message: r.convertMessageFromOpenAI(&res.Choices[0].Message),
	}

	if res.Usage.PromptTokens != 0 || res.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(res.Usage.PromptTokens),
				CompletionTokens: int32(res.Usage.CompletionTokens),
			},
		}
	}
	return
}

type openaiChatStreamClient struct {
	id       string
	req      *biz.ChatReq
	upstream *ssestream.Stream[openai.ChatCompletionChunk]
}

func (c openaiChatStreamClient) Recv() (resp *biz.ChatResp, err error) {
	if !c.upstream.Next() {
		if err = c.upstream.Err(); err != nil {
			return
		}
		err = io.EOF
		return
	}

	chunk := c.upstream.Current()
	resp = &biz.ChatResp{
		Id: c.req.Id,
		Message: &v1.Message{
			Id:   c.id,
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Content: &v1.Content_Text{
						Text: chunk.Choices[0].Delta.Content,
					},
				},
			},
		},
	}

	if chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0 {
		resp.Statistics = &v1.Statistics{
			Usage: &v1.Statistics_Usage{
				PromptTokens:     int32(chunk.Usage.PromptTokens),
				CompletionTokens: int32(chunk.Usage.CompletionTokens),
			},
		}
	}
	return
}

func (r *OpenAIChatRepo) ChatStream(ctx context.Context, req *biz.ChatReq) (client biz.ChatStreamClient, err error) {
	stream := r.client.Chat.Completions.NewStreaming(
		ctx,
		r.convertRequestToOpenAI(req),
	)

	id, err := uuid.NewUUID()
	if err != nil {
		return
	}

	client = &openaiChatStreamClient{
		id:       id.String(),
		req:      req,
		upstream: stream,
	}
	return
}
