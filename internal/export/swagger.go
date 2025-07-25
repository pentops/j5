package export

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type Document struct {
	OpenAPI    string       `json:"openapi"`
	Info       DocumentInfo `json:"info"`
	Paths      PathSet      `json:"paths"`
	Components Components   `json:"components"`
}

type DocumentInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Components struct {
	Schemas         map[string]*Schema `json:"schemas"`
	SecuritySchemes map[string]any     `json:"securitySchemes"`
}

type OperationHeader struct {
	Method string `json:"-"`
	Path   string `json:"-"`

	OperationID  string             `json:"operationId,omitempty"`
	Summary      string             `json:"summary,omitempty"`
	Description  string             `json:"description,omitempty"`
	DisplayOrder int                `json:"x-display-order"`
	Parameters   []SwaggerParameter `json:"parameters,omitempty"`

	GrpcServiceName string `json:"x-grpc-service"`
	GrpcMethodName  string `json:"x-grpc-method"`
}

type Operation struct {
	OperationHeader
	RequestBody  *RequestBody `json:"requestBody,omitempty"`
	ResponseBody *Response    `json:"-"`
	Responses    *ResponseSet `json:"responses,omitempty"`
}

func (oo Operation) MapKey() string {
	return oo.Method
}

type ResponseSet []Response

func (rs ResponseSet) MarshalJSON() ([]byte, error) {
	return OrderedMap[Response](rs).MarshalJSON()
}

type RequestBody struct {
	Description string           `json:"description,omitempty"`
	Required    bool             `json:"required,omitempty"`
	Content     OperationContent `json:"content"`
}
type Response struct {
	Code        int              `json:"-"`
	Description string           `json:"description"`
	Content     OperationContent `json:"content"`
}

func (rs Response) MapKey() string {
	return strconv.Itoa(rs.Code)
}

type OperationContent struct {
	JSON *OperationSchema `json:"application/json,omitempty"`
}

type OperationSchema struct {
	Schema *Schema `json:"schema"`
}

type SwaggerParameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema"`
}

func (dd *Document) addService(service *client_j5pb.Service) error {
	for _, method := range service.Methods {
		err := dd.addMethod(service, method)
		if err != nil {
			return fmt.Errorf("method %s: %w", method.Method.FullGrpcName, err)
		}
	}
	return nil
}

var methodShortString = map[schema_j5pb.HTTPMethod]string{
	schema_j5pb.HTTPMethod_HTTP_METHOD_GET:    "get",
	schema_j5pb.HTTPMethod_HTTP_METHOD_POST:   "post",
	schema_j5pb.HTTPMethod_HTTP_METHOD_PUT:    "put",
	schema_j5pb.HTTPMethod_HTTP_METHOD_DELETE: "delete",
	schema_j5pb.HTTPMethod_HTTP_METHOD_PATCH:  "patch",
}

// formatPathParameters replaces path parameters in the format ":param" with "{param}"
func formatPathParameters(path string, pathParameters []*schema_j5pb.ObjectProperty) (string, error) {
	if len(pathParameters) == 0 {
		return path, nil
	}

	pathParts := strings.Split(path, "/")
	for _, param := range pathParameters {
		found := false
		for i, part := range pathParts {
			// Check if the part matches the parameter name
			if part == ":"+param.Name {
				// Replace the path parameter with a placeholder
				pathParts[i] = "{" + param.Name + "}"
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("path parameter %s not found in path %s", param.Name, path)
		}
	}

	return strings.Join(pathParts, "/"), nil
}

func (dd *Document) addMethod(service *client_j5pb.Service, method *client_j5pb.Method) error {

	operation := &Operation{
		OperationHeader: OperationHeader{
			Method:          methodShortString[method.Method.HttpMethod],
			Path:            method.Method.HttpPath,
			OperationID:     method.Method.FullGrpcName,
			GrpcMethodName:  method.Method.Name,
			GrpcServiceName: service.Name,

			Parameters: make([]SwaggerParameter, 0, len(method.Request.PathParameters)+len(method.Request.QueryParameters)),
		},
	}

	for _, property := range method.Request.PathParameters {
		schema, err := convertSchema(property.Schema)
		if err != nil {
			return fmt.Errorf("path param %s: %w", property.Name, err)
		}
		operation.Parameters = append(operation.Parameters, SwaggerParameter{
			Name:        property.Name,
			In:          "path",
			Description: property.Description,
			Required:    true,
			Schema:      schema,
		})
	}

	for _, property := range method.Request.QueryParameters {
		schema, err := convertSchema(property.Schema)
		if err != nil {
			return fmt.Errorf("query param %s: %w", property.Name, err)
		}
		operation.Parameters = append(operation.Parameters, SwaggerParameter{
			Name:        property.Name,
			In:          "query",
			Description: property.Description,
			Required:    property.Required,
			Schema:      schema,
		})
	}

	if method.Request.Body != nil {
		requestSchema, err := convertObjectItem(method.Request.Body)
		if err != nil {
			return err
		}
		operation.RequestBody = &RequestBody{
			Required: true,
			Content: OperationContent{
				JSON: &OperationSchema{
					Schema: requestSchema,
				},
			},
		}
	}

	responseSchema, err := convertObjectItem(method.ResponseBody)
	if err != nil {
		return fmt.Errorf("response body: %w", err)
	}
	operation.Responses = &ResponseSet{{
		Code:        200,
		Description: "OK",
		Content: OperationContent{
			JSON: &OperationSchema{
				Schema: responseSchema,
			},
		},
	}}

	found := false
	for _, pathItem := range dd.Paths {
		if pathItem.MapKey() == method.Method.HttpPath {
			pathItem.AddOperation(operation)
			found = true
			break
		}
	}

	operation.Path, err = formatPathParameters(operation.Path, method.Request.PathParameters)
	if err != nil {
		return err
	}

	if !found {
		pathItem := &PathItem{operation}
		dd.Paths = append(dd.Paths, pathItem)
	}

	return nil
}

type PathSet []*PathItem

func (ps PathSet) MarshalJSON() ([]byte, error) {
	return OrderedMap[*PathItem](ps).MarshalJSON()
}

type PathItem []*Operation

func (pi PathItem) MarshalJSON() ([]byte, error) {
	return OrderedMap[*Operation](pi).MarshalJSON()
}

func (pi *PathItem) AddOperation(op *Operation) {
	*pi = append(*pi, op)
}

func (pi PathItem) MapKey() string {
	if len(pi) == 0 {
		return ""
	}
	return pi[0].Path
}

type MapItem interface {
	MapKey() string
}

type OrderedMap[T MapItem] []T

func (om OrderedMap[T]) MarshalJSON() ([]byte, error) {
	fields := make([]string, len(om))
	for idx, field := range om {
		val, err := json.Marshal(field)
		if err != nil {
			return nil, err
		}
		keyString := field.MapKey()
		key, _ := json.Marshal(keyString)
		fields[idx] = string(key) + ":" + string(val)
	}
	outStr := "{" + strings.Join(fields, ",") + "}"
	return []byte(outStr), nil
}
