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
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

func (s *Server) handleMessageCompletion(httpCtx http.Context) (err error) {
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

		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			util.EmitEvent(ctx, s.otelLogger, util.EventServerReqReceived, requestBody)
			streamServer := &messageStreamServer{
				ctx:     ctx,
				httpCtx: httpCtx,
			}
			if s.otelLogger != nil {
				streamServer.buffer = &bytes.Buffer{}
			}
			err := s.chatSvc.ChatStream(req.(*v1.ChatReq), streamServer)
			if err == nil {
				streamServer.sendMessageStopEvent()
			}
			if s.otelLogger != nil {
				util.EmitEvent(ctx, s.otelLogger, util.EventServerRespSent, streamServer.buffer.Bytes())
			}
			return nil, err
		})

		_, err = m(httpCtx, req)
	} else {
		var emitCtx context.Context = httpCtx
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			emitCtx = ctx
			util.EmitEvent(ctx, s.otelLogger, util.EventServerReqReceived, requestBody)
			return s.chatSvc.Chat(ctx, req.(*v1.ChatReq))
		})

		resp, err := m(httpCtx, req)
		if err != nil {
			return err
		}

		anthropicResp := convertChatRespToAnthropic(resp.(*v1.ChatResp))
		respBytes, err := json.Marshal(anthropicResp)
		if err != nil {
			return err
		}

		util.EmitEvent(emitCtx, s.otelLogger, util.EventServerRespSent, respBytes)

		err = httpCtx.Blob(200, "application/json", respBytes)
		if err != nil {
			return err
		}
	}

	return
}
