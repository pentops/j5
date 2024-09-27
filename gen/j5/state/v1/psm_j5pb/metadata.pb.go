// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/state/v1/metadata.proto

package psm_j5pb

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	auth_j5pb "github.com/pentops/j5/gen/j5/auth/v1/auth_j5pb"
	_ "github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
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

type StateMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Time of the first event on the state machine
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// Time of the most recent event on the state machine
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	// Sequcence number of the most recent event on the state machine
	LastSequence uint64 `protobuf:"varint,3,opt,name=last_sequence,json=lastSequence,proto3" json:"last_sequence,omitempty"`
}

func (x *StateMetadata) Reset() {
	*x = StateMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StateMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StateMetadata) ProtoMessage() {}

func (x *StateMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StateMetadata.ProtoReflect.Descriptor instead.
func (*StateMetadata) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *StateMetadata) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *StateMetadata) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *StateMetadata) GetLastSequence() uint64 {
	if x != nil {
		return x.LastSequence
	}
	return 0
}

type EventMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId string `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	// Sequence within the state machine. Discrete, 1,2,3
	Sequence  uint64                 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Cause     *Cause                 `protobuf:"bytes,4,opt,name=cause,proto3" json:"cause,omitempty"`
}

func (x *EventMetadata) Reset() {
	*x = EventMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventMetadata) ProtoMessage() {}

func (x *EventMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventMetadata.ProtoReflect.Descriptor instead.
func (*EventMetadata) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{1}
}

func (x *EventMetadata) GetEventId() string {
	if x != nil {
		return x.EventId
	}
	return ""
}

func (x *EventMetadata) GetSequence() uint64 {
	if x != nil {
		return x.Sequence
	}
	return 0
}

func (x *EventMetadata) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *EventMetadata) GetCause() *Cause {
	if x != nil {
		return x.Cause
	}
	return nil
}

type EventPublishMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId string `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	// Sequence within the state machine. Discrete, 1,2,3
	Sequence  uint64                 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Cause     *Cause                 `protobuf:"bytes,4,opt,name=cause,proto3" json:"cause,omitempty"`
}

func (x *EventPublishMetadata) Reset() {
	*x = EventPublishMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventPublishMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventPublishMetadata) ProtoMessage() {}

func (x *EventPublishMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventPublishMetadata.ProtoReflect.Descriptor instead.
func (*EventPublishMetadata) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{2}
}

func (x *EventPublishMetadata) GetEventId() string {
	if x != nil {
		return x.EventId
	}
	return ""
}

func (x *EventPublishMetadata) GetSequence() uint64 {
	if x != nil {
		return x.Sequence
	}
	return 0
}

func (x *EventPublishMetadata) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *EventPublishMetadata) GetCause() *Cause {
	if x != nil {
		return x.Cause
	}
	return nil
}

// Events are caused by either an actor external to the boundary, an application
// within the boundary, the state machine itself,
type Cause struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*Cause_PsmEvent
	//	*Cause_Command
	//	*Cause_ExternalEvent
	//	*Cause_Reply
	Type isCause_Type `protobuf_oneof:"type"`
}

func (x *Cause) Reset() {
	*x = Cause{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cause) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cause) ProtoMessage() {}

func (x *Cause) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cause.ProtoReflect.Descriptor instead.
func (*Cause) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{3}
}

func (m *Cause) GetType() isCause_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Cause) GetPsmEvent() *PSMEventCause {
	if x, ok := x.GetType().(*Cause_PsmEvent); ok {
		return x.PsmEvent
	}
	return nil
}

func (x *Cause) GetCommand() *auth_j5pb.Action {
	if x, ok := x.GetType().(*Cause_Command); ok {
		return x.Command
	}
	return nil
}

func (x *Cause) GetExternalEvent() *ExternalEventCause {
	if x, ok := x.GetType().(*Cause_ExternalEvent); ok {
		return x.ExternalEvent
	}
	return nil
}

func (x *Cause) GetReply() *ReplyCause {
	if x, ok := x.GetType().(*Cause_Reply); ok {
		return x.Reply
	}
	return nil
}

type isCause_Type interface {
	isCause_Type()
}

type Cause_PsmEvent struct {
	PsmEvent *PSMEventCause `protobuf:"bytes,1,opt,name=psm_event,json=psmEvent,proto3,oneof"`
}

