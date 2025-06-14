// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/source/v1/image.proto

package source_j5pb

import (
	reflect "reflect"
	sync "sync"

	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Image is a parsed source image, similar to google.protobuf.Descriptor but
// with the J5 config, and some non-proto files
type SourceImage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	File            []*descriptorpb.FileDescriptorProto `protobuf:"bytes,1,rep,name=file,proto3" json:"file,omitempty"`
	Packages        []*PackageInfo                      `protobuf:"bytes,2,rep,name=packages,proto3" json:"packages,omitempty"`
	Prose           []*ProseFile                        `protobuf:"bytes,3,rep,name=prose,proto3" json:"prose,omitempty"`
	Options         *FakeOptions                        `protobuf:"bytes,4,opt,name=options,proto3" json:"options,omitempty"`
	SourceFilenames []string                            `protobuf:"bytes,6,rep,name=source_filenames,json=sourceFilenames,proto3" json:"source_filenames,omitempty"`
	SourceName      string                              `protobuf:"bytes,8,opt,name=source_name,json=sourceName,proto3" json:"source_name,omitempty"`
	Version         *string                             `protobuf:"bytes,7,opt,name=version,proto3,oneof" json:"version,omitempty"`
}

func (x *SourceImage) Reset() {
	*x = SourceImage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SourceImage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SourceImage) ProtoMessage() {}

func (x *SourceImage) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SourceImage.ProtoReflect.Descriptor instead.
func (*SourceImage) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{0}
}

func (x *SourceImage) GetFile() []*descriptorpb.FileDescriptorProto {
	if x != nil {
		return x.File
	}
	return nil
}

func (x *SourceImage) GetPackages() []*PackageInfo {
	if x != nil {
		return x.Packages
	}
	return nil
}

func (x *SourceImage) GetProse() []*ProseFile {
	if x != nil {
		return x.Prose
	}
	return nil
}

func (x *SourceImage) GetOptions() *FakeOptions {
	if x != nil {
		return x.Options
	}
	return nil
}

func (x *SourceImage) GetSourceFilenames() []string {
	if x != nil {
		return x.SourceFilenames
	}
	return nil
}

func (x *SourceImage) GetSourceName() string {
	if x != nil {
		return x.SourceName
	}
	return ""
}

func (x *SourceImage) GetVersion() string {
	if x != nil && x.Version != nil {
		return *x.Version
	}
	return ""
}

// DEPRECATED: This isn't required.
type FakeOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SubPackages []*FakeOptions_SubPackageType `protobuf:"bytes,1,rep,name=sub_packages,json=subPackages,proto3" json:"sub_packages,omitempty"`
	Go          *FakeOptions_GoPackageOptions `protobuf:"bytes,2,opt,name=go,proto3,oneof" json:"go,omitempty"`
}

func (x *FakeOptions) Reset() {
	*x = FakeOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FakeOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeOptions) ProtoMessage() {}

func (x *FakeOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeOptions.ProtoReflect.Descriptor instead.
func (*FakeOptions) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{1}
}

func (x *FakeOptions) GetSubPackages() []*FakeOptions_SubPackageType {
	if x != nil {
		return x.SubPackages
	}
	return nil
}

func (x *FakeOptions) GetGo() *FakeOptions_GoPackageOptions {
	if x != nil {
		return x.Go
	}
	return nil
}

type ProseFile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path    string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Content []byte `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *ProseFile) Reset() {
	*x = ProseFile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProseFile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProseFile) ProtoMessage() {}

func (x *ProseFile) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProseFile.ProtoReflect.Descriptor instead.
func (*ProseFile) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{2}
}

func (x *ProseFile) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *ProseFile) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type CommitInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner   string                 `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repo    string                 `protobuf:"bytes,2,opt,name=repo,proto3" json:"repo,omitempty"`
	Hash    string                 `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	Time    *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=time,proto3" json:"time,omitempty"`
	Aliases []string               `protobuf:"bytes,5,rep,name=aliases,proto3" json:"aliases,omitempty"`
}

func (x *CommitInfo) Reset() {
	*x = CommitInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitInfo) ProtoMessage() {}

func (x *CommitInfo) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitInfo.ProtoReflect.Descriptor instead.
func (*CommitInfo) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{3}
}

