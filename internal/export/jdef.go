package export

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type API struct {
	Packages []*Package            `json:"packages"`
	Schemas  map[string]*Schema    `json:"definitions"`
	Metadata *schema_j5pb.Metadata `json:"metadata"`
}

type Package struct {
	Label string `json:"label"`
	Name  string `json:"name"`

	Introduction string       `json:"introduction,omitempty"`
	Methods      []*Method    `json:"methods"`
	Events       []*EventSpec `json:"events"`
}

type Method struct {
	GrpcServiceName string `json:"grpcServiceName"`
	GrpcMethodName  string `json:"grpcMethodName"`
	FullGrpcName    string `json:"fullGrpcName"`

	HTTPMethod      string           `json:"httpMethod"`
	HTTPPath        string           `json:"httpPath"`
	RequestBody     *Schema          `json:"requestBody,omitempty"`
	ResponseBody    *Schema          `json:"responseBody,omitempty"`
	QueryParameters []*JdefParameter `json:"queryParameters,omitempty"`
	PathParameters  []*JdefParameter `json:"pathParameters,omitempty"`
}

type EventSpec struct {
	Name   string  `json:"name"`
	Schema *Schema `json:"schema,omitempty"`
}

type JdefParameter struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Schema      Schema `json:"schema"`
}

func FromProto(protoSchema *schema_j5pb.API) (*API, error) {
	out := &API{
		Packages: make([]*Package, len(protoSchema.Packages)),
		Schemas:  make(map[string]*Schema),
		Metadata: protoSchema.Metadata,
	}
	for idx, protoPackage := range protoSchema.Packages {
		pkg, err := fromProtoPackage(protoPackage)
		if err != nil {
			return nil, err
		}
		out.Packages[idx] = pkg

		for key, protoSchema := range protoPackage.Schemas {
			schema, err := ConvertRootSchema(protoSchema)
			if err != nil {
				return nil, err
			}
			out.Schemas[key] = schema
		}

	}
	return out, nil
}

func fromProtoPackage(protoPackage *schema_j5pb.Package) (*Package, error) {
	out := &Package{
		Label: protoPackage.Label,
		Name:  protoPackage.Name,

		Introduction: protoPackage.Prose,
	}
	out.Methods = make([]*Method, 0)
	for _, protoService := range protoPackage.Services {
		for _, protoMethod := range protoService.Methods {
			method, err := fromProtoMethod(protoService, protoMethod)
			if err != nil {
				return nil, err
			}
			out.Methods = append(out.Methods, method)
		}
	}
	out.Events = make([]*EventSpec, len(protoPackage.Events))
	for idx, protoEvent := range protoPackage.Events {
		event := fromProtoEvent(protoEvent)
		out.Events[idx] = event
	}
	return out, nil
}

var methodShortString = map[schema_j5pb.HTTPMethod]string{
	schema_j5pb.HTTPMethod_HTTP_METHOD_GET:    "get",
	schema_j5pb.HTTPMethod_HTTP_METHOD_POST:   "post",
	schema_j5pb.HTTPMethod_HTTP_METHOD_PUT:    "put",
	schema_j5pb.HTTPMethod_HTTP_METHOD_DELETE: "delete",
	schema_j5pb.HTTPMethod_HTTP_METHOD_PATCH:  "patch",
}

func fromProtoMethod(protoService *schema_j5pb.Service, protoMethod *schema_j5pb.Method) (*Method, error) {
	out := &Method{
		GrpcServiceName: protoService.Name,
		GrpcMethodName:  protoMethod.Name,
		FullGrpcName:    protoMethod.FullGrpcName,

		HTTPMethod: methodShortString[protoMethod.HttpMethod],
		HTTPPath:   protoMethod.HttpPath,
	}
	if protoMethod.Request.Body != nil {
		schema, err := convertObjectItem(protoMethod.Request.Body)
		if err != nil {
			return nil, err
		}
		out.RequestBody = schema
	}

	out.PathParameters = make([]*JdefParameter, len(protoMethod.Request.PathParameters))
	for idx, property := range protoMethod.Request.PathParameters {
		schema, err := convertSchema(property.Schema)
		if err != nil {
			return nil, fmt.Errorf("path param %s: %w", property.Name, err)
		}
		out.PathParameters[idx] = &JdefParameter{
			Name:        property.Name,
			Description: property.Description,
			Required:    true,
			Schema:      *schema,
		}
	}

	out.QueryParameters = make([]*JdefParameter, len(protoMethod.Request.QueryParameters))
	for idx, property := range protoMethod.Request.QueryParameters {
		schema, err := convertSchema(property.Schema)
		if err != nil {
			return nil, fmt.Errorf("query param %s: %w", property.Name, err)
		}
		out.QueryParameters[idx] = &JdefParameter{
			Name:        property.Name,
			Description: property.Description,
			Required:    property.Required,
			Schema:      *schema,
		}
	}

	responseSchema, err := convertObjectItem(protoMethod.ResponseBody)
	if err != nil {
		return nil, err
	}
	out.ResponseBody = responseSchema
	return out, nil
}

func fromProtoEvent(protoEvent *schema_j5pb.EventSpec) *EventSpec {
	ref := fmt.Sprintf("#/definitions/%s", protoEvent.Schema)
	out := &EventSpec{
		Name: protoEvent.Name,
		Schema: &Schema{
			Ref: &ref,
		},
	}

	return out
}
