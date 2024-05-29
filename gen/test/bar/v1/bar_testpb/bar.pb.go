// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: test/bar/v1/bar.proto

package bar_testpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BarEnum int32

const (
	BarEnum_BAR_ENUM_UNSPECIFIED BarEnum = 0
	BarEnum_BAR_ENUM_FOO         BarEnum = 1
	BarEnum_BAR_ENUM_BAR         BarEnum = 2
)

// Enum value maps for BarEnum.
var (
	BarEnum_name = map[int32]string{
		0: "BAR_ENUM_UNSPECIFIED",
		1: "BAR_ENUM_FOO",
		2: "BAR_ENUM_BAR",
	}
	BarEnum_value = map[string]int32{
		"BAR_ENUM_UNSPECIFIED": 0,
		"BAR_ENUM_FOO":         1,
		"BAR_ENUM_BAR":         2,
	}
)

func (x BarEnum) Enum() *BarEnum {
	p := new(BarEnum)
	*p = x
	return p
}

func (x BarEnum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BarEnum) Descriptor() protoreflect.EnumDescriptor {
	return file_test_bar_v1_bar_proto_enumTypes[0].Descriptor()
}

func (BarEnum) Type() protoreflect.EnumType {
	return &file_test_bar_v1_bar_proto_enumTypes[0]
}

func (x BarEnum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BarEnum.Descriptor instead.
func (BarEnum) EnumDescriptor() ([]byte, []int) {
	return file_test_bar_v1_bar_proto_rawDescGZIP(), []int{0}
}

type Bar struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Field string `protobuf:"bytes,3,opt,name=field,proto3" json:"field,omitempty"`
}

func (x *Bar) Reset() {
	*x = Bar{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_bar_v1_bar_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Bar) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Bar) ProtoMessage() {}

func (x *Bar) ProtoReflect() protoreflect.Message {
	mi := &file_test_bar_v1_bar_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Bar.ProtoReflect.Descriptor instead.
func (*Bar) Descriptor() ([]byte, []int) {
	return file_test_bar_v1_bar_proto_rawDescGZIP(), []int{0}
}

func (x *Bar) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Bar) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Bar) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

var File_test_bar_v1_bar_proto protoreflect.FileDescriptor

var file_test_bar_v1_bar_proto_rawDesc = []byte{
	0x0a, 0x15, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x62, 0x61, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x61,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x62, 0x61,
	0x72, 0x2e, 0x76, 0x31, 0x22, 0x3f, 0x0a, 0x03, 0x42, 0x61, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x66, 0x69, 0x65, 0x6c, 0x64, 0x2a, 0x47, 0x0a, 0x07, 0x42, 0x61, 0x72, 0x45, 0x6e, 0x75, 0x6d,
	0x12, 0x18, 0x0a, 0x14, 0x42, 0x41, 0x52, 0x5f, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x42, 0x41,
	0x52, 0x5f, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x46, 0x4f, 0x4f, 0x10, 0x01, 0x12, 0x10, 0x0a, 0x0c,
	0x42, 0x41, 0x52, 0x5f, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x42, 0x41, 0x52, 0x10, 0x02, 0x42, 0x37,
	0x5a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e,
	0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e,
	0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x62, 0x61, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x61, 0x72,
	0x5f, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_test_bar_v1_bar_proto_rawDescOnce sync.Once
	file_test_bar_v1_bar_proto_rawDescData = file_test_bar_v1_bar_proto_rawDesc
)

func file_test_bar_v1_bar_proto_rawDescGZIP() []byte {
	file_test_bar_v1_bar_proto_rawDescOnce.Do(func() {
		file_test_bar_v1_bar_proto_rawDescData = protoimpl.X.CompressGZIP(file_test_bar_v1_bar_proto_rawDescData)
	})
	return file_test_bar_v1_bar_proto_rawDescData
}

var file_test_bar_v1_bar_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_test_bar_v1_bar_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_test_bar_v1_bar_proto_goTypes = []interface{}{
	(BarEnum)(0), // 0: test.bar.v1.BarEnum
	(*Bar)(nil),  // 1: test.bar.v1.Bar
}
var file_test_bar_v1_bar_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_test_bar_v1_bar_proto_init() }
func file_test_bar_v1_bar_proto_init() {
	if File_test_bar_v1_bar_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_test_bar_v1_bar_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Bar); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_test_bar_v1_bar_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_test_bar_v1_bar_proto_goTypes,
		DependencyIndexes: file_test_bar_v1_bar_proto_depIdxs,
		EnumInfos:         file_test_bar_v1_bar_proto_enumTypes,
		MessageInfos:      file_test_bar_v1_bar_proto_msgTypes,
	}.Build()
	File_test_bar_v1_bar_proto = out.File
	file_test_bar_v1_bar_proto_rawDesc = nil
	file_test_bar_v1_bar_proto_goTypes = nil
	file_test_bar_v1_bar_proto_depIdxs = nil
}
