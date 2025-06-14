// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/list/v1/query.proto

package list_j5pb

import (
	reflect "reflect"
	sync "sync"

	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type QueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Searches []*Search `protobuf:"bytes,1,rep,name=searches,proto3" json:"searches,omitempty"`
	Sorts    []*Sort   `protobuf:"bytes,2,rep,name=sorts,proto3" json:"sorts,omitempty"`
	Filters  []*Filter `protobuf:"bytes,3,rep,name=filters,proto3" json:"filters,omitempty"`
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{0}
}

func (x *QueryRequest) GetSearches() []*Search {
	if x != nil {
		return x.Searches
	}
	return nil
}

func (x *QueryRequest) GetSorts() []*Sort {
	if x != nil {
		return x.Sorts
	}
	return nil
}

func (x *QueryRequest) GetFilters() []*Filter {
	if x != nil {
		return x.Filters
	}
	return nil
}

type Search struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Field string `protobuf:"bytes,1,opt,name=field,proto3" json:"field,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Search) Reset() {
	*x = Search{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Search) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Search) ProtoMessage() {}

func (x *Search) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Search.ProtoReflect.Descriptor instead.
func (*Search) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{1}
}

func (x *Search) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

func (x *Search) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type Sort struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Field      string `protobuf:"bytes,1,opt,name=field,proto3" json:"field,omitempty"`
	Descending bool   `protobuf:"varint,2,opt,name=descending,proto3" json:"descending,omitempty"`
}

func (x *Sort) Reset() {
	*x = Sort{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sort) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sort) ProtoMessage() {}

func (x *Sort) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sort.ProtoReflect.Descriptor instead.
func (*Sort) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{2}
}

func (x *Sort) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

func (x *Sort) GetDescending() bool {
	if x != nil {
		return x.Descending
	}
	return false
}

type Filter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*Filter_Field
	//	*Filter_And
	//	*Filter_Or
	Type isFilter_Type `protobuf_oneof:"type"`
}

func (x *Filter) Reset() {
	*x = Filter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Filter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Filter) ProtoMessage() {}

func (x *Filter) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Filter.ProtoReflect.Descriptor instead.
func (*Filter) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{3}
}

func (m *Filter) GetType() isFilter_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Filter) GetField() *Field {
	if x, ok := x.GetType().(*Filter_Field); ok {
		return x.Field
	}
	return nil
}

func (x *Filter) GetAnd() *And {
	if x, ok := x.GetType().(*Filter_And); ok {
		return x.And
	}
	return nil
}

func (x *Filter) GetOr() *Or {
	if x, ok := x.GetType().(*Filter_Or); ok {
		return x.Or
	}
	return nil
}

type isFilter_Type interface {
	isFilter_Type()
}

type Filter_Field struct {
	Field *Field `protobuf:"bytes,1,opt,name=field,proto3,oneof"`
}

type Filter_And struct {
	And *And `protobuf:"bytes,2,opt,name=and,proto3,oneof"`
}

type Filter_Or struct {
	Or *Or `protobuf:"bytes,3,opt,name=or,proto3,oneof"`
}

func (*Filter_Field) isFilter_Type() {}

func (*Filter_And) isFilter_Type() {}

func (*Filter_Or) isFilter_Type() {}

type And struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filters []*Filter `protobuf:"bytes,1,rep,name=filters,proto3" json:"filters,omitempty"`
}

func (x *And) Reset() {
	*x = And{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *And) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*And) ProtoMessage() {}

func (x *And) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use And.ProtoReflect.Descriptor instead.
func (*And) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{4}
}

func (x *And) GetFilters() []*Filter {
	if x != nil {
		return x.Filters
	}
	return nil
}

type Or struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filters []*Filter `protobuf:"bytes,1,rep,name=filters,proto3" json:"filters,omitempty"`
}

func (x *Or) Reset() {
	*x = Or{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Or) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Or) ProtoMessage() {}

func (x *Or) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Or.ProtoReflect.Descriptor instead.
func (*Or) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{5}
}

func (x *Or) GetFilters() []*Filter {
	if x != nil {
		return x.Filters
	}
	return nil
}

type Field struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type *FieldType `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *Field) Reset() {
	*x = Field{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Field) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Field) ProtoMessage() {}

