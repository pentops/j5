package jdef

import (
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/schema/swagger"
)

func FromProto(protoSchema *schema_j5pb.API) (*API, error) {
	out := &API{
		Packages: make([]*Package, len(protoSchema.Packages)),
		Schemas:  make(map[string]*swagger.Schema),
		Metadata: protoSchema.Metadata,
	}
	for idx, protoPackage := range protoSchema.Packages {
		pkg, err := fromProtoPackage(protoPackage)
		if err != nil {
			return nil, err
		}
		out.Packages[idx] = pkg
	}
	for key, protoSchema := range protoSchema.Schemas {
		schema, err := fromProtoSchema(protoSchema)
		if err != nil {
			return nil, err
		}
		out.Schemas[key] = schema
	}

	return out, nil
}

func fromProtoPackage(protoPackage *schema_j5pb.Package) (*Package, error) {
	out := &Package{
		Label:  protoPackage.Label,
		Name:   protoPackage.Name,
		Hidden: protoPackage.Hidden,

		Introduction: protoPackage.Introduction,
	}
	out.Methods = make([]*Method, len(protoPackage.Methods))
	for idx, protoMethod := range protoPackage.Methods {
		method, err := fromProtoMethod(protoMethod)
		if err != nil {
			return nil, err
		}
		out.Methods[idx] = method
	}
	out.Events = make([]*EventSpec, len(protoPackage.Events))
	for idx, protoEvent := range protoPackage.Events {
		event, err := fromProtoEvent(protoEvent)
		if err != nil {
			return nil, err
		}
		out.Events[idx] = event
	}
	return out, nil
}

func fromProtoMethod(protoMethod *schema_j5pb.Method) (*Method, error) {
	out := &Method{
		GrpcServiceName: protoMethod.GrpcServiceName,
		GrpcMethodName:  protoMethod.GrpcMethodName,
		FullGrpcName:    protoMethod.FullGrpcName,

		HTTPMethod: protoMethod.HttpMethod,
		HTTPPath:   protoMethod.HttpPath,
	}
	if protoMethod.RequestBody != nil {
		schema, err := fromProtoSchema(protoMethod.RequestBody)
		if err != nil {
			return nil, err
		}
		out.RequestBody = schema
	}
	if protoMethod.ResponseBody != nil {
		schema, err := fromProtoSchema(protoMethod.ResponseBody)
		if err != nil {
			return nil, err
		}
		out.ResponseBody = schema
	}

	out.QueryParameters = make([]*Parameter, len(protoMethod.QueryParameters))
	for idx, protoParam := range protoMethod.QueryParameters {
		param, err := fromProtoParameter(protoParam)
		if err != nil {
			return nil, err
		}
		out.QueryParameters[idx] = param
	}
	out.PathParameters = make([]*Parameter, len(protoMethod.PathParameters))
	for idx, protoParam := range protoMethod.PathParameters {
		param, err := fromProtoParameter(protoParam)
		if err != nil {
			return nil, err
		}
		out.PathParameters[idx] = param
	}
	return out, nil
}

func fromProtoEvent(protoEvent *schema_j5pb.EventSpec) (*EventSpec, error) {
	out := &EventSpec{
		Name: protoEvent.Name,
	}
	if protoEvent.StateSchema != nil {
		schema, err := fromProtoSchema(protoEvent.StateSchema)
		if err != nil {
			return nil, err
		}
		out.StateSchema = schema
	}
	if protoEvent.EventSchema != nil {
		schema, err := fromProtoSchema(protoEvent.EventSchema)
		if err != nil {
			return nil, err
		}
		out.EventSchema = schema
	}

	return out, nil
}

func fromProtoParameter(protoParam *schema_j5pb.Parameter) (*Parameter, error) {
	out := &Parameter{
		Name:        protoParam.Name,
		Description: protoParam.Description,
		Required:    protoParam.Required,
	}
	schema, err := fromProtoSchema(protoParam.Schema)
	if err != nil {
		return nil, err
	}
	out.Schema = *schema
	return out, nil
}

func fromProtoSchema(protoSchema *schema_j5pb.Schema) (*swagger.Schema, error) {
	return swagger.ConvertSchema(protoSchema)
}

type API struct {
	Packages []*Package                 `json:"packages"`
	Schemas  map[string]*swagger.Schema `json:"schemas"`
	Metadata *schema_j5pb.Metadata      `json:"metadata"`
}

type Package struct {
	Label  string `json:"label"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`

	Introduction string       `json:"introduction,omitempty"`
	Methods      []*Method    `json:"methods"`
	Events       []*EventSpec `json:"events"`
}

type Method struct {
	GrpcServiceName string `json:"grpcServiceName"`
	GrpcMethodName  string `json:"grpcMethodName"`
	FullGrpcName    string `json:"fullGrpcName"`

	HTTPMethod      string          `json:"httpMethod"`
	HTTPPath        string          `json:"httpPath"`
	RequestBody     *swagger.Schema `json:"requestBody,omitempty"`
	ResponseBody    *swagger.Schema `json:"responseBody,omitempty"`
	QueryParameters []*Parameter    `json:"queryParameters,omitempty"`
	PathParameters  []*Parameter    `json:"pathParameters,omitempty"`
}

type EventSpec struct {
	Name        string          `json:"name"`
	StateSchema *swagger.Schema `json:"stateSchema,omitempty"`
	EventSchema *swagger.Schema `json:"eventSchema,omitempty"`
}

type Parameter struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      swagger.Schema `json:"schema"`
}