func (x *CommitInfo) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *CommitInfo) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *CommitInfo) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *CommitInfo) GetTime() *timestamppb.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *CommitInfo) GetAliases() []string {
	if x != nil {
		return x.Aliases
	}
	return nil
}

type PackageInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Label string `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Prose string `protobuf:"bytes,3,opt,name=prose,proto3" json:"prose,omitempty"`
}

func (x *PackageInfo) Reset() {
	*x = PackageInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PackageInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackageInfo) ProtoMessage() {}

func (x *PackageInfo) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackageInfo.ProtoReflect.Descriptor instead.
func (*PackageInfo) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{4}
}

func (x *PackageInfo) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *PackageInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PackageInfo) GetProse() string {
	if x != nil {
		return x.Prose
	}
	return ""
}

type FakeOptions_SubPackageType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *FakeOptions_SubPackageType) Reset() {
	*x = FakeOptions_SubPackageType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FakeOptions_SubPackageType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeOptions_SubPackageType) ProtoMessage() {}

func (x *FakeOptions_SubPackageType) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeOptions_SubPackageType.ProtoReflect.Descriptor instead.
func (*FakeOptions_SubPackageType) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{1, 0}
}

func (x *FakeOptions_SubPackageType) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type FakeOptions_GoPackageOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix       string   `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	TrimPrefixes []string `protobuf:"bytes,2,rep,name=trim_prefixes,json=trimPrefixes,proto3" json:"trim_prefixes,omitempty"`
	Suffix       *string  `protobuf:"bytes,3,opt,name=suffix,proto3,oneof" json:"suffix,omitempty"`
}

func (x *FakeOptions_GoPackageOptions) Reset() {
	*x = FakeOptions_GoPackageOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_image_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FakeOptions_GoPackageOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FakeOptions_GoPackageOptions) ProtoMessage() {}

func (x *FakeOptions_GoPackageOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_image_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FakeOptions_GoPackageOptions.ProtoReflect.Descriptor instead.
func (*FakeOptions_GoPackageOptions) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_image_proto_rawDescGZIP(), []int{1, 1}
}

func (x *FakeOptions_GoPackageOptions) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *FakeOptions_GoPackageOptions) GetTrimPrefixes() []string {
	if x != nil {
		return x.TrimPrefixes
	}
	return nil
}

func (x *FakeOptions_GoPackageOptions) GetSuffix() string {
	if x != nil && x.Suffix != nil {
		return *x.Suffix
	}
	return ""
}

var File_j5_source_v1_image_proto protoreflect.FileDescriptor

var file_j5_source_v1_image_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6a, 0x35, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x69,
	0x6d, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6a, 0x35, 0x2e, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd9, 0x02, 0x0a, 0x0b, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x38, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x04, 0x66, 0x69,
	0x6c, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x08, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x12, 0x2d, 0x0a, 0x05, 0x70, 0x72, 0x6f,
	0x73, 0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x73, 0x65, 0x46, 0x69, 0x6c,
	0x65, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6a, 0x35, 0x2e, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x61, 0x6b, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x29, 0x0a,
	0x10, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x46,
	0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x22, 0xc1, 0x02, 0x0a, 0x0b, 0x46, 0x61, 0x6b, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x12, 0x4b, 0x0a, 0x0c, 0x73, 0x75, 0x62, 0x5f, 0x70, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6a, 0x35, 0x2e,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x61, 0x6b, 0x65, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x53, 0x75, 0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x0b, 0x73, 0x75, 0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x12, 0x3f, 0x0a, 0x02, 0x67, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e,
	0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x61, 0x6b,
	0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x47, 0x6f, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x48, 0x00, 0x52, 0x02, 0x67, 0x6f, 0x88,
	0x01, 0x01, 0x1a, 0x24, 0x0a, 0x0e, 0x53, 0x75, 0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x1a, 0x77, 0x0a, 0x10, 0x47, 0x6f, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x16, 0x0a, 0x06,
	0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72,
	0x65, 0x66, 0x69, 0x78, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x72, 0x69, 0x6d, 0x5f, 0x70, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x74, 0x72, 0x69,
	0x6d, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x12, 0x1b, 0x0a, 0x06, 0x73, 0x75, 0x66,
	0x66, 0x69, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x73, 0x75, 0x66,
	0x66, 0x69, 0x78, 0x88, 0x01, 0x01, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x73, 0x75, 0x66, 0x66, 0x69,
	0x78, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x67, 0x6f, 0x22, 0x39, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x73,
	0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x22, 0xb4, 0x01, 0x0a, 0x0a, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x1c, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72,
	0x12, 0x1a, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06,
	0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x12, 0x1a, 0x0a, 0x04,
	0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8,
	0x01, 0x01, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x36, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x6c, 0x69, 0x61, 0x73, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x6c, 0x69, 0x61, 0x73, 0x65, 0x73, 0x22, 0x4d, 0x0a, 0x0b, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62,
	0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x73, 0x65, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f,
	0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_source_v1_image_proto_rawDescOnce sync.Once
	file_j5_source_v1_image_proto_rawDescData = file_j5_source_v1_image_proto_rawDesc
)

