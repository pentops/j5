// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: j5/schema/v1/document.proto

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

type API struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Packages []*Package         `protobuf:"bytes,1,rep,name=packages,proto3" json:"packages,omitempty"`
	Schemas  map[string]*Schema `protobuf:"bytes,2,rep,name=schemas,proto3" json:"schemas,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *API) Reset() {
	*x = API{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_document_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *API) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*API) ProtoMessage() {}

func (x *API) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_document_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use API.ProtoReflect.Descriptor instead.
func (*API) Descriptor() ([]byte, []int) {
	return file_j5_schema_v1_document_proto_rawDescGZIP(), []int{0}
}

func (x *API) GetPackages() []*Package {
	if x != nil {
		return x.Packages
	}
	return nil
}

func (x *API) GetSchemas() map[string]*Schema {
	if x != nil {
		return x.Schemas
	}
	return nil
}

type Package struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Label        string       `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	Name         string       `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Hidden       bool         `protobuf:"varint,3,opt,name=hidden,proto3" json:"hidden,omitempty"`
	Introduction string       `protobuf:"bytes,4,opt,name=introduction,proto3" json:"introduction,omitempty"`
	Methods      []*Method    `protobuf:"bytes,5,rep,name=methods,proto3" json:"methods,omitempty"`
	Entities     []*Entity    `protobuf:"bytes,6,rep,name=entities,proto3" json:"entities,omitempty"`
	Events       []*EventSpec `protobuf:"bytes,7,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *Package) Reset() {
	*x = Package{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_document_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Package) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Package) ProtoMessage() {}

func (x *Package) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_document_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Package.ProtoReflect.Descriptor instead.
func (*Package) Descriptor() ([]byte, []int) {
	return file_j5_schema_v1_document_proto_rawDescGZIP(), []int{1}
}

func (x *Package) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *Package) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Package) GetHidden() bool {
	if x != nil {
		return x.Hidden
	}
	return false
}

func (x *Package) GetIntroduction() string {
	if x != nil {
		return x.Introduction
	}
	return ""
}

func (x *Package) GetMethods() []*Method {
	if x != nil {
		return x.Methods
	}
	return nil
}

func (x *Package) GetEntities() []*Entity {
	if x != nil {
		return x.Entities
	}
	return nil
}

func (x *Package) GetEvents() []*EventSpec {
	if x != nil {
		return x.Events
	}
	return nil
}

type Method struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GrpcServiceName string       `protobuf:"bytes,1,opt,name=grpc_service_name,json=grpcServiceName,proto3" json:"grpc_service_name,omitempty"`
	GrpcMethodName  string       `protobuf:"bytes,2,opt,name=grpc_method_name,json=grpcMethodName,proto3" json:"grpc_method_name,omitempty"`
	FullGrpcName    string       `protobuf:"bytes,3,opt,name=full_grpc_name,json=fullGrpcName,proto3" json:"full_grpc_name,omitempty"`
	HttpMethod      string       `protobuf:"bytes,4,opt,name=http_method,json=httpMethod,proto3" json:"http_method,omitempty"`
	HttpPath        string       `protobuf:"bytes,5,opt,name=http_path,json=httpPath,proto3" json:"http_path,omitempty"`
	RequestBody     *Schema      `protobuf:"bytes,6,opt,name=request_body,json=requestBody,proto3" json:"request_body,omitempty"`
	ResponseBody    *Schema      `protobuf:"bytes,7,opt,name=response_body,json=responseBody,proto3" json:"response_body,omitempty"`
	QueryParameters []*Parameter `protobuf:"bytes,8,rep,name=query_parameters,json=queryParameters,proto3" json:"query_parameters,omitempty"`
	PathParameters  []*Parameter `protobuf:"bytes,9,rep,name=path_parameters,json=pathParameters,proto3" json:"path_parameters,omitempty"`
}

func (x *Method) Reset() {
	*x = Method{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_document_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Method) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Method) ProtoMessage() {}

func (x *Method) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_document_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Method.ProtoReflect.Descriptor instead.
func (*Method) Descriptor() ([]byte, []int) {
	return file_j5_schema_v1_document_proto_rawDescGZIP(), []int{2}
}

func (x *Method) GetGrpcServiceName() string {
	if x != nil {
		return x.GrpcServiceName
	}
	return ""
}

func (x *Method) GetGrpcMethodName() string {
	if x != nil {
		return x.GrpcMethodName
	}
	return ""
}

func (x *Method) GetFullGrpcName() string {
	if x != nil {
		return x.FullGrpcName
	}
	return ""
}

func (x *Method) GetHttpMethod() string {
	if x != nil {
		return x.HttpMethod
	}
	return ""
}

func (x *Method) GetHttpPath() string {
	if x != nil {
		return x.HttpPath
	}
	return ""
}

func (x *Method) GetRequestBody() *Schema {
	if x != nil {
		return x.RequestBody
	}
	return nil
}

