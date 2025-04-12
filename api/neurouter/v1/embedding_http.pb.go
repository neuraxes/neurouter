// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.8.2
// - protoc             v3.21.12
// source: neurouter/v1/embedding.proto

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationEmbeddingEmbed = "/neurouter.v1.Embedding/Embed"

type EmbeddingHTTPServer interface {
	Embed(context.Context, *EmbedReq) (*EmbedResp, error)
}

func RegisterEmbeddingHTTPServer(s *http.Server, srv EmbeddingHTTPServer) {
	r := s.Route("/")
	r.POST("/v1/embed", _Embedding_Embed0_HTTP_Handler(srv))
}

func _Embedding_Embed0_HTTP_Handler(srv EmbeddingHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in EmbedReq
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationEmbeddingEmbed)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Embed(ctx, req.(*EmbedReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*EmbedResp)
		return ctx.Result(200, reply)
	}
}

type EmbeddingHTTPClient interface {
	Embed(ctx context.Context, req *EmbedReq, opts ...http.CallOption) (rsp *EmbedResp, err error)
}

type EmbeddingHTTPClientImpl struct {
	cc *http.Client
}

func NewEmbeddingHTTPClient(client *http.Client) EmbeddingHTTPClient {
	return &EmbeddingHTTPClientImpl{client}
}

func (c *EmbeddingHTTPClientImpl) Embed(ctx context.Context, in *EmbedReq, opts ...http.CallOption) (*EmbedResp, error) {
	var out EmbedResp
	pattern := "/v1/embed"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationEmbeddingEmbed))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
