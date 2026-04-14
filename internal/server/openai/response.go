// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/openai/openai-go/v3/responses"
	"github.com/tidwall/gjson"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

func (s *Server) handleCreateResponse(httpCtx http.Context) (err error) {
	requestBody, err := io.ReadAll(httpCtx.Request().Body)
	if err != nil {
		return
	}

	var openAIReq responses.ResponseNewParams
	err = json.Unmarshal(requestBody, &openAIReq)
	if err != nil {
		return err
	}

	req := convertReqFromResponse(&openAIReq)

	if gjson.GetBytes(requestBody, "stream").Bool() {
		httpCtx.Response().Header().Set("Content-Type", "text/event-stream")
		httpCtx.Response().Header().Set("Cache-Control", "no-cache")
		httpCtx.Response().Header().Set("Connection", "keep-alive")

		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			util.EmitEvent(ctx, s.otelLogger, util.EventServerReqReceived, requestBody)
			streamServer := &responseStreamServer{
				ctx:     ctx,
				httpCtx: httpCtx,
			}
			if s.otelLogger != nil {
				streamServer.buffer = &bytes.Buffer{}
			}
			err := s.chatSvc.ChatStream(req.(*v1.ChatReq), streamServer)
			if err == nil {
				err = streamServer.sendDone()
			}
			if s.otelLogger != nil {
				util.EmitEvent(ctx, s.otelLogger, util.EventServerRespSent, streamServer.buffer.Bytes())
			}
			return nil, err
		})
		_, err = m(httpCtx, req)
	} else {
		var eventCtx context.Context = httpCtx
		m := httpCtx.Middleware(func(ctx context.Context, req any) (any, error) {
			eventCtx = ctx
			util.EmitEvent(ctx, s.otelLogger, util.EventServerReqReceived, requestBody)
			return s.chatSvc.Chat(ctx, req.(*v1.ChatReq))
		})

		resp, err := m(httpCtx, req)
		if err != nil {
			return err
		}

		openAIResp := convertRespToResponse(resp.(*v1.ChatResp))

		respBytes, err := json.Marshal(openAIResp)
		if err != nil {
			return err
		}

		util.EmitEvent(eventCtx, s.otelLogger, util.EventServerRespSent, respBytes)

		err = httpCtx.Blob(200, "application/json", respBytes)
		if err != nil {
			return err
		}
	}

	return
}
