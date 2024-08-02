// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/source/v1/package.proto

package source_j5pb

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	auth_j5pb "github.com/pentops/j5/gen/j5/auth/v1/auth_j5pb"
	client_j5pb "github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	schema_j5pb "github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
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

	Packages []*Package `protobuf:"bytes,1,rep,name=packages,proto3" json:"packages,omitempty"`
}

func (x *API) Reset() {
	*x = API{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *API) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*API) ProtoMessage() {}

func (x *API) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[0]
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
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{0}
}

func (x *API) GetPackages() []*Package {
	if x != nil {
		return x.Packages
	}
	return nil
}

type Package struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Label string `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	// name of the versioned parent package, e.g. "j5.source.v1"
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// markdown formatted prose
	Prose       string                             `protobuf:"bytes,3,opt,name=prose,proto3" json:"prose,omitempty"`
	SubPackages []*SubPackage                      `protobuf:"bytes,4,rep,name=sub_packages,json=subPackages,proto3" json:"sub_packages,omitempty"`
	Schemas     map[string]*schema_j5pb.RootSchema `protobuf:"bytes,8,rep,name=schemas,proto3" json:"schemas,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Package) Reset() {
	*x = Package{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Package) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Package) ProtoMessage() {}

func (x *Package) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[1]
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
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{1}
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

func (x *Package) GetProse() string {
	if x != nil {
		return x.Prose
	}
	return ""
}

func (x *Package) GetSubPackages() []*SubPackage {
	if x != nil {
		return x.SubPackages
	}
	return nil
}

func (x *Package) GetSchemas() map[string]*schema_j5pb.RootSchema {
	if x != nil {
		return x.Schemas
	}
	return nil
}

type SubPackage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string                             `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Services []*Service                         `protobuf:"bytes,2,rep,name=services,proto3" json:"services,omitempty"`
	Topics   []*Topic                           `protobuf:"bytes,3,rep,name=topics,proto3" json:"topics,omitempty"`
	Schemas  map[string]*schema_j5pb.RootSchema `protobuf:"bytes,8,rep,name=schemas,proto3" json:"schemas,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *SubPackage) Reset() {
	*x = SubPackage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubPackage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubPackage) ProtoMessage() {}

func (x *SubPackage) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubPackage.ProtoReflect.Descriptor instead.
func (*SubPackage) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{2}
}

func (x *SubPackage) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SubPackage) GetServices() []*Service {
	if x != nil {
		return x.Services
	}
	return nil
}

func (x *SubPackage) GetTopics() []*Topic {
	if x != nil {
		return x.Topics
	}
	return nil
}

