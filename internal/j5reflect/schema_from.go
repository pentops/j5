package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func RootSchemaFromDesc(pkg *Package, schema *schema_j5pb.RootSchema) (RootSchema, error) {
	switch st := schema.Type.(type) {
	case *schema_j5pb.RootSchema_Object:
		item, err := objectSchemaFromDesc(pkg, st.Object)
		if err != nil {
			return nil, err
		}
		return item, nil

	case *schema_j5pb.RootSchema_Oneof:
		item, err := oneofSchemaFromDesc(pkg, st.Oneof)
		if err != nil {
			return nil, err
		}
		return item, nil

	case *schema_j5pb.RootSchema_Enum:
		itemSchema := enumSchemaFromDesc(pkg, st.Enum)

		return itemSchema, nil
	}

	return nil, fmt.Errorf("expected root schema, got %T", schema.Type)
}

func schemaFromDesc(pkg *Package, schema *schema_j5pb.Schema) (FieldSchema, error) {
	if pkg == nil {
		return nil, fmt.Errorf("package is nil")
	}

	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	switch st := schema.Type.(type) {

	case *schema_j5pb.Schema_Object:
		switch inner := st.Object.Schema.(type) {
		case *schema_j5pb.ObjectAsField_Object:
			item, err := objectSchemaFromDesc(pkg, inner.Object)
			if err != nil {
				return nil, err
			}
			return &ObjectAsFieldSchema{
				Ref:   item.AsRef(),
				Rules: st.Object.Rules,
			}, nil
		case *schema_j5pb.ObjectAsField_Ref:
			return &ObjectAsFieldSchema{
				Ref: &RefSchema{
					Package: inner.Ref.Package,
					Schema:  inner.Ref.Schema,
				},
				Rules: st.Object.Rules,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported oneof schema type %T", inner)
		}

	case *schema_j5pb.Schema_Oneof:
		switch inner := st.Oneof.Schema.(type) {
		case *schema_j5pb.OneofAsField_Oneof:
			item, err := oneofSchemaFromDesc(pkg, inner.Oneof)
			if err != nil {
				return nil, err
			}
			return &OneofAsFieldSchema{
				Ref:   item.AsRef(),
				Rules: st.Oneof.Rules,
			}, nil
		case *schema_j5pb.OneofAsField_Ref:
			return &OneofAsFieldSchema{
				Ref: &RefSchema{
					Package: inner.Ref.Package,
					Schema:  inner.Ref.Schema,
				},
				Rules: st.Oneof.Rules,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported oneof schema type %T", inner)
		}

	case *schema_j5pb.Schema_Enum:
		switch inner := st.Enum.Schema.(type) {
		case *schema_j5pb.EnumAsField_Enum:
			item := enumSchemaFromDesc(pkg, inner.Enum)
			return &EnumAsFieldSchema{
				Ref:   item.AsRef(),
				Rules: st.Enum.Rules,
			}, nil
		case *schema_j5pb.EnumAsField_Ref:
			return &EnumAsFieldSchema{
				Ref: &RefSchema{
					Package: inner.Ref.Package,
					Schema:  inner.Ref.Schema,
				},
				Rules: st.Enum.Rules,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported enum schema type %T", inner)
		}

	case *schema_j5pb.Schema_Array:
		itemSchema, err := schemaFromDesc(pkg, st.Array.Items)
		if err != nil {
			return nil, wrapError(err, "items")
		}
		return &ArraySchema{
			Rules:  st.Array.Rules,
			Schema: itemSchema,
		}, nil

	case *schema_j5pb.Schema_Map:
		valueSchema, err := schemaFromDesc(pkg, st.Map.ItemSchema)
		if err != nil {
			return nil, wrapError(err, "items")
		}
		return &MapSchema{
			Rules:  st.Map.Rules,
			Schema: valueSchema,
		}, nil

	case *schema_j5pb.Schema_Boolean:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.BoolKind,
		}, nil

	case *schema_j5pb.Schema_String_:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.StringKind,
		}, nil

	case *schema_j5pb.Schema_Integer:
		intKind, ok := intKinds[st.Integer.Format]
		if !ok {
			return nil, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}
		return &ScalarSchema{
			Proto: schema,
			Kind:  intKind,
		}, nil

	case *schema_j5pb.Schema_Float:
		floatKind, ok := floatKinds[st.Float.Format]
		if !ok {
			return nil, fmt.Errorf("unsupported float format %v", st.Float.Format)
		}
		return &ScalarSchema{
			Proto: schema,
			Kind:  floatKind,
		}, nil

	case *schema_j5pb.Schema_Any:
		return &AnySchema{}, nil

	default:
		return nil, fmt.Errorf("unsupported descriptor schema type %T", st)
	}
}

var floatKinds = map[schema_j5pb.Float_Format]protoreflect.Kind{
	schema_j5pb.Float_FORMAT_FLOAT32: protoreflect.FloatKind,
	schema_j5pb.Float_FORMAT_FLOAT64: protoreflect.DoubleKind,
}

var intKinds = map[schema_j5pb.Integer_Format]protoreflect.Kind{
	schema_j5pb.Integer_FORMAT_INT32:  protoreflect.Int32Kind,
	schema_j5pb.Integer_FORMAT_INT64:  protoreflect.Int64Kind,
	schema_j5pb.Integer_FORMAT_UINT32: protoreflect.Uint32Kind,
	schema_j5pb.Integer_FORMAT_UINT64: protoreflect.Uint64Kind,
}

func objectSchemaFromDesc(pkg *Package, sch *schema_j5pb.Object) (*ObjectSchema, error) {
	properties := make([]*ObjectProperty, len(sch.Properties))
	for i, prop := range sch.Properties {
		var err error
		properties[i], err = objectPropertyFromDesc(pkg, prop)
		if err != nil {
			return nil, err
		}
	}

	return &ObjectSchema{
		Properties: properties,
		SchemaRoot: SchemaRoot{
			Description: sch.Description,
			Name:        sch.Name,
			Package:     pkg.Name,
		},
	}, nil
}

func oneofSchemaFromDesc(pkg *Package, sch *schema_j5pb.Oneof) (*OneofSchema, error) {
	properties := make([]*ObjectProperty, len(sch.Properties))
	for i, prop := range sch.Properties {
		var err error
		properties[i], err = objectPropertyFromDesc(pkg, prop)
		if err != nil {
			return nil, err
		}
	}

	return &OneofSchema{
		Properties: properties,
		SchemaRoot: SchemaRoot{
			Description: sch.Description,
			Name:        sch.Name,
			Package:     pkg.Name,
		},
	}, nil
}

func enumSchemaFromDesc(pkg *Package, sch *schema_j5pb.Enum) *EnumSchema {
	return &EnumSchema{
		NamePrefix: sch.Prefix,
		SchemaRoot: SchemaRoot{
			Description: sch.Description,
			Name:        sch.Name,
			Package:     pkg.Name,
		},
		Options: sch.Options,
	}
}

func objectPropertyFromDesc(pkg *Package, prop *schema_j5pb.ObjectProperty) (*ObjectProperty, error) {
	protoField := make([]protoreflect.FieldNumber, len(prop.ProtoField))
	for i, field := range prop.ProtoField {
		protoField[i] = protoreflect.FieldNumber(field)
	}
	propSchema, err := schemaFromDesc(pkg, prop.Schema)
	if err != nil {
		return nil, wrapError(err, "properties", prop.Name)
	}

	return &ObjectProperty{
		Schema:             propSchema,
		ProtoField:         protoField,
		JSONName:           prop.Name,
		Required:           prop.Required,
		ReadOnly:           prop.ReadOnly,
		WriteOnly:          prop.WriteOnly,
		ExplicitlyOptional: prop.ExplicitlyOptional,
		Description:        prop.Description,
	}, nil
}

type ValidationError struct {
	Path []string
	Err  error
}

func newValidationError(err error, path ...string) *ValidationError {
	return &ValidationError{
		Path: path,
		Err:  err,
	}
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", strings.Join(v.Path, "."), v.Err)
}

func (v *ValidationError) Unwrap() error {
	return v.Err
}

func (v *ValidationError) prefix(prefix ...string) *ValidationError {
	return &ValidationError{
		Path: append(prefix, v.Path...),
		Err:  v.Err,
	}
}

func wrapError(err error, path ...string) error {
	valErr, ok := err.(*ValidationError)
	if !ok {
		return newValidationError(err, path...)
	}
	return valErr.prefix(path...)
}
