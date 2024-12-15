package openai

import (
	"context"
	"encoding/json"
	"io"
	nethttp "net/http"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/sashabaranov/go-openai"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
)

type chatStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx http.Context
}

func (c *chatStreamServer) Context() context.Context {
	return c.ctx
}

func (c *chatStreamServer) Send(resp *v1.ChatResp) error {
	chunk := &openai.ChatCompletionStreamResponse{
		ID: resp.Message.Id,
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Role:    openai.ChatMessageRoleAssistant,
					Content: resp.Message.Contents[0].GetText(),
				},
			},
		},
	}
	if resp.Statistics != nil {
		chunk.Usage.PromptTokens = int(resp.Statistics.Usage.PromptTokens)
		chunk.Usage.CompletionTokens = int(resp.Statistics.Usage.CompletionTokens)
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	c.ctx.Response().Write([]byte("data: "))
	c.ctx.Response().Write(chunkJson)
	c.ctx.Response().Write([]byte("\n\n"))
	c.ctx.Response().(http.Flusher).Flush()
	return nil
}

func handleChatCompletion(ctx http.Context, svc v1.ChatServer) error {
	requestBody, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	openAIReq := openai.ChatCompletionRequest{}
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertChatCompletionRequestFromOpenAI(&openAIReq)

	if openAIReq.Stream {
		err = svc.ChatStream(req, &chatStreamServer{ctx: ctx})
	} else {
		resp, err := svc.Chat(ctx, req)
		if err != nil {
			return err
		}

		openAIResp := &openai.ChatCompletionResponse{
			ID: resp.Message.Id,
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: resp.Message.Contents[0].GetText(),
					},
				},
			},
		}
		if resp.Statistics != nil {
			openAIResp.Usage.PromptTokens = int(resp.Statistics.Usage.PromptTokens)
			openAIResp.Usage.CompletionTokens = int(resp.Statistics.Usage.CompletionTokens)
		}

		openAiRespJson, err := json.Marshal(openAIResp)
		if err != nil {
			return err
		}

		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(nethttp.StatusOK)
		_, err = ctx.Response().Write(openAiRespJson)
		if err != nil {
			return err
		}
	}

	return err
}