func file_j5_source_v1_image_proto_rawDescGZIP() []byte {
	file_j5_source_v1_image_proto_rawDescOnce.Do(func() {
		file_j5_source_v1_image_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_source_v1_image_proto_rawDescData)
	})
	return file_j5_source_v1_image_proto_rawDescData
}

var file_j5_source_v1_image_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_j5_source_v1_image_proto_goTypes = []any{
	(*SourceImage)(nil),                      // 0: j5.source.v1.SourceImage
	(*FakeOptions)(nil),                      // 1: j5.source.v1.FakeOptions
	(*ProseFile)(nil),                        // 2: j5.source.v1.ProseFile
	(*CommitInfo)(nil),                       // 3: j5.source.v1.CommitInfo
	(*PackageInfo)(nil),                      // 4: j5.source.v1.PackageInfo
	(*FakeOptions_SubPackageType)(nil),       // 5: j5.source.v1.FakeOptions.SubPackageType
	(*FakeOptions_GoPackageOptions)(nil),     // 6: j5.source.v1.FakeOptions.GoPackageOptions
	(*descriptorpb.FileDescriptorProto)(nil), // 7: google.protobuf.FileDescriptorProto
	(*timestamppb.Timestamp)(nil),            // 8: google.protobuf.Timestamp
}
var file_j5_source_v1_image_proto_depIdxs = []int32{
	7, // 0: j5.source.v1.SourceImage.file:type_name -> google.protobuf.FileDescriptorProto
	4, // 1: j5.source.v1.SourceImage.packages:type_name -> j5.source.v1.PackageInfo
	2, // 2: j5.source.v1.SourceImage.prose:type_name -> j5.source.v1.ProseFile
	1, // 3: j5.source.v1.SourceImage.options:type_name -> j5.source.v1.FakeOptions
	5, // 4: j5.source.v1.FakeOptions.sub_packages:type_name -> j5.source.v1.FakeOptions.SubPackageType
	6, // 5: j5.source.v1.FakeOptions.go:type_name -> j5.source.v1.FakeOptions.GoPackageOptions
	8, // 6: j5.source.v1.CommitInfo.time:type_name -> google.protobuf.Timestamp
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_j5_source_v1_image_proto_init() }
func file_j5_source_v1_image_proto_init() {
	if File_j5_source_v1_image_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_source_v1_image_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SourceImage); i {
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
		file_j5_source_v1_image_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*FakeOptions); i {
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
		file_j5_source_v1_image_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ProseFile); i {
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
		file_j5_source_v1_image_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*CommitInfo); i {
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
		file_j5_source_v1_image_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*PackageInfo); i {
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
		file_j5_source_v1_image_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*FakeOptions_SubPackageType); i {
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
		file_j5_source_v1_image_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*FakeOptions_GoPackageOptions); i {
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
	file_j5_source_v1_image_proto_msgTypes[0].OneofWrappers = []any{}
	file_j5_source_v1_image_proto_msgTypes[1].OneofWrappers = []any{}
	file_j5_source_v1_image_proto_msgTypes[6].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_source_v1_image_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_source_v1_image_proto_goTypes,
		DependencyIndexes: file_j5_source_v1_image_proto_depIdxs,
		MessageInfos:      file_j5_source_v1_image_proto_msgTypes,
	}.Build()
	File_j5_source_v1_image_proto = out.File
	file_j5_source_v1_image_proto_rawDesc = nil
	file_j5_source_v1_image_proto_goTypes = nil
	file_j5_source_v1_image_proto_depIdxs = nil
}
