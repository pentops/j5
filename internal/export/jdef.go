package export

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type API struct {
	Packages []*Package            `json:"packages"`
	Schemas  map[string]*Schema    `json:"definitions"`
	Metadata *schema_j5pb.Metadata `json:"metadata"`
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
			schema, err := fromProtoSchema(protoSchema)
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
		Label:  protoPackage.Label,
		Name:   protoPackage.Name,
		Hidden: protoPackage.Hidden,

		Introduction: protoPackage.Introduction,
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
		event, err := fromProtoEvent(protoEvent)
		if err != nil {
			return nil, err
		}
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
	if protoMethod.RequestBody != nil {
		schema, err := fromProtoSchema(protoMethod.RequestBody)
		if err != nil {
			return nil, err
		}
		out.RequestBody = schema
	}

	pathParameterNames := map[string]struct{}{}
	pathParts := strings.Split(protoMethod.HttpPath, "/")
	for _, part := range pathParts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		fieldName := strings.TrimPrefix(part, ":")
		pathParameterNames[fieldName] = struct{}{}
	}

	pathProperties := make([]*schema_j5pb.ObjectProperty, 0)
	bodyProperties := make([]*schema_j5pb.ObjectProperty, 0)

	requestSchema := protoMethod.RequestBody.GetObject()
	if requestSchema == nil {
		return nil, fmt.Errorf("request body was not an object: %T", protoMethod.RequestBody)
	}
	for _, prop := range requestSchema.Properties {
		_, isPath := pathParameterNames[prop.Name]
		if isPath {
			pathProperties = append(pathProperties, prop)
		} else {
			bodyProperties = append(bodyProperties, prop)
		}
	}

	out.PathParameters = make([]*JdefParameter, len(pathProperties))
	for idx, property := range pathProperties {
		schema, err := fromProtoSchema(property.Schema)
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

	if protoMethod.HttpMethod == schema_j5pb.HTTPMethod_HTTP_METHOD_GET {
		out.QueryParameters = make([]*JdefParameter, len(bodyProperties))
		for idx, property := range bodyProperties {
			schema, err := fromProtoSchema(property.Schema)
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
	} else {
		newRequest := &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Object{
				Object: &schema_j5pb.Object{
					Properties:  bodyProperties,
					Name:        requestSchema.Name,
					Description: requestSchema.Description,
				},
			},
		}
		requestSchema, err := fromProtoSchema(newRequest)
		if err != nil {
			return nil, err
		}
		out.RequestBody = requestSchema
	}

	responseSchema, err := fromProtoSchema(protoMethod.ResponseBody)
	if err != nil {
		return nil, err
	}
	out.ResponseBody = responseSchema
	return out, nil
}

func fromProtoEvent(protoEvent *schema_j5pb.EventSpec) (*EventSpec, error) {
	out := &EventSpec{
		Name: protoEvent.Name,
	}
	schema, err := fromProtoSchema(protoEvent.Schema)
	if err != nil {
		return nil, err
	}
	out.Schema = schema

	return out, nil
}

func fromProtoSchema(protoSchema *schema_j5pb.Schema) (*Schema, error) {
	return ConvertSchema(protoSchema)
}
