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

package server

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/contrib/middleware/jwt/v3"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	jwt5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"

	"github.com/neuraxes/neurouter/internal/conf"
)

// jwtAuth returns a JWT auth middleware.
func jwtAuth(c *conf.Server) middleware.Middleware {
	jwtSecret := c.GetJwtKey()
	if jwtSecret == "" {
		return nil
	}
	return jwt.Server(func(token *jwt5.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
}

// createStreamInterceptor applies middleware to streaming RPCs.
func createStreamInterceptor(ms ...middleware.Middleware) grpc.StreamServerInterceptor {
	chain := middleware.Chain(ms...)

	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var req any
		h := func(ctx context.Context, _ any) (any, error) {
			return nil, handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx, req: &req})
		}
		_, err := chain(h)(ss.Context(), &capturedReq{msg: &req})
		return err
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
	req *any
}

func (w *wrappedStream) Context() context.Context {
	if w.ctx != nil {
		return w.ctx
	}
	return w.ServerStream.Context()
}

func (w *wrappedStream) RecvMsg(m any) error {
	err := w.ServerStream.RecvMsg(m)
	if err == nil && *w.req == nil {
		*w.req = m
	}
	return err
}

// capturedReq is evaluated by logging.Server after the handler returns, once
// RecvMsg has populated the inbound message.
type capturedReq struct {
	msg *any
}

func (c *capturedReq) Redact() string {
	if c == nil || c.msg == nil || *c.msg == nil {
		return ""
	}
	req := *c.msg
	if redacter, ok := req.(logging.Redacter); ok {
		return redacter.Redact()
	}
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}
