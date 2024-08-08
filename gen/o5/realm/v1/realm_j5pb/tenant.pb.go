// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/realm/v1/tenant.proto

package realm_j5pb

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	_ "github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	psm_pb "github.com/pentops/j5/gen/psm/state/v1/psm_pb"
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

type TenantStatus int32

const (
	TenantStatus_TENANT_STATUS_UNSPECIFIED TenantStatus = 0
	TenantStatus_TENANT_STATUS_ACTIVE      TenantStatus = 1
)

// Enum value maps for TenantStatus.
var (
	TenantStatus_name = map[int32]string{
		0: "TENANT_STATUS_UNSPECIFIED",
		1: "TENANT_STATUS_ACTIVE",
	}
	TenantStatus_value = map[string]int32{
		"TENANT_STATUS_UNSPECIFIED": 0,
		"TENANT_STATUS_ACTIVE":      1,
	}
)

func (x TenantStatus) Enum() *TenantStatus {
	p := new(TenantStatus)
	*p = x
	return p
}

func (x TenantStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TenantStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_j5_realm_v1_tenant_proto_enumTypes[0].Descriptor()
}

func (TenantStatus) Type() protoreflect.EnumType {
	return &file_j5_realm_v1_tenant_proto_enumTypes[0]
}

func (x TenantStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TenantStatus.Descriptor instead.
func (TenantStatus) EnumDescriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{0}
}

type TenantKeys struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TenantId   string `protobuf:"bytes,1,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	RealmId    string `protobuf:"bytes,2,opt,name=realm_id,json=realmId,proto3" json:"realm_id,omitempty"`
	TenantType string `protobuf:"bytes,3,opt,name=tenant_type,json=tenantType,proto3" json:"tenant_type,omitempty"`
}

func (x *TenantKeys) Reset() {
	*x = TenantKeys{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantKeys) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantKeys) ProtoMessage() {}

func (x *TenantKeys) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantKeys.ProtoReflect.Descriptor instead.
func (*TenantKeys) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{0}
}

func (x *TenantKeys) GetTenantId() string {
	if x != nil {
		return x.TenantId
	}
	return ""
}

func (x *TenantKeys) GetRealmId() string {
	if x != nil {
		return x.RealmId
	}
	return ""
}

func (x *TenantKeys) GetTenantType() string {
	if x != nil {
		return x.TenantType
	}
	return ""
}

type TenantStateData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Spec *TenantSpec `protobuf:"bytes,1,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *TenantStateData) Reset() {
	*x = TenantStateData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantStateData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantStateData) ProtoMessage() {}

func (x *TenantStateData) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantStateData.ProtoReflect.Descriptor instead.
func (*TenantStateData) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{1}
}

func (x *TenantStateData) GetSpec() *TenantSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

type TenantState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata *psm_pb.StateMetadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Keys     *TenantKeys           `protobuf:"bytes,2,opt,name=keys,proto3" json:"keys,omitempty"`
	Status   TenantStatus          `protobuf:"varint,3,opt,name=status,proto3,enum=j5.realm.v1.TenantStatus" json:"status,omitempty"`
	Data     *TenantStateData      `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *TenantState) Reset() {
	*x = TenantState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantState) ProtoMessage() {}

func (x *TenantState) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantState.ProtoReflect.Descriptor instead.
func (*TenantState) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{2}
}

func (x *TenantState) GetMetadata() *psm_pb.StateMetadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *TenantState) GetKeys() *TenantKeys {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *TenantState) GetStatus() TenantStatus {
	if x != nil {
		return x.Status
	}
	return TenantStatus_TENANT_STATUS_UNSPECIFIED
}

func (x *TenantState) GetData() *TenantStateData {
	if x != nil {
		return x.Data
	}
	return nil
}

type TenantSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Key-value pairs of metadata interpreted in the context of the realm and
	// tenant-type within the realm
	Metadata map[string]string `protobuf:"bytes,10,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *TenantSpec) Reset() {
	*x = TenantSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantSpec) ProtoMessage() {}

func (x *TenantSpec) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantSpec.ProtoReflect.Descriptor instead.
func (*TenantSpec) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{3}
}

