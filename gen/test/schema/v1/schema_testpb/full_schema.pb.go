// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: test/schema/v1/full_schema.proto

package schema_testpb

import (
	_ "github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
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

type Enum int32

const (
	Enum_ENUM_UNSPECIFIED Enum = 0
	Enum_ENUM_VALUE1      Enum = 1
	Enum_ENUM_VALUE2      Enum = 2
)

// Enum value maps for Enum.
var (
	Enum_name = map[int32]string{
		0: "ENUM_UNSPECIFIED",
		1: "ENUM_VALUE1",
		2: "ENUM_VALUE2",
	}
	Enum_value = map[string]int32{
		"ENUM_UNSPECIFIED": 0,
		"ENUM_VALUE1":      1,
		"ENUM_VALUE2":      2,
	}
)

func (x Enum) Enum() *Enum {
	p := new(Enum)
	*p = x
	return p
}

func (x Enum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Enum) Descriptor() protoreflect.EnumDescriptor {
	return file_test_schema_v1_full_schema_proto_enumTypes[0].Descriptor()
}

func (Enum) Type() protoreflect.EnumType {
	return &file_test_schema_v1_full_schema_proto_enumTypes[0]
}

func (x Enum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Enum.Descriptor instead.
func (Enum) EnumDescriptor() ([]byte, []int) {
	return file_test_schema_v1_full_schema_proto_rawDescGZIP(), []int{0}
}

type FullSchema struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SString         string                   `protobuf:"bytes,1,opt,name=s_string,json=sString,proto3" json:"s_string,omitempty"`
	OString         *string                  `protobuf:"bytes,2,opt,name=o_string,json=oString,proto3,oneof" json:"o_string,omitempty"`
	RString         []string                 `protobuf:"bytes,3,rep,name=r_string,json=rString,proto3" json:"r_string,omitempty"`
	SFloat          float32                  `protobuf:"fixed32,4,opt,name=s_float,json=sFloat,proto3" json:"s_float,omitempty"`
	OFloat          *float32                 `protobuf:"fixed32,5,opt,name=o_float,json=oFloat,proto3,oneof" json:"o_float,omitempty"`
	RFloat          []float32                `protobuf:"fixed32,6,rep,packed,name=r_float,json=rFloat,proto3" json:"r_float,omitempty"`
	Ts              *timestamppb.Timestamp   `protobuf:"bytes,7,opt,name=ts,proto3" json:"ts,omitempty"`
	RTs             []*timestamppb.Timestamp `protobuf:"bytes,8,rep,name=r_ts,json=rTs,proto3" json:"r_ts,omitempty"`
	SBar            *Bar                     `protobuf:"bytes,9,opt,name=s_bar,json=sBar,proto3" json:"s_bar,omitempty"`
	RBars           []*Bar                   `protobuf:"bytes,10,rep,name=r_bars,json=rBars,proto3" json:"r_bars,omitempty"`
	Enum            Enum                     `protobuf:"varint,11,opt,name=enum,proto3,enum=test.schema.v1.Enum" json:"enum,omitempty"`
	REnum           []Enum                   `protobuf:"varint,12,rep,packed,name=r_enum,json=rEnum,proto3,enum=test.schema.v1.Enum" json:"r_enum,omitempty"`
	SBytes          []byte                   `protobuf:"bytes,13,opt,name=s_bytes,json=sBytes,proto3" json:"s_bytes,omitempty"`
	RBytes          [][]byte                 `protobuf:"bytes,14,rep,name=r_bytes,json=rBytes,proto3" json:"r_bytes,omitempty"`
	MapStringString map[string]string        `protobuf:"bytes,15,rep,name=map_string_string,json=mapStringString,proto3" json:"map_string_string,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Types that are assignable to AnonOneof:
	//
	//	*FullSchema_AOneofString
	//	*FullSchema_AOneofBar
	//	*FullSchema_AOneofFloat
	//	*FullSchema_AOneofEnum
	AnonOneof isFullSchema_AnonOneof `protobuf_oneof:"anon_oneof"`
	// Types that are assignable to ExposedOneof:
	//
	//	*FullSchema_ExposedString
	ExposedOneof        isFullSchema_ExposedOneof `protobuf_oneof:"exposed_oneof"`
	WrappedOneof        *WrappedOneof             `protobuf:"bytes,16,opt,name=wrapped_oneof,json=wrappedOneof,proto3" json:"wrapped_oneof,omitempty"`
	WrappedOneofs       []*WrappedOneof           `protobuf:"bytes,17,rep,name=wrapped_oneofs,json=wrappedOneofs,proto3" json:"wrapped_oneofs,omitempty"`
	Flattened           *FlattenedMessage         `protobuf:"bytes,18,opt,name=flattened,proto3" json:"flattened,omitempty"`
	NestedExposedOneof  *NestedExposed            `protobuf:"bytes,19,opt,name=nested_exposed_oneof,json=nestedExposedOneof,proto3" json:"nested_exposed_oneof,omitempty"`
	NestedExposedOneofs []*NestedExposed          `protobuf:"bytes,20,rep,name=nested_exposed_oneofs,json=nestedExposedOneofs,proto3" json:"nested_exposed_oneofs,omitempty"`
}

func (x *FullSchema) Reset() {
	*x = FullSchema{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_schema_v1_full_schema_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FullSchema) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FullSchema) ProtoMessage() {}

func (x *FullSchema) ProtoReflect() protoreflect.Message {
	mi := &file_test_schema_v1_full_schema_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FullSchema.ProtoReflect.Descriptor instead.
func (*FullSchema) Descriptor() ([]byte, []int) {
	return file_test_schema_v1_full_schema_proto_rawDescGZIP(), []int{0}
}

func (x *FullSchema) GetSString() string {
	if x != nil {
		return x.SString
	}
	return ""
}

func (x *FullSchema) GetOString() string {
	if x != nil && x.OString != nil {
		return *x.OString
	}
	return ""
}

func (x *FullSchema) GetRString() []string {
	if x != nil {
		return x.RString
	}
	return nil
}

func (x *FullSchema) GetSFloat() float32 {
	if x != nil {
		return x.SFloat
	}
	return 0
}

func (x *FullSchema) GetOFloat() float32 {
	if x != nil && x.OFloat != nil {
		return *x.OFloat
	}
	return 0
}

func (x *FullSchema) GetRFloat() []float32 {
	if x != nil {
		return x.RFloat
	}
	return nil
}

func (x *FullSchema) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *FullSchema) GetRTs() []*timestamppb.Timestamp {
	if x != nil {
		return x.RTs
	}
	return nil
}

func (x *FullSchema) GetSBar() *Bar {
	if x != nil {
		return x.SBar
	}
	return nil
}

func (x *FullSchema) GetRBars() []*Bar {
	if x != nil {
		return x.RBars
	}
	return nil
}

func (x *FullSchema) GetEnum() Enum {
	if x != nil {
		return x.Enum
	}
	return Enum_ENUM_UNSPECIFIED
}

func (x *FullSchema) GetREnum() []Enum {
	if x != nil {
		return x.REnum
	}
	return nil
}

func (x *FullSchema) GetSBytes() []byte {
	if x != nil {
		return x.SBytes
	}
	return nil
}

func (x *FullSchema) GetRBytes() [][]byte {
	if x != nil {
		return x.RBytes
	}
	return nil
}

func (x *FullSchema) GetMapStringString() map[string]string {
	if x != nil {
		return x.MapStringString
	}
	return nil
}

func (m *FullSchema) GetAnonOneof() isFullSchema_AnonOneof {
	if m != nil {
		return m.AnonOneof
	}
	return nil
}

func (x *FullSchema) GetAOneofString() string {
	if x, ok := x.GetAnonOneof().(*FullSchema_AOneofString); ok {
		return x.AOneofString
	}
	return ""
}

func (x *FullSchema) GetAOneofBar() *Bar {
	if x, ok := x.GetAnonOneof().(*FullSchema_AOneofBar); ok {
		return x.AOneofBar
	}
	return nil
}

func (x *FullSchema) GetAOneofFloat() float32 {
	if x, ok := x.GetAnonOneof().(*FullSchema_AOneofFloat); ok {
		return x.AOneofFloat
	}
	return 0
}

func (x *FullSchema) GetAOneofEnum() Enum {
	if x, ok := x.GetAnonOneof().(*FullSchema_AOneofEnum); ok {
		return x.AOneofEnum
	}
	return Enum_ENUM_UNSPECIFIED
}

func (m *FullSchema) GetExposedOneof() isFullSchema_ExposedOneof {
	if m != nil {
		return m.ExposedOneof
	}
	return nil
}

func (x *FullSchema) GetExposedString() string {
	if x, ok := x.GetExposedOneof().(*FullSchema_ExposedString); ok {
		return x.ExposedString
	}
	return ""
}

func (x *FullSchema) GetWrappedOneof() *WrappedOneof {
	if x != nil {
		return x.WrappedOneof
	}
	return nil
}

func (x *FullSchema) GetWrappedOneofs() []*WrappedOneof {
	if x != nil {
		return x.WrappedOneofs
	}
	return nil
}

func (x *FullSchema) GetFlattened() *FlattenedMessage {
	if x != nil {
		return x.Flattened
	}
	return nil
}

func (x *FullSchema) GetNestedExposedOneof() *NestedExposed {
	if x != nil {
		return x.NestedExposedOneof
	}
	return nil
}

func (x *FullSchema) GetNestedExposedOneofs() []*NestedExposed {
	if x != nil {
		return x.NestedExposedOneofs
	}
	return nil
}

type isFullSchema_AnonOneof interface {
	isFullSchema_AnonOneof()
}

type FullSchema_AOneofString struct {
	AOneofString string `protobuf:"bytes,100,opt,name=a_oneof_string,json=aOneofString,proto3,oneof"`
}

type FullSchema_AOneofBar struct {
	AOneofBar *Bar `protobuf:"bytes,101,opt,name=a_oneof_bar,json=aOneofBar,proto3,oneof"`
}

type FullSchema_AOneofFloat struct {
	AOneofFloat float32 `protobuf:"fixed32,102,opt,name=a_oneof_float,json=aOneofFloat,proto3,oneof"`
}

type FullSchema_AOneofEnum struct {
	AOneofEnum Enum `protobuf:"varint,103,opt,name=a_oneof_enum,json=aOneofEnum,proto3,enum=test.schema.v1.Enum,oneof"`
}

func (*FullSchema_AOneofString) isFullSchema_AnonOneof() {}

func (*FullSchema_AOneofBar) isFullSchema_AnonOneof() {}

func (*FullSchema_AOneofFloat) isFullSchema_AnonOneof() {}

func (*FullSchema_AOneofEnum) isFullSchema_AnonOneof() {}

type isFullSchema_ExposedOneof interface {
	isFullSchema_ExposedOneof()
}

type FullSchema_ExposedString struct {
	ExposedString string `protobuf:"bytes,200,opt,name=exposed_string,json=exposedString,proto3,oneof"`
}

func (*FullSchema_ExposedString) isFullSchema_ExposedOneof() {}

type WrappedOneof struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*WrappedOneof_WOneofString
	//	*WrappedOneof_WOneofBar
	//	*WrappedOneof_WOneofFloat
	//	*WrappedOneof_WOneofEnum
	Type isWrappedOneof_Type `protobuf_oneof:"type"`
}

func (x *WrappedOneof) Reset() {
	*x = WrappedOneof{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_schema_v1_full_schema_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WrappedOneof) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WrappedOneof) ProtoMessage() {}

func (x *WrappedOneof) ProtoReflect() protoreflect.Message {
	mi := &file_test_schema_v1_full_schema_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WrappedOneof.ProtoReflect.Descriptor instead.
func (*WrappedOneof) Descriptor() ([]byte, []int) {
	return file_test_schema_v1_full_schema_proto_rawDescGZIP(), []int{1}
}

func (m *WrappedOneof) GetType() isWrappedOneof_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *WrappedOneof) GetWOneofString() string {
	if x, ok := x.GetType().(*WrappedOneof_WOneofString); ok {
		return x.WOneofString
	}
	return ""
}

func (x *WrappedOneof) GetWOneofBar() *Bar {
	if x, ok := x.GetType().(*WrappedOneof_WOneofBar); ok {
		return x.WOneofBar
	}
	return nil
}

func (x *WrappedOneof) GetWOneofFloat() float32 {
	if x, ok := x.GetType().(*WrappedOneof_WOneofFloat); ok {
		return x.WOneofFloat
	}
	return 0
}

func (x *WrappedOneof) GetWOneofEnum() Enum {
	if x, ok := x.GetType().(*WrappedOneof_WOneofEnum); ok {
		return x.WOneofEnum
	}
	return Enum_ENUM_UNSPECIFIED
}

type isWrappedOneof_Type interface {
	isWrappedOneof_Type()
}

type WrappedOneof_WOneofString struct {
	WOneofString string `protobuf:"bytes,1,opt,name=w_oneof_string,json=wOneofString,proto3,oneof"`
}

type WrappedOneof_WOneofBar struct {
	WOneofBar *Bar `protobuf:"bytes,2,opt,name=w_oneof_bar,json=wOneofBar,proto3,oneof"`
}

type WrappedOneof_WOneofFloat struct {
	WOneofFloat float32 `protobuf:"fixed32,3,opt,name=w_oneof_float,json=wOneofFloat,proto3,oneof"`
}

type WrappedOneof_WOneofEnum struct {
	WOneofEnum Enum `protobuf:"varint,4,opt,name=w_oneof_enum,json=wOneofEnum,proto3,enum=test.schema.v1.Enum,oneof"`
}

func (*WrappedOneof_WOneofString) isWrappedOneof_Type() {}

func (*WrappedOneof_WOneofBar) isWrappedOneof_Type() {}

func (*WrappedOneof_WOneofFloat) isWrappedOneof_Type() {}

func (*WrappedOneof_WOneofEnum) isWrappedOneof_Type() {}

type NestedExposed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*NestedExposed_De1
	//	*NestedExposed_De2
	//	*NestedExposed_De3
	Type isNestedExposed_Type `protobuf_oneof:"type"`
}

func (x *NestedExposed) Reset() {
	*x = NestedExposed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_schema_v1_full_schema_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NestedExposed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NestedExposed) ProtoMessage() {}

func (x *NestedExposed) ProtoReflect() protoreflect.Message {
	mi := &file_test_schema_v1_full_schema_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NestedExposed.ProtoReflect.Descriptor instead.
func (*NestedExposed) Descriptor() ([]byte, []int) {
	return file_test_schema_v1_full_schema_proto_rawDescGZIP(), []int{2}
}

func (m *NestedExposed) GetType() isNestedExposed_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *NestedExposed) GetDe1() string {
	if x, ok := x.GetType().(*NestedExposed_De1); ok {
		return x.De1
	}
	return ""
}

func (x *NestedExposed) GetDe2() string {
	if x, ok := x.GetType().(*NestedExposed_De2); ok {
		return x.De2
	}
	return ""
}

func (x *NestedExposed) GetDe3() *NestedExposed {
	if x, ok := x.GetType().(*NestedExposed_De3); ok {
		return x.De3
	}
	return nil
}

type isNestedExposed_Type interface {
	isNestedExposed_Type()
}

type NestedExposed_De1 struct {
	De1 string `protobuf:"bytes,101,opt,name=de1,proto3,oneof"`
}

type NestedExposed_De2 struct {
	De2 string `protobuf:"bytes,102,opt,name=de2,proto3,oneof"`
}

type NestedExposed_De3 struct {
	De3 *NestedExposed `protobuf:"bytes,103,opt,name=de3,proto3,oneof"`
}

func (*NestedExposed_De1) isNestedExposed_Type() {}

func (*NestedExposed_De2) isNestedExposed_Type() {}

func (*NestedExposed_De3) isNestedExposed_Type() {}

type Bar struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Bar) Reset() {
	*x = Bar{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_schema_v1_full_schema_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Bar) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Bar) ProtoMessage() {}

func (x *Bar) ProtoReflect() protoreflect.Message {
	mi := &file_test_schema_v1_full_schema_proto_msgTypes[3]
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
	return file_test_schema_v1_full_schema_proto_rawDescGZIP(), []int{3}
}

func (x *Bar) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type FlattenedMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FieldFromFlattened   string `protobuf:"bytes,1,opt,name=field_from_flattened,json=fieldFromFlattened,proto3" json:"field_from_flattened,omitempty"`
	Field_2FromFlattened string `protobuf:"bytes,2,opt,name=field_2_from_flattened,json=field2FromFlattened,proto3" json:"field_2_from_flattened,omitempty"`
}

func (x *FlattenedMessage) Reset() {
	*x = FlattenedMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_schema_v1_full_schema_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FlattenedMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FlattenedMessage) ProtoMessage() {}

func (x *FlattenedMessage) ProtoReflect() protoreflect.Message {
	mi := &file_test_schema_v1_full_schema_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FlattenedMessage.ProtoReflect.Descriptor instead.
func (*FlattenedMessage) Descriptor() ([]byte, []int) {
	return file_test_schema_v1_full_schema_proto_rawDescGZIP(), []int{4}
}

func (x *FlattenedMessage) GetFieldFromFlattened() string {
	if x != nil {
		return x.FieldFromFlattened
	}
	return ""
}

func (x *FlattenedMessage) GetField_2FromFlattened() string {
	if x != nil {
		return x.Field_2FromFlattened
	}
	return ""
}

var File_test_schema_v1_full_schema_proto protoreflect.FileDescriptor

var file_test_schema_v1_full_schema_proto_rawDesc = []byte{
	0x0a, 0x20, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31,
	0x2f, 0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e,
	0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x6a, 0x35, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xae, 0x0a, 0x0a, 0x0a, 0x46, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12,
	0x19, 0x0a, 0x08, 0x73, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x1e, 0x0a, 0x08, 0x6f, 0x5f,
	0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x07,
	0x6f, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x88, 0x01, 0x01, 0x12, 0x19, 0x0a, 0x08, 0x72, 0x5f,
	0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x72, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x02, 0x52, 0x06, 0x73, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x12, 0x1c,
	0x0a, 0x07, 0x6f, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x48,
	0x03, 0x52, 0x06, 0x6f, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x88, 0x01, 0x01, 0x12, 0x17, 0x0a, 0x07,
	0x72, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x03, 0x28, 0x02, 0x52, 0x06, 0x72,
	0x46, 0x6c, 0x6f, 0x61, 0x74, 0x12, 0x2a, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74,
	0x73, 0x12, 0x2d, 0x0a, 0x04, 0x72, 0x5f, 0x74, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x03, 0x72, 0x54, 0x73,
	0x12, 0x28, 0x0a, 0x05, 0x73, 0x5f, 0x62, 0x61, 0x72, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x61, 0x72, 0x52, 0x04, 0x73, 0x42, 0x61, 0x72, 0x12, 0x2a, 0x0a, 0x06, 0x72, 0x5f,
	0x62, 0x61, 0x72, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x74, 0x65, 0x73,
	0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x61, 0x72, 0x52,
	0x05, 0x72, 0x42, 0x61, 0x72, 0x73, 0x12, 0x28, 0x0a, 0x04, 0x65, 0x6e, 0x75, 0x6d, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x6e, 0x75, 0x6d, 0x52, 0x04, 0x65, 0x6e, 0x75, 0x6d,
	0x12, 0x2b, 0x0a, 0x06, 0x72, 0x5f, 0x65, 0x6e, 0x75, 0x6d, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0e,
	0x32, 0x14, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x6e, 0x75, 0x6d, 0x52, 0x05, 0x72, 0x45, 0x6e, 0x75, 0x6d, 0x12, 0x17, 0x0a,
	0x07, 0x73, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06,
	0x73, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x17, 0x0a, 0x07, 0x72, 0x5f, 0x62, 0x79, 0x74, 0x65,
	0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12,
	0x5b, 0x0a, 0x11, 0x6d, 0x61, 0x70, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x74, 0x65, 0x73,
	0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6c, 0x6c,
	0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0f, 0x6d, 0x61, 0x70,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x26, 0x0a, 0x0e,
	0x61, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x64,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0c, 0x61, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x12, 0x35, 0x0a, 0x0b, 0x61, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f,
	0x62, 0x61, 0x72, 0x18, 0x65, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x74, 0x65, 0x73, 0x74,
	0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x61, 0x72, 0x48, 0x00,
	0x52, 0x09, 0x61, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x42, 0x61, 0x72, 0x12, 0x24, 0x0a, 0x0d, 0x61,
	0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x18, 0x66, 0x20, 0x01,
	0x28, 0x02, 0x48, 0x00, 0x52, 0x0b, 0x61, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x46, 0x6c, 0x6f, 0x61,
	0x74, 0x12, 0x38, 0x0a, 0x0c, 0x61, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x65, 0x6e, 0x75,
	0x6d, 0x18, 0x67, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x6e, 0x75, 0x6d, 0x48, 0x00, 0x52,
	0x0a, 0x61, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x45, 0x6e, 0x75, 0x6d, 0x12, 0x28, 0x0a, 0x0e, 0x65,
	0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0xc8, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x0d, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x41, 0x0a, 0x0d, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x64,
	0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x18, 0x10, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x72,
	0x61, 0x70, 0x70, 0x65, 0x64, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x52, 0x0c, 0x77, 0x72, 0x61, 0x70,
	0x70, 0x65, 0x64, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x12, 0x43, 0x0a, 0x0e, 0x77, 0x72, 0x61, 0x70,
	0x70, 0x65, 0x64, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x73, 0x18, 0x11, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x57, 0x72, 0x61, 0x70, 0x70, 0x65, 0x64, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x52, 0x0d,
	0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x64, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x73, 0x12, 0x49, 0x0a,
	0x09, 0x66, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x65, 0x64, 0x18, 0x12, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x42, 0x09, 0xc2, 0xff, 0x8e, 0x02, 0x04, 0x0a, 0x02, 0x08, 0x01, 0x52, 0x09, 0x66,
	0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x65, 0x64, 0x12, 0x4f, 0x0a, 0x14, 0x6e, 0x65, 0x73, 0x74,
	0x65, 0x64, 0x5f, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66,
	0x18, 0x13, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x45, 0x78,
	0x70, 0x6f, 0x73, 0x65, 0x64, 0x52, 0x12, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x45, 0x78, 0x70,
	0x6f, 0x73, 0x65, 0x64, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x12, 0x51, 0x0a, 0x15, 0x6e, 0x65, 0x73,
	0x74, 0x65, 0x64, 0x5f, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x5f, 0x6f, 0x6e, 0x65, 0x6f,
	0x66, 0x73, 0x18, 0x14, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64,
	0x45, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x52, 0x13, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x45,
	0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x73, 0x1a, 0x42, 0x0a, 0x14,
	0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x42, 0x0c, 0x0a, 0x0a, 0x61, 0x6e, 0x6f, 0x6e, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x42, 0x18,
	0x0a, 0x0d, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x12,
	0x07, 0xc2, 0xff, 0x8e, 0x02, 0x02, 0x08, 0x01, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x6f, 0x5f, 0x73,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x6f, 0x5f, 0x66, 0x6c, 0x6f, 0x61,
	0x74, 0x22, 0xde, 0x01, 0x0a, 0x0c, 0x57, 0x72, 0x61, 0x70, 0x70, 0x65, 0x64, 0x4f, 0x6e, 0x65,
	0x6f, 0x66, 0x12, 0x26, 0x0a, 0x0e, 0x77, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0c, 0x77, 0x4f,
	0x6e, 0x65, 0x6f, 0x66, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x35, 0x0a, 0x0b, 0x77, 0x5f,
	0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x62, 0x61, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x61, 0x72, 0x48, 0x00, 0x52, 0x09, 0x77, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x42, 0x61,
	0x72, 0x12, 0x24, 0x0a, 0x0d, 0x77, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x66, 0x6c, 0x6f,
	0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x48, 0x00, 0x52, 0x0b, 0x77, 0x4f, 0x6e, 0x65,
	0x6f, 0x66, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x12, 0x38, 0x0a, 0x0c, 0x77, 0x5f, 0x6f, 0x6e, 0x65,
	0x6f, 0x66, 0x5f, 0x65, 0x6e, 0x75, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x45,
	0x6e, 0x75, 0x6d, 0x48, 0x00, 0x52, 0x0a, 0x77, 0x4f, 0x6e, 0x65, 0x6f, 0x66, 0x45, 0x6e, 0x75,
	0x6d, 0x3a, 0x07, 0xc2, 0xff, 0x8e, 0x02, 0x02, 0x08, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x22, 0x7b, 0x0a, 0x0d, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x45, 0x78, 0x70, 0x6f,
	0x73, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x03, 0x64, 0x65, 0x31, 0x18, 0x65, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x03, 0x64, 0x65, 0x31, 0x12, 0x12, 0x0a, 0x03, 0x64, 0x65, 0x32, 0x18, 0x66,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x64, 0x65, 0x32, 0x12, 0x31, 0x0a, 0x03, 0x64,
	0x65, 0x33, 0x18, 0x67, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64,
	0x45, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x48, 0x00, 0x52, 0x03, 0x64, 0x65, 0x33, 0x42, 0x0f,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x07, 0xc2, 0xff, 0x8e, 0x02, 0x02, 0x08, 0x01, 0x22,
	0x15, 0x0a, 0x03, 0x42, 0x61, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x79, 0x0a, 0x10, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65,
	0x6e, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x30, 0x0a, 0x14, 0x66, 0x69,
	0x65, 0x6c, 0x64, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x5f, 0x66, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e,
	0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x46,
	0x72, 0x6f, 0x6d, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x65, 0x64, 0x12, 0x33, 0x0a, 0x16,
	0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x32, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x5f, 0x66, 0x6c, 0x61,
	0x74, 0x74, 0x65, 0x6e, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x66, 0x69,
	0x65, 0x6c, 0x64, 0x32, 0x46, 0x72, 0x6f, 0x6d, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x65,
	0x64, 0x2a, 0x3e, 0x0a, 0x04, 0x45, 0x6e, 0x75, 0x6d, 0x12, 0x14, 0x0a, 0x10, 0x45, 0x4e, 0x55,
	0x4d, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x0f, 0x0a, 0x0b, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x31, 0x10, 0x01,
	0x12, 0x0f, 0x0a, 0x0b, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x32, 0x10,
	0x02, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74,
	0x65, 0x73, 0x74, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_test_schema_v1_full_schema_proto_rawDescOnce sync.Once
	file_test_schema_v1_full_schema_proto_rawDescData = file_test_schema_v1_full_schema_proto_rawDesc
)

func file_test_schema_v1_full_schema_proto_rawDescGZIP() []byte {
	file_test_schema_v1_full_schema_proto_rawDescOnce.Do(func() {
		file_test_schema_v1_full_schema_proto_rawDescData = protoimpl.X.CompressGZIP(file_test_schema_v1_full_schema_proto_rawDescData)
	})
	return file_test_schema_v1_full_schema_proto_rawDescData
}

var file_test_schema_v1_full_schema_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_test_schema_v1_full_schema_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_test_schema_v1_full_schema_proto_goTypes = []any{
	(Enum)(0),                     // 0: test.schema.v1.Enum
	(*FullSchema)(nil),            // 1: test.schema.v1.FullSchema
	(*WrappedOneof)(nil),          // 2: test.schema.v1.WrappedOneof
	(*NestedExposed)(nil),         // 3: test.schema.v1.NestedExposed
	(*Bar)(nil),                   // 4: test.schema.v1.Bar
	(*FlattenedMessage)(nil),      // 5: test.schema.v1.FlattenedMessage
	nil,                           // 6: test.schema.v1.FullSchema.MapStringStringEntry
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
}
var file_test_schema_v1_full_schema_proto_depIdxs = []int32{
	7,  // 0: test.schema.v1.FullSchema.ts:type_name -> google.protobuf.Timestamp
	7,  // 1: test.schema.v1.FullSchema.r_ts:type_name -> google.protobuf.Timestamp
	4,  // 2: test.schema.v1.FullSchema.s_bar:type_name -> test.schema.v1.Bar
	4,  // 3: test.schema.v1.FullSchema.r_bars:type_name -> test.schema.v1.Bar
	0,  // 4: test.schema.v1.FullSchema.enum:type_name -> test.schema.v1.Enum
	0,  // 5: test.schema.v1.FullSchema.r_enum:type_name -> test.schema.v1.Enum
	6,  // 6: test.schema.v1.FullSchema.map_string_string:type_name -> test.schema.v1.FullSchema.MapStringStringEntry
	4,  // 7: test.schema.v1.FullSchema.a_oneof_bar:type_name -> test.schema.v1.Bar
	0,  // 8: test.schema.v1.FullSchema.a_oneof_enum:type_name -> test.schema.v1.Enum
	2,  // 9: test.schema.v1.FullSchema.wrapped_oneof:type_name -> test.schema.v1.WrappedOneof
	2,  // 10: test.schema.v1.FullSchema.wrapped_oneofs:type_name -> test.schema.v1.WrappedOneof
	5,  // 11: test.schema.v1.FullSchema.flattened:type_name -> test.schema.v1.FlattenedMessage
	3,  // 12: test.schema.v1.FullSchema.nested_exposed_oneof:type_name -> test.schema.v1.NestedExposed
	3,  // 13: test.schema.v1.FullSchema.nested_exposed_oneofs:type_name -> test.schema.v1.NestedExposed
	4,  // 14: test.schema.v1.WrappedOneof.w_oneof_bar:type_name -> test.schema.v1.Bar
	0,  // 15: test.schema.v1.WrappedOneof.w_oneof_enum:type_name -> test.schema.v1.Enum
	3,  // 16: test.schema.v1.NestedExposed.de3:type_name -> test.schema.v1.NestedExposed
	17, // [17:17] is the sub-list for method output_type
	17, // [17:17] is the sub-list for method input_type
	17, // [17:17] is the sub-list for extension type_name
	17, // [17:17] is the sub-list for extension extendee
	0,  // [0:17] is the sub-list for field type_name
}

func init() { file_test_schema_v1_full_schema_proto_init() }
func file_test_schema_v1_full_schema_proto_init() {
	if File_test_schema_v1_full_schema_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_test_schema_v1_full_schema_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*FullSchema); i {
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
		file_test_schema_v1_full_schema_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*WrappedOneof); i {
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
		file_test_schema_v1_full_schema_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*NestedExposed); i {
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
		file_test_schema_v1_full_schema_proto_msgTypes[3].Exporter = func(v any, i int) any {
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
		file_test_schema_v1_full_schema_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*FlattenedMessage); i {
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
	file_test_schema_v1_full_schema_proto_msgTypes[0].OneofWrappers = []any{
		(*FullSchema_AOneofString)(nil),
		(*FullSchema_AOneofBar)(nil),
		(*FullSchema_AOneofFloat)(nil),
		(*FullSchema_AOneofEnum)(nil),
		(*FullSchema_ExposedString)(nil),
	}
	file_test_schema_v1_full_schema_proto_msgTypes[1].OneofWrappers = []any{
		(*WrappedOneof_WOneofString)(nil),
		(*WrappedOneof_WOneofBar)(nil),
		(*WrappedOneof_WOneofFloat)(nil),
		(*WrappedOneof_WOneofEnum)(nil),
	}
	file_test_schema_v1_full_schema_proto_msgTypes[2].OneofWrappers = []any{
		(*NestedExposed_De1)(nil),
		(*NestedExposed_De2)(nil),
		(*NestedExposed_De3)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_test_schema_v1_full_schema_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_test_schema_v1_full_schema_proto_goTypes,
		DependencyIndexes: file_test_schema_v1_full_schema_proto_depIdxs,
		EnumInfos:         file_test_schema_v1_full_schema_proto_enumTypes,
		MessageInfos:      file_test_schema_v1_full_schema_proto_msgTypes,
	}.Build()
	File_test_schema_v1_full_schema_proto = out.File
	file_test_schema_v1_full_schema_proto_rawDesc = nil
	file_test_schema_v1_full_schema_proto_goTypes = nil
	file_test_schema_v1_full_schema_proto_depIdxs = nil
}