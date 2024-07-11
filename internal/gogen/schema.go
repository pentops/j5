package gogen

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
)

type builder struct {
	fileSet *FileSet
	options Options
	//schemas SchemaResolver
}

func (bb *builder) buildTypeName(schema j5reflect.FieldSchema) (*DataType, error) {

	switch schemaType := schema.(type) {

	case *j5reflect.ObjectFieldSchema:
		if err := bb.addObject(schemaType.Schema()); err != nil {
			return nil, fmt.Errorf("referencedType in %s: %w", schemaType.Ref.FullName(), err)
		}

		objectPackage, err := bb.options.ReferenceGoPackage(schemaType.Ref.Package)
		if err != nil {
			return nil, fmt.Errorf("referredType in %s: %w", schemaType.Ref.FullName(), err)
		}

		return &DataType{
			Name:      goTypeName(schemaType.Ref.Schema),
			GoPackage: objectPackage,
			J5Package: schemaType.Ref.Package,
			Pointer:   true,
		}, nil

	case *j5reflect.OneofFieldSchema:
		if err := bb.addOneofWrapper(schemaType.Schema()); err != nil {
			return nil, fmt.Errorf("referencedType in %s: %w", schemaType.Ref.FullName(), err)
		}

		objectPackage, err := bb.options.ReferenceGoPackage(schemaType.Ref.Package)
		if err != nil {
			return nil, fmt.Errorf("referredType in %s: %w", schemaType.Ref.FullName(), err)
		}

		return &DataType{
			Name:      goTypeName(schemaType.Ref.Schema),
			GoPackage: objectPackage,
			Pointer:   true,
			J5Package: schemaType.Ref.Package,
		}, nil

	case *j5reflect.EnumFieldSchema:
		return &DataType{
			Name:    "string",
			Pointer: false,
		}, nil

	case *j5reflect.ArraySchema:
		itemType, err := bb.buildTypeName(schemaType.Schema)
		if err != nil {
			return nil, err
		}

		return &DataType{
			Name:      itemType.Name,
			Pointer:   itemType.Pointer,
			J5Package: itemType.J5Package,
			GoPackage: itemType.GoPackage,
			Slice:     true,
		}, nil

	case *j5reflect.MapSchema:
		valueType, err := bb.buildTypeName(schemaType.Schema)
		if err != nil {
			return nil, fmt.Errorf("map value: %w", err)
		}

		return &DataType{
			Name:    fmt.Sprintf("map[string]%s", valueType.Name),
			Pointer: false,
		}, nil

	case *j5reflect.AnySchema:
		return &DataType{
			Name:    "interface{}",
			Pointer: false,
		}, nil

	case *j5reflect.ScalarSchema:
		asProto := schemaType.Proto

		switch schemaType := asProto.Type.(type) {
		case *schema_j5pb.Schema_String_:
			item := schemaType.String_
			if item.Format == nil {
				return &DataType{
					Name:    "string",
					Pointer: false,
				}, nil
			}

			switch *item.Format {
			case "uuid", "date", "email", "uri":
				return &DataType{
					Name:    "string",
					Pointer: false,
				}, nil
			case "date-time":
				return &DataType{
					Name:      "Time",
					Pointer:   true,
					GoPackage: "time",
				}, nil
			case "byte":
				return &DataType{
					Name:    "[]byte",
					Pointer: false,
				}, nil
			default:
				return nil, fmt.Errorf("Unknown string format: %s", *item.Format)
			}

		case *schema_j5pb.Schema_Float:
			return &DataType{
				Name:    goFloatTypes[schemaType.Float.Format],
				Pointer: false,
			}, nil

		case *schema_j5pb.Schema_Integer:
			return &DataType{
				Name:    goIntTypes[schemaType.Integer.Format],
				Pointer: false,
			}, nil

		case *schema_j5pb.Schema_Boolean:
			return &DataType{
				Name:    "bool",
				Pointer: false,
			}, nil

		default:
			return nil, fmt.Errorf("Unknown scalar type: %T", schemaType)
		}

	default:
		return nil, fmt.Errorf("Unknown type for Go Gen: %T\n", schema)
	}

}

