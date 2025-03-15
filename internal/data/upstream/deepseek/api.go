package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *ChatRepo) CreateChatCompletion(ctx context.Context, req *ChatRequest) (resp *ChatResponse, err error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		err = fmt.Errorf("failed to marshal request: %w", err)
		return
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.config.BaseUrl+"/chat/completions",
		bytes.NewBuffer(reqBytes),
	)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+r.config.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("failed to send request: %w", err)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		var errResp struct {
			Error *Error `json:"error"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&errResp); err != nil {
			err = fmt.Errorf("failed to decode error response: %w", err)
			return
		}
		err = fmt.Errorf("DeepSeek API error: %w", errResp.Error)
		return
	}

	resp = &ChatResponse{}
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return
}

func (r *ChatRepo) CreateChatCompletionStream(ctx context.Context, req *ChatRequest) (httpResp *http.Response, err error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		err = fmt.Errorf("failed to marshal request: %w", err)
		return
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		r.config.BaseUrl+"/chat/completions",
		bytes.NewBuffer(reqBytes),
	)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+r.config.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	httpResp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("failed to send request: %w", err)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp struct {
			Error *Error `json:"error"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&errResp); err != nil {
			err = fmt.Errorf("failed to decode error response: %w", err)
			return
		}
		err = fmt.Errorf("DeepSeek API error: %w", errResp.Error)
		return
	}

	return
}
