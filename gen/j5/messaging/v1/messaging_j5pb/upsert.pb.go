// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/messaging/v1/upsert.proto

package messaging_j5pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type UpsertMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EntityId  string                 `protobuf:"bytes,1,opt,name=entity_id,json=entityId,proto3" json:"entity_id,omitempty"`
	EventId   string                 `protobuf:"bytes,2,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *UpsertMetadata) Reset() {
	*x = UpsertMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_messaging_v1_upsert_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertMetadata) ProtoMessage() {}

func (x *UpsertMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_j5_messaging_v1_upsert_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertMetadata.ProtoReflect.Descriptor instead.
func (*UpsertMetadata) Descriptor() ([]byte, []int) {
	return file_j5_messaging_v1_upsert_proto_rawDescGZIP(), []int{0}
}

func (x *UpsertMetadata) GetEntityId() string {
	if x != nil {
		return x.EntityId
	}
	return ""
}

func (x *UpsertMetadata) GetEventId() string {
	if x != nil {
		return x.EventId
	}
	return ""
}

func (x *UpsertMetadata) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

var File_j5_messaging_v1_upsert_proto protoreflect.FileDescriptor

var file_j5_messaging_v1_upsert_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6a, 0x35, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2f, 0x76,
	0x31, 0x2f, 0x75, 0x70, 0x73, 0x65, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f,
	0x6a, 0x35, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x82, 0x01, 0x0a, 0x0e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x1b, 0x0a, 0x09, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x49, 0x64,
	0x12, 0x19, 0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x38, 0x0a, 0x09, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x3a, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2f,
	0x76, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x5f, 0x6a, 0x35, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_messaging_v1_upsert_proto_rawDescOnce sync.Once
	file_j5_messaging_v1_upsert_proto_rawDescData = file_j5_messaging_v1_upsert_proto_rawDesc
)

func file_j5_messaging_v1_upsert_proto_rawDescGZIP() []byte {
	file_j5_messaging_v1_upsert_proto_rawDescOnce.Do(func() {
		file_j5_messaging_v1_upsert_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_messaging_v1_upsert_proto_rawDescData)
	})
	return file_j5_messaging_v1_upsert_proto_rawDescData
}

var file_j5_messaging_v1_upsert_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_j5_messaging_v1_upsert_proto_goTypes = []any{
	(*UpsertMetadata)(nil),        // 0: j5.messaging.v1.UpsertMetadata
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_j5_messaging_v1_upsert_proto_depIdxs = []int32{
	1, // 0: j5.messaging.v1.UpsertMetadata.timestamp:type_name -> google.protobuf.Timestamp
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_j5_messaging_v1_upsert_proto_init() }
func file_j5_messaging_v1_upsert_proto_init() {
	if File_j5_messaging_v1_upsert_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_messaging_v1_upsert_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*UpsertMetadata); i {
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
			RawDescriptor: file_j5_messaging_v1_upsert_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_messaging_v1_upsert_proto_goTypes,
		DependencyIndexes: file_j5_messaging_v1_upsert_proto_depIdxs,
		MessageInfos:      file_j5_messaging_v1_upsert_proto_msgTypes,
	}.Build()
	File_j5_messaging_v1_upsert_proto = out.File
	file_j5_messaging_v1_upsert_proto_rawDesc = nil
	file_j5_messaging_v1_upsert_proto_goTypes = nil
	file_j5_messaging_v1_upsert_proto_depIdxs = nil
}
