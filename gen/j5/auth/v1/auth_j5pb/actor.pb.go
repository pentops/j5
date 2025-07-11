// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: j5/auth/v1/actor.proto

package auth_j5pb

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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

type Action struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Method         string       `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Actor          *Actor       `protobuf:"bytes,2,opt,name=actor,proto3" json:"actor,omitempty"`
	Fingerprint    *Fingerprint `protobuf:"bytes,3,opt,name=fingerprint,proto3" json:"fingerprint,omitempty"`
	IdempotencyKey string       `protobuf:"bytes,4,opt,name=idempotency_key,json=idempotencyKey,proto3" json:"idempotency_key,omitempty"`
}

func (x *Action) Reset() {
	*x = Action{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Action) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Action) ProtoMessage() {}

func (x *Action) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Action.ProtoReflect.Descriptor instead.
func (*Action) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{0}
}

func (x *Action) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Action) GetActor() *Actor {
	if x != nil {
		return x.Actor
	}
	return nil
}

func (x *Action) GetFingerprint() *Fingerprint {
	if x != nil {
		return x.Fingerprint
	}
	return nil
}

func (x *Action) GetIdempotencyKey() string {
	if x != nil {
		return x.IdempotencyKey
	}
	return ""
}

type Fingerprint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The IP address of the client as best as can be determined
	IpAddress *string `protobuf:"bytes,1,opt,name=ip_address,json=ipAddress,proto3,oneof" json:"ip_address,omitempty"`
	// The provided user agent string of the client.
	UserAgent *string `protobuf:"bytes,2,opt,name=user_agent,json=userAgent,proto3,oneof" json:"user_agent,omitempty"`
}

func (x *Fingerprint) Reset() {
	*x = Fingerprint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Fingerprint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fingerprint) ProtoMessage() {}

func (x *Fingerprint) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fingerprint.ProtoReflect.Descriptor instead.
func (*Fingerprint) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{1}
}

func (x *Fingerprint) GetIpAddress() string {
	if x != nil && x.IpAddress != nil {
		return *x.IpAddress
	}
	return ""
}

func (x *Fingerprint) GetUserAgent() string {
	if x != nil && x.UserAgent != nil {
		return *x.UserAgent
	}
	return ""
}

type Actor struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique identifier of the actor, derived from the various actor type
	// methods.
	SubjectId string `protobuf:"bytes,1,opt,name=subject_id,json=subjectId,proto3" json:"subject_id,omitempty"`
	// Free string identifying the type of the actor, e.g. 'user', 'service',
	// defined by the authenticating system. (subject IDs must still be unique without
	// considering subject_type)
	SubjectType          string                `protobuf:"bytes,2,opt,name=subject_type,json=subjectType,proto3" json:"subject_type,omitempty"`
	AuthenticationMethod *AuthenticationMethod `protobuf:"bytes,3,opt,name=authentication_method,json=authenticationMethod,proto3" json:"authentication_method,omitempty"`
	Claim                *Claim                `protobuf:"bytes,4,opt,name=claim,proto3" json:"claim,omitempty"`
	// Arbitrary tags that are defined by the authorizing system to quickly
	// identify the user e.g. the user's email address, API Key Name, etc.
	// Must not be used in authorization logic, and should not be used as a
	// the primary source of the actor's identity.
	ActorTags map[string]string `protobuf:"bytes,5,rep,name=actor_tags,json=actorTags,proto3" json:"actor_tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Actor) Reset() {
	*x = Actor{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Actor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Actor) ProtoMessage() {}

func (x *Actor) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Actor.ProtoReflect.Descriptor instead.
func (*Actor) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{2}
}

func (x *Actor) GetSubjectId() string {
	if x != nil {
		return x.SubjectId
	}
	return ""
}

func (x *Actor) GetSubjectType() string {
	if x != nil {
		return x.SubjectType
	}
	return ""
}

func (x *Actor) GetAuthenticationMethod() *AuthenticationMethod {
	if x != nil {
		return x.AuthenticationMethod
	}
	return nil
}

