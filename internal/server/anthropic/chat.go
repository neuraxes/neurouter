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

package anthropic

import (
	"context"
	"encoding/json"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

func handleMessageCompletion(httpCtx http.Context, svc v1.ChatServer) (err error) {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return
	}

	anthropicReq := anthropic.MessageNewParams{}
	err = json.Unmarshal(requestBody, &anthropicReq)
	if err != nil {
		return
	}

	req := convertChatReqFromAnthropic(&anthropicReq)

	if httpCtx.Request().Header.Get("Accept") == "text/event-stream" ||
		gjson.GetBytes(requestBody, "stream").Bool() {
		httpCtx.Response().Header().Set("Content-Type", "text/event-stream")
		httpCtx.Response().Header().Set("Cache-Control", "no-cache")
		httpCtx.Response().Header().Set("Connection", "keep-alive")

		streamServer := &messageStreamServer{httpCtx: httpCtx}

		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			err := svc.ChatStream(req.(*v1.ChatReq), streamServer)
			// Send final events after streaming is complete
			if err == nil {
				streamServer.sendMessageStopEvent()
			}
			return nil, err
		})

		_, err = m(httpCtx, req)
	} else {
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			return svc.Chat(ctx, req.(*v1.ChatReq))
		})

		resp, err := m(httpCtx, req)
		if err != nil {
			return err
		}

		return httpCtx.Result(200, convertChatRespToAnthropic(resp.(*v1.ChatResp)))
	}

	return
}
