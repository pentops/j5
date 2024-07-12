// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/schema/v1/entity.proto

package schema_j5pb

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

type Entity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Object     *Object  `protobuf:"bytes,1,opt,name=object,proto3" json:"object,omitempty"`
	PrimaryKey []string `protobuf:"bytes,2,rep,name=primary_key,json=primaryKey,proto3" json:"primary_key,omitempty"`
}

func (x *Entity) Reset() {
	*x = Entity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_entity_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entity) ProtoMessage() {}

func (x *Entity) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_entity_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entity.ProtoReflect.Descriptor instead.
func (*Entity) Descriptor() ([]byte, []int) {
	return file_j5_schema_v1_entity_proto_rawDescGZIP(), []int{0}
}

func (x *Entity) GetObject() *Object {
	if x != nil {
		return x.Object
	}
	return nil
}

func (x *Entity) GetPrimaryKey() []string {
	if x != nil {
		return x.PrimaryKey
	}
	return nil
}

var File_j5_schema_v1_entity_proto protoreflect.FileDescriptor

var file_j5_schema_v1_entity_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6a, 0x35, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6a, 0x35, 0x2e,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x1a, 0x19, 0x6a, 0x35, 0x2f, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x57, 0x0a, 0x06, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x2c,
	0x0a, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x1f, 0x0a, 0x0b,
	0x70, 0x72, 0x69, 0x6d, 0x61, 0x72, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x0a, 0x70, 0x72, 0x69, 0x6d, 0x61, 0x72, 0x79, 0x4b, 0x65, 0x79, 0x42, 0x34, 0x5a,
	0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74,
	0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x5f, 0x6a,
	0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_schema_v1_entity_proto_rawDescOnce sync.Once
	file_j5_schema_v1_entity_proto_rawDescData = file_j5_schema_v1_entity_proto_rawDesc
)

func file_j5_schema_v1_entity_proto_rawDescGZIP() []byte {
	file_j5_schema_v1_entity_proto_rawDescOnce.Do(func() {
		file_j5_schema_v1_entity_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_schema_v1_entity_proto_rawDescData)
	})
	return file_j5_schema_v1_entity_proto_rawDescData
}

var file_j5_schema_v1_entity_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_j5_schema_v1_entity_proto_goTypes = []any{
	(*Entity)(nil), // 0: j5.schema.v1.Entity
	(*Object)(nil), // 1: j5.schema.v1.Object
}
var file_j5_schema_v1_entity_proto_depIdxs = []int32{
	1, // 0: j5.schema.v1.Entity.object:type_name -> j5.schema.v1.Object
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_j5_schema_v1_entity_proto_init() }
func file_j5_schema_v1_entity_proto_init() {
	if File_j5_schema_v1_entity_proto != nil {
		return
	}
	file_j5_schema_v1_schema_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_j5_schema_v1_entity_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Entity); i {
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
			RawDescriptor: file_j5_schema_v1_entity_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_schema_v1_entity_proto_goTypes,
		DependencyIndexes: file_j5_schema_v1_entity_proto_depIdxs,
		MessageInfos:      file_j5_schema_v1_entity_proto_msgTypes,
	}.Build()
	File_j5_schema_v1_entity_proto = out.File
	file_j5_schema_v1_entity_proto_rawDesc = nil
	file_j5_schema_v1_entity_proto_goTypes = nil
	file_j5_schema_v1_entity_proto_depIdxs = nil
}