func (x *Actor) GetClaim() *Claim {
	if x != nil {
		return x.Claim
	}
	return nil
}

func (x *Actor) GetActorTags() map[string]string {
	if x != nil {
		return x.ActorTags
	}
	return nil
}

type AuthenticationMethod struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*AuthenticationMethod_Jwt
	//	*AuthenticationMethod_Session_
	//	*AuthenticationMethod_External_
	Type isAuthenticationMethod_Type `protobuf_oneof:"type"`
}

func (x *AuthenticationMethod) Reset() {
	*x = AuthenticationMethod{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationMethod) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationMethod) ProtoMessage() {}

func (x *AuthenticationMethod) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationMethod.ProtoReflect.Descriptor instead.
func (*AuthenticationMethod) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{3}
}

func (m *AuthenticationMethod) GetType() isAuthenticationMethod_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *AuthenticationMethod) GetJwt() *AuthenticationMethod_JWT {
	if x, ok := x.GetType().(*AuthenticationMethod_Jwt); ok {
		return x.Jwt
	}
	return nil
}

func (x *AuthenticationMethod) GetSession() *AuthenticationMethod_Session {
	if x, ok := x.GetType().(*AuthenticationMethod_Session_); ok {
		return x.Session
	}
	return nil
}

func (x *AuthenticationMethod) GetExternal() *AuthenticationMethod_External {
	if x, ok := x.GetType().(*AuthenticationMethod_External_); ok {
		return x.External
	}
	return nil
}

type isAuthenticationMethod_Type interface {
	isAuthenticationMethod_Type()
}

type AuthenticationMethod_Jwt struct {
	Jwt *AuthenticationMethod_JWT `protobuf:"bytes,1,opt,name=jwt,proto3,oneof"`
}

type AuthenticationMethod_Session_ struct {
	Session *AuthenticationMethod_Session `protobuf:"bytes,2,opt,name=session,proto3,oneof"`
}

type AuthenticationMethod_External_ struct {
	External *AuthenticationMethod_External `protobuf:"bytes,3,opt,name=external,proto3,oneof"`
}

func (*AuthenticationMethod_Jwt) isAuthenticationMethod_Type() {}

func (*AuthenticationMethod_Session_) isAuthenticationMethod_Type() {}

func (*AuthenticationMethod_External_) isAuthenticationMethod_Type() {}

// A Claim is a Realm Tenant + Scope, identifying the tenant the user belongs
// to, and what they can do.
type Claim struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RealmId    string   `protobuf:"bytes,1,opt,name=realm_id,json=realmId,proto3" json:"realm_id,omitempty"`
	TenantType string   `protobuf:"bytes,2,opt,name=tenant_type,json=tenantType,proto3" json:"tenant_type,omitempty"`
	TenantId   string   `protobuf:"bytes,3,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	Scopes     []string `protobuf:"bytes,4,rep,name=scopes,proto3" json:"scopes,omitempty"`
}

func (x *Claim) Reset() {
	*x = Claim{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Claim) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Claim) ProtoMessage() {}

func (x *Claim) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Claim.ProtoReflect.Descriptor instead.
func (*Claim) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{4}
}

func (x *Claim) GetRealmId() string {
	if x != nil {
		return x.RealmId
	}
	return ""
}

func (x *Claim) GetTenantType() string {
	if x != nil {
		return x.TenantType
	}
	return ""
}

func (x *Claim) GetTenantId() string {
	if x != nil {
		return x.TenantId
	}
	return ""
}

func (x *Claim) GetScopes() []string {
	if x != nil {
		return x.Scopes
	}
	return nil
}

type AuthenticationMethod_JWT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JwtId    string                 `protobuf:"bytes,1,opt,name=jwt_id,json=jwtId,proto3" json:"jwt_id,omitempty"`
	Issuer   string                 `protobuf:"bytes,2,opt,name=issuer,proto3" json:"issuer,omitempty"`
	IssuedAt *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=issued_at,json=issuedAt,proto3" json:"issued_at,omitempty"`
}