type Cause_Command struct {
	Command *auth_j5pb.Action `protobuf:"bytes,2,opt,name=command,proto3,oneof"`
}

type Cause_ExternalEvent struct {
	ExternalEvent *ExternalEventCause `protobuf:"bytes,3,opt,name=external_event,json=externalEvent,proto3,oneof"`
}

type Cause_Reply struct {
	Reply *ReplyCause `protobuf:"bytes,4,opt,name=reply,proto3,oneof"`
}

func (*Cause_PsmEvent) isCause_Type() {}

func (*Cause_Command) isCause_Type() {}

func (*Cause_ExternalEvent) isCause_Type() {}

func (*Cause_Reply) isCause_Type() {}

// The event was caused by a transition in this or another state machine
type PSMEventCause struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The ID of the event that caused this event
	EventId string `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	// The identity of the state machine for the event.
	// {package}.{name}, where name is the annotated name in
	// j5.state.v1.(State|Event)ObjectOptions.name
	// e.g. "foo.bar.v1.foobar" (not foo.bar.v1.FooBarState)
	StateMachine string `protobuf:"bytes,2,opt,name=state_machine,json=stateMachine,proto3" json:"state_machine,omitempty"`
}

func (x *PSMEventCause) Reset() {
	*x = PSMEventCause{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PSMEventCause) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PSMEventCause) ProtoMessage() {}

func (x *PSMEventCause) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PSMEventCause.ProtoReflect.Descriptor instead.
func (*PSMEventCause) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{4}
}

func (x *PSMEventCause) GetEventId() string {
	if x != nil {
		return x.EventId
	}
	return ""
}

func (x *PSMEventCause) GetStateMachine() string {
	if x != nil {
		return x.StateMachine
	}
	return ""
}

// An external system replying to a side-effect request, e.g. where the
// application sends a request to a vendor system, and the vendor replies.
type ReplyCause struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request *PSMEventCause `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	// When true, the reply was via an event input from the external system, rather than
	// a simple request-reply like a HTTP call. This means the event was matched
	// using lookup keys in the vendor's event, and therefore not guaranteed to be
	// linked to the correct state machine.
	Async bool `protobuf:"varint,2,opt,name=async,proto3" json:"async,omitempty"`
}

func (x *ReplyCause) Reset() {
	*x = ReplyCause{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReplyCause) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplyCause) ProtoMessage() {}

func (x *ReplyCause) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplyCause.ProtoReflect.Descriptor instead.
func (*ReplyCause) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{5}
}

func (x *ReplyCause) GetRequest() *PSMEventCause {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *ReplyCause) GetAsync() bool {
	if x != nil {
		return x.Async
	}
	return false
}

// The event was caused by an external event, e.g. a webhook, a message from a queue, etc.
type ExternalEventCause struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the external system that caused the event. No specific format
	// or rules.
	SystemName string `protobuf:"bytes,1,opt,name=system_name,json=systemName,proto3" json:"system_name,omitempty"`
	// The name of the event in the external system. No specific format or rules.
	EventName string `protobuf:"bytes,2,opt,name=event_name,json=eventName,proto3" json:"event_name,omitempty"`
	// The ID of the event in the external system as defined by that system.
	// ID generation must consistently derivable from the source event.
	// Do not make up IDs from the// current system time or random
	// Leave nil if an acceptable unique ID is not available.
	ExternalId *string `protobuf:"bytes,3,opt,name=external_id,json=externalId,proto3,oneof" json:"external_id,omitempty"`
}

func (x *ExternalEventCause) Reset() {
	*x = ExternalEventCause{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_state_v1_metadata_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExternalEventCause) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExternalEventCause) ProtoMessage() {}

