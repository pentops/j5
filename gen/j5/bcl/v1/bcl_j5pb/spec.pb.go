// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/bcl/v1/spec.proto

package bcl_j5pb

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

type Schema struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Blocks []*Block `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
}

func (x *Schema) Reset() {
	*x = Schema{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_bcl_v1_spec_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Schema) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Schema) ProtoMessage() {}

func (x *Schema) ProtoReflect() protoreflect.Message {
	mi := &file_j5_bcl_v1_spec_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Schema.ProtoReflect.Descriptor instead.
func (*Schema) Descriptor() ([]byte, []int) {
	return file_j5_bcl_v1_spec_proto_rawDescGZIP(), []int{0}
}

func (x *Schema) GetBlocks() []*Block {
	if x != nil {
		return x.Blocks
	}
	return nil
}

// message SchemaFile {
// Schema schema = 1;
// SourceLocation source_location = 2;
// }
type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The full name (i.e. protoreflect's FullName) of the schema this block
	// defines.
	SchemaName       string       `protobuf:"bytes,1,opt,name=schema_name,json=schemaName,proto3" json:"schema_name,omitempty"`
	Name             *Tag         `protobuf:"bytes,3,opt,name=name,proto3,oneof" json:"name,omitempty"`
	TypeSelect       *Tag         `protobuf:"bytes,4,opt,name=type_select,json=typeSelect,proto3,oneof" json:"type_select,omitempty"`
	Qualifier        *Tag         `protobuf:"bytes,5,opt,name=qualifier,proto3,oneof" json:"qualifier,omitempty"`
	DescriptionField *string      `protobuf:"bytes,6,opt,name=description_field,json=descriptionField,proto3,oneof" json:"description_field,omitempty"`
	Alias            []*Alias     `protobuf:"bytes,10,rep,name=alias,proto3" json:"alias,omitempty"`
	ScalarSplit      *ScalarSplit `protobuf:"bytes,8,opt,name=scalar_split,json=scalarSplit,proto3" json:"scalar_split,omitempty"`
	// When true, fields in the block which are not mentioned in tags or children
	// are not settable.
	OnlyExplicit bool `protobuf:"varint,9,opt,name=only_explicit,json=onlyExplicit,proto3" json:"only_explicit,omitempty"`
}

func (x *Block) Reset() {
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_bcl_v1_spec_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_j5_bcl_v1_spec_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_j5_bcl_v1_spec_proto_rawDescGZIP(), []int{1}
}

func (x *Block) GetSchemaName() string {
	if x != nil {
		return x.SchemaName
	}
	return ""
}

func (x *Block) GetName() *Tag {
	if x != nil {
		return x.Name
	}
	return nil
}

func (x *Block) GetTypeSelect() *Tag {
	if x != nil {
		return x.TypeSelect
	}
	return nil
}

func (x *Block) GetQualifier() *Tag {
	if x != nil {
		return x.Qualifier
	}
	return nil
}

func (x *Block) GetDescriptionField() string {
	if x != nil && x.DescriptionField != nil {
		return *x.DescriptionField
	}
	return ""
}

func (x *Block) GetAlias() []*Alias {
	if x != nil {
		return x.Alias
	}
	return nil
}

func (x *Block) GetScalarSplit() *ScalarSplit {
	if x != nil {
		return x.ScalarSplit
	}
	return nil
}

func (x *Block) GetOnlyExplicit() bool {
	if x != nil {
		return x.OnlyExplicit
	}
	return false
}

// A Path is the nested field names from a root node to a child node. All path
// elements are strings, which is the field names, map keys, or in theory list
// index string-numbers.
type Path struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path []string `protobuf:"bytes,1,rep,name=path,proto3" json:"path,omitempty"`
}

func (x *Path) Reset() {
	*x = Path{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_bcl_v1_spec_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Path) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Path) ProtoMessage() {}

func (x *Path) ProtoReflect() protoreflect.Message {
	mi := &file_j5_bcl_v1_spec_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Path.ProtoReflect.Descriptor instead.
func (*Path) Descriptor() ([]byte, []int) {
	return file_j5_bcl_v1_spec_proto_rawDescGZIP(), []int{2}
}

func (x *Path) GetPath() []string {
	if x != nil {
		return x.Path
	}
	return nil
}

// A Tag defines the behavior of the block header components for the type.
type Tag struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FieldName string `protobuf:"bytes,1,opt,name=field_name,json=fieldName,proto3" json:"field_name,omitempty"` // can use aliases.
	IsBlock   bool   `protobuf:"varint,2,opt,name=is_block,json=isBlock,proto3" json:"is_block,omitempty"`
	Optional  bool   `protobuf:"varint,3,opt,name=optional,proto3" json:"optional,omitempty"`
	// When set, a leading '!' on the tag sets a boolean to true at the given
	// path. (e.g. setting required=true). When not set, bang is illegal. You
	// shouldn't bang where it's illegal.
	BangBool *string `protobuf:"bytes,4,opt,name=bang_bool,json=bangBool,proto3,oneof" json:"bang_bool,omitempty"`
	// Same as bang_bool, but for a ?. Still 'true', e.g. 'optional=true)
	QuestionBool *string `protobuf:"bytes,5,opt,name=question_bool,json=questionBool,proto3,oneof" json:"question_bool,omitempty"`
}

func (x *Tag) Reset() {
	*x = Tag{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_bcl_v1_spec_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tag) ProtoMessage() {}

func (x *Tag) ProtoReflect() protoreflect.Message {
	mi := &file_j5_bcl_v1_spec_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tag.ProtoReflect.Descriptor instead.
func (*Tag) Descriptor() ([]byte, []int) {
	return file_j5_bcl_v1_spec_proto_rawDescGZIP(), []int{3}
}

func (x *Tag) GetFieldName() string {
	if x != nil {
		return x.FieldName
	}
	return ""
}

func (x *Tag) GetIsBlock() bool {
	if x != nil {
		return x.IsBlock
	}
	return false
}

func (x *Tag) GetOptional() bool {
	if x != nil {
		return x.Optional
	}
	return false
}

func (x *Tag) GetBangBool() string {
	if x != nil && x.BangBool != nil {
		return *x.BangBool
	}
	return ""
}

func (x *Tag) GetQuestionBool() string {
	if x != nil && x.QuestionBool != nil {
		return *x.QuestionBool
	}
	return ""
}

type Alias struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Path []string `protobuf:"bytes,2,rep,name=path,proto3" json:"path,omitempty"`
}

func (x *Alias) Reset() {
	*x = Alias{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_bcl_v1_spec_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Alias) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Alias) ProtoMessage() {}

func (x *Alias) ProtoReflect() protoreflect.Message {
	mi := &file_j5_bcl_v1_spec_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Alias.ProtoReflect.Descriptor instead.
func (*Alias) Descriptor() ([]byte, []int) {
	return file_j5_bcl_v1_spec_proto_rawDescGZIP(), []int{4}
}

func (x *Alias) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Alias) GetPath() []string {
	if x != nil {
		return x.Path
	}
	return nil
}

// ScalarSplit is a way to set a block (container/object type) from a scalar,
// either an array of scalars or a single scalar string.
type ScalarSplit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Delimiter *string `protobuf:"bytes,1,opt,name=delimiter,proto3,oneof" json:"delimiter,omitempty"` // When the value is a string, split it by this delimiter into array of strings and continue
	// When true, the first element is the rightmost element, walking left to right.
	RightToLeft bool `protobuf:"varint,2,opt,name=right_to_left,json=rightToLeft,proto3" json:"right_to_left,omitempty"`
	// Fields are popped one by one, and set to the values at the paths specified
	// in required. If there are not enough values, an error is raised.
	RequiredFields []*Path `protobuf:"bytes,3,rep,name=required_fields,json=requiredFields,proto3" json:"required_fields,omitempty"`
	// After popping all required fields, the remaining values are added to
	// optional fields one by one. If we run out of values, that's fine here.
	OptionalFields []*Path `protobuf:"bytes,4,rep,name=optional_fields,json=optionalFields,proto3" json:"optional_fields,omitempty"`
	// If there are still remaining values after the optional fields, the
	// remaining values are concatenated (as strings) using delimiter and all
	// added to this one field. If there are no remaining values, this field is
	// not touched. If there are remaining values and this field is not set, an
	// error is raised.
	RemainderField *Path `protobuf:"bytes,5,opt,name=remainder_field,json=remainderField,proto3,oneof" json:"remainder_field,omitempty"`
}

func (x *ScalarSplit) Reset() {
	*x = ScalarSplit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_bcl_v1_spec_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScalarSplit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScalarSplit) ProtoMessage() {}

func (x *ScalarSplit) ProtoReflect() protoreflect.Message {
	mi := &file_j5_bcl_v1_spec_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScalarSplit.ProtoReflect.Descriptor instead.
func (*ScalarSplit) Descriptor() ([]byte, []int) {
	return file_j5_bcl_v1_spec_proto_rawDescGZIP(), []int{5}
}

func (x *ScalarSplit) GetDelimiter() string {
	if x != nil && x.Delimiter != nil {
		return *x.Delimiter
	}
	return ""
}

func (x *ScalarSplit) GetRightToLeft() bool {
	if x != nil {
		return x.RightToLeft
	}
	return false
}

func (x *ScalarSplit) GetRequiredFields() []*Path {
	if x != nil {
		return x.RequiredFields
	}
	return nil
}

func (x *ScalarSplit) GetOptionalFields() []*Path {
	if x != nil {
		return x.OptionalFields
	}
	return nil
}

func (x *ScalarSplit) GetRemainderField() *Path {
	if x != nil {
		return x.RemainderField
	}
	return nil
}

var File_j5_bcl_v1_spec_proto protoreflect.FileDescriptor

var file_j5_bcl_v1_spec_proto_rawDesc = []byte{
	0x0a, 0x14, 0x6a, 0x35, 0x2f, 0x62, 0x63, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x70, 0x65, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76,
	0x31, 0x22, 0x32, 0x0a, 0x06, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x28, 0x0a, 0x06, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6a, 0x35,
	0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x06, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x22, 0xb1, 0x03, 0x0a, 0x05, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12,
	0x1f, 0x0a, 0x0b, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x27, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x48, 0x00,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x34, 0x0a, 0x0b, 0x74, 0x79, 0x70,
	0x65, 0x5f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x48, 0x01,
	0x52, 0x0a, 0x74, 0x79, 0x70, 0x65, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x88, 0x01, 0x01, 0x12,
	0x31, 0x0a, 0x09, 0x71, 0x75, 0x61, 0x6c, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x61, 0x67, 0x48, 0x02, 0x52, 0x09, 0x71, 0x75, 0x61, 0x6c, 0x69, 0x66, 0x69, 0x65, 0x72, 0x88,
	0x01, 0x01, 0x12, 0x30, 0x0a, 0x11, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x03, 0x52,
	0x10, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x05, 0x61, 0x6c, 0x69, 0x61, 0x73, 0x18, 0x0a, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x41, 0x6c, 0x69, 0x61, 0x73, 0x52, 0x05, 0x61, 0x6c, 0x69, 0x61, 0x73, 0x12, 0x39, 0x0a, 0x0c,
	0x73, 0x63, 0x61, 0x6c, 0x61, 0x72, 0x5f, 0x73, 0x70, 0x6c, 0x69, 0x74, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x63, 0x61, 0x6c, 0x61, 0x72, 0x53, 0x70, 0x6c, 0x69, 0x74, 0x52, 0x0b, 0x73, 0x63, 0x61, 0x6c,
	0x61, 0x72, 0x53, 0x70, 0x6c, 0x69, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x6f, 0x6e, 0x6c, 0x79, 0x5f,
	0x65, 0x78, 0x70, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c,
	0x6f, 0x6e, 0x6c, 0x79, 0x45, 0x78, 0x70, 0x6c, 0x69, 0x63, 0x69, 0x74, 0x42, 0x07, 0x0a, 0x05,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x71, 0x75, 0x61, 0x6c, 0x69, 0x66,
	0x69, 0x65, 0x72, 0x42, 0x14, 0x0a, 0x12, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x22, 0x1a, 0x0a, 0x04, 0x50, 0x61, 0x74,
	0x68, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0xc7, 0x01, 0x0a, 0x03, 0x54, 0x61, 0x67, 0x12, 0x1d, 0x0a,
	0x0a, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08,
	0x69, 0x73, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x69, 0x73, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1a, 0x0a, 0x08, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x61, 0x6c, 0x12, 0x20, 0x0a, 0x09, 0x62, 0x61, 0x6e, 0x67, 0x5f, 0x62, 0x6f, 0x6f, 0x6c,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x08, 0x62, 0x61, 0x6e, 0x67, 0x42, 0x6f,
	0x6f, 0x6c, 0x88, 0x01, 0x01, 0x12, 0x28, 0x0a, 0x0d, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x62, 0x6f, 0x6f, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x0c,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x6f, 0x6f, 0x6c, 0x88, 0x01, 0x01, 0x42,
	0x0c, 0x0a, 0x0a, 0x5f, 0x62, 0x61, 0x6e, 0x67, 0x5f, 0x62, 0x6f, 0x6f, 0x6c, 0x42, 0x10, 0x0a,
	0x0e, 0x5f, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x62, 0x6f, 0x6f, 0x6c, 0x22,
	0x2f, 0x0a, 0x05, 0x41, 0x6c, 0x69, 0x61, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x22, 0xa9, 0x02, 0x0a, 0x0b, 0x53, 0x63, 0x61, 0x6c, 0x61, 0x72, 0x53, 0x70, 0x6c, 0x69, 0x74,
	0x12, 0x21, 0x0a, 0x09, 0x64, 0x65, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x64, 0x65, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72,
	0x88, 0x01, 0x01, 0x12, 0x22, 0x0a, 0x0d, 0x72, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x74, 0x6f, 0x5f,
	0x6c, 0x65, 0x66, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x72, 0x69, 0x67, 0x68,
	0x74, 0x54, 0x6f, 0x4c, 0x65, 0x66, 0x74, 0x12, 0x38, 0x0a, 0x0f, 0x72, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x64, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x74,
	0x68, 0x52, 0x0e, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x12, 0x38, 0x0a, 0x0f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x66, 0x69,
	0x65, 0x6c, 0x64, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6a, 0x35, 0x2e,
	0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x74, 0x68, 0x52, 0x0e, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x12, 0x3d, 0x0a, 0x0f, 0x72,
	0x65, 0x6d, 0x61, 0x69, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6a, 0x35, 0x2e, 0x62, 0x63, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x61, 0x74, 0x68, 0x48, 0x01, 0x52, 0x0e, 0x72, 0x65, 0x6d, 0x61, 0x69, 0x6e, 0x64,
	0x65, 0x72, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x88, 0x01, 0x01, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x64,
	0x65, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x65, 0x72, 0x42, 0x12, 0x0a, 0x10, 0x5f, 0x72, 0x65, 0x6d,
	0x61, 0x69, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x42, 0x2e, 0x5a, 0x2c,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f,
	0x70, 0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x62, 0x63, 0x6c,
	0x2f, 0x76, 0x31, 0x2f, 0x62, 0x63, 0x6c, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_bcl_v1_spec_proto_rawDescOnce sync.Once
	file_j5_bcl_v1_spec_proto_rawDescData = file_j5_bcl_v1_spec_proto_rawDesc
)

func file_j5_bcl_v1_spec_proto_rawDescGZIP() []byte {
	file_j5_bcl_v1_spec_proto_rawDescOnce.Do(func() {
		file_j5_bcl_v1_spec_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_bcl_v1_spec_proto_rawDescData)
	})
	return file_j5_bcl_v1_spec_proto_rawDescData
}

var file_j5_bcl_v1_spec_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_j5_bcl_v1_spec_proto_goTypes = []any{
	(*Schema)(nil),      // 0: j5.bcl.v1.Schema
	(*Block)(nil),       // 1: j5.bcl.v1.Block
	(*Path)(nil),        // 2: j5.bcl.v1.Path
	(*Tag)(nil),         // 3: j5.bcl.v1.Tag
	(*Alias)(nil),       // 4: j5.bcl.v1.Alias
	(*ScalarSplit)(nil), // 5: j5.bcl.v1.ScalarSplit
}
var file_j5_bcl_v1_spec_proto_depIdxs = []int32{
	1, // 0: j5.bcl.v1.Schema.blocks:type_name -> j5.bcl.v1.Block
	3, // 1: j5.bcl.v1.Block.name:type_name -> j5.bcl.v1.Tag
	3, // 2: j5.bcl.v1.Block.type_select:type_name -> j5.bcl.v1.Tag
	3, // 3: j5.bcl.v1.Block.qualifier:type_name -> j5.bcl.v1.Tag
	4, // 4: j5.bcl.v1.Block.alias:type_name -> j5.bcl.v1.Alias
	5, // 5: j5.bcl.v1.Block.scalar_split:type_name -> j5.bcl.v1.ScalarSplit
	2, // 6: j5.bcl.v1.ScalarSplit.required_fields:type_name -> j5.bcl.v1.Path
	2, // 7: j5.bcl.v1.ScalarSplit.optional_fields:type_name -> j5.bcl.v1.Path
	2, // 8: j5.bcl.v1.ScalarSplit.remainder_field:type_name -> j5.bcl.v1.Path
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_j5_bcl_v1_spec_proto_init() }
func file_j5_bcl_v1_spec_proto_init() {
	if File_j5_bcl_v1_spec_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_bcl_v1_spec_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Schema); i {
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
		file_j5_bcl_v1_spec_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Block); i {
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
		file_j5_bcl_v1_spec_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*Path); i {
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
		file_j5_bcl_v1_spec_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*Tag); i {
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
		file_j5_bcl_v1_spec_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*Alias); i {
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
		file_j5_bcl_v1_spec_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*ScalarSplit); i {
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
	file_j5_bcl_v1_spec_proto_msgTypes[1].OneofWrappers = []any{}
	file_j5_bcl_v1_spec_proto_msgTypes[3].OneofWrappers = []any{}
	file_j5_bcl_v1_spec_proto_msgTypes[5].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_bcl_v1_spec_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_bcl_v1_spec_proto_goTypes,
		DependencyIndexes: file_j5_bcl_v1_spec_proto_depIdxs,
		MessageInfos:      file_j5_bcl_v1_spec_proto_msgTypes,
	}.Build()
	File_j5_bcl_v1_spec_proto = out.File
	file_j5_bcl_v1_spec_proto_rawDesc = nil
	file_j5_bcl_v1_spec_proto_goTypes = nil
	file_j5_bcl_v1_spec_proto_depIdxs = nil
}