func (x *AuthenticationMethod_JWT) Reset() {
	*x = AuthenticationMethod_JWT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationMethod_JWT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationMethod_JWT) ProtoMessage() {}

func (x *AuthenticationMethod_JWT) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationMethod_JWT.ProtoReflect.Descriptor instead.
func (*AuthenticationMethod_JWT) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{3, 0}
}

func (x *AuthenticationMethod_JWT) GetJwtId() string {
	if x != nil {
		return x.JwtId
	}
	return ""
}

func (x *AuthenticationMethod_JWT) GetIssuer() string {
	if x != nil {
		return x.Issuer
	}
	return ""
}

func (x *AuthenticationMethod_JWT) GetIssuedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.IssuedAt
	}
	return nil
}

type AuthenticationMethod_Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The identity of the system which stored and evaluated the session.
	SessionManager string `protobuf:"bytes,1,opt,name=session_manager,json=sessionManager,proto3" json:"session_manager,omitempty"`
	// The session ID as defined by the session manager
	SessionId string `protobuf:"bytes,2,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	// The time at which the session was verified by the session manager.
	VerifiedAt *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=verified_at,json=verifiedAt,proto3" json:"verified_at,omitempty"`
	// The time at which the session began at the session manager. (e.g. the
	// time a refresh token was used to create a new session)
	AuthenticatedAt *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=authenticated_at,json=authenticatedAt,proto3" json:"authenticated_at,omitempty"`
}

func (x *AuthenticationMethod_Session) Reset() {
	*x = AuthenticationMethod_Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationMethod_Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationMethod_Session) ProtoMessage() {}

func (x *AuthenticationMethod_Session) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationMethod_Session.ProtoReflect.Descriptor instead.
func (*AuthenticationMethod_Session) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{3, 1}
}

func (x *AuthenticationMethod_Session) GetSessionManager() string {
	if x != nil {
		return x.SessionManager
	}
	return ""
}

func (x *AuthenticationMethod_Session) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *AuthenticationMethod_Session) GetVerifiedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.VerifiedAt
	}
	return nil
}

func (x *AuthenticationMethod_Session) GetAuthenticatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.AuthenticatedAt
	}
	return nil
}

