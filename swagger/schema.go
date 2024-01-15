package swagger

import (
	"encoding/json"
	"fmt"

	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
)

// Schema is a JSON Schema wrapper for any of the high level types.
// Only one will be set
type Schema struct {
	Ref         *string   `json:"$ref,omitempty"`
	OneOf       []*Schema `json:"oneOf,omitempty"`
	AnyOf       []*Schema `json:"anyOf,omitempty"`
	*SchemaItem           // anonymous
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.Ref != nil {
		return json.Marshal(map[string]string{
			"$ref": *s.Ref,
		})
	}
	if s.OneOf != nil {
		return json.Marshal(s.OneOf)
	}
	if s.AnyOf != nil {
		return json.Marshal(s.AnyOf)
	}
	return json.Marshal(s.SchemaItem)
}

func convertSchema(schema *jsonapi_pb.Schema) (*Schema, error) {
	out := &Schema{}
	return out, fmt.Errorf("not implemented")
}

type SchemaItem struct {
	Type        SchemaType
	Description string
}

func (si SchemaItem) MarshalJSON() ([]byte, error) {
	base := map[string]interface{}{}
	base["type"] = si.Type.TypeName()
	if si.Description != "" {
		base["description"] = si.Description
	}

	if err := jsonStructFields(si.Type, base); err != nil {
		return nil, err
	}

	return json.Marshal(base)
}

type SchemaType interface {
	TypeName() string
}

type EmptySchemaItem struct{}

func (ri EmptySchemaItem) TypeName() string {
	return "object"
}

type StringItem struct {
	Format    string           `json:"format,omitempty"`
	Example   string           `json:"example,omitempty"`
	Pattern   string           `json:"pattern,omitempty"`
	MinLength Optional[uint64] `json:"minLength,omitempty"`
	MaxLength Optional[uint64] `json:"maxLength,omitempty"`
}

func (ri StringItem) TypeName() string {
	return "string"
}

// EnumItem represents a PROTO enum in Swagger, so can only be a string
type EnumItem struct {
	Extended []EnumValueDescription `json:"x-enum"`
	Enum     []string               `json:"enum,omitempty"`
}

type EnumValueDescription struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Description string `json:"description"`
}

func (ri EnumItem) TypeName() string {
	return "string"
}

type NumberItem struct {
	Format           string            `json:"format,omitempty"`
	ExclusiveMaximum Optional[bool]    `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum Optional[bool]    `json:"exclusiveMinimum,omitempty"`
	Minimum          Optional[float64] `json:"minimum,omitempty"`
	Maximum          Optional[float64] `json:"maximum,omitempty"`
	MultipleOf       Optional[float64] `json:"multipleOf,omitempty"`
}

func (ri NumberItem) TypeName() string {
	return "number"
}

type IntegerItem struct {
	Format           string          `json:"format,omitempty"`
	ExclusiveMaximum Optional[bool]  `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum Optional[bool]  `json:"exclusiveMinimum,omitempty"`
	Minimum          Optional[int64] `json:"minimum,omitempty"`
	Maximum          Optional[int64] `json:"maximum,omitempty"`
	MultipleOf       Optional[int64] `json:"multipleOf,omitempty"`
}

func (ri IntegerItem) TypeName() string {
	return "integer"
}

type BooleanItem struct {
	Const Optional[bool] `json:"const,omitempty"`
}

func (ri BooleanItem) TypeName() string {
	return "boolean"
}

type ArrayItem struct {
	Items       SchemaItem       `json:"items,omitempty"`
	MinItems    Optional[uint64] `json:"minItems,omitempty"`
	MaxItems    Optional[uint64] `json:"maxItems,omitempty"`
	UniqueItems Optional[bool]   `json:"uniqueItems,omitempty"`
}

func (ri ArrayItem) TypeName() string {
	return "array"
}

type MapSchemaItem struct {
	ValueProperty SchemaItem `json:"additionalProperties,omitempty"`
	KeyProperty   SchemaItem `json:"x-key-property,omitempty"` // Only used for maps
}

func (mi MapSchemaItem) TypeName() string {
	return "object"
}

type ObjectItem struct {
	Properties map[string]*ObjectProperty `json:"properties,omitempty"`
	Required   []string                   `json:"required,omitempty"`

	FullProtoName string `json:"x-proto-full-name"`
	ProtoName     string `json:"x-proto-name"`
	IsOneof       bool   `json:"x-is-oneof,omitempty"`

	GoPackageName string `json:"-"`
	GoTypeName    string `json:"-"`
	GRPCPackage   string `json:"-"`

	MinProperties Optional[uint64] `json:"minProperties,omitempty"`
	MaxProperties Optional[uint64] `json:"maxProperties,omitempty"`
}

func (ri *ObjectItem) TypeName() string {
	return "object"
}

type ObjectProperty struct {
	Schema
	ReadOnly         bool   `json:"readOnly,omitempty"`
	WriteOnly        bool   `json:"writeOnly,omitempty"`
	Description      string `json:"description,omitempty"`
	ProtoFieldName   string `json:"x-proto-name,omitempty"`
	ProtoFieldNumber int    `json:"x-proto-number,omitempty"`
	Optional         bool   `json:"x-proto-optional"` // The proto field is marked as optional, go code etc should use a pointer
}
