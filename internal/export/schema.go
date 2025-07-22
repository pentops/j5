package export

import (
	"encoding/json"
	"fmt"
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
		return json.Marshal(map[string]any{
			"oneOf": s.OneOf,
		})
	}
	if s.AnyOf != nil {
		return json.Marshal(map[string]any{
			"oneOf": s.AnyOf,
		})
	}
	return json.Marshal(s.SchemaItem)
}

type SchemaItem struct {
	Type SchemaType
}

func (si SchemaItem) MarshalJSON() ([]byte, error) {
	if si.Type == nil {
		return nil, fmt.Errorf("no type set")
	}

	base := map[string]any{}
	base["type"] = si.Type.TypeName()

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
	Format    Optional[string] `json:"format"`
	Example   Optional[string] `json:"example"`
	Pattern   Optional[string] `json:"pattern"`
	MinLength Optional[uint64] `json:"minLength"`
	MaxLength Optional[uint64] `json:"maxLength"`
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

type FloatItem struct {
	Format           string            `json:"format,omitempty"`
	ExclusiveMaximum Optional[bool]    `json:"exclusiveMaximum"`
	ExclusiveMinimum Optional[bool]    `json:"exclusiveMinimum"`
	Minimum          Optional[float64] `json:"minimum"`
	Maximum          Optional[float64] `json:"maximum"`
	MultipleOf       Optional[float64] `json:"multipleOf"`
}

func (ri FloatItem) TypeName() string {
	return "number"
}

type IntegerItem struct {
	Format           string          `json:"format,omitempty"`
	ExclusiveMaximum Optional[bool]  `json:"exclusiveMaximum"`
	ExclusiveMinimum Optional[bool]  `json:"exclusiveMinimum"`
	Minimum          Optional[int64] `json:"minimum"`
	Maximum          Optional[int64] `json:"maximum"`
	MultipleOf       Optional[int64] `json:"multipleOf"`
}

func (ri IntegerItem) TypeName() string {
	return "integer"
}

type BoolItem struct {
	Const Optional[bool] `json:"const"`
}

func (ri BoolItem) TypeName() string {
	return "boolean"
}

type ArrayItem struct {
	Items       *Schema          `json:"items,omitempty"`
	MinItems    Optional[uint64] `json:"minItems"`
	MaxItems    Optional[uint64] `json:"maxItems"`
	UniqueItems Optional[bool]   `json:"uniqueItems"`
}

func (ri ArrayItem) TypeName() string {
	return "array"
}

type MapSchemaItem struct {
	ValueProperty *Schema `json:"additionalProperties,omitempty"`
	KeyProperty   *Schema `json:"x-key-property,omitempty"` // Only used for maps
}

func (mi MapSchemaItem) TypeName() string {
	return "object"
}

type AnySchemaItem struct {
	AdditionalProperties bool `json:"additionalProperties"`
}

func (ri AnySchemaItem) TypeName() string {
	return "object"
}

type ObjectItem struct {
	Name          string                     `json:"x-name"`
	Description   string                     `json:"description,omitempty"`
	IsOneof       bool                       `json:"x-is-oneof,omitempty"`
	Properties    map[string]*ObjectProperty `json:"properties,omitempty"`
	Required      []string                   `json:"required,omitempty"`
	MinProperties Optional[uint64]           `json:"minProperties"`
	MaxProperties Optional[uint64]           `json:"maxProperties"`
}

func (ri *ObjectItem) TypeName() string {
	return "object"
}

type ObjectProperty struct {
	*Schema
	ReadOnly    bool   `json:"readOnly,omitempty"`
	WriteOnly   bool   `json:"writeOnly,omitempty"`
	Description string `json:"description,omitempty"`
	Optional    bool   `json:"x-proto-optional"` // The proto field is marked as optional, go code etc should use a pointer
}

func (op *ObjectProperty) MarshalJSON() ([]byte, error) {

	base := map[string]any{}

	if op.Ref == nil {
		// Sibling values are ignored and cause a warning with $ref is used
		base["readOnly"] = op.ReadOnly
		base["writeOnly"] = op.WriteOnly
	}

	if op.Description != "" {
		base["description"] = op.Description
	}
	if op.Optional {
		base["x-proto-optional"] = op.Optional
	}

	if op.Schema == nil {
		return nil, fmt.Errorf("no schema on object property")
	}

	if op.Ref != nil {
		base["$ref"] = *op.Ref
	} else if op.OneOf != nil {
		base["oneOf"] = op.OneOf
	} else if op.AnyOf != nil {
		base["anyOf"] = op.AnyOf
	} else if op.SchemaItem != nil {
		si := op.SchemaItem
		if si.Type == nil {
			return nil, fmt.Errorf("no type set")
		}
		base["type"] = si.Type.TypeName()
		if err := jsonStructFields(si.Type, base); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no valid child in object property")
	}
	return json.Marshal(base)

}