func (x *SubPackage) GetSchemas() map[string]*schema_j5pb.RootSchema {
	if x != nil {
		return x.Schemas
	}
	return nil
}

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string                    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Methods     []*Method                 `protobuf:"bytes,3,rep,name=methods,proto3" json:"methods,omitempty"`
	Type        *ServiceType              `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	DefaultAuth *auth_j5pb.MethodAuthType `protobuf:"bytes,5,opt,name=default_auth,json=defaultAuth,proto3" json:"default_auth,omitempty"`
}

func (x *Service) Reset() {
	*x = Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{3}
}

func (x *Service) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Service) GetMethods() []*Method {
	if x != nil {
		return x.Methods
	}
	return nil
}

func (x *Service) GetType() *ServiceType {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *Service) GetDefaultAuth() *auth_j5pb.MethodAuthType {
	if x != nil {
		return x.DefaultAuth
	}
	return nil
}

type ServiceType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*ServiceType_StateEntityQuery_
	//	*ServiceType_StateEntityCommand_
	Type isServiceType_Type `protobuf_oneof:"type"`
}

func (x *ServiceType) Reset() {
	*x = ServiceType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceType) ProtoMessage() {}

func (x *ServiceType) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceType.ProtoReflect.Descriptor instead.
func (*ServiceType) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{4}
}

func (m *ServiceType) GetType() isServiceType_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *ServiceType) GetStateEntityQuery() *ServiceType_StateEntityQuery {
	if x, ok := x.GetType().(*ServiceType_StateEntityQuery_); ok {
		return x.StateEntityQuery
	}
	return nil
}

func (x *ServiceType) GetStateEntityCommand() *ServiceType_StateEntityCommand {
	if x, ok := x.GetType().(*ServiceType_StateEntityCommand_); ok {
		return x.StateEntityCommand
	}
	return nil
}

type isServiceType_Type interface {
	isServiceType_Type()
}

type ServiceType_StateEntityQuery_ struct {
	StateEntityQuery *ServiceType_StateEntityQuery `protobuf:"bytes,1,opt,name=state_entity_query,json=stateEntityQuery,proto3,oneof"`
}

type ServiceType_StateEntityCommand_ struct {
	StateEntityCommand *ServiceType_StateEntityCommand `protobuf:"bytes,2,opt,name=state_entity_command,json=stateEntityCommand,proto3,oneof"`
}

func (*ServiceType_StateEntityQuery_) isServiceType_Type() {}

func (*ServiceType_StateEntityCommand_) isServiceType_Type() {}

type Method struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name           string                    `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	FullGrpcName   string                    `protobuf:"bytes,3,opt,name=full_grpc_name,json=fullGrpcName,proto3" json:"full_grpc_name,omitempty"`
	HttpMethod     client_j5pb.HTTPMethod    `protobuf:"varint,4,opt,name=http_method,json=httpMethod,proto3,enum=j5.client.v1.HTTPMethod" json:"http_method,omitempty"`
	HttpPath       string                    `protobuf:"bytes,5,opt,name=http_path,json=httpPath,proto3" json:"http_path,omitempty"`
	RequestSchema  string                    `protobuf:"bytes,6,opt,name=request_schema,json=requestSchema,proto3" json:"request_schema,omitempty"`
	ResponseSchema string                    `protobuf:"bytes,7,opt,name=response_schema,json=responseSchema,proto3" json:"response_schema,omitempty"`
	Auth           *auth_j5pb.MethodAuthType `protobuf:"bytes,8,opt,name=auth,proto3" json:"auth,omitempty"`
}

func (x *Method) Reset() {
	*x = Method{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Method) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Method) ProtoMessage() {}

func (x *Method) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[5]
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
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{5}
}

func (x *Method) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Method) GetFullGrpcName() string {
	if x != nil {
		return x.FullGrpcName
	}
	return ""
}

func (x *Method) GetHttpMethod() client_j5pb.HTTPMethod {
	if x != nil {
		return x.HttpMethod
	}
	return client_j5pb.HTTPMethod(0)
}

func (x *Method) GetHttpPath() string {
	if x != nil {
		return x.HttpPath
	}
	return ""
}

func (x *Method) GetRequestSchema() string {
	if x != nil {
		return x.RequestSchema
	}
	return ""
}

func (x *Method) GetResponseSchema() string {
	if x != nil {
		return x.ResponseSchema
	}
	return ""
}

func (x *Method) GetAuth() *auth_j5pb.MethodAuthType {
	if x != nil {
		return x.Auth
	}
	return nil
}

type Topic struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// name as specified in proto, e.g. "FooTopic"
	Name     string          `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Messages []*TopicMessage `protobuf:"bytes,3,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *Topic) Reset() {
	*x = Topic{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Topic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Topic) ProtoMessage() {}

func (x *Topic) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Topic.ProtoReflect.Descriptor instead.
func (*Topic) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{6}
}

func (x *Topic) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Topic) GetMessages() []*TopicMessage {
	if x != nil {
		return x.Messages
	}
	return nil
}

type TopicMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	FullGrpcName string `protobuf:"bytes,2,opt,name=full_grpc_name,json=fullGrpcName,proto3" json:"full_grpc_name,omitempty"`
	Schema       string `protobuf:"bytes,3,opt,name=schema,proto3" json:"schema,omitempty"`
}

func (x *TopicMessage) Reset() {
	*x = TopicMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TopicMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TopicMessage) ProtoMessage() {}

func (x *TopicMessage) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TopicMessage.ProtoReflect.Descriptor instead.
func (*TopicMessage) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{7}
}

