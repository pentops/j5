// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: j5/ext/v1/annotations.proto

package ext_j5pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StringFormat int32

const (
	StringFormat_STRING_FORMAT_UNSPECIFIED StringFormat = 0
	StringFormat_STRING_FORMAT_DATE        StringFormat = 1
)

// Enum value maps for StringFormat.
var (
	StringFormat_name = map[int32]string{
		0: "STRING_FORMAT_UNSPECIFIED",
		1: "STRING_FORMAT_DATE",
	}
	StringFormat_value = map[string]int32{
		"STRING_FORMAT_UNSPECIFIED": 0,
		"STRING_FORMAT_DATE":        1,
	}
)

func (x StringFormat) Enum() *StringFormat {
	p := new(StringFormat)
	*p = x
	return p
}

func (x StringFormat) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (StringFormat) Descriptor() protoreflect.EnumDescriptor {
	return file_j5_ext_v1_annotations_proto_enumTypes[0].Descriptor()
}

func (StringFormat) Type() protoreflect.EnumType {
	return &file_j5_ext_v1_annotations_proto_enumTypes[0]
}

func (x StringFormat) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use StringFormat.Descriptor instead.
func (StringFormat) EnumDescriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{0}
}

type MessageOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// When true, all fields in this message should be wrapped in a single oneof
	// field. The message will show in json-schema as-is but with the
	// x-oneof flag set.
	IsOneofWrapper bool `protobuf:"varint,1,opt,name=is_oneof_wrapper,json=isOneofWrapper,proto3" json:"is_oneof_wrapper,omitempty"`
}

func (x *MessageOptions) Reset() {
	*x = MessageOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_ext_v1_annotations_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageOptions) ProtoMessage() {}

func (x *MessageOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_ext_v1_annotations_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageOptions.ProtoReflect.Descriptor instead.
func (*MessageOptions) Descriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{0}
}

func (x *MessageOptions) GetIsOneofWrapper() bool {
	if x != nil {
		return x.IsOneofWrapper
	}
	return false
}

type OneofOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// When true, the oneof is exposed as a field in the parent message, rather
	// than being a validation rule.
	// Will show in json-schema as an object with the x-oneof flag set.
	Expose bool `protobuf:"varint,1,opt,name=expose,proto3" json:"expose,omitempty"`
}

func (x *OneofOptions) Reset() {
	*x = OneofOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_ext_v1_annotations_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OneofOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OneofOptions) ProtoMessage() {}

func (x *OneofOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_ext_v1_annotations_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OneofOptions.ProtoReflect.Descriptor instead.
func (*OneofOptions) Descriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{1}
}

func (x *OneofOptions) GetExpose() bool {
	if x != nil {
		return x.Expose
	}
	return false
}

type MethodOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Label  string `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	Hidden bool   `protobuf:"varint,2,opt,name=hidden,proto3" json:"hidden,omitempty"`
}

func (x *MethodOptions) Reset() {
	*x = MethodOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_ext_v1_annotations_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MethodOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MethodOptions) ProtoMessage() {}

func (x *MethodOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_ext_v1_annotations_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MethodOptions.ProtoReflect.Descriptor instead.
func (*MethodOptions) Descriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{2}
}

func (x *MethodOptions) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *MethodOptions) GetHidden() bool {
	if x != nil {
		return x.Hidden
	}
	return false
}

type FieldOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*FieldOptions_String_
	//	*FieldOptions_Message
	Type isFieldOptions_Type `protobuf_oneof:"type"`
}

func (x *FieldOptions) Reset() {
	*x = FieldOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_ext_v1_annotations_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldOptions) ProtoMessage() {}

func (x *FieldOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_ext_v1_annotations_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldOptions.ProtoReflect.Descriptor instead.
func (*FieldOptions) Descriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{3}
}

func (m *FieldOptions) GetType() isFieldOptions_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *FieldOptions) GetString_() *StringFieldOptions {
	if x, ok := x.GetType().(*FieldOptions_String_); ok {
		return x.String_
	}
	return nil
}

func (x *FieldOptions) GetMessage() *MessageFieldOptions {
	if x, ok := x.GetType().(*FieldOptions_Message); ok {
		return x.Message
	}
	return nil
}

type isFieldOptions_Type interface {
	isFieldOptions_Type()
}

type FieldOptions_String_ struct {
	String_ *StringFieldOptions `protobuf:"bytes,1,opt,name=string,proto3,oneof"`
}

type FieldOptions_Message struct {
	Message *MessageFieldOptions `protobuf:"bytes,2,opt,name=message,proto3,oneof"`
}

func (*FieldOptions_String_) isFieldOptions_Type() {}

func (*FieldOptions_Message) isFieldOptions_Type() {}

type StringFieldOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Format StringFormat `protobuf:"varint,1,opt,name=format,proto3,enum=j5.ext.v1.StringFormat" json:"format,omitempty"`
}

func (x *StringFieldOptions) Reset() {
	*x = StringFieldOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_ext_v1_annotations_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringFieldOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringFieldOptions) ProtoMessage() {}

func (x *StringFieldOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_ext_v1_annotations_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringFieldOptions.ProtoReflect.Descriptor instead.
func (*StringFieldOptions) Descriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{4}
}

func (x *StringFieldOptions) GetFormat() StringFormat {
	if x != nil {
		return x.Format
	}
	return StringFormat_STRING_FORMAT_UNSPECIFIED
}

type MessageFieldOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// When true, the fields of the child message are flattened into the parent message.
	Flatten bool `protobuf:"varint,1,opt,name=flatten,proto3" json:"flatten,omitempty"`
}

func (x *MessageFieldOptions) Reset() {
	*x = MessageFieldOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_ext_v1_annotations_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageFieldOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageFieldOptions) ProtoMessage() {}

func (x *MessageFieldOptions) ProtoReflect() protoreflect.Message {
	mi := &file_j5_ext_v1_annotations_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageFieldOptions.ProtoReflect.Descriptor instead.
func (*MessageFieldOptions) Descriptor() ([]byte, []int) {
	return file_j5_ext_v1_annotations_proto_rawDescGZIP(), []int{5}
}

func (x *MessageFieldOptions) GetFlatten() bool {
	if x != nil {
		return x.Flatten
	}
	return false
}

var file_j5_ext_v1_annotations_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*MessageOptions)(nil),
		Field:         90443353,
		Name:          "j5.ext.v1.message",
		Tag:           "bytes,90443353,opt,name=message",
		Filename:      "j5/ext/v1/annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.OneofOptions)(nil),
		ExtensionType: (*OneofOptions)(nil),
		Field:         90443355,
		Name:          "j5.ext.v1.oneof",
		Tag:           "bytes,90443355,opt,name=oneof",
		Filename:      "j5/ext/v1/annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*MethodOptions)(nil),
		Field:         90443356,
		Name:          "j5.ext.v1.method",
		Tag:           "bytes,90443356,opt,name=method",
		Filename:      "j5/ext/v1/annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*FieldOptions)(nil),
		Field:         90443357,
		Name:          "j5.ext.v1.field",
		Tag:           "bytes,90443357,opt,name=field",
		Filename:      "j5/ext/v1/annotations.proto",
	},
}

// Extension fields to descriptorpb.MessageOptions.
var (
	// optional j5.ext.v1.MessageOptions message = 90443353;
	E_Message = &file_j5_ext_v1_annotations_proto_extTypes[0]
)

// Extension fields to descriptorpb.OneofOptions.
var (
	// optional j5.ext.v1.OneofOptions oneof = 90443355;
	E_Oneof = &file_j5_ext_v1_annotations_proto_extTypes[1]
)

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional j5.ext.v1.MethodOptions method = 90443356;
	E_Method = &file_j5_ext_v1_annotations_proto_extTypes[2]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional j5.ext.v1.FieldOptions field = 90443357;
	E_Field = &file_j5_ext_v1_annotations_proto_extTypes[3]
)

var File_j5_ext_v1_annotations_proto protoreflect.FileDescriptor

var file_j5_ext_v1_annotations_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x6a, 0x35, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6a,
	0x35, 0x2e, 0x65, 0x78, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3a, 0x0a, 0x0e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x28, 0x0a, 0x10,
	0x69, 0x73, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x69, 0x73, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x57,
	0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x22, 0x26, 0x0a, 0x0c, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x22, 0x3d,
	0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x69, 0x64, 0x64, 0x65, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x68, 0x69, 0x64, 0x64, 0x65, 0x6e, 0x22, 0x8b, 0x01,
	0x0a, 0x0c, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x37,
	0x0a, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d,
	0x2e, 0x6a, 0x35, 0x2e, 0x65, 0x78, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x48, 0x00, 0x52,
	0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x3a, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x6a, 0x35, 0x2e, 0x65, 0x78,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x48, 0x00, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x45, 0x0a, 0x12, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x2f, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x65, 0x78, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x74, 0x22, 0x2f, 0x0a, 0x13, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x6c, 0x61,
	0x74, 0x74, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x66, 0x6c, 0x61, 0x74,
	0x74, 0x65, 0x6e, 0x2a, 0x45, 0x0a, 0x0c, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x46, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x12, 0x1d, 0x0a, 0x19, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x5f, 0x46, 0x4f,
	0x52, 0x4d, 0x41, 0x54, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44,
	0x10, 0x00, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x5f, 0x46, 0x4f, 0x52,
	0x4d, 0x41, 0x54, 0x5f, 0x44, 0x41, 0x54, 0x45, 0x10, 0x01, 0x3a, 0x57, 0x0a, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd9, 0x9c, 0x90, 0x2b, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x6a, 0x35, 0x2e, 0x65, 0x78, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x3a, 0x4f, 0x0a, 0x05, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x12, 0x1d, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4f,
	0x6e, 0x65, 0x6f, 0x66, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xdb, 0x9c, 0x90, 0x2b,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x65, 0x78, 0x74, 0x2e, 0x76, 0x31,
	0x2e, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x05, 0x6f,
	0x6e, 0x65, 0x6f, 0x66, 0x3a, 0x53, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x1e,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xdc,
	0x9c, 0x90, 0x2b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6a, 0x35, 0x2e, 0x65, 0x78, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x3a, 0x4f, 0x0a, 0x05, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0xdd, 0x9c, 0x90, 0x2b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e,
	0x65, 0x78, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x52, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73,
	0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f,
	0x65, 0x78, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x78, 0x74, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_ext_v1_annotations_proto_rawDescOnce sync.Once
	file_j5_ext_v1_annotations_proto_rawDescData = file_j5_ext_v1_annotations_proto_rawDesc
)

func file_j5_ext_v1_annotations_proto_rawDescGZIP() []byte {
	file_j5_ext_v1_annotations_proto_rawDescOnce.Do(func() {
		file_j5_ext_v1_annotations_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_ext_v1_annotations_proto_rawDescData)
	})
	return file_j5_ext_v1_annotations_proto_rawDescData
}

var file_j5_ext_v1_annotations_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_j5_ext_v1_annotations_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_j5_ext_v1_annotations_proto_goTypes = []interface{}{
	(StringFormat)(0),                   // 0: j5.ext.v1.StringFormat
	(*MessageOptions)(nil),              // 1: j5.ext.v1.MessageOptions
	(*OneofOptions)(nil),                // 2: j5.ext.v1.OneofOptions
	(*MethodOptions)(nil),               // 3: j5.ext.v1.MethodOptions
	(*FieldOptions)(nil),                // 4: j5.ext.v1.FieldOptions
	(*StringFieldOptions)(nil),          // 5: j5.ext.v1.StringFieldOptions
	(*MessageFieldOptions)(nil),         // 6: j5.ext.v1.MessageFieldOptions
	(*descriptorpb.MessageOptions)(nil), // 7: google.protobuf.MessageOptions
	(*descriptorpb.OneofOptions)(nil),   // 8: google.protobuf.OneofOptions
	(*descriptorpb.MethodOptions)(nil),  // 9: google.protobuf.MethodOptions
	(*descriptorpb.FieldOptions)(nil),   // 10: google.protobuf.FieldOptions
}
var file_j5_ext_v1_annotations_proto_depIdxs = []int32{
	5,  // 0: j5.ext.v1.FieldOptions.string:type_name -> j5.ext.v1.StringFieldOptions
	6,  // 1: j5.ext.v1.FieldOptions.message:type_name -> j5.ext.v1.MessageFieldOptions
	0,  // 2: j5.ext.v1.StringFieldOptions.format:type_name -> j5.ext.v1.StringFormat
	7,  // 3: j5.ext.v1.message:extendee -> google.protobuf.MessageOptions
	8,  // 4: j5.ext.v1.oneof:extendee -> google.protobuf.OneofOptions
	9,  // 5: j5.ext.v1.method:extendee -> google.protobuf.MethodOptions
	10, // 6: j5.ext.v1.field:extendee -> google.protobuf.FieldOptions
	1,  // 7: j5.ext.v1.message:type_name -> j5.ext.v1.MessageOptions
	2,  // 8: j5.ext.v1.oneof:type_name -> j5.ext.v1.OneofOptions
	3,  // 9: j5.ext.v1.method:type_name -> j5.ext.v1.MethodOptions
	4,  // 10: j5.ext.v1.field:type_name -> j5.ext.v1.FieldOptions
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	7,  // [7:11] is the sub-list for extension type_name
	3,  // [3:7] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_j5_ext_v1_annotations_proto_init() }
func file_j5_ext_v1_annotations_proto_init() {
	if File_j5_ext_v1_annotations_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_ext_v1_annotations_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageOptions); i {
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
		file_j5_ext_v1_annotations_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OneofOptions); i {
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
		file_j5_ext_v1_annotations_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MethodOptions); i {
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
		file_j5_ext_v1_annotations_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldOptions); i {
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
		file_j5_ext_v1_annotations_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringFieldOptions); i {
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
		file_j5_ext_v1_annotations_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageFieldOptions); i {
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
	file_j5_ext_v1_annotations_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*FieldOptions_String_)(nil),
		(*FieldOptions_Message)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_ext_v1_annotations_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 4,
			NumServices:   0,
		},
		GoTypes:           file_j5_ext_v1_annotations_proto_goTypes,
		DependencyIndexes: file_j5_ext_v1_annotations_proto_depIdxs,
		EnumInfos:         file_j5_ext_v1_annotations_proto_enumTypes,
		MessageInfos:      file_j5_ext_v1_annotations_proto_msgTypes,
		ExtensionInfos:    file_j5_ext_v1_annotations_proto_extTypes,
	}.Build()
	File_j5_ext_v1_annotations_proto = out.File
	file_j5_ext_v1_annotations_proto_rawDesc = nil
	file_j5_ext_v1_annotations_proto_goTypes = nil
	file_j5_ext_v1_annotations_proto_depIdxs = nil
}
