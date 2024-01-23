package swagger

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pentops/jsonapi/gen/j5/v1/schema_j5pb"
)

type Document struct {
	OpenAPI    string     `json:"openapi"`
	Info       Info       `json:"info"`
	Paths      PathSet    `json:"paths"`
	Components Components `json:"components"`
}

func (dd *Document) GetSchema(name string) (*Schema, bool) {
	name = strings.TrimPrefix(name, "#/components/schemas/")
	schema, ok := dd.Components.Schemas[name]
	return schema, ok
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Components struct {
	Schemas         map[string]*Schema     `json:"schemas"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes"`
}

type OperationHeader struct {
	Method string `json:"-"`
	Path   string `json:"-"`

	OperationID  string      `json:"operationId,omitempty"`
	Summary      string      `json:"summary,omitempty"`
	Description  string      `json:"description,omitempty"`
	DisplayOrder int         `json:"x-display-order"`
	Parameters   []Parameter `json:"parameters,omitempty"`

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

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema"`
}

func (dd *Document) addMethod(method *schema_j5pb.Method) error {

	parameters := make([]Parameter, 0)
	for _, param := range method.PathParameters {
		schema, err := ConvertSchema(param.Schema)
		if err != nil {
			return fmt.Errorf("path param %s: %w", param.Name, err)
		}
		parameters = append(parameters, Parameter{
			Name:        param.Name,
			In:          "path",
			Description: param.Description,
			Required:    true,
			Schema:      schema,
		})
	}

	for _, param := range method.QueryParameters {
		schema, err := ConvertSchema(param.Schema)
		if err != nil {
			return fmt.Errorf("query param %s: %w", param.Name, err)
		}
		parameters = append(parameters, Parameter{
			Name:        param.Name,
			In:          "query",
			Description: param.Description,
			Required:    param.Required,
			Schema:      schema,
		})
	}

	operation := &Operation{
		OperationHeader: OperationHeader{
			Method:          method.HttpMethod,
			Path:            method.HttpPath,
			OperationID:     method.FullGrpcName,
			GrpcMethodName:  method.GrpcMethodName,
			GrpcServiceName: method.GrpcServiceName,

			Parameters: parameters,
		},
	}

	responseSchema, err := ConvertSchema(method.ResponseBody)
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

	if method.RequestBody != nil {
		requestSchema, err := ConvertSchema(method.RequestBody)
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

	found := false
	for _, pathItem := range dd.Paths {
		if pathItem.MapKey() == method.HttpPath {
			pathItem.AddOperation(operation)
			found = true
			break
		}
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
