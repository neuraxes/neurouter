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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v3.21.12
// source: neurouter/v1/chat.proto

package v1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Role int32

const (
	Role_SYSTEM Role = 0
	Role_USER   Role = 1
	Role_MODEL  Role = 2
	Role_TOOL   Role = 3
)

// Enum value maps for Role.
var (
	Role_name = map[int32]string{
		0: "SYSTEM",
		1: "USER",
		2: "MODEL",
		3: "TOOL",
	}
	Role_value = map[string]int32{
		"SYSTEM": 0,
		"USER":   1,
		"MODEL":  2,
		"TOOL":   3,
	}
)

func (x Role) Enum() *Role {
	p := new(Role)
	*p = x
	return p
}

func (x Role) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Role) Descriptor() protoreflect.EnumDescriptor {
	return file_neurouter_v1_chat_proto_enumTypes[0].Descriptor()
}

func (Role) Type() protoreflect.EnumType {
	return &file_neurouter_v1_chat_proto_enumTypes[0]
}

func (x Role) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Role.Descriptor instead.
func (Role) EnumDescriptor() ([]byte, []int) {
	return file_neurouter_v1_chat_proto_rawDescGZIP(), []int{0}
}

type ToolCall struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Types that are assignable to Tool:
	//
	//	*ToolCall_Function
	Tool isToolCall_Tool `protobuf_oneof:"tool"`
}

func (x *ToolCall) Reset() {
	*x = ToolCall{}
	mi := &file_neurouter_v1_chat_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToolCall) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToolCall) ProtoMessage() {}

func (x *ToolCall) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_chat_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToolCall.ProtoReflect.Descriptor instead.
func (*ToolCall) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_chat_proto_rawDescGZIP(), []int{0}
}

func (x *ToolCall) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (m *ToolCall) GetTool() isToolCall_Tool {
	if m != nil {
		return m.Tool
	}
	return nil
}

func (x *ToolCall) GetFunction() *ToolCall_FunctionCall {
	if x, ok := x.GetTool().(*ToolCall_Function); ok {
		return x.Function
	}
	return nil
}

type isToolCall_Tool interface {
	isToolCall_Tool()
}

type ToolCall_Function struct {
	Function *ToolCall_FunctionCall `protobuf:"bytes,2,opt,name=function,proto3,oneof"`
}

func (*ToolCall_Function) isToolCall_Tool() {}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique identifier of the message
	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Role Role   `protobuf:"varint,2,opt,name=role,proto3,enum=neurouter.v1.Role" json:"role,omitempty"`
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// The multi-modality content
	Contents []*Content `protobuf:"bytes,4,rep,name=contents,proto3" json:"contents,omitempty"`
	// A set of tool calls raised by the message
	ToolCalls []*ToolCall `protobuf:"bytes,5,rep,name=tool_calls,json=toolCalls,proto3" json:"tool_calls,omitempty"`
	// Indicate the message is a response to a tool call
	ToolCallId string `protobuf:"bytes,6,opt,name=tool_call_id,json=toolCallId,proto3" json:"tool_call_id,omitempty"`
	// The reasoning content before final response
	ReasoningContent string `protobuf:"bytes,7,opt,name=reasoning_content,json=reasoningContent,proto3" json:"reasoning_content,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	mi := &file_neurouter_v1_chat_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_chat_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_chat_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Message) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_SYSTEM
}

func (x *Message) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Message) GetContents() []*Content {
	if x != nil {
		return x.Contents
	}
	return nil
}

func (x *Message) GetToolCalls() []*ToolCall {
	if x != nil {
		return x.ToolCalls
	}
	return nil
}

func (x *Message) GetToolCallId() string {
	if x != nil {
		return x.ToolCallId
	}
	return ""
}

func (x *Message) GetReasoningContent() string {
	if x != nil {
		return x.ReasoningContent
	}
	return ""
}

type ChatReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Model    string            `protobuf:"bytes,2,opt,name=model,proto3" json:"model,omitempty"`
	Config   *GenerationConfig `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	Messages []*Message        `protobuf:"bytes,4,rep,name=messages,proto3" json:"messages,omitempty"`
	Tools    []*Tool           `protobuf:"bytes,5,rep,name=tools,proto3" json:"tools,omitempty"`
}

func (x *ChatReq) Reset() {
	*x = ChatReq{}
	mi := &file_neurouter_v1_chat_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatReq) ProtoMessage() {}

