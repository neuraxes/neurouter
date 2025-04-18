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

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: neurouter/v1/embedding.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Embedding_Embed_FullMethodName = "/neurouter.v1.Embedding/Embed"
)

// EmbeddingClient is the client API for Embedding service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EmbeddingClient interface {
	Embed(ctx context.Context, in *EmbedReq, opts ...grpc.CallOption) (*EmbedResp, error)
}

type embeddingClient struct {
	cc grpc.ClientConnInterface
}

func NewEmbeddingClient(cc grpc.ClientConnInterface) EmbeddingClient {
	return &embeddingClient{cc}
}

func (c *embeddingClient) Embed(ctx context.Context, in *EmbedReq, opts ...grpc.CallOption) (*EmbedResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmbedResp)
	err := c.cc.Invoke(ctx, Embedding_Embed_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EmbeddingServer is the server API for Embedding service.
// All implementations must embed UnimplementedEmbeddingServer
// for forward compatibility.
type EmbeddingServer interface {
	Embed(context.Context, *EmbedReq) (*EmbedResp, error)
	mustEmbedUnimplementedEmbeddingServer()
}

// UnimplementedEmbeddingServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedEmbeddingServer struct{}

func (UnimplementedEmbeddingServer) Embed(context.Context, *EmbedReq) (*EmbedResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Embed not implemented")
}
func (UnimplementedEmbeddingServer) mustEmbedUnimplementedEmbeddingServer() {}
func (UnimplementedEmbeddingServer) testEmbeddedByValue()                   {}

// UnsafeEmbeddingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EmbeddingServer will
// result in compilation errors.
type UnsafeEmbeddingServer interface {
	mustEmbedUnimplementedEmbeddingServer()
}

func RegisterEmbeddingServer(s grpc.ServiceRegistrar, srv EmbeddingServer) {
	// If the following call pancis, it indicates UnimplementedEmbeddingServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Embedding_ServiceDesc, srv)
}

func _Embedding_Embed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmbedReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmbeddingServer).Embed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Embedding_Embed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmbeddingServer).Embed(ctx, req.(*EmbedReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Embedding_ServiceDesc is the grpc.ServiceDesc for Embedding service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Embedding_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "neurouter.v1.Embedding",
	HandlerType: (*EmbeddingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Embed",
			Handler:    _Embedding_Embed_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "neurouter/v1/embedding.proto",
}