func (x *ExternalEventCause) ProtoReflect() protoreflect.Message {
	mi := &file_j5_state_v1_metadata_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExternalEventCause.ProtoReflect.Descriptor instead.
func (*ExternalEventCause) Descriptor() ([]byte, []int) {
	return file_j5_state_v1_metadata_proto_rawDescGZIP(), []int{6}
}

func (x *ExternalEventCause) GetSystemName() string {
	if x != nil {
		return x.SystemName
	}
	return ""
}

func (x *ExternalEventCause) GetEventName() string {
	if x != nil {
		return x.EventName
	}
	return ""
}

func (x *ExternalEventCause) GetExternalId() string {
	if x != nil && x.ExternalId != nil {
		return *x.ExternalId
	}
	return ""
}

var File_j5_state_v1_metadata_proto protoreflect.FileDescriptor

var file_j5_state_v1_metadata_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6a, 0x35, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6a, 0x35,
	0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x6a, 0x35, 0x2f, 0x61, 0x75,
	0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1c, 0x6a, 0x35, 0x2f, 0x6c, 0x69, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xca, 0x01, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x12, 0x4a, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x42, 0x0f, 0x8a, 0xf7, 0x98, 0xc6, 0x02, 0x09, 0xf2, 0x01, 0x06, 0x5a, 0x04, 0x08, 0x01,
	0x10, 0x01, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x48, 0x0a,
	0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0d, 0x8a,
	0xf7, 0x98, 0xc6, 0x02, 0x07, 0xf2, 0x01, 0x04, 0x5a, 0x02, 0x08, 0x01, 0x52, 0x09, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x5f,
	0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c,
	0x6c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x22, 0xcb, 0x01, 0x0a,
	0x0d, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x23,
	0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x08, 0xba, 0x48, 0x05, 0x72, 0x03, 0xb0, 0x01, 0x01, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x12,
	0x4f, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x15,
	0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x8a, 0xf7, 0x98, 0xc6, 0x02, 0x09, 0xf2, 0x01, 0x06, 0x5a,
	0x04, 0x08, 0x01, 0x10, 0x01, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x28, 0x0a, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61,
	0x75, 0x73, 0x65, 0x52, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x22, 0xd2, 0x01, 0x0a, 0x14, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x23, 0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xba, 0x48, 0x05, 0x72, 0x03, 0xb0, 0x01, 0x01, 0x52,
	0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x71, 0x75,
	0x65, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x73, 0x65, 0x71, 0x75,
	0x65, 0x6e, 0x63, 0x65, 0x12, 0x4f, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x42, 0x15, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x8a, 0xf7, 0x98, 0xc6, 0x02,
	0x09, 0xf2, 0x01, 0x06, 0x5a, 0x04, 0x08, 0x01, 0x10, 0x01, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x28, 0x0a, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x61, 0x75, 0x73, 0x65, 0x52, 0x05, 0x63, 0x61, 0x75, 0x73, 0x65, 0x22,
	0xf5, 0x01, 0x0a, 0x05, 0x43, 0x61, 0x75, 0x73, 0x65, 0x12, 0x39, 0x0a, 0x09, 0x70, 0x73, 0x6d,
	0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6a,
	0x35, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x53, 0x4d, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x43, 0x61, 0x75, 0x73, 0x65, 0x48, 0x00, 0x52, 0x08, 0x70, 0x73, 0x6d, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x2e, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x76, 0x31, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x07, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x12, 0x48, 0x0a, 0x0e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6a,
	0x35, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x43, 0x61, 0x75, 0x73, 0x65, 0x48, 0x00, 0x52,
	0x0d, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x2f,
	0x0a, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x6a, 0x35, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x43, 0x61, 0x75, 0x73, 0x65, 0x48, 0x00, 0x52, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x42,
	0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x59, 0x0a, 0x0d, 0x50, 0x53, 0x4d, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x43, 0x61, 0x75, 0x73, 0x65, 0x12, 0x23, 0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xba, 0x48, 0x05, 0x72,
	0x03, 0xb0, 0x01, 0x01, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x23, 0x0a,
	0x0d, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x74, 0x61, 0x74, 0x65, 0x4d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x22, 0x58, 0x0a, 0x0a, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x43, 0x61, 0x75, 0x73, 0x65,
	0x12, 0x34, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x6a, 0x35, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x50, 0x53, 0x4d, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x43, 0x61, 0x75, 0x73, 0x65, 0x52, 0x07, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x73, 0x79, 0x6e, 0x63, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x61, 0x73, 0x79, 0x6e, 0x63, 0x22, 0x8a, 0x01, 0x0a,
	0x12, 0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x43, 0x61,
	0x75, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0b, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0a, 0x65, 0x78, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x49, 0x64, 0x88, 0x01, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x65, 0x78,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f,
	0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2f,
	0x76, 0x31, 0x2f, 0x70, 0x73, 0x6d, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_j5_state_v1_metadata_proto_rawDescOnce sync.Once
	file_j5_state_v1_metadata_proto_rawDescData = file_j5_state_v1_metadata_proto_rawDesc
)