func (x *ChatReq) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_chat_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatReq.ProtoReflect.Descriptor instead.
func (*ChatReq) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_chat_proto_rawDescGZIP(), []int{2}
}

func (x *ChatReq) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ChatReq) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *ChatReq) GetConfig() *GenerationConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ChatReq) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

func (x *ChatReq) GetTools() []*Tool {
	if x != nil {
		return x.Tools
	}
	return nil
}

type ChatResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id         string      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Message    *Message    `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Statistics *Statistics `protobuf:"bytes,3,opt,name=statistics,proto3" json:"statistics,omitempty"`
}

func (x *ChatResp) Reset() {
	*x = ChatResp{}
	mi := &file_neurouter_v1_chat_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatResp) ProtoMessage() {}

func (x *ChatResp) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_chat_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatResp.ProtoReflect.Descriptor instead.
func (*ChatResp) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_chat_proto_rawDescGZIP(), []int{3}
}

func (x *ChatResp) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ChatResp) GetMessage() *Message {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *ChatResp) GetStatistics() *Statistics {
	if x != nil {
		return x.Statistics
	}
	return nil
}

type ToolCall_FunctionCall struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Arguments string `protobuf:"bytes,2,opt,name=arguments,proto3" json:"arguments,omitempty"`
}

func (x *ToolCall_FunctionCall) Reset() {
	*x = ToolCall_FunctionCall{}
	mi := &file_neurouter_v1_chat_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToolCall_FunctionCall) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToolCall_FunctionCall) ProtoMessage() {}

func (x *ToolCall_FunctionCall) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_chat_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToolCall_FunctionCall.ProtoReflect.Descriptor instead.
func (*ToolCall_FunctionCall) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_chat_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ToolCall_FunctionCall) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ToolCall_FunctionCall) GetArguments() string {
	if x != nil {
		return x.Arguments
	}
	return ""
}

var File_neurouter_v1_chat_proto protoreflect.FileDescriptor

var file_neurouter_v1_chat_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63,
	0x68, 0x61, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6e, 0x65, 0x75, 0x72, 0x6f,
	0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xa7, 0x01, 0x0a, 0x08, 0x54, 0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x41, 0x0a,
	0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x43, 0x61, 0x6c, 0x6c, 0x48, 0x00, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x1a, 0x40, 0x0a, 0x0c, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x61, 0x6c, 0x6c,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x6f, 0x6f, 0x6c, 0x22, 0x8e, 0x02, 0x0a, 0x07, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x26, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x31, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x35, 0x0a, 0x0a, 0x74, 0x6f, 0x6f, 0x6c, 0x5f, 0x63, 0x61,
	0x6c, 0x6c, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6e, 0x65, 0x75, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c,
	0x6c, 0x52, 0x09, 0x74, 0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x73, 0x12, 0x20, 0x0a, 0x0c,
	0x74, 0x6f, 0x6f, 0x6c, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x74, 0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x49, 0x64, 0x12, 0x2b,
	0x0a, 0x11, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x72, 0x65, 0x61, 0x73, 0x6f,
	0x6e, 0x69, 0x6e, 0x67, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0xc4, 0x01, 0x0a, 0x07,
	0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x71, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x36, 0x0a,
	0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e,
	0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x31, 0x0a, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x28, 0x0a, 0x05, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x52, 0x05, 0x74, 0x6f, 0x6f,
	0x6c, 0x73, 0x22, 0x85, 0x01, 0x0a, 0x08, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x2f, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x15, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x38, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x52, 0x0a,
	0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x2a, 0x31, 0x0a, 0x04, 0x52, 0x6f,
	0x6c, 0x65, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x59, 0x53, 0x54, 0x45, 0x4d, 0x10, 0x00, 0x12, 0x08,
	0x0a, 0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x4d, 0x4f, 0x44, 0x45,
	0x4c, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x54, 0x4f, 0x4f, 0x4c, 0x10, 0x03, 0x32, 0x9b, 0x01,
	0x0a, 0x04, 0x43, 0x68, 0x61, 0x74, 0x12, 0x52, 0x0a, 0x04, 0x43, 0x68, 0x61, 0x74, 0x12, 0x15,
	0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68,
	0x61, 0x74, 0x52, 0x65, 0x71, 0x1a, 0x16, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x22, 0x1b, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x15, 0x12, 0x13, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x68, 0x61, 0x74, 0x2f,
	0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3f, 0x0a, 0x0a, 0x43, 0x68,
	0x61, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x15, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f,
	0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x71, 0x1a,
	0x16, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43,
	0x68, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x30, 0x01, 0x42, 0x33, 0x5a, 0x31, 0x67,
	0x69, 0x74, 0x2e, 0x78, 0x64, 0x65, 0x61, 0x2e, 0x78, 0x79, 0x7a, 0x2f, 0x54, 0x75, 0x72, 0x69,
	0x6e, 0x67, 0x2f, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_neurouter_v1_chat_proto_rawDescOnce sync.Once
	file_neurouter_v1_chat_proto_rawDescData = file_neurouter_v1_chat_proto_rawDesc
)

func file_neurouter_v1_chat_proto_rawDescGZIP() []byte {
	file_neurouter_v1_chat_proto_rawDescOnce.Do(func() {
		file_neurouter_v1_chat_proto_rawDescData = protoimpl.X.CompressGZIP(file_neurouter_v1_chat_proto_rawDescData)
	})
	return file_neurouter_v1_chat_proto_rawDescData
}

var file_neurouter_v1_chat_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_neurouter_v1_chat_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_neurouter_v1_chat_proto_goTypes = []any{
	(Role)(0),                     // 0: neurouter.v1.Role
	(*ToolCall)(nil),              // 1: neurouter.v1.ToolCall
	(*Message)(nil),               // 2: neurouter.v1.Message
	(*ChatReq)(nil),               // 3: neurouter.v1.ChatReq
	(*ChatResp)(nil),              // 4: neurouter.v1.ChatResp
	(*ToolCall_FunctionCall)(nil), // 5: neurouter.v1.ToolCall.FunctionCall
	(*Content)(nil),               // 6: neurouter.v1.Content
	(*GenerationConfig)(nil),      // 7: neurouter.v1.GenerationConfig
	(*Tool)(nil),                  // 8: neurouter.v1.Tool
	(*Statistics)(nil),            // 9: neurouter.v1.Statistics
}
var file_neurouter_v1_chat_proto_depIdxs = []int32{
	5,  // 0: neurouter.v1.ToolCall.function:type_name -> neurouter.v1.ToolCall.FunctionCall
	0,  // 1: neurouter.v1.Message.role:type_name -> neurouter.v1.Role
	6,  // 2: neurouter.v1.Message.contents:type_name -> neurouter.v1.Content
	1,  // 3: neurouter.v1.Message.tool_calls:type_name -> neurouter.v1.ToolCall
	7,  // 4: neurouter.v1.ChatReq.config:type_name -> neurouter.v1.GenerationConfig
	2,  // 5: neurouter.v1.ChatReq.messages:type_name -> neurouter.v1.Message
	8,  // 6: neurouter.v1.ChatReq.tools:type_name -> neurouter.v1.Tool
	2,  // 7: neurouter.v1.ChatResp.message:type_name -> neurouter.v1.Message
	9,  // 8: neurouter.v1.ChatResp.statistics:type_name -> neurouter.v1.Statistics
	3,  // 9: neurouter.v1.Chat.Chat:input_type -> neurouter.v1.ChatReq
	3,  // 10: neurouter.v1.Chat.ChatStream:input_type -> neurouter.v1.ChatReq
	4,  // 11: neurouter.v1.Chat.Chat:output_type -> neurouter.v1.ChatResp
	4,  // 12: neurouter.v1.Chat.ChatStream:output_type -> neurouter.v1.ChatResp
	11, // [11:13] is the sub-list for method output_type
	9,  // [9:11] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_neurouter_v1_chat_proto_init() }
func file_neurouter_v1_chat_proto_init() {
	if File_neurouter_v1_chat_proto != nil {
		return
	}
	file_neurouter_v1_common_proto_init()
	file_neurouter_v1_chat_proto_msgTypes[0].OneofWrappers = []any{
		(*ToolCall_Function)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_neurouter_v1_chat_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_neurouter_v1_chat_proto_goTypes,
		DependencyIndexes: file_neurouter_v1_chat_proto_depIdxs,
		EnumInfos:         file_neurouter_v1_chat_proto_enumTypes,
		MessageInfos:      file_neurouter_v1_chat_proto_msgTypes,
	}.Build()
	File_neurouter_v1_chat_proto = out.File
	file_neurouter_v1_chat_proto_rawDesc = nil
	file_neurouter_v1_chat_proto_goTypes = nil
	file_neurouter_v1_chat_proto_depIdxs = nil
}
