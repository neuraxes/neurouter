// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v3.21.12
// source: laas/v1/router.proto

package v1

import (
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
	return file_laas_v1_router_proto_enumTypes[0].Descriptor()
}

func (Role) Type() protoreflect.EnumType {
	return &file_laas_v1_router_proto_enumTypes[0]
}

func (x Role) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Role.Descriptor instead.
func (Role) EnumDescriptor() ([]byte, []int) {
	return file_laas_v1_router_proto_rawDescGZIP(), []int{0}
}

type ToolCall struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Types that are assignable to Tool:
	//
	//	*ToolCall_Function_
	Tool isToolCall_Tool `protobuf_oneof:"tool"`
}

func (x *ToolCall) Reset() {
	*x = ToolCall{}
	mi := &file_laas_v1_router_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToolCall) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToolCall) ProtoMessage() {}

func (x *ToolCall) ProtoReflect() protoreflect.Message {
	mi := &file_laas_v1_router_proto_msgTypes[0]
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
	return file_laas_v1_router_proto_rawDescGZIP(), []int{0}
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

func (x *ToolCall) GetFunction() *ToolCall_Function {
	if x, ok := x.GetTool().(*ToolCall_Function_); ok {
		return x.Function
	}
	return nil
}

type isToolCall_Tool interface {
	isToolCall_Tool()
}

type ToolCall_Function_ struct {
	Function *ToolCall_Function `protobuf:"bytes,2,opt,name=function,proto3,oneof"`
}

func (*ToolCall_Function_) isToolCall_Tool() {}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Role Role   `protobuf:"varint,2,opt,name=role,proto3,enum=laas.router.v1.Role" json:"role,omitempty"`
	// The multi-modality content
	Contents []*Content `protobuf:"bytes,3,rep,name=contents,proto3" json:"contents,omitempty"`
	// A set of tool calls raised by the message
	ToolCalls []*ToolCall `protobuf:"bytes,4,rep,name=tool_calls,json=toolCalls,proto3" json:"tool_calls,omitempty"`
	// Indicate the message is a response to a tool call
	ToolCallId string `protobuf:"bytes,5,opt,name=tool_call_id,json=toolCallId,proto3" json:"tool_call_id,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	mi := &file_laas_v1_router_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_laas_v1_router_proto_msgTypes[1]
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
	return file_laas_v1_router_proto_rawDescGZIP(), []int{1}
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
	mi := &file_laas_v1_router_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatReq) ProtoMessage() {}

func (x *ChatReq) ProtoReflect() protoreflect.Message {
	mi := &file_laas_v1_router_proto_msgTypes[2]
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
	return file_laas_v1_router_proto_rawDescGZIP(), []int{2}
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
	mi := &file_laas_v1_router_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatResp) ProtoMessage() {}

func (x *ChatResp) ProtoReflect() protoreflect.Message {
	mi := &file_laas_v1_router_proto_msgTypes[3]
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
	return file_laas_v1_router_proto_rawDescGZIP(), []int{3}
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

type ToolCall_Function struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Arguments string `protobuf:"bytes,2,opt,name=arguments,proto3" json:"arguments,omitempty"`
}

func (x *ToolCall_Function) Reset() {
	*x = ToolCall_Function{}
	mi := &file_laas_v1_router_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ToolCall_Function) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToolCall_Function) ProtoMessage() {}

func (x *ToolCall_Function) ProtoReflect() protoreflect.Message {
	mi := &file_laas_v1_router_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToolCall_Function.ProtoReflect.Descriptor instead.
func (*ToolCall_Function) Descriptor() ([]byte, []int) {
	return file_laas_v1_router_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ToolCall_Function) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ToolCall_Function) GetArguments() string {
	if x != nil {
		return x.Arguments
	}
	return ""
}

var File_laas_v1_router_proto protoreflect.FileDescriptor

var file_laas_v1_router_proto_rawDesc = []byte{
	0x0a, 0x14, 0x6c, 0x61, 0x61, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x14, 0x6c, 0x61, 0x61, 0x73, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa1, 0x01, 0x0a,
	0x08, 0x54, 0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3f, 0x0a, 0x08, 0x66, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6c, 0x61,
	0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6f,
	0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00,
	0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x3c, 0x0a, 0x08, 0x46, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x72,
	0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61,
	0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x6f, 0x6f, 0x6c,
	0x22, 0xd3, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x28, 0x0a, 0x04,
	0x72, 0x6f, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x6c, 0x61, 0x61,
	0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x6f, 0x6c, 0x65,
	0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x33, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e,
	0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x37, 0x0a, 0x0a, 0x74,
	0x6f, 0x6f, 0x6c, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x09, 0x74, 0x6f, 0x6f, 0x6c, 0x43,
	0x61, 0x6c, 0x6c, 0x73, 0x12, 0x20, 0x0a, 0x0c, 0x74, 0x6f, 0x6f, 0x6c, 0x5f, 0x63, 0x61, 0x6c,
	0x6c, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x6f, 0x6f, 0x6c,
	0x43, 0x61, 0x6c, 0x6c, 0x49, 0x64, 0x22, 0xca, 0x01, 0x0a, 0x07, 0x43, 0x68, 0x61, 0x74, 0x52,
	0x65, 0x71, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x38, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e,
	0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x33, 0x0a, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x2a, 0x0a, 0x05, 0x74, 0x6f, 0x6f, 0x6c, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f,
	0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x6f, 0x6f, 0x6c, 0x52, 0x05, 0x74, 0x6f,
	0x6f, 0x6c, 0x73, 0x22, 0x89, 0x01, 0x0a, 0x08, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x31, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x3a, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74,
	0x69, 0x63, 0x73, 0x52, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x2a,
	0x31, 0x0a, 0x04, 0x52, 0x6f, 0x6c, 0x65, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x59, 0x53, 0x54, 0x45,
	0x4d, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x01, 0x12, 0x09, 0x0a,
	0x05, 0x4d, 0x4f, 0x44, 0x45, 0x4c, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x54, 0x4f, 0x4f, 0x4c,
	0x10, 0x03, 0x32, 0x88, 0x01, 0x0a, 0x04, 0x43, 0x68, 0x61, 0x74, 0x12, 0x3b, 0x0a, 0x04, 0x43,
	0x68, 0x61, 0x74, 0x12, 0x17, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x71, 0x1a, 0x18, 0x2e, 0x6c,
	0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68,
	0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x0a, 0x43, 0x68, 0x61, 0x74,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x17, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f,
	0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x71, 0x1a,
	0x18, 0x2e, 0x6c, 0x61, 0x61, 0x73, 0x2e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x68, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x30, 0x01, 0x42, 0x2b, 0x5a,
	0x29, 0x67, 0x69, 0x74, 0x2e, 0x78, 0x64, 0x65, 0x61, 0x2e, 0x78, 0x79, 0x7a, 0x2f, 0x54, 0x75,
	0x72, 0x69, 0x6e, 0x67, 0x2f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x6c, 0x61, 0x61, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_laas_v1_router_proto_rawDescOnce sync.Once
	file_laas_v1_router_proto_rawDescData = file_laas_v1_router_proto_rawDesc
)

func file_laas_v1_router_proto_rawDescGZIP() []byte {
	file_laas_v1_router_proto_rawDescOnce.Do(func() {
		file_laas_v1_router_proto_rawDescData = protoimpl.X.CompressGZIP(file_laas_v1_router_proto_rawDescData)
	})
	return file_laas_v1_router_proto_rawDescData
}

var file_laas_v1_router_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_laas_v1_router_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_laas_v1_router_proto_goTypes = []any{
	(Role)(0),                 // 0: laas.router.v1.Role
	(*ToolCall)(nil),          // 1: laas.router.v1.ToolCall
	(*Message)(nil),           // 2: laas.router.v1.Message
	(*ChatReq)(nil),           // 3: laas.router.v1.ChatReq
	(*ChatResp)(nil),          // 4: laas.router.v1.ChatResp
	(*ToolCall_Function)(nil), // 5: laas.router.v1.ToolCall.Function
	(*Content)(nil),           // 6: laas.router.v1.Content
	(*GenerationConfig)(nil),  // 7: laas.router.v1.GenerationConfig
	(*Tool)(nil),              // 8: laas.router.v1.Tool
	(*Statistics)(nil),        // 9: laas.router.v1.Statistics
}
var file_laas_v1_router_proto_depIdxs = []int32{
	5,  // 0: laas.router.v1.ToolCall.function:type_name -> laas.router.v1.ToolCall.Function
	0,  // 1: laas.router.v1.Message.role:type_name -> laas.router.v1.Role
	6,  // 2: laas.router.v1.Message.contents:type_name -> laas.router.v1.Content
	1,  // 3: laas.router.v1.Message.tool_calls:type_name -> laas.router.v1.ToolCall
	7,  // 4: laas.router.v1.ChatReq.config:type_name -> laas.router.v1.GenerationConfig
	2,  // 5: laas.router.v1.ChatReq.messages:type_name -> laas.router.v1.Message
	8,  // 6: laas.router.v1.ChatReq.tools:type_name -> laas.router.v1.Tool
	2,  // 7: laas.router.v1.ChatResp.message:type_name -> laas.router.v1.Message
	9,  // 8: laas.router.v1.ChatResp.statistics:type_name -> laas.router.v1.Statistics
	3,  // 9: laas.router.v1.Chat.Chat:input_type -> laas.router.v1.ChatReq
	3,  // 10: laas.router.v1.Chat.ChatStream:input_type -> laas.router.v1.ChatReq
	4,  // 11: laas.router.v1.Chat.Chat:output_type -> laas.router.v1.ChatResp
	4,  // 12: laas.router.v1.Chat.ChatStream:output_type -> laas.router.v1.ChatResp
	11, // [11:13] is the sub-list for method output_type
	9,  // [9:11] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_laas_v1_router_proto_init() }
func file_laas_v1_router_proto_init() {
	if File_laas_v1_router_proto != nil {
		return
	}
	file_laas_v1_common_proto_init()
	file_laas_v1_router_proto_msgTypes[0].OneofWrappers = []any{
		(*ToolCall_Function_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_laas_v1_router_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_laas_v1_router_proto_goTypes,
		DependencyIndexes: file_laas_v1_router_proto_depIdxs,
		EnumInfos:         file_laas_v1_router_proto_enumTypes,
		MessageInfos:      file_laas_v1_router_proto_msgTypes,
	}.Build()
	File_laas_v1_router_proto = out.File
	file_laas_v1_router_proto_rawDesc = nil
	file_laas_v1_router_proto_goTypes = nil
	file_laas_v1_router_proto_depIdxs = nil
}