func file_j5_state_v1_metadata_proto_rawDescGZIP() []byte {
	file_j5_state_v1_metadata_proto_rawDescOnce.Do(func() {
		file_j5_state_v1_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_state_v1_metadata_proto_rawDescData)
	})
	return file_j5_state_v1_metadata_proto_rawDescData
}

var file_j5_state_v1_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_j5_state_v1_metadata_proto_goTypes = []any{
	(*StateMetadata)(nil),         // 0: j5.state.v1.StateMetadata
	(*EventMetadata)(nil),         // 1: j5.state.v1.EventMetadata
	(*EventPublishMetadata)(nil),  // 2: j5.state.v1.EventPublishMetadata
	(*Cause)(nil),                 // 3: j5.state.v1.Cause
	(*PSMEventCause)(nil),         // 4: j5.state.v1.PSMEventCause
	(*ReplyCause)(nil),            // 5: j5.state.v1.ReplyCause
	(*ExternalEventCause)(nil),    // 6: j5.state.v1.ExternalEventCause
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
	(*auth_j5pb.Action)(nil),      // 8: j5.auth.v1.Action
}
var file_j5_state_v1_metadata_proto_depIdxs = []int32{
	7,  // 0: j5.state.v1.StateMetadata.created_at:type_name -> google.protobuf.Timestamp
	7,  // 1: j5.state.v1.StateMetadata.updated_at:type_name -> google.protobuf.Timestamp
	7,  // 2: j5.state.v1.EventMetadata.timestamp:type_name -> google.protobuf.Timestamp
	3,  // 3: j5.state.v1.EventMetadata.cause:type_name -> j5.state.v1.Cause
	7,  // 4: j5.state.v1.EventPublishMetadata.timestamp:type_name -> google.protobuf.Timestamp
	3,  // 5: j5.state.v1.EventPublishMetadata.cause:type_name -> j5.state.v1.Cause
	4,  // 6: j5.state.v1.Cause.psm_event:type_name -> j5.state.v1.PSMEventCause
	8,  // 7: j5.state.v1.Cause.command:type_name -> j5.auth.v1.Action
	6,  // 8: j5.state.v1.Cause.external_event:type_name -> j5.state.v1.ExternalEventCause
	5,  // 9: j5.state.v1.Cause.reply:type_name -> j5.state.v1.ReplyCause
	4,  // 10: j5.state.v1.ReplyCause.request:type_name -> j5.state.v1.PSMEventCause
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_j5_state_v1_metadata_proto_init() }
func file_j5_state_v1_metadata_proto_init() {
	if File_j5_state_v1_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_state_v1_metadata_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*StateMetadata); i {
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
		file_j5_state_v1_metadata_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*EventMetadata); i {
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
		file_j5_state_v1_metadata_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*EventPublishMetadata); i {
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
		file_j5_state_v1_metadata_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*Cause); i {
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
		file_j5_state_v1_metadata_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*PSMEventCause); i {
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
		file_j5_state_v1_metadata_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*ReplyCause); i {
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
		file_j5_state_v1_metadata_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ExternalEventCause); i {
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
	file_j5_state_v1_metadata_proto_msgTypes[3].OneofWrappers = []any{
		(*Cause_PsmEvent)(nil),
		(*Cause_Command)(nil),
		(*Cause_ExternalEvent)(nil),
		(*Cause_Reply)(nil),
	}
	file_j5_state_v1_metadata_proto_msgTypes[6].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_state_v1_metadata_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_state_v1_metadata_proto_goTypes,
		DependencyIndexes: file_j5_state_v1_metadata_proto_depIdxs,
		MessageInfos:      file_j5_state_v1_metadata_proto_msgTypes,
	}.Build()
	File_j5_state_v1_metadata_proto = out.File
	file_j5_state_v1_metadata_proto_rawDesc = nil
	file_j5_state_v1_metadata_proto_goTypes = nil
	file_j5_state_v1_metadata_proto_depIdxs = nil
}