func (x *TenantSpec) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TenantSpec) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type TenantEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata *psm_pb.EventMetadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Keys     *TenantKeys           `protobuf:"bytes,2,opt,name=keys,proto3" json:"keys,omitempty"`
	Event    *TenantEventType      `protobuf:"bytes,3,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *TenantEvent) Reset() {
	*x = TenantEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantEvent) ProtoMessage() {}

func (x *TenantEvent) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantEvent.ProtoReflect.Descriptor instead.
func (*TenantEvent) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{4}
}

func (x *TenantEvent) GetMetadata() *psm_pb.EventMetadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *TenantEvent) GetKeys() *TenantKeys {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *TenantEvent) GetEvent() *TenantEventType {
	if x != nil {
		return x.Event
	}
	return nil
}

type TenantEventType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*TenantEventType_Created_
	//	*TenantEventType_Updated_
	Type isTenantEventType_Type `protobuf_oneof:"type"`
}

func (x *TenantEventType) Reset() {
	*x = TenantEventType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantEventType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantEventType) ProtoMessage() {}

func (x *TenantEventType) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantEventType.ProtoReflect.Descriptor instead.
func (*TenantEventType) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{5}
}

func (m *TenantEventType) GetType() isTenantEventType_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *TenantEventType) GetCreated() *TenantEventType_Created {
	if x, ok := x.GetType().(*TenantEventType_Created_); ok {
		return x.Created
	}
	return nil
}

func (x *TenantEventType) GetUpdated() *TenantEventType_Updated {
	if x, ok := x.GetType().(*TenantEventType_Updated_); ok {
		return x.Updated
	}
	return nil
}

type isTenantEventType_Type interface {
	isTenantEventType_Type()
}

type TenantEventType_Created_ struct {
	Created *TenantEventType_Created `protobuf:"bytes,1,opt,name=created,proto3,oneof"`
}

type TenantEventType_Updated_ struct {
	Updated *TenantEventType_Updated `protobuf:"bytes,2,opt,name=updated,proto3,oneof"`
}

func (*TenantEventType_Created_) isTenantEventType_Type() {}

func (*TenantEventType_Updated_) isTenantEventType_Type() {}

type TenantEventType_Created struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Spec *TenantSpec `protobuf:"bytes,1,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *TenantEventType_Created) Reset() {
	*x = TenantEventType_Created{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantEventType_Created) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantEventType_Created) ProtoMessage() {}

func (x *TenantEventType_Created) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantEventType_Created.ProtoReflect.Descriptor instead.
func (*TenantEventType_Created) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{5, 0}
}

func (x *TenantEventType_Created) GetSpec() *TenantSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