func (x *TopicMessage) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TopicMessage) GetFullGrpcName() string {
	if x != nil {
		return x.FullGrpcName
	}
	return ""
}

func (x *TopicMessage) GetSchema() string {
	if x != nil {
		return x.Schema
	}
	return ""
}

type ServiceType_StateEntityQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entity string `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
}

func (x *ServiceType_StateEntityQuery) Reset() {
	*x = ServiceType_StateEntityQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceType_StateEntityQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceType_StateEntityQuery) ProtoMessage() {}

func (x *ServiceType_StateEntityQuery) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceType_StateEntityQuery.ProtoReflect.Descriptor instead.
func (*ServiceType_StateEntityQuery) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{4, 0}
}

func (x *ServiceType_StateEntityQuery) GetEntity() string {
	if x != nil {
		return x.Entity
	}
	return ""
}

type ServiceType_StateEntityCommand struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entity string `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
}

func (x *ServiceType_StateEntityCommand) Reset() {
	*x = ServiceType_StateEntityCommand{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_source_v1_package_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceType_StateEntityCommand) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceType_StateEntityCommand) ProtoMessage() {}

func (x *ServiceType_StateEntityCommand) ProtoReflect() protoreflect.Message {
	mi := &file_j5_source_v1_package_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceType_StateEntityCommand.ProtoReflect.Descriptor instead.
func (*ServiceType_StateEntityCommand) Descriptor() ([]byte, []int) {
	return file_j5_source_v1_package_proto_rawDescGZIP(), []int{4, 1}
}

func (x *ServiceType_StateEntityCommand) GetEntity() string {
	if x != nil {
		return x.Entity
	}
	return ""
}

var File_j5_source_v1_package_proto protoreflect.FileDescriptor

var file_j5_source_v1_package_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6a, 0x35, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6a, 0x35,
	0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x6a, 0x35, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x19, 0x6a, 0x35, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x6a, 0x35, 0x2f,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x38, 0x0a, 0x03, 0x41, 0x50, 0x49, 0x12, 0x31, 0x0a,
	0x08, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x15, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x08, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73,
	0x22, 0xba, 0x02, 0x0a, 0x07, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62,
	0x65, 0x6c, 0x12, 0x32, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x1e, 0xba, 0x48, 0x1b, 0x72, 0x19, 0x32, 0x17, 0x5e, 0x28, 0x5b, 0x61, 0x2d, 0x7a, 0x30,
	0x2d, 0x39, 0x5f, 0x5d, 0x2b, 0x5c, 0x2e, 0x29, 0x76, 0x5b, 0x30, 0x2d, 0x39, 0x5d, 0x2b, 0x24,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x73, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x0c,
	0x73, 0x75, 0x62, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x75, 0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x0b, 0x73, 0x75,
	0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x12, 0x3c, 0x0a, 0x07, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6a, 0x35, 0x2e,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x1a, 0x54, 0x0a, 0x0c, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2e, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x6f, 0x6f, 0x74, 0x53, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x97, 0x02,
	0x0a, 0x0a, 0x53, 0x75, 0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x31, 0x0a, 0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x12, 0x2b, 0x0a, 0x06, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x52, 0x06, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x73,
	0x12, 0x3f, 0x0a, 0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x75, 0x62, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x2e, 0x53, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x73, 0x1a, 0x54, 0x0a, 0x0c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x2e, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x6f, 0x6f, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xd7, 0x01, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x1a, 0xba, 0x48, 0x17, 0x72, 0x15, 0x32, 0x13, 0x5e, 0x5b, 0x41, 0x2d, 0x5a, 0x5d,
	0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a, 0x30, 0x2d, 0x39, 0x5d, 0x2a, 0x24, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x73, 0x12, 0x2d, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x61, 0x75,
	0x74, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75,
	0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x41, 0x75, 0x74, 0x68,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x0b, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x41, 0x75, 0x74,
	0x68, 0x22, 0xad, 0x02, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x5a, 0x0a, 0x12, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e,
	0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x51, 0x75, 0x65, 0x72, 0x79, 0x48, 0x00, 0x52, 0x10, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x60, 0x0a,
	0x14, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x63, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x6a, 0x35,
	0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x48, 0x00, 0x52, 0x12, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x1a,
	0x2a, 0x0a, 0x10, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x51, 0x75,
	0x65, 0x72, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x1a, 0x2c, 0x0a, 0x12, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x22, 0xef, 0x02, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x2e, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1a, 0xba, 0x48, 0x17, 0x72,
	0x15, 0x32, 0x13, 0x5e, 0x5b, 0x61, 0x2d, 0x7a, 0x5d, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a,
	0x30, 0x2d, 0x39, 0x5d, 0x2a, 0x24, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e,
	0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x67, 0x72, 0x70, 0x63, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x75, 0x6c, 0x6c, 0x47, 0x72, 0x70, 0x63, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x46, 0x0a, 0x0b, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x6a, 0x35, 0x2e, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x4d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x42, 0x0b, 0xba, 0x48, 0x08, 0x82, 0x01, 0x05, 0x10, 0x01, 0x22, 0x01, 0x00, 0x52, 0x0a,
	0x68, 0x74, 0x74, 0x70, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x37, 0x0a, 0x09, 0x68, 0x74,
	0x74, 0x70, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1a, 0xba,
	0x48, 0x17, 0x72, 0x15, 0x32, 0x13, 0x5e, 0x28, 0x5c, 0x2f, 0x3a, 0x3f, 0x5b, 0x61, 0x2d, 0x7a,
	0x30, 0x2d, 0x39, 0x5f, 0x5d, 0x2b, 0x29, 0x2b, 0x24, 0x52, 0x08, 0x68, 0x74, 0x74, 0x70, 0x50,
	0x61, 0x74, 0x68, 0x12, 0x2d, 0x0a, 0x0e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03,
	0xc8, 0x01, 0x01, 0x52, 0x0d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x12, 0x2f, 0x0a, 0x0f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x5f, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03,
	0xc8, 0x01, 0x01, 0x52, 0x0e, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x12, 0x2e, 0x0a, 0x04, 0x61, 0x75, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x41, 0x75, 0x74, 0x68, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x61,
	0x75, 0x74, 0x68, 0x22, 0x53, 0x0a, 0x05, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x36, 0x0a, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22, 0x60, 0x0a, 0x0c, 0x54, 0x6f, 0x70, 0x69,
	0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e,
	0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x67, 0x72, 0x70, 0x63, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x75, 0x6c, 0x6c, 0x47, 0x72, 0x70, 0x63, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73,
	0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6a, 0x35, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_source_v1_package_proto_rawDescOnce sync.Once
	file_j5_source_v1_package_proto_rawDescData = file_j5_source_v1_package_proto_rawDesc
)

