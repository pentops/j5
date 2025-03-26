// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/config/v1/mods.proto

package config_j5pb

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

type ImageMod struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*ImageMod_GoPackageNames_
	Type isImageMod_Type `protobuf_oneof:"type"`
}

func (x *ImageMod) Reset() {
	*x = ImageMod{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_config_v1_mods_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageMod) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageMod) ProtoMessage() {}

func (x *ImageMod) ProtoReflect() protoreflect.Message {
	mi := &file_j5_config_v1_mods_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageMod.ProtoReflect.Descriptor instead.
func (*ImageMod) Descriptor() ([]byte, []int) {
	return file_j5_config_v1_mods_proto_rawDescGZIP(), []int{0}
}

func (m *ImageMod) GetType() isImageMod_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *ImageMod) GetGoPackageNames() *ImageMod_GoPackageNames {
	if x, ok := x.GetType().(*ImageMod_GoPackageNames_); ok {
		return x.GoPackageNames
	}
	return nil
}

type isImageMod_Type interface {
	isImageMod_Type()
}

type ImageMod_GoPackageNames_ struct {
	GoPackageNames *ImageMod_GoPackageNames `protobuf:"bytes,2,opt,name=go_package_names,json=goPackageNames,proto3,oneof"`
}

func (*ImageMod_GoPackageNames_) isImageMod_Type() {}

// This sets the option go_package = "{prefix}/foo/bar/v1/bar{suffix}"
// Go packages are named "{prefix}/{package_root}/{version}/{name}{suffix}"
// PackageRoot is everything up to the Version part: foo/bar
// Name is the part just before the version: bar
// Suffix comes from the suffixes map.
// Prefix comes from the Prefix field
// Package part replace '.' with '/'.
type ImageMod_GoPackageNames struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	// Maps the sub package name to a suffix.
	// If a suffix is not found, default is to take the first letter.
	// multiple sub-packages with the same suffix will 'work' but probably best
	// avoided.
	// Empty string is the package root, which also maps to just _pb by default.
	Suffixes map[string]string `protobuf:"bytes,4,rep,name=suffixes,proto3" json:"suffixes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// These are stripped off the start package name before running the rest,
	// but does not act as a filter - if the package doesn't begin with the
	// prefix it is left as-is.
	TrimPrefixes []string `protobuf:"bytes,2,rep,name=trim_prefixes,json=trimPrefixes,proto3" json:"trim_prefixes,omitempty"`
}

func (x *ImageMod_GoPackageNames) Reset() {
	*x = ImageMod_GoPackageNames{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_config_v1_mods_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageMod_GoPackageNames) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageMod_GoPackageNames) ProtoMessage() {}

func (x *ImageMod_GoPackageNames) ProtoReflect() protoreflect.Message {
	mi := &file_j5_config_v1_mods_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageMod_GoPackageNames.ProtoReflect.Descriptor instead.
func (*ImageMod_GoPackageNames) Descriptor() ([]byte, []int) {
	return file_j5_config_v1_mods_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ImageMod_GoPackageNames) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *ImageMod_GoPackageNames) GetSuffixes() map[string]string {
	if x != nil {
		return x.Suffixes
	}
	return nil
}

func (x *ImageMod_GoPackageNames) GetTrimPrefixes() []string {
	if x != nil {
		return x.TrimPrefixes
	}
	return nil
}

var File_j5_config_v1_mods_proto protoreflect.FileDescriptor

var file_j5_config_v1_mods_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6a, 0x35, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x6d,
	0x6f, 0x64, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6a, 0x35, 0x2e, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x22, 0xc3, 0x02, 0x0a, 0x08, 0x49, 0x6d, 0x61, 0x67,
	0x65, 0x4d, 0x6f, 0x64, 0x12, 0x51, 0x0a, 0x10, 0x67, 0x6f, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x6a, 0x35, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6d,
	0x61, 0x67, 0x65, 0x4d, 0x6f, 0x64, 0x2e, 0x47, 0x6f, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x73, 0x48, 0x00, 0x52, 0x0e, 0x67, 0x6f, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x1a, 0xdb, 0x01, 0x0a, 0x0e, 0x47, 0x6f, 0x50, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72,
	0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66,
	0x69, 0x78, 0x12, 0x4f, 0x0a, 0x08, 0x73, 0x75, 0x66, 0x66, 0x69, 0x78, 0x65, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x6a, 0x35, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4d, 0x6f, 0x64, 0x2e, 0x47, 0x6f, 0x50,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x2e, 0x53, 0x75, 0x66, 0x66,
	0x69, 0x78, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x75, 0x66, 0x66, 0x69,
	0x78, 0x65, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x72, 0x69, 0x6d, 0x5f, 0x70, 0x72, 0x65, 0x66,
	0x69, 0x78, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x74, 0x72, 0x69, 0x6d,
	0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x1a, 0x3b, 0x0a, 0x0d, 0x53, 0x75, 0x66, 0x66,
	0x69, 0x78, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x42, 0x34, 0x5a,
	0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74,
	0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x6a,
	0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_config_v1_mods_proto_rawDescOnce sync.Once
	file_j5_config_v1_mods_proto_rawDescData = file_j5_config_v1_mods_proto_rawDesc
)

func file_j5_config_v1_mods_proto_rawDescGZIP() []byte {
	file_j5_config_v1_mods_proto_rawDescOnce.Do(func() {
		file_j5_config_v1_mods_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_config_v1_mods_proto_rawDescData)
	})
	return file_j5_config_v1_mods_proto_rawDescData
}

var file_j5_config_v1_mods_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_j5_config_v1_mods_proto_goTypes = []any{
	(*ImageMod)(nil),                // 0: j5.config.v1.ImageMod
	(*ImageMod_GoPackageNames)(nil), // 1: j5.config.v1.ImageMod.GoPackageNames
	nil,                             // 2: j5.config.v1.ImageMod.GoPackageNames.SuffixesEntry
}
var file_j5_config_v1_mods_proto_depIdxs = []int32{
	1, // 0: j5.config.v1.ImageMod.go_package_names:type_name -> j5.config.v1.ImageMod.GoPackageNames
	2, // 1: j5.config.v1.ImageMod.GoPackageNames.suffixes:type_name -> j5.config.v1.ImageMod.GoPackageNames.SuffixesEntry
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_j5_config_v1_mods_proto_init() }
func file_j5_config_v1_mods_proto_init() {
	if File_j5_config_v1_mods_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_config_v1_mods_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*ImageMod); i {
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
		file_j5_config_v1_mods_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*ImageMod_GoPackageNames); i {
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
	file_j5_config_v1_mods_proto_msgTypes[0].OneofWrappers = []any{
		(*ImageMod_GoPackageNames_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_config_v1_mods_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_config_v1_mods_proto_goTypes,
		DependencyIndexes: file_j5_config_v1_mods_proto_depIdxs,
		MessageInfos:      file_j5_config_v1_mods_proto_msgTypes,
	}.Build()
	File_j5_config_v1_mods_proto = out.File
	file_j5_config_v1_mods_proto_rawDesc = nil
	file_j5_config_v1_mods_proto_goTypes = nil
	file_j5_config_v1_mods_proto_depIdxs = nil
}