func (x *Field) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Field.ProtoReflect.Descriptor instead.
func (*Field) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{6}
}

func (x *Field) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Field) GetType() *FieldType {
	if x != nil {
		return x.Type
	}
	return nil
}

type FieldType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*FieldType_Value
	//	*FieldType_Range
	Type isFieldType_Type `protobuf_oneof:"type"`
}

func (x *FieldType) Reset() {
	*x = FieldType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldType) ProtoMessage() {}

func (x *FieldType) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldType.ProtoReflect.Descriptor instead.
func (*FieldType) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{7}
}

func (m *FieldType) GetType() isFieldType_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *FieldType) GetValue() string {
	if x, ok := x.GetType().(*FieldType_Value); ok {
		return x.Value
	}
	return ""
}

func (x *FieldType) GetRange() *Range {
	if x, ok := x.GetType().(*FieldType_Range); ok {
		return x.Range
	}
	return nil
}

type isFieldType_Type interface {
	isFieldType_Type()
}

type FieldType_Value struct {
	Value string `protobuf:"bytes,2,opt,name=value,proto3,oneof"`
}

type FieldType_Range struct {
	Range *Range `protobuf:"bytes,3,opt,name=range,proto3,oneof"`
}

func (*FieldType_Value) isFieldType_Type() {}

func (*FieldType_Range) isFieldType_Type() {}

type Range struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Min string `protobuf:"bytes,1,opt,name=min,proto3" json:"min,omitempty"`
	Max string `protobuf:"bytes,2,opt,name=max,proto3" json:"max,omitempty"`
}

func (x *Range) Reset() {
	*x = Range{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_list_v1_query_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Range) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Range) ProtoMessage() {}

func (x *Range) ProtoReflect() protoreflect.Message {
	mi := &file_j5_list_v1_query_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Range.ProtoReflect.Descriptor instead.
func (*Range) Descriptor() ([]byte, []int) {
	return file_j5_list_v1_query_proto_rawDescGZIP(), []int{8}
}

func (x *Range) GetMin() string {
	if x != nil {
		return x.Min
	}
	return ""
}

func (x *Range) GetMax() string {
	if x != nil {
		return x.Max
	}
	return ""
}

var File_j5_list_v1_query_proto protoreflect.FileDescriptor

var file_j5_list_v1_query_proto_rawDesc = []byte{
	0x0a, 0x16, 0x6a, 0x35, 0x2f, 0x6c, 0x69, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x71, 0x75, 0x65,
	0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73,
	0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x94, 0x01, 0x0a, 0x0c, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x2e, 0x0a, 0x08, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x08, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x12, 0x26, 0x0a, 0x05, 0x73, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x10, 0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x6f, 0x72, 0x74, 0x52, 0x05, 0x73, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x2c, 0x0a, 0x07, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6a, 0x35,
	0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52,
	0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x22, 0x34, 0x0a, 0x06, 0x53, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3c,
	0x0a, 0x04, 0x53, 0x6f, 0x72, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1e, 0x0a, 0x0a,
	0x64, 0x65, 0x73, 0x63, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0a, 0x64, 0x65, 0x73, 0x63, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x22, 0x82, 0x01, 0x0a,
	0x06, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x48, 0x00, 0x52, 0x05, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x12, 0x23, 0x0a, 0x03, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x6e, 0x64,
	0x48, 0x00, 0x52, 0x03, 0x61, 0x6e, 0x64, 0x12, 0x20, 0x0a, 0x02, 0x6f, 0x72, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76, 0x31,
	0x2e, 0x4f, 0x72, 0x48, 0x00, 0x52, 0x02, 0x6f, 0x72, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x22, 0x33, 0x0a, 0x03, 0x41, 0x6e, 0x64, 0x12, 0x2c, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x74,
	0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6a, 0x35, 0x2e, 0x6c,
	0x69, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x07, 0x66,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x22, 0x32, 0x0a, 0x02, 0x4f, 0x72, 0x12, 0x2c, 0x0a, 0x07,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x74, 0x65,
	0x72, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x22, 0x4e, 0x0a, 0x05, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x31, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e,
	0x76, 0x31, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x42, 0x06, 0xba, 0x48,
	0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x5d, 0x0a, 0x09, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x29, 0x0a, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11,
	0x2e, 0x6a, 0x35, 0x2e, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61, 0x6e, 0x67,
	0x65, 0x48, 0x00, 0x52, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x42, 0x0d, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x05, 0xba, 0x48, 0x02, 0x08, 0x01, 0x22, 0x2b, 0x0a, 0x05, 0x52, 0x61, 0x6e,
	0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6d, 0x69, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6d, 0x61, 0x78, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f,
	0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x6c, 0x69, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x6c,
	0x69, 0x73, 0x74, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_list_v1_query_proto_rawDescOnce sync.Once
	file_j5_list_v1_query_proto_rawDescData = file_j5_list_v1_query_proto_rawDesc
)

func file_j5_list_v1_query_proto_rawDescGZIP() []byte {
	file_j5_list_v1_query_proto_rawDescOnce.Do(func() {
		file_j5_list_v1_query_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_list_v1_query_proto_rawDescData)
	})
	return file_j5_list_v1_query_proto_rawDescData
}

