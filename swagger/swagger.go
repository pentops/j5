package swagger

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pentops/custom-proto-api/jsonapi"
)

type Document struct {
	OpenAPI    string     `json:"openapi"`
	Info       Info       `json:"info"`
	Paths      PathSet    `json:"paths"`
	Components Components `json:"components"`
}

func (dd *Document) GetSchema(name string) (*jsonapi.SchemaItem, bool) {
	name = strings.TrimPrefix(name, "#/components/schemas/")
	schema, ok := dd.Components.Schemas[name]
	return schema, ok
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
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

type Components struct {
	Schemas         map[string]*jsonapi.SchemaItem `json:"schemas"`
	SecuritySchemes map[string]interface{}         `json:"securitySchemes"`
}

type OperationHeader struct {
	Method string `json:"-"`
	Path   string `json:"-"`

	OperationID  string      `json:"operationId,omitempty"`
	Summary      string      `json:"summary,omitempty"`
	Description  string      `json:"description,omitempty"`
	DisplayOrder int         `json:"x-display-order"`
	Parameters   []Parameter `json:"parameters,omitempty"`
}

type Operation struct {
	OperationHeader
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	Responses   *ResponseSet `json:"responses,omitempty"`
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

type OperationContent struct {
	JSON *OperationSchema `json:"application/json,omitempty"`
}

type OperationSchema struct {
	Schema jsonapi.SchemaItem `json:"schema"`
}

func (rs Response) MapKey() string {
	return strconv.Itoa(rs.Code)
}

func (oo Operation) MapKey() string {
	return oo.Method
}

type Parameter struct {
	Name        string             `json:"name"`
	In          string             `json:"in"`
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Schema      jsonapi.SchemaItem `json:"schema"`
}