func (x *Method) GetResponseBody() *Schema {
	if x != nil {
		return x.ResponseBody
	}
	return nil
}

func (x *Method) GetQueryParameters() []*Parameter {
	if x != nil {
		return x.QueryParameters
	}
	return nil
}

func (x *Method) GetPathParameters() []*Parameter {
	if x != nil {
		return x.PathParameters
	}
	return nil
}

type EventSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string  `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	StateSchema *Schema `protobuf:"bytes,2,opt,name=state_schema,json=stateSchema,proto3" json:"state_schema,omitempty"`
	EventSchema *Schema `protobuf:"bytes,3,opt,name=event_schema,json=eventSchema,proto3" json:"event_schema,omitempty"`
}

func (x *EventSpec) Reset() {
	*x = EventSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_document_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventSpec) ProtoMessage() {}

func (x *EventSpec) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_document_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventSpec.ProtoReflect.Descriptor instead.
func (*EventSpec) Descriptor() ([]byte, []int) {
	return file_j5_schema_v1_document_proto_rawDescGZIP(), []int{3}
}

func (x *EventSpec) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *EventSpec) GetStateSchema() *Schema {
	if x != nil {
		return x.StateSchema
	}
	return nil
}

func (x *EventSpec) GetEventSchema() *Schema {
	if x != nil {
		return x.EventSchema
	}
	return nil
}

type Entity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StateSchema *Schema `protobuf:"bytes,1,opt,name=state_schema,json=stateSchema,proto3" json:"state_schema,omitempty"`
	EventSchema *Schema `protobuf:"bytes,2,opt,name=event_schema,json=eventSchema,proto3" json:"event_schema,omitempty"`
}

func (x *Entity) Reset() {
	*x = Entity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_document_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entity) ProtoMessage() {}

func (x *Entity) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_document_proto_msgTypes[4]
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
	return file_j5_schema_v1_document_proto_rawDescGZIP(), []int{4}
}

func (x *Entity) GetStateSchema() *Schema {
	if x != nil {
		return x.StateSchema
	}
	return nil
}

func (x *Entity) GetEventSchema() *Schema {
	if x != nil {
		return x.EventSchema
	}
	return nil
}

type Parameter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string  `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description string  `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Required    bool    `protobuf:"varint,4,opt,name=required,proto3" json:"required,omitempty"`
	Schema      *Schema `protobuf:"bytes,5,opt,name=schema,proto3" json:"schema,omitempty"`
}

func (x *Parameter) Reset() {
	*x = Parameter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_schema_v1_document_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Parameter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Parameter) ProtoMessage() {}

func (x *Parameter) ProtoReflect() protoreflect.Message {
	mi := &file_j5_schema_v1_document_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Parameter.ProtoReflect.Descriptor instead.
func (*Parameter) Descriptor() ([]byte, []int) {
	return file_j5_schema_v1_document_proto_rawDescGZIP(), []int{5}
}

func (x *Parameter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Parameter) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Parameter) GetRequired() bool {
	if x != nil {
		return x.Required
	}
	return false
}

func (x *Parameter) GetSchema() *Schema {
	if x != nil {
		return x.Schema
	}
	return nil
}

var File_j5_schema_v1_document_proto protoreflect.FileDescriptor

var file_j5_schema_v1_document_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x6a, 0x35, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x64,
	0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6a,
	0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x1a, 0x19, 0x6a, 0x35, 0x2f,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc4, 0x01, 0x0a, 0x03, 0x41, 0x50, 0x49, 0x12, 0x31,
	0x0a, 0x08, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x15, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e,
	0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x08, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x12, 0x38, 0x0a, 0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x50, 0x49, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x1a, 0x50, 0x0a, 0x0c, 0x53,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6a,
	0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x82, 0x02,
	0x0a, 0x07, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62,
	0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x69, 0x64, 0x64, 0x65, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x06, 0x68, 0x69, 0x64, 0x64, 0x65, 0x6e, 0x12, 0x22, 0x0a, 0x0c, 0x69,
	0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x69, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x2e, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e,
	0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x73, 0x12,
	0x30, 0x0a, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31,
	0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65,
	0x73, 0x12, 0x2f, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31,
	0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x70, 0x65, 0x63, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x22, 0xbc, 0x03, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x2a, 0x0a,
	0x11, 0x67, 0x72, 0x70, 0x63, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x67, 0x72, 0x70, 0x63, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x67, 0x72, 0x70,
	0x63, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0e, 0x67, 0x72, 0x70, 0x63, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x67, 0x72, 0x70, 0x63,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x75, 0x6c,
	0x6c, 0x47, 0x72, 0x70, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x68, 0x74, 0x74,
	0x70, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x68, 0x74, 0x74, 0x70, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x68, 0x74,
	0x74, 0x70, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68,
	0x74, 0x74, 0x70, 0x50, 0x61, 0x74, 0x68, 0x12, 0x37, 0x0a, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x5f, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e,
	0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x52, 0x0b, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x42, 0x6f, 0x64, 0x79,
	0x12, 0x39, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x62, 0x6f, 0x64,
	0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x0c, 0x72,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x42, 0x0a, 0x10, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18,
	0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x52, 0x0f,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x12,
	0x40, 0x0a, 0x0f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65,
	0x72, 0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65,
	0x72, 0x52, 0x0e, 0x70, 0x61, 0x74, 0x68, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72,
	0x73, 0x22, 0x91, 0x01, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x70, 0x65, 0x63, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x37, 0x0a, 0x0c, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52,
	0x0b, 0x73, 0x74, 0x61, 0x74, 0x65, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x37, 0x0a, 0x0c,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x0b, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x53,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x7a, 0x0a, 0x06, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12,
	0x37, 0x0a, 0x0c, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x0b, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x37, 0x0a, 0x0c, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x52, 0x0b, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x22, 0x8b, 0x01, 0x0a, 0x09, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65,
	0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65,
	0x64, 0x12, 0x2c, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x42,
	0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65,
	0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_j5_schema_v1_document_proto_rawDescOnce sync.Once
	file_j5_schema_v1_document_proto_rawDescData = file_j5_schema_v1_document_proto_rawDesc
)

func file_j5_schema_v1_document_proto_rawDescGZIP() []byte {
	file_j5_schema_v1_document_proto_rawDescOnce.Do(func() {
		file_j5_schema_v1_document_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_schema_v1_document_proto_rawDescData)
	})
	return file_j5_schema_v1_document_proto_rawDescData
}

var file_j5_schema_v1_document_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_j5_schema_v1_document_proto_goTypes = []interface{}{
	(*API)(nil),       // 0: j5.schema.v1.API
	(*Package)(nil),   // 1: j5.schema.v1.Package
	(*Method)(nil),    // 2: j5.schema.v1.Method
	(*EventSpec)(nil), // 3: j5.schema.v1.EventSpec
	(*Entity)(nil),    // 4: j5.schema.v1.Entity
	(*Parameter)(nil), // 5: j5.schema.v1.Parameter
	nil,               // 6: j5.schema.v1.API.SchemasEntry
	(*Schema)(nil),    // 7: j5.schema.v1.Schema
}
var file_j5_schema_v1_document_proto_depIdxs = []int32{
	1,  // 0: j5.schema.v1.API.packages:type_name -> j5.schema.v1.Package
	6,  // 1: j5.schema.v1.API.schemas:type_name -> j5.schema.v1.API.SchemasEntry
	2,  // 2: j5.schema.v1.Package.methods:type_name -> j5.schema.v1.Method
	4,  // 3: j5.schema.v1.Package.entities:type_name -> j5.schema.v1.Entity
	3,  // 4: j5.schema.v1.Package.events:type_name -> j5.schema.v1.EventSpec
	7,  // 5: j5.schema.v1.Method.request_body:type_name -> j5.schema.v1.Schema
	7,  // 6: j5.schema.v1.Method.response_body:type_name -> j5.schema.v1.Schema
	5,  // 7: j5.schema.v1.Method.query_parameters:type_name -> j5.schema.v1.Parameter
	5,  // 8: j5.schema.v1.Method.path_parameters:type_name -> j5.schema.v1.Parameter
	7,  // 9: j5.schema.v1.EventSpec.state_schema:type_name -> j5.schema.v1.Schema
	7,  // 10: j5.schema.v1.EventSpec.event_schema:type_name -> j5.schema.v1.Schema
	7,  // 11: j5.schema.v1.Entity.state_schema:type_name -> j5.schema.v1.Schema
	7,  // 12: j5.schema.v1.Entity.event_schema:type_name -> j5.schema.v1.Schema
	7,  // 13: j5.schema.v1.Parameter.schema:type_name -> j5.schema.v1.Schema
	7,  // 14: j5.schema.v1.API.SchemasEntry.value:type_name -> j5.schema.v1.Schema
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_j5_schema_v1_document_proto_init() }
func file_j5_schema_v1_document_proto_init() {
	if File_j5_schema_v1_document_proto != nil {
		return
	}
	file_j5_schema_v1_schema_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_j5_schema_v1_document_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*API); i {
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
		file_j5_schema_v1_document_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Package); i {
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
		file_j5_schema_v1_document_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Method); i {
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
		file_j5_schema_v1_document_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventSpec); i {
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
		file_j5_schema_v1_document_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
		file_j5_schema_v1_document_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Parameter); i {
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
			RawDescriptor: file_j5_schema_v1_document_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_schema_v1_document_proto_goTypes,
		DependencyIndexes: file_j5_schema_v1_document_proto_depIdxs,
		MessageInfos:      file_j5_schema_v1_document_proto_msgTypes,
	}.Build()
	File_j5_schema_v1_document_proto = out.File
	file_j5_schema_v1_document_proto_rawDesc = nil
	file_j5_schema_v1_document_proto_goTypes = nil
	file_j5_schema_v1_document_proto_depIdxs = nil
}