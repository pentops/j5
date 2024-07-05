package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func RootSchemaFromDesc(pkg *Package, schema *schema_j5pb.Schema) (RootSchema, error) {
	item, err := SchemaFromDesc(pkg, schema)
	if err != nil {
		return nil, err
	}
	switch it := item.(type) {
	case *ObjectSchema:
		return it, nil
	case *OneofSchema:
		return it, nil
	case *EnumSchema:
		return it, nil
	}
	return nil, fmt.Errorf("expected root schema, got %T", item)
}

func SchemaFromDesc(pkg *Package, schema *schema_j5pb.Schema) (Schema, error) {
	if pkg == nil {
		return nil, fmt.Errorf("package is nil")
	}

	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	switch st := schema.Type.(type) {

	case *schema_j5pb.Schema_Ref:
		item := &RefSchema{
			Package: st.Ref.Package,
			Schema:  st.Ref.Schema,
			// To will be resolved later
		}
		return item, nil
	case *schema_j5pb.Schema_Object:
		item, err := objectSchemaFromDesc(pkg, st.Object)
		if err != nil {
			return nil, err
		}
		return item, nil

	case *schema_j5pb.Schema_Oneof:
		item, err := oneofSchemaFromDesc(pkg, st.Oneof)
		if err != nil {
			return nil, err
		}
		return item, nil

	case *schema_j5pb.Schema_Enum:
		itemSchema := enumSchemaFromDesc(pkg, st.Enum)

		return itemSchema, nil
	case *schema_j5pb.Schema_Array:
		itemSchema, err := SchemaFromDesc(pkg, st.Array.Items)
		if err != nil {
			return nil, wrapError(err, "items")
		}
		return &ArraySchema{
			Rules:  st.Array.Rules,
			Schema: itemSchema,
		}, nil

	case *schema_j5pb.Schema_Map:
		valueSchema, err := SchemaFromDesc(pkg, st.Map.ItemSchema)
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
		Rules:      sch.Rules,
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
		Rules:      sch.Rules,
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
	propSchema, err := SchemaFromDesc(pkg, prop.Schema)
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
