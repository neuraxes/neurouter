package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http/status"
	jwt5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// jwtAuth returns a JWT auth middleware.
//
// It reads the JWT key from the environment variable JWT_KEY.
func jwtAuth() middleware.Middleware {
	jwtSecret := os.Getenv("JWT_KEY")
	if jwtSecret == "" {
		return nil
	}
	return jwt.Server(func(token *jwt5.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
}

// createStreamInterceptor creates a gRPC server stream interceptor that implement middleware for streams.
func createStreamInterceptor(logger log.Logger) grpc.StreamServerInterceptor {
	// For request
	requestMiddlewares := []middleware.Middleware{
		recovery.Recovery(),
	}

	if j := jwtAuth(); j != nil {
		requestMiddlewares = append(requestMiddlewares, j)
	}

	m := middleware.Chain(requestMiddlewares...)
	logHelper := log.NewHelper(logger)

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		req := new(any)

		if info.FullMethod == "/neurouter.v1.Chat/ChatStream" {
			ss = &wrappedStream{ss, req}
		}

		h := func(_ context.Context, _ any) (any, error) {
			code := int32(status.FromGRPCCode(codes.OK))
			kind := transport.KindGRPC.String()
			operation := info.FullMethod
			startTime := time.Now()

			err := handler(srv, ss)

			reason := ""
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level := log.LevelInfo
			stack := ""
			if err != nil {
				level = log.LevelError
				stack = fmt.Sprintf("%+v", err)
			}
			reqStr := ""
			if r := *req; r != nil {
				if redacter, ok := r.(logging.Redacter); ok {
					reqStr = redacter.Redact()
				} else if stringer, ok := r.(fmt.Stringer); ok {
					reqStr = stringer.String()
				} else {
					reqStr = fmt.Sprintf("%+v", req)
				}
			}

			logHelper.WithContext(ss.Context()).Log(
				level,
				"kind", "server",
				"component", kind,
				"operation", operation,
				"args", reqStr,
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return nil, err
		}

		_, err := m(h)(ss.Context(), nil)

		return err
	}
}

type wrappedStream struct {
	grpc.ServerStream
	req *any
}

func (w *wrappedStream) RecvMsg(m any) error {
	*w.req = m // Save request message for logging
	return w.ServerStream.RecvMsg(m)
}
