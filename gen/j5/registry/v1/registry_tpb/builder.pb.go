// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/registry/v1/topic/builder.proto

package registry_tpb

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	messaging_j5pb "github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	source_j5pb "github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BuildStatus int32

const (
	BuildStatus_BUILD_STATUS_UNSPECIFIED BuildStatus = 0
	BuildStatus_BUILD_STATUS_IN_PROGRESS BuildStatus = 1
	BuildStatus_BUILD_STATUS_SUCCESS     BuildStatus = 2
	BuildStatus_BUILD_STATUS_FAILURE     BuildStatus = 3
)

// Enum value maps for BuildStatus.
var (
	BuildStatus_name = map[int32]string{
		0: "BUILD_STATUS_UNSPECIFIED",
		1: "BUILD_STATUS_IN_PROGRESS",
		2: "BUILD_STATUS_SUCCESS",
		3: "BUILD_STATUS_FAILURE",
	}
	BuildStatus_value = map[string]int32{
		"BUILD_STATUS_UNSPECIFIED": 0,
		"BUILD_STATUS_IN_PROGRESS": 1,
		"BUILD_STATUS_SUCCESS":     2,
		"BUILD_STATUS_FAILURE":     3,
	}
)

func (x BuildStatus) Enum() *BuildStatus {
	p := new(BuildStatus)
	*p = x
	return p
}

func (x BuildStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BuildStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_j5_registry_v1_topic_builder_proto_enumTypes[0].Descriptor()
}

func (BuildStatus) Type() protoreflect.EnumType {
	return &file_j5_registry_v1_topic_builder_proto_enumTypes[0]
}

func (x BuildStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BuildStatus.Descriptor instead.
func (BuildStatus) EnumDescriptor() ([]byte, []int) {
	return file_j5_registry_v1_topic_builder_proto_rawDescGZIP(), []int{0}
}

type PublishMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request *messaging_j5pb.RequestMetadata `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	Commit  *source_j5pb.CommitInfo         `protobuf:"bytes,2,opt,name=commit,proto3" json:"commit,omitempty"`
	Bundle  string                          `protobuf:"bytes,3,opt,name=bundle,proto3" json:"bundle,omitempty"`
	Name    string                          `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *PublishMessage) Reset() {
	*x = PublishMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublishMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishMessage) ProtoMessage() {}

func (x *PublishMessage) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishMessage.ProtoReflect.Descriptor instead.
func (*PublishMessage) Descriptor() ([]byte, []int) {
	return file_j5_registry_v1_topic_builder_proto_rawDescGZIP(), []int{0}
}

func (x *PublishMessage) GetRequest() *messaging_j5pb.RequestMetadata {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *PublishMessage) GetCommit() *source_j5pb.CommitInfo {
	if x != nil {
		return x.Commit
	}
	return nil
}

func (x *PublishMessage) GetBundle() string {
	if x != nil {
		return x.Bundle
	}
	return ""
}

func (x *PublishMessage) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type BuildAPIMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request *messaging_j5pb.RequestMetadata `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	Commit  *source_j5pb.CommitInfo         `protobuf:"bytes,2,opt,name=commit,proto3" json:"commit,omitempty"`
	Bundle  string                          `protobuf:"bytes,3,opt,name=bundle,proto3" json:"bundle,omitempty"`
}

func (x *BuildAPIMessage) Reset() {
	*x = BuildAPIMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildAPIMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildAPIMessage) ProtoMessage() {}

func (x *BuildAPIMessage) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildAPIMessage.ProtoReflect.Descriptor instead.
func (*BuildAPIMessage) Descriptor() ([]byte, []int) {
	return file_j5_registry_v1_topic_builder_proto_rawDescGZIP(), []int{1}
}

func (x *BuildAPIMessage) GetRequest() *messaging_j5pb.RequestMetadata {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *BuildAPIMessage) GetCommit() *source_j5pb.CommitInfo {
	if x != nil {
		return x.Commit
	}
	return nil
}

func (x *BuildAPIMessage) GetBundle() string {
	if x != nil {
		return x.Bundle
	}
	return ""
}

type J5BuildStatusMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request *messaging_j5pb.RequestMetadata `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	Status  BuildStatus                     `protobuf:"varint,2,opt,name=status,proto3,enum=j5.registry.v1.topic.BuildStatus" json:"status,omitempty"`
	Output  *BuildOutput                    `protobuf:"bytes,3,opt,name=output,proto3" json:"output,omitempty"`
}

func (x *J5BuildStatusMessage) Reset() {
	*x = J5BuildStatusMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *J5BuildStatusMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*J5BuildStatusMessage) ProtoMessage() {}