var file_j5_list_v1_query_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_j5_list_v1_query_proto_goTypes = []any{
	(*QueryRequest)(nil), // 0: j5.list.v1.QueryRequest
	(*Search)(nil),       // 1: j5.list.v1.Search
	(*Sort)(nil),         // 2: j5.list.v1.Sort
	(*Filter)(nil),       // 3: j5.list.v1.Filter
	(*And)(nil),          // 4: j5.list.v1.And
	(*Or)(nil),           // 5: j5.list.v1.Or
	(*Field)(nil),        // 6: j5.list.v1.Field
	(*FieldType)(nil),    // 7: j5.list.v1.FieldType
	(*Range)(nil),        // 8: j5.list.v1.Range
}
var file_j5_list_v1_query_proto_depIdxs = []int32{
	1,  // 0: j5.list.v1.QueryRequest.searches:type_name -> j5.list.v1.Search
	2,  // 1: j5.list.v1.QueryRequest.sorts:type_name -> j5.list.v1.Sort
	3,  // 2: j5.list.v1.QueryRequest.filters:type_name -> j5.list.v1.Filter
	6,  // 3: j5.list.v1.Filter.field:type_name -> j5.list.v1.Field
	4,  // 4: j5.list.v1.Filter.and:type_name -> j5.list.v1.And
	5,  // 5: j5.list.v1.Filter.or:type_name -> j5.list.v1.Or
	3,  // 6: j5.list.v1.And.filters:type_name -> j5.list.v1.Filter
	3,  // 7: j5.list.v1.Or.filters:type_name -> j5.list.v1.Filter
	7,  // 8: j5.list.v1.Field.type:type_name -> j5.list.v1.FieldType
	8,  // 9: j5.list.v1.FieldType.range:type_name -> j5.list.v1.Range
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_j5_list_v1_query_proto_init() }
func file_j5_list_v1_query_proto_init() {
	if File_j5_list_v1_query_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_list_v1_query_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*QueryRequest); i {
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
		file_j5_list_v1_query_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Search); i {
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
		file_j5_list_v1_query_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*Sort); i {
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
		file_j5_list_v1_query_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*Filter); i {
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
		file_j5_list_v1_query_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*And); i {
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
		file_j5_list_v1_query_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*Or); i {
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
		file_j5_list_v1_query_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*Field); i {
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
		file_j5_list_v1_query_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*FieldType); i {
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
		file_j5_list_v1_query_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*Range); i {
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
	file_j5_list_v1_query_proto_msgTypes[3].OneofWrappers = []any{
		(*Filter_Field)(nil),
		(*Filter_And)(nil),
		(*Filter_Or)(nil),
	}
	file_j5_list_v1_query_proto_msgTypes[7].OneofWrappers = []any{
		(*FieldType_Value)(nil),
		(*FieldType_Range)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_list_v1_query_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_list_v1_query_proto_goTypes,
		DependencyIndexes: file_j5_list_v1_query_proto_depIdxs,
		MessageInfos:      file_j5_list_v1_query_proto_msgTypes,
	}.Build()
	File_j5_list_v1_query_proto = out.File
	file_j5_list_v1_query_proto_rawDesc = nil
	file_j5_list_v1_query_proto_goTypes = nil
	file_j5_list_v1_query_proto_depIdxs = nil
}
