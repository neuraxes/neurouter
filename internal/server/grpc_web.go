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
	nethttp "net/http"
	"slices"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"github.com/neuraxes/neurouter/internal/conf"
)

// GRPCWebFilter is an HTTP filter that serves browser gRPC-Web requests by translating them into in-process gRPC calls.
// It is nil when gRPC-Web support is disabled.
type GRPCWebFilter http.FilterFunc

// NewGRPCWebFilter builds a GRPCWebFilter that wraps the underlying gRPC server. It returns nil when gRPC-Web support is
// disabled in the configuration.
func NewGRPCWebFilter(c *conf.Server, grpcSrv *grpc.Server) GRPCWebFilter {
	if c.Http.GrpcWeb != nil && !c.Http.GetGrpcWeb() {
		return nil
	}

	allowedOrigins := c.Http.Cors.GetAllowedOrigins()
	originFunc := func(string) bool { return true }
	if len(allowedOrigins) > 0 && !slices.Contains(allowedOrigins, "*") {
		allowed := make(map[string]struct{}, len(allowedOrigins))
		for _, o := range allowedOrigins {
			allowed[o] = struct{}{}
		}
		originFunc = func(origin string) bool {
			_, ok := allowed[origin]
			return ok
		}
	}

	wrapped := grpcweb.WrapServer(
		grpcSrv.Server,
		grpcweb.WithOriginFunc(originFunc),
	)

	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			if wrapped.IsGrpcWebRequest(r) || wrapped.IsAcceptableGrpcCorsRequest(r) {
				wrapped.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