func file_j5_source_v1_package_proto_rawDescGZIP() []byte {
	file_j5_source_v1_package_proto_rawDescOnce.Do(func() {
		file_j5_source_v1_package_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_source_v1_package_proto_rawDescData)
	})
	return file_j5_source_v1_package_proto_rawDescData
}

var file_j5_source_v1_package_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_j5_source_v1_package_proto_goTypes = []any{
	(*API)(nil),                            // 0: j5.source.v1.API
	(*Package)(nil),                        // 1: j5.source.v1.Package
	(*SubPackage)(nil),                     // 2: j5.source.v1.SubPackage
	(*Service)(nil),                        // 3: j5.source.v1.Service
	(*ServiceType)(nil),                    // 4: j5.source.v1.ServiceType
	(*Method)(nil),                         // 5: j5.source.v1.Method
	(*Topic)(nil),                          // 6: j5.source.v1.Topic
	(*TopicMessage)(nil),                   // 7: j5.source.v1.TopicMessage
	nil,                                    // 8: j5.source.v1.Package.SchemasEntry
	nil,                                    // 9: j5.source.v1.SubPackage.SchemasEntry
	(*ServiceType_StateEntityQuery)(nil),   // 10: j5.source.v1.ServiceType.StateEntityQuery
	(*ServiceType_StateEntityCommand)(nil), // 11: j5.source.v1.ServiceType.StateEntityCommand
	(*auth_j5pb.MethodAuthType)(nil),       // 12: j5.auth.v1.MethodAuthType
	(client_j5pb.HTTPMethod)(0),            // 13: j5.client.v1.HTTPMethod
	(*schema_j5pb.RootSchema)(nil),         // 14: j5.schema.v1.RootSchema
}
var file_j5_source_v1_package_proto_depIdxs = []int32{
	1,  // 0: j5.source.v1.API.packages:type_name -> j5.source.v1.Package
	2,  // 1: j5.source.v1.Package.sub_packages:type_name -> j5.source.v1.SubPackage
	8,  // 2: j5.source.v1.Package.schemas:type_name -> j5.source.v1.Package.SchemasEntry
	3,  // 3: j5.source.v1.SubPackage.services:type_name -> j5.source.v1.Service
	6,  // 4: j5.source.v1.SubPackage.topics:type_name -> j5.source.v1.Topic
	9,  // 5: j5.source.v1.SubPackage.schemas:type_name -> j5.source.v1.SubPackage.SchemasEntry
	5,  // 6: j5.source.v1.Service.methods:type_name -> j5.source.v1.Method
	4,  // 7: j5.source.v1.Service.type:type_name -> j5.source.v1.ServiceType
	12, // 8: j5.source.v1.Service.default_auth:type_name -> j5.auth.v1.MethodAuthType
	10, // 9: j5.source.v1.ServiceType.state_entity_query:type_name -> j5.source.v1.ServiceType.StateEntityQuery
	11, // 10: j5.source.v1.ServiceType.state_entity_command:type_name -> j5.source.v1.ServiceType.StateEntityCommand
	13, // 11: j5.source.v1.Method.http_method:type_name -> j5.client.v1.HTTPMethod
	12, // 12: j5.source.v1.Method.auth:type_name -> j5.auth.v1.MethodAuthType
	7,  // 13: j5.source.v1.Topic.messages:type_name -> j5.source.v1.TopicMessage
	14, // 14: j5.source.v1.Package.SchemasEntry.value:type_name -> j5.schema.v1.RootSchema
	14, // 15: j5.source.v1.SubPackage.SchemasEntry.value:type_name -> j5.schema.v1.RootSchema
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_j5_source_v1_package_proto_init() }
func file_j5_source_v1_package_proto_init() {
	if File_j5_source_v1_package_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_source_v1_package_proto_msgTypes[0].Exporter = func(v any, i int) any {
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
		file_j5_source_v1_package_proto_msgTypes[1].Exporter = func(v any, i int) any {
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
		file_j5_source_v1_package_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*SubPackage); i {
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
		file_j5_source_v1_package_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*Service); i {
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
		file_j5_source_v1_package_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*ServiceType); i {
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
		file_j5_source_v1_package_proto_msgTypes[5].Exporter = func(v any, i int) any {
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
		file_j5_source_v1_package_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*Topic); i {
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
		file_j5_source_v1_package_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*TopicMessage); i {
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
		file_j5_source_v1_package_proto_msgTypes[10].Exporter = func(v any, i int) any {
			switch v := v.(*ServiceType_StateEntityQuery); i {
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
		file_j5_source_v1_package_proto_msgTypes[11].Exporter = func(v any, i int) any {
			switch v := v.(*ServiceType_StateEntityCommand); i {
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
	file_j5_source_v1_package_proto_msgTypes[4].OneofWrappers = []any{
		(*ServiceType_StateEntityQuery_)(nil),
		(*ServiceType_StateEntityCommand_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_source_v1_package_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_source_v1_package_proto_goTypes,
		DependencyIndexes: file_j5_source_v1_package_proto_depIdxs,
		MessageInfos:      file_j5_source_v1_package_proto_msgTypes,
	}.Build()
	File_j5_source_v1_package_proto = out.File
	file_j5_source_v1_package_proto_rawDesc = nil
	file_j5_source_v1_package_proto_goTypes = nil
	file_j5_source_v1_package_proto_depIdxs = nil
}