type AuthenticationMethod_External struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SystemName string            `protobuf:"bytes,1,opt,name=system_name,json=systemName,proto3" json:"system_name,omitempty"`
	Metadata   map[string]string `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *AuthenticationMethod_External) Reset() {
	*x = AuthenticationMethod_External{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_auth_v1_actor_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationMethod_External) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationMethod_External) ProtoMessage() {}

func (x *AuthenticationMethod_External) ProtoReflect() protoreflect.Message {
	mi := &file_j5_auth_v1_actor_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationMethod_External.ProtoReflect.Descriptor instead.
func (*AuthenticationMethod_External) Descriptor() ([]byte, []int) {
	return file_j5_auth_v1_actor_proto_rawDescGZIP(), []int{3, 2}
}

func (x *AuthenticationMethod_External) GetSystemName() string {
	if x != nil {
		return x.SystemName
	}
	return ""
}

func (x *AuthenticationMethod_External) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

var File_j5_auth_v1_actor_proto protoreflect.FileDescriptor

var file_j5_auth_v1_actor_proto_rawDesc = []byte{
	0x0a, 0x16, 0x6a, 0x35, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x63, 0x74,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74,
	0x68, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xbd, 0x01, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1e, 0x0a,
	0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba,
	0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x2f, 0x0a,
	0x05, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6a,
	0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x42,
	0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x05, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x39,
	0x0a, 0x0b, 0x66, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31,
	0x2e, 0x46, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x52, 0x0b, 0x66, 0x69,
	0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x64, 0x65,
	0x6d, 0x70, 0x6f, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0e, 0x69, 0x64, 0x65, 0x6d, 0x70, 0x6f, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x4b,
	0x65, 0x79, 0x22, 0x73, 0x0a, 0x0b, 0x46, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e,
	0x74, 0x12, 0x22, 0x0a, 0x0a, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x69, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x88, 0x01, 0x01, 0x12, 0x22, 0x0a, 0x0a, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x61, 0x67,
	0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x09, 0x75, 0x73, 0x65,
	0x72, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x88, 0x01, 0x01, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x69, 0x70,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x75, 0x73, 0x65,
	0x72, 0x5f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x22, 0xd0, 0x02, 0x0a, 0x05, 0x41, 0x63, 0x74, 0x6f,
	0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64,
	0x12, 0x21, 0x0a, 0x0c, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x55, 0x0a, 0x15, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e,
	0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65,
	0x74, 0x68, 0x6f, 0x64, 0x52, 0x14, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x2f, 0x0a, 0x05, 0x63, 0x6c,
	0x61, 0x69, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6a, 0x35, 0x2e, 0x61,
	0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x42, 0x06, 0xba, 0x48,
	0x03, 0xc8, 0x01, 0x01, 0x52, 0x05, 0x63, 0x6c, 0x61, 0x69, 0x6d, 0x12, 0x3f, 0x0a, 0x0a, 0x61,
	0x63, 0x74, 0x6f, 0x72, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x20, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x74,
	0x6f, 0x72, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x09, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x54, 0x61, 0x67, 0x73, 0x1a, 0x3c, 0x0a, 0x0e,
	0x41, 0x63, 0x74, 0x6f, 0x72, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xee, 0x05, 0x0a, 0x14, 0x41,
	0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x12, 0x38, 0x0a, 0x03, 0x6a, 0x77, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x24, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x2e, 0x4a, 0x57, 0x54, 0x48, 0x00, 0x52, 0x03, 0x6a, 0x77, 0x74, 0x12, 0x44, 0x0a,
	0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28,
	0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68,
	0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x07, 0x73, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x12, 0x47, 0x0a, 0x08, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x2e, 0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x48, 0x00, 0x52, 0x08, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x1a, 0x6d, 0x0a, 0x03,
	0x4a, 0x57, 0x54, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x77, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x77, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x73,
	0x73, 0x75, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x69, 0x73, 0x73, 0x75,
	0x65, 0x72, 0x12, 0x37, 0x0a, 0x09, 0x69, 0x73, 0x73, 0x75, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x08, 0x69, 0x73, 0x73, 0x75, 0x65, 0x64, 0x41, 0x74, 0x1a, 0xd5, 0x01, 0x0a, 0x07,
	0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12,
	0x3b, 0x0a, 0x0b, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x0a, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x74, 0x12, 0x45, 0x0a, 0x10,
	0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x1a, 0xbd, 0x01, 0x0a, 0x08, 0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x53, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x6a, 0x35, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x76, 0x31,
	0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x2e, 0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x8d, 0x01, 0x0a, 0x05,
	0x43, 0x6c, 0x61, 0x69, 0x6d, 0x12, 0x19, 0x0a, 0x08, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x49, 0x64,
	0x12, 0x34, 0x0a, 0x0b, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x13, 0xba, 0x48, 0x10, 0x72, 0x0e, 0x32, 0x0c, 0x5e, 0x5b,
	0x61, 0x2d, 0x7a, 0x30, 0x2d, 0x39, 0x5f, 0x5d, 0x2b, 0x24, 0x52, 0x0a, 0x74, 0x65, 0x6e, 0x61,
	0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x65, 0x6e, 0x61, 0x6e,
	0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x06, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x73, 0x42, 0x30, 0x5a, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70,
	0x73, 0x2f, 0x6a, 0x35, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x6a, 0x35, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_j5_auth_v1_actor_proto_rawDescOnce sync.Once
	file_j5_auth_v1_actor_proto_rawDescData = file_j5_auth_v1_actor_proto_rawDesc
)

func file_j5_auth_v1_actor_proto_rawDescGZIP() []byte {
	file_j5_auth_v1_actor_proto_rawDescOnce.Do(func() {
		file_j5_auth_v1_actor_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_auth_v1_actor_proto_rawDescData)
	})
	return file_j5_auth_v1_actor_proto_rawDescData
}

var file_j5_auth_v1_actor_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_j5_auth_v1_actor_proto_goTypes = []any{
	(*Action)(nil),                        // 0: j5.auth.v1.Action
	(*Fingerprint)(nil),                   // 1: j5.auth.v1.Fingerprint
	(*Actor)(nil),                         // 2: j5.auth.v1.Actor
	(*AuthenticationMethod)(nil),          // 3: j5.auth.v1.AuthenticationMethod
	(*Claim)(nil),                         // 4: j5.auth.v1.Claim
	nil,                                   // 5: j5.auth.v1.Actor.ActorTagsEntry
	(*AuthenticationMethod_JWT)(nil),      // 6: j5.auth.v1.AuthenticationMethod.JWT
	(*AuthenticationMethod_Session)(nil),  // 7: j5.auth.v1.AuthenticationMethod.Session
	(*AuthenticationMethod_External)(nil), // 8: j5.auth.v1.AuthenticationMethod.External
	nil,                                   // 9: j5.auth.v1.AuthenticationMethod.External.MetadataEntry
	(*timestamppb.Timestamp)(nil),         // 10: google.protobuf.Timestamp
}
var file_j5_auth_v1_actor_proto_depIdxs = []int32{
	2,  // 0: j5.auth.v1.Action.actor:type_name -> j5.auth.v1.Actor
	1,  // 1: j5.auth.v1.Action.fingerprint:type_name -> j5.auth.v1.Fingerprint
	3,  // 2: j5.auth.v1.Actor.authentication_method:type_name -> j5.auth.v1.AuthenticationMethod
	4,  // 3: j5.auth.v1.Actor.claim:type_name -> j5.auth.v1.Claim
	5,  // 4: j5.auth.v1.Actor.actor_tags:type_name -> j5.auth.v1.Actor.ActorTagsEntry
	6,  // 5: j5.auth.v1.AuthenticationMethod.jwt:type_name -> j5.auth.v1.AuthenticationMethod.JWT
	7,  // 6: j5.auth.v1.AuthenticationMethod.session:type_name -> j5.auth.v1.AuthenticationMethod.Session
	8,  // 7: j5.auth.v1.AuthenticationMethod.external:type_name -> j5.auth.v1.AuthenticationMethod.External
	10, // 8: j5.auth.v1.AuthenticationMethod.JWT.issued_at:type_name -> google.protobuf.Timestamp
	10, // 9: j5.auth.v1.AuthenticationMethod.Session.verified_at:type_name -> google.protobuf.Timestamp
	10, // 10: j5.auth.v1.AuthenticationMethod.Session.authenticated_at:type_name -> google.protobuf.Timestamp
	9,  // 11: j5.auth.v1.AuthenticationMethod.External.metadata:type_name -> j5.auth.v1.AuthenticationMethod.External.MetadataEntry
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_j5_auth_v1_actor_proto_init() }
func file_j5_auth_v1_actor_proto_init() {
	if File_j5_auth_v1_actor_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_auth_v1_actor_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Action); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Fingerprint); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*Actor); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*AuthenticationMethod); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*Claim); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*AuthenticationMethod_JWT); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*AuthenticationMethod_Session); i {
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
		file_j5_auth_v1_actor_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*AuthenticationMethod_External); i {
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
	file_j5_auth_v1_actor_proto_msgTypes[1].OneofWrappers = []any{}
	file_j5_auth_v1_actor_proto_msgTypes[3].OneofWrappers = []any{
		(*AuthenticationMethod_Jwt)(nil),
		(*AuthenticationMethod_Session_)(nil),
		(*AuthenticationMethod_External_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_auth_v1_actor_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_j5_auth_v1_actor_proto_goTypes,
		DependencyIndexes: file_j5_auth_v1_actor_proto_depIdxs,
		MessageInfos:      file_j5_auth_v1_actor_proto_msgTypes,
	}.Build()
	File_j5_auth_v1_actor_proto = out.File
	file_j5_auth_v1_actor_proto_rawDesc = nil
	file_j5_auth_v1_actor_proto_goTypes = nil
	file_j5_auth_v1_actor_proto_depIdxs = nil
}