var goFloatTypes = map[schema_j5pb.FloatField_Format]string{
	schema_j5pb.FloatField_FORMAT_FLOAT32: "float32",
	schema_j5pb.FloatField_FORMAT_FLOAT64: "float64",
}

var goIntTypes = map[schema_j5pb.IntegerField_Format]string{
	schema_j5pb.IntegerField_FORMAT_INT32:  "int32",
	schema_j5pb.IntegerField_FORMAT_INT64:  "int64",
	schema_j5pb.IntegerField_FORMAT_UINT32: "uint32",
	schema_j5pb.IntegerField_FORMAT_UINT64: "uint64",
}

func (bb *builder) jsonField(property *j5reflect.ObjectProperty) (*Field, error) {

	tags := map[string]string{}

	tags["json"] = property.JSONName
	if !property.Required {
		tags["json"] += ",omitempty"
	}

	dataType, err := bb.buildTypeName(property.Schema)
	if err != nil {
		return nil, fmt.Errorf("building field %s: %w", property.JSONName, err)
	}

	if !dataType.Pointer && !dataType.Slice && property.ExplicitlyOptional {
		dataType.Pointer = true
	}

	return &Field{
		Name:     goTypeName(property.JSONName),
		DataType: *dataType,
		Tags:     tags,
		Property: property,
	}, nil

}

func (bb *builder) addObject(object *j5reflect.ObjectSchema) error {
	gen, err := bb.fileForPackage(object.Package)
	if err != nil {
		return err
	}
	if gen == nil {
		return nil
	}

	typeName := goTypeName(object.Name)
	_, ok := gen.types[object.Name]
	if ok {
		return nil
	}

	structType := &Struct{
		Name: typeName,
		Comment: fmt.Sprintf(
			"Proto: %s",
			object.FullName(),
		),
	}
	gen.types[object.Name] = structType

	for _, property := range object.Properties {
		field, err := bb.jsonField(property)
		if err != nil {
			return fmt.Errorf("object %s: %w", object.FullName(), err)
		}
		structType.Fields = append(structType.Fields, field)
	}

	return nil
}

func (bb *builder) addOneofWrapper(wrapper *j5reflect.OneofSchema) error {
	gen, err := bb.fileForPackage(wrapper.Package)
	if err != nil {
		return err
	}
	if gen == nil {
		return nil
	}

	_, ok := gen.types[wrapper.Name]
	if ok {
		return nil
	}

	comment := fmt.Sprintf(
		"Proto Message: %s", wrapper.FullName(),
	)

	structType := &Struct{
		Name:    goTypeName(wrapper.Name),
		Comment: comment,
	}
	gen.types[wrapper.Name] = structType

	keyMethod := &Function{
		Name: "OneofKey",
		Returns: []*Parameter{{
			DataType: DataType{
				Name:    "string",
				Pointer: false,
			}},
		},
		StringGen: gen.ChildGen(),
	}

	valueMethod := &Function{
		Name: "Type",
		Returns: []*Parameter{{
			DataType: DataType{
				Name:    "interface{}",
				Pointer: false,
			}},
		},
		StringGen: gen.ChildGen(),
	}

	structType.Fields = append(structType.Fields, &Field{
		Name:     "J5TypeKey",
		DataType: DataType{Name: "string", Pointer: false},
		Tags:     map[string]string{"json": "!type,omitempty"},
	})

	for _, property := range wrapper.Properties {
		field, err := bb.jsonField(property)
		if err != nil {
			return fmt.Errorf("object %s: %w", wrapper.FullName(), err)
		}
		field.DataType.Pointer = true
		structType.Fields = append(structType.Fields, field)
		keyMethod.P("if s.", field.Name, " != nil {")
		keyMethod.P("  return \"", property.JSONName, "\"")
		keyMethod.P("}")
		valueMethod.P("if s.", field.Name, " != nil {")
		valueMethod.P("  return s.", field.Name)
		valueMethod.P("}")
	}
	keyMethod.P("return \"\"")
	valueMethod.P("return nil")

	structType.Methods = append(structType.Methods, keyMethod, valueMethod)

	return nil
}
