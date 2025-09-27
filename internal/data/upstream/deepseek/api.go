// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// httpClient is the minimal interface for sending HTTP requests.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r *upstream) CreateChatCompletion(ctx context.Context, req *ChatRequest) (resp *ChatResponse, err error) {
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

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("failed to send request: %w", err)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		var errResp ErrorResp
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

func (r *upstream) CreateChatCompletionStream(ctx context.Context, req *ChatRequest) (httpResp *http.Response, err error) {
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

	httpResp, err = r.client.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("failed to send request: %w", err)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		// Close the body on error to avoid leaking the connection
		defer httpResp.Body.Close()
		var errResp ErrorResp
		if err = json.NewDecoder(httpResp.Body).Decode(&errResp); err != nil {
			err = fmt.Errorf("failed to decode error response: %w", err)
			return
		}
		err = fmt.Errorf("DeepSeek API error: %w", errResp.Error)
		return
	}

	return
}