type TenantEventType_Updated struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Spec *TenantSpec `protobuf:"bytes,1,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *TenantEventType_Updated) Reset() {
	*x = TenantEventType_Updated{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_realm_v1_tenant_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TenantEventType_Updated) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TenantEventType_Updated) ProtoMessage() {}

func (x *TenantEventType_Updated) ProtoReflect() protoreflect.Message {
	mi := &file_j5_realm_v1_tenant_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TenantEventType_Updated.ProtoReflect.Descriptor instead.
func (*TenantEventType_Updated) Descriptor() ([]byte, []int) {
	return file_j5_realm_v1_tenant_proto_rawDescGZIP(), []int{5, 1}
}

func (x *TenantEventType_Updated) GetSpec() *TenantSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

var File_j5_realm_v1_tenant_proto protoreflect.FileDescriptor

var file_j5_realm_v1_tenant_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6a, 0x35, 0x2f, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x65,
	0x6e, 0x61, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6a, 0x35, 0x2e, 0x72,
	0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x6a, 0x35, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x76, 0x31, 0x2f,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1b, 0x70, 0x73, 0x6d, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa9,
	0x01, 0x0a, 0x0a, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x2c, 0x0a,
	0x09, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x0f, 0xba, 0x48, 0x05, 0x72, 0x03, 0xb0, 0x01, 0x01, 0xea, 0x85, 0x8f, 0x02, 0x02, 0x08,
	0x01, 0x52, 0x08, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x08, 0x72,
	0x65, 0x61, 0x6c, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xba,
	0x48, 0x05, 0x72, 0x03, 0xb0, 0x01, 0x01, 0x52, 0x07, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x49, 0x64,
	0x12, 0x39, 0x0a, 0x0b, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xba, 0x48, 0x10, 0x72, 0x0e, 0x32, 0x0c, 0x5e, 0x5b,
	0x61, 0x2d, 0x7a, 0x30, 0x2d, 0x39, 0x2d, 0x5d, 0x2b, 0x24, 0xea, 0x85, 0x8f, 0x02, 0x00, 0x52,
	0x0a, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x3a, 0x0d, 0xea, 0x85, 0x8f,
	0x02, 0x08, 0x0a, 0x06, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x22, 0x3e, 0x0a, 0x0f, 0x54, 0x65,
	0x6e, 0x61, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x2b, 0x0a,
	0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35,
	0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74,
	0x53, 0x70, 0x65, 0x63, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x22, 0x81, 0x02, 0x0a, 0x0b, 0x54,
	0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x3f, 0x0a, 0x08, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x70,
	0x73, 0x6d, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01,
	0x01, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x3c, 0x0a, 0x04, 0x6b,
	0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x72,
	0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x4b, 0x65,
	0x79, 0x73, 0x42, 0x0f, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0xc2, 0xff, 0x8e, 0x02, 0x04, 0x0a,
	0x02, 0x08, 0x01, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x39, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x6a, 0x35, 0x2e, 0x72,
	0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x38, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31,
	0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61,
	0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xa0,
	0x01, 0x0a, 0x0a, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x70, 0x65, 0x63, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x41, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x0a, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x70, 0x65, 0x63, 0x2e, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0xc8, 0x01, 0x0a, 0x0b, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x3f, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x70, 0x73, 0x6d, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x3c, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x65, 0x6e, 0x61, 0x6e, 0x74, 0x4b, 0x65, 0x79, 0x73, 0x42, 0x0f, 0xba, 0x48, 0x03, 0xc8, 0x01,
	0x01, 0xc2, 0xff, 0x8e, 0x02, 0x04, 0x0a, 0x02, 0x08, 0x01, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73,
	0x12, 0x3a, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65,
	0x6e, 0x61, 0x6e, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x42, 0x06, 0xba,
	0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x8d, 0x02, 0x0a,
	0x0f, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x40, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x24, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e,
	0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x48, 0x00, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x12, 0x40, 0x0a, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70,
	0x65, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x48, 0x00, 0x52, 0x07, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x64, 0x1a, 0x36, 0x0a, 0x07, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12,
	0x2b, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61,
	0x6e, 0x74, 0x53, 0x70, 0x65, 0x63, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x1a, 0x36, 0x0a, 0x07,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x12, 0x2b, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x61, 0x6c, 0x6d,
	0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x70, 0x65, 0x63, 0x52, 0x04,
	0x73, 0x70, 0x65, 0x63, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x2a, 0x47, 0x0a, 0x0c,
	0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a, 0x19,
	0x54, 0x45, 0x4e, 0x41, 0x4e, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x18, 0x0a, 0x14, 0x54,
	0x45, 0x4e, 0x41, 0x4e, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x41, 0x43, 0x54,
	0x49, 0x56, 0x45, 0x10, 0x01, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x6f, 0x35, 0x2f, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x72,
	0x65, 0x61, 0x6c, 0x6d, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_j5_realm_v1_tenant_proto_rawDescOnce sync.Once
	file_j5_realm_v1_tenant_proto_rawDescData = file_j5_realm_v1_tenant_proto_rawDesc
)

func file_j5_realm_v1_tenant_proto_rawDescGZIP() []byte {
	file_j5_realm_v1_tenant_proto_rawDescOnce.Do(func() {
		file_j5_realm_v1_tenant_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_realm_v1_tenant_proto_rawDescData)
	})
	return file_j5_realm_v1_tenant_proto_rawDescData
}

var file_j5_realm_v1_tenant_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_j5_realm_v1_tenant_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_j5_realm_v1_tenant_proto_goTypes = []any{
	(TenantStatus)(0),               // 0: j5.realm.v1.TenantStatus
	(*TenantKeys)(nil),              // 1: j5.realm.v1.TenantKeys
	(*TenantStateData)(nil),         // 2: j5.realm.v1.TenantStateData
	(*TenantState)(nil),             // 3: j5.realm.v1.TenantState
	(*TenantSpec)(nil),              // 4: j5.realm.v1.TenantSpec
	(*TenantEvent)(nil),             // 5: j5.realm.v1.TenantEvent
	(*TenantEventType)(nil),         // 6: j5.realm.v1.TenantEventType
	nil,                             // 7: j5.realm.v1.TenantSpec.MetadataEntry
	(*TenantEventType_Created)(nil), // 8: j5.realm.v1.TenantEventType.Created
	(*TenantEventType_Updated)(nil), // 9: j5.realm.v1.TenantEventType.Updated
	(*psm_pb.StateMetadata)(nil),    // 10: psm.state.v1.StateMetadata
	(*psm_pb.EventMetadata)(nil),    // 11: psm.state.v1.EventMetadata
}
var file_j5_realm_v1_tenant_proto_depIdxs = []int32{
	4,  // 0: j5.realm.v1.TenantStateData.spec:type_name -> j5.realm.v1.TenantSpec
	10, // 1: j5.realm.v1.TenantState.metadata:type_name -> psm.state.v1.StateMetadata
	1,  // 2: j5.realm.v1.TenantState.keys:type_name -> j5.realm.v1.TenantKeys
	0,  // 3: j5.realm.v1.TenantState.status:type_name -> j5.realm.v1.TenantStatus
	2,  // 4: j5.realm.v1.TenantState.data:type_name -> j5.realm.v1.TenantStateData
	7,  // 5: j5.realm.v1.TenantSpec.metadata:type_name -> j5.realm.v1.TenantSpec.MetadataEntry
	11, // 6: j5.realm.v1.TenantEvent.metadata:type_name -> psm.state.v1.EventMetadata
	1,  // 7: j5.realm.v1.TenantEvent.keys:type_name -> j5.realm.v1.TenantKeys
	6,  // 8: j5.realm.v1.TenantEvent.event:type_name -> j5.realm.v1.TenantEventType
	8,  // 9: j5.realm.v1.TenantEventType.created:type_name -> j5.realm.v1.TenantEventType.Created
	9,  // 10: j5.realm.v1.TenantEventType.updated:type_name -> j5.realm.v1.TenantEventType.Updated
	4,  // 11: j5.realm.v1.TenantEventType.Created.spec:type_name -> j5.realm.v1.TenantSpec
	4,  // 12: j5.realm.v1.TenantEventType.Updated.spec:type_name -> j5.realm.v1.TenantSpec
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_j5_realm_v1_tenant_proto_init() }
func file_j5_realm_v1_tenant_proto_init() {
	if File_j5_realm_v1_tenant_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_realm_v1_tenant_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*TenantKeys); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*TenantStateData); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*TenantState); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*TenantSpec); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*TenantEvent); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*TenantEventType); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*TenantEventType_Created); i {
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
		file_j5_realm_v1_tenant_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*TenantEventType_Updated); i {
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
	file_j5_realm_v1_tenant_proto_msgTypes[5].OneofWrappers = []any{
		(*TenantEventType_Created_)(nil),
		(*TenantEventType_Updated_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_realm_v1_tenant_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_realm_v1_tenant_proto_goTypes,
		DependencyIndexes: file_j5_realm_v1_tenant_proto_depIdxs,
		EnumInfos:         file_j5_realm_v1_tenant_proto_enumTypes,
		MessageInfos:      file_j5_realm_v1_tenant_proto_msgTypes,
	}.Build()
	File_j5_realm_v1_tenant_proto = out.File
	file_j5_realm_v1_tenant_proto_rawDesc = nil
	file_j5_realm_v1_tenant_proto_goTypes = nil
	file_j5_realm_v1_tenant_proto_depIdxs = nil
}