func (x *J5BuildStatusMessage) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use J5BuildStatusMessage.ProtoReflect.Descriptor instead.
func (*J5BuildStatusMessage) Descriptor() ([]byte, []int) {
	return file_j5_registry_v1_topic_builder_proto_rawDescGZIP(), []int{2}
}

func (x *J5BuildStatusMessage) GetRequest() *messaging_j5pb.RequestMetadata {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *J5BuildStatusMessage) GetStatus() BuildStatus {
	if x != nil {
		return x.Status
	}
	return BuildStatus_BUILD_STATUS_UNSPECIFIED
}

func (x *J5BuildStatusMessage) GetOutput() *BuildOutput {
	if x != nil {
		return x.Output
	}
	return nil
}

type BuildOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Title   string  `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Summary string  `protobuf:"bytes,2,opt,name=summary,proto3" json:"summary,omitempty"`
	Text    *string `protobuf:"bytes,3,opt,name=text,proto3,oneof" json:"text,omitempty"`
}

func (x *BuildOutput) Reset() {
	*x = BuildOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildOutput) ProtoMessage() {}

func (x *BuildOutput) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_v1_topic_builder_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildOutput.ProtoReflect.Descriptor instead.
func (*BuildOutput) Descriptor() ([]byte, []int) {
	return file_j5_registry_v1_topic_builder_proto_rawDescGZIP(), []int{3}
}

func (x *BuildOutput) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *BuildOutput) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *BuildOutput) GetText() string {
	if x != nil && x.Text != nil {
		return *x.Text
	}
	return ""
}

var File_j5_registry_v1_topic_builder_proto protoreflect.FileDescriptor

var file_j5_registry_v1_topic_builder_proto_rawDesc = []byte{
	0x0a, 0x22, 0x6a, 0x35, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x2f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x6a, 0x35, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69,
	0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6a, 0x35, 0x2f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x71, 0x72, 0x65, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x6a, 0x35, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xba, 0x01, 0x0a, 0x0e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x3a, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6a, 0x35, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69,
	0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x38,
	0x0a, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18,
	0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01,
	0x52, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x6e, 0x64,
	0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75, 0x6e, 0x64, 0x6c, 0x65,
	0x12, 0x1a, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06,
	0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x9f, 0x01, 0x0a,
	0x0f, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x41, 0x50, 0x49, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x3a, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x6a, 0x35, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x06,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6a,
	0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x06,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x6e, 0x64, 0x6c, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x22, 0xdf,
	0x01, 0x0a, 0x14, 0x4a, 0x35, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x42, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6a, 0x35, 0x2e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8,
	0x01, 0x01, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x48, 0x0a, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x6a, 0x35,
	0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x74, 0x6f, 0x70,
	0x69, 0x63, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x0d,
	0xba, 0x48, 0x0a, 0xc8, 0x01, 0x01, 0x82, 0x01, 0x04, 0x10, 0x01, 0x20, 0x00, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x39, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x2e, 0x42, 0x75, 0x69,
	0x6c, 0x64, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x22, 0x5f, 0x0a, 0x0b, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12,
	0x17, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x04, 0x74, 0x65, 0x78, 0x74, 0x88, 0x01, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x74, 0x65, 0x78,
	0x74, 0x2a, 0x7d, 0x0a, 0x0b, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x1c, 0x0a, 0x18, 0x42, 0x55, 0x49, 0x4c, 0x44, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1c,
	0x0a, 0x18, 0x42, 0x55, 0x49, 0x4c, 0x44, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x49,
	0x4e, 0x5f, 0x50, 0x52, 0x4f, 0x47, 0x52, 0x45, 0x53, 0x53, 0x10, 0x01, 0x12, 0x18, 0x0a, 0x14,
	0x42, 0x55, 0x49, 0x4c, 0x44, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x53, 0x55, 0x43,
	0x43, 0x45, 0x53, 0x53, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x42, 0x55, 0x49, 0x4c, 0x44, 0x5f,
	0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x55, 0x52, 0x45, 0x10, 0x03,
	0x32, 0xc7, 0x01, 0x0a, 0x13, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x49, 0x0a, 0x07, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x73, 0x68, 0x12, 0x24, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x73, 0x68, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x00, 0x12, 0x4b, 0x0a, 0x08, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x41, 0x50, 0x49, 0x12,
	0x25, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x41, 0x50, 0x49, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00,
	0x1a, 0x18, 0xd2, 0xa2, 0xf5, 0xe4, 0x02, 0x12, 0x1a, 0x10, 0x0a, 0x0e, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2d, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x32, 0x83, 0x01, 0x0a, 0x0f, 0x42,
	0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x55,
	0x0a, 0x0d, 0x4a, 0x35, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x2a, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x2e, 0x4a, 0x35, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x00, 0x1a, 0x19, 0xd2, 0xa2, 0xf5, 0xe4, 0x02, 0x13, 0x22, 0x11, 0x0a,
	0x0f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2d, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70,
	0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35,
	0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x5f, 0x74, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_j5_registry_v1_topic_builder_proto_rawDescOnce sync.Once
	file_j5_registry_v1_topic_builder_proto_rawDescData = file_j5_registry_v1_topic_builder_proto_rawDesc
)

func file_j5_registry_v1_topic_builder_proto_rawDescGZIP() []byte {
	file_j5_registry_v1_topic_builder_proto_rawDescOnce.Do(func() {
		file_j5_registry_v1_topic_builder_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_registry_v1_topic_builder_proto_rawDescData)
	})
	return file_j5_registry_v1_topic_builder_proto_rawDescData
}

var file_j5_registry_v1_topic_builder_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_j5_registry_v1_topic_builder_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_j5_registry_v1_topic_builder_proto_goTypes = []any{
	(BuildStatus)(0),                       // 0: j5.registry.v1.topic.BuildStatus
	(*PublishMessage)(nil),                 // 1: j5.registry.v1.topic.PublishMessage
	(*BuildAPIMessage)(nil),                // 2: j5.registry.v1.topic.BuildAPIMessage
	(*J5BuildStatusMessage)(nil),           // 3: j5.registry.v1.topic.J5BuildStatusMessage
	(*BuildOutput)(nil),                    // 4: j5.registry.v1.topic.BuildOutput
	(*messaging_j5pb.RequestMetadata)(nil), // 5: j5.messaging.v1.RequestMetadata
	(*source_j5pb.CommitInfo)(nil),         // 6: j5.source.v1.CommitInfo
	(*emptypb.Empty)(nil),                  // 7: google.protobuf.Empty
}
var file_j5_registry_v1_topic_builder_proto_depIdxs = []int32{
	5,  // 0: j5.registry.v1.topic.PublishMessage.request:type_name -> j5.messaging.v1.RequestMetadata
	6,  // 1: j5.registry.v1.topic.PublishMessage.commit:type_name -> j5.source.v1.CommitInfo
	5,  // 2: j5.registry.v1.topic.BuildAPIMessage.request:type_name -> j5.messaging.v1.RequestMetadata
	6,  // 3: j5.registry.v1.topic.BuildAPIMessage.commit:type_name -> j5.source.v1.CommitInfo
	5,  // 4: j5.registry.v1.topic.J5BuildStatusMessage.request:type_name -> j5.messaging.v1.RequestMetadata
	0,  // 5: j5.registry.v1.topic.J5BuildStatusMessage.status:type_name -> j5.registry.v1.topic.BuildStatus
	4,  // 6: j5.registry.v1.topic.J5BuildStatusMessage.output:type_name -> j5.registry.v1.topic.BuildOutput
	1,  // 7: j5.registry.v1.topic.BuilderRequestTopic.Publish:input_type -> j5.registry.v1.topic.PublishMessage
	2,  // 8: j5.registry.v1.topic.BuilderRequestTopic.BuildAPI:input_type -> j5.registry.v1.topic.BuildAPIMessage
	3,  // 9: j5.registry.v1.topic.BuildReplyTopic.J5BuildStatus:input_type -> j5.registry.v1.topic.J5BuildStatusMessage
	7,  // 10: j5.registry.v1.topic.BuilderRequestTopic.Publish:output_type -> google.protobuf.Empty
	7,  // 11: j5.registry.v1.topic.BuilderRequestTopic.BuildAPI:output_type -> google.protobuf.Empty
	7,  // 12: j5.registry.v1.topic.BuildReplyTopic.J5BuildStatus:output_type -> google.protobuf.Empty
	10, // [10:13] is the sub-list for method output_type
	7,  // [7:10] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_j5_registry_v1_topic_builder_proto_init() }
func file_j5_registry_v1_topic_builder_proto_init() {
	if File_j5_registry_v1_topic_builder_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_registry_v1_topic_builder_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*PublishMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_j5_registry_v1_topic_builder_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*BuildAPIMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_j5_registry_v1_topic_builder_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*J5BuildStatusMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_j5_registry_v1_topic_builder_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*BuildOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_j5_registry_v1_topic_builder_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_registry_v1_topic_builder_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_j5_registry_v1_topic_builder_proto_goTypes,
		DependencyIndexes: file_j5_registry_v1_topic_builder_proto_depIdxs,
		EnumInfos:         file_j5_registry_v1_topic_builder_proto_enumTypes,
		MessageInfos:      file_j5_registry_v1_topic_builder_proto_msgTypes,
	}.Build()
	File_j5_registry_v1_topic_builder_proto = out.File
	file_j5_registry_v1_topic_builder_proto_rawDesc = nil
	file_j5_registry_v1_topic_builder_proto_goTypes = nil
	file_j5_registry_v1_topic_builder_proto_depIdxs = nil
}
