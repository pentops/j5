// Package j5reflect links j5schema with concrete go values, generated in proto.
//
// ## Element types
//
// ### Property
//
// A Property is a is a j5.schema.v1.ObjectProperty
// Represented in go as a field in a go struct
//
// The parent context is a pointer to a Go struct, and a field type
// definition (protoreflect MessageValue and Field), which allows
// new values to be created where they were otherwise null/unset
//
// All Properties are also Fields
//
// ### Field
//
// A field is a value held in any context where it can be accessed and updated, but not necessarily created
//
// All Properties are also Fields, and are mutable by modifying the parent Object/Oneof.
//
// Elements of Arrays and Maps are Fields but not Properties.
//
// - Message Fields are settable using message.Set(field, value)
// - Map Fields are updatable using the Key
// - Array Fields are updatable using the Index
//
// ### Value
//
// A Value is a field detached from its parent. It can't be set or mutated, it
// cab be read, or can be sent in as the value of a set and append to leaf fields
//
// All J6 schema types are values

package j5reflect

import (
	"github.com/pentops/j5/internal/j5schema"
)

type ArrayField interface {
	Field
	Range(func(Field) error) error
	//ItemSchema() j5schema.FieldSchema
}

type MutableArrayField interface {
	ArrayField
	NewElement() Field
}

type ArrayOfObjectField interface {
	ArrayOfContainerField
	NewObjectElement() (Object, int, error)
}

type ArrayOfOneofField interface {
	ArrayOfContainerField
	NewOneofElement() (Oneof, int, error)
}

type ArrayOfScalarField interface {
	ArrayField
	AppendGoScalar(value interface{}) (int, error)
	AppendASTValue(ASTValue) (int, error)
}

type MapField interface {
	Field
	Range(func(string, Field) error) error
	ItemSchema() j5schema.FieldSchema
}

type MutableMapField interface {
	MapField
	NewValue(key string) Field
}

type MapOfObjectField interface {
	NewObjectValue(key string) (Oneof, error)
}

type MapOfScalarField interface {
	SetGoScalar(key string, value interface{}) error
}

type RangeCallback func(Field) error