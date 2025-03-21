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
// source: neurouter/v1/model.proto

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

type ModelSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Provider string `protobuf:"bytes,3,opt,name=provider,proto3" json:"provider,omitempty"`
}

func (x *ModelSpec) Reset() {
	*x = ModelSpec{}
	mi := &file_neurouter_v1_model_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModelSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModelSpec) ProtoMessage() {}

func (x *ModelSpec) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_model_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModelSpec.ProtoReflect.Descriptor instead.
func (*ModelSpec) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_model_proto_rawDescGZIP(), []int{0}
}

func (x *ModelSpec) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ModelSpec) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ModelSpec) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

type ListModelReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListModelReq) Reset() {
	*x = ListModelReq{}
	mi := &file_neurouter_v1_model_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListModelReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListModelReq) ProtoMessage() {}

func (x *ListModelReq) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_model_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListModelReq.ProtoReflect.Descriptor instead.
func (*ListModelReq) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_model_proto_rawDescGZIP(), []int{1}
}

type ListModelResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Models []*ModelSpec `protobuf:"bytes,1,rep,name=models,proto3" json:"models,omitempty"`
}

func (x *ListModelResp) Reset() {
	*x = ListModelResp{}
	mi := &file_neurouter_v1_model_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListModelResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListModelResp) ProtoMessage() {}

func (x *ListModelResp) ProtoReflect() protoreflect.Message {
	mi := &file_neurouter_v1_model_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListModelResp.ProtoReflect.Descriptor instead.
func (*ListModelResp) Descriptor() ([]byte, []int) {
	return file_neurouter_v1_model_proto_rawDescGZIP(), []int{2}
}

func (x *ListModelResp) GetModels() []*ModelSpec {
	if x != nil {
		return x.Models
	}
	return nil
}

var File_neurouter_v1_model_proto protoreflect.FileDescriptor

var file_neurouter_v1_model_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6e, 0x65, 0x75, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4b, 0x0a, 0x09, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x53,
	0x70, 0x65, 0x63, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x22, 0x0e, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x52, 0x65, 0x71, 0x22, 0x40, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x52, 0x65, 0x73, 0x70, 0x12, 0x2f, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x53, 0x70, 0x65, 0x63, 0x52, 0x06, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x32, 0x61, 0x0a, 0x05, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x58,
	0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x1a, 0x2e, 0x6e, 0x65,
	0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4d,
	0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x1a, 0x1b, 0x2e, 0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x52, 0x65, 0x73, 0x70, 0x22, 0x12, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f, 0x76,
	0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x2e,
	0x78, 0x64, 0x65, 0x61, 0x2e, 0x78, 0x79, 0x7a, 0x2f, 0x54, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x2f,
	0x6e, 0x65, 0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6e, 0x65,
	0x75, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_neurouter_v1_model_proto_rawDescOnce sync.Once
	file_neurouter_v1_model_proto_rawDescData = file_neurouter_v1_model_proto_rawDesc
)

func file_neurouter_v1_model_proto_rawDescGZIP() []byte {
	file_neurouter_v1_model_proto_rawDescOnce.Do(func() {
		file_neurouter_v1_model_proto_rawDescData = protoimpl.X.CompressGZIP(file_neurouter_v1_model_proto_rawDescData)
	})
	return file_neurouter_v1_model_proto_rawDescData
}

var file_neurouter_v1_model_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_neurouter_v1_model_proto_goTypes = []any{
	(*ModelSpec)(nil),     // 0: neurouter.v1.ModelSpec
	(*ListModelReq)(nil),  // 1: neurouter.v1.ListModelReq
	(*ListModelResp)(nil), // 2: neurouter.v1.ListModelResp
}
var file_neurouter_v1_model_proto_depIdxs = []int32{
	0, // 0: neurouter.v1.ListModelResp.models:type_name -> neurouter.v1.ModelSpec
	1, // 1: neurouter.v1.Model.ListModel:input_type -> neurouter.v1.ListModelReq
	2, // 2: neurouter.v1.Model.ListModel:output_type -> neurouter.v1.ListModelResp
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_neurouter_v1_model_proto_init() }
func file_neurouter_v1_model_proto_init() {
	if File_neurouter_v1_model_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_neurouter_v1_model_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_neurouter_v1_model_proto_goTypes,
		DependencyIndexes: file_neurouter_v1_model_proto_depIdxs,
		MessageInfos:      file_neurouter_v1_model_proto_msgTypes,
	}.Build()
	File_neurouter_v1_model_proto = out.File
	file_neurouter_v1_model_proto_rawDesc = nil
	file_neurouter_v1_model_proto_goTypes = nil
	file_neurouter_v1_model_proto_depIdxs = nil
}
