package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func PackageSetFromClientAPI(api *client_j5pb.API) (*SchemaSet, error) {
	pkgSet := newSchemaSet()

	for _, apiPackage := range api.Packages {
		pkg := pkgSet.Package(apiPackage.Name)

		for name, schema := range apiPackage.Schemas {
			refSchema := &RefSchema{
				Package: pkg,
				Schema:  name,
			}
			pkg.Schemas[name] = refSchema

			to, err := pkg.buildRoot(schema)
			if err != nil {
				return nil, patherr.Wrap(err, "schema", name)
			}

			refSchema.To = to
		}
	}

	for _, pkg := range pkgSet.Packages {
		if err := pkg.assertAllRefsLink(); err != nil {
			return nil, patherr.Wrap(err, "package", pkg.Name)
		}
	}

	return pkgSet, nil
}

func PackageSetFromSourceAPI(api *source_j5pb.API) (*SchemaSet, error) {
	pkgSet := newSchemaSet()

	for _, apiPackage := range api.Packages {
		pkg := pkgSet.Package(apiPackage.Name)
		if err := pkg.buildSchemas(apiPackage.Schemas); err != nil {
			return nil, patherr.Wrap(err, pkg.Name)
		}

		for _, subPkg := range apiPackage.SubPackages {
			pkg := pkgSet.Package(fmt.Sprintf("%s.%s", apiPackage.Name, subPkg.Name))
			if err := pkg.buildSchemas(subPkg.Schemas); err != nil {
				return nil, patherr.Wrap(err, pkg.Name)
			}

		}
	}

	for _, pkg := range pkgSet.Packages {
		if err := pkg.assertAllRefsLink(); err != nil {
			return nil, fmt.Errorf("asserting links on package from source API: %w", patherr.Wrap(err, pkg.Name))
		}
	}

	return pkgSet, nil
}

// AnonymousObjectFromSchema converts the schema object but does not add it to
// the package set. This is used for dynamic request and reply entities.
func (ps *SchemaSet) AnonymousObjectFromSchema(packageName string, schema *schema_j5pb.Object) (*ObjectSchema, error) {
	pkg := ps.Package(packageName)
	return pkg.objectSchemaFromDesc(schema)
}

func (pkg *Package) buildSchemas(src map[string]*schema_j5pb.RootSchema) error {
	for name, schema := range src {
		refSchema, _ := pkg.PackageSet.refTo(pkg.Name, name)
		pkg.Schemas[name] = refSchema

		if refSchema.To != nil {
			return fmt.Errorf("schema %q already exists in package %q", name, pkg.Name)
		}

		to, err := pkg.buildRoot(schema)
		if err != nil {
			return patherr.Wrap(err, "schema", name)
		}

		refSchema.To = to
	}

	return nil
}

func (pkg *Package) buildRoot(schema *schema_j5pb.RootSchema) (RootSchema, error) {
	switch st := schema.Type.(type) {
	case *schema_j5pb.RootSchema_Object:
		item, err := pkg.objectSchemaFromDesc(st.Object)
		if err != nil {
			return nil, err
		}
		return item, nil

	case *schema_j5pb.RootSchema_Oneof:
		item, err := pkg.oneofSchemaFromDesc(st.Oneof)
		if err != nil {
			return nil, err
		}
		return item, nil

	case *schema_j5pb.RootSchema_Enum:
		itemSchema := pkg.enumSchemaFromDesc(st.Enum)

		return itemSchema, nil
	}

	return nil, fmt.Errorf("expected root schema, got %T", schema.Type)
}

func (pkg *Package) schemaFromDesc(schema *schema_j5pb.Field) (FieldSchema, error) {

	switch st := schema.Type.(type) {

	case *schema_j5pb.Field_Object:
		switch inner := st.Object.Schema.(type) {
		case *schema_j5pb.ObjectField_Object:
			item, err := pkg.objectSchemaFromDesc(inner.Object)
			if err != nil {
				return nil, err
			}
			return &ObjectField{
				Ref:     item.AsRef(),
				Rules:   st.Object.Rules,
				Flatten: st.Object.Flatten,
			}, nil
		case *schema_j5pb.ObjectField_Ref:
			ref, _ := pkg.PackageSet.refTo(inner.Ref.Package, inner.Ref.Schema)
			return &ObjectField{
				Ref:   ref,
				Rules: st.Object.Rules,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported oneof schema type %T", inner)
		}

	case *schema_j5pb.Field_Oneof:
		switch inner := st.Oneof.Schema.(type) {
		case *schema_j5pb.OneofField_Oneof:
			item, err := pkg.oneofSchemaFromDesc(inner.Oneof)
			if err != nil {
				return nil, err
			}
			return &OneofField{
				Ref:   item.AsRef(),
				Rules: st.Oneof.Rules,
			}, nil
		case *schema_j5pb.OneofField_Ref:
			ref, _ := pkg.PackageSet.refTo(inner.Ref.Package, inner.Ref.Schema)
			return &OneofField{
				Ref:   ref,
				Rules: st.Oneof.Rules,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported oneof schema type %T", inner)
		}

	case *schema_j5pb.Field_Enum:
		switch inner := st.Enum.Schema.(type) {
		case *schema_j5pb.EnumField_Enum:
			item := pkg.enumSchemaFromDesc(inner.Enum)
			return &EnumField{
				Ref:   item.AsRef(),
				Rules: st.Enum.Rules,
			}, nil
		case *schema_j5pb.EnumField_Ref:
			ref, _ := pkg.PackageSet.refTo(inner.Ref.Package, inner.Ref.Schema)
			return &EnumField{
				Ref:       ref,
				Rules:     st.Enum.Rules,
				ListRules: st.Enum.ListRules,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported enum schema type %T", inner)
		}

	case *schema_j5pb.Field_Array:
		itemSchema, err := pkg.schemaFromDesc(st.Array.Items)
		if err != nil {
			return nil, patherr.Wrap(err, "items")
		}
		return &ArrayField{
			Rules:  st.Array.Rules,
			Schema: itemSchema,
		}, nil

	case *schema_j5pb.Field_Map:
		valueSchema, err := pkg.schemaFromDesc(st.Map.ItemSchema)
		if err != nil {
			return nil, patherr.Wrap(err, "items")
		}
		return &MapField{
			Rules:  st.Map.Rules,
			Schema: valueSchema,
		}, nil

	case *schema_j5pb.Field_Timestamp:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.MessageKind,
		}, nil

	case *schema_j5pb.Field_Boolean:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.BoolKind,
		}, nil

	case *schema_j5pb.Field_String_:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.StringKind,
		}, nil

	case *schema_j5pb.Field_Key:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.StringKind,
		}, nil

	case *schema_j5pb.Field_Integer:
		intKind, ok := intKinds[st.Integer.Format]
		if !ok {
			return nil, fmt.Errorf("unsupported integer format %v", st.Integer.Format)
		}
		return &ScalarSchema{
			Proto: schema,
			Kind:  intKind,
		}, nil

	case *schema_j5pb.Field_Float:
		floatKind, ok := floatKinds[st.Float.Format]
		if !ok {
			return nil, fmt.Errorf("unsupported float format %v", st.Float.Format)
		}
		return &ScalarSchema{
			Proto: schema,
			Kind:  floatKind,
		}, nil

	case *schema_j5pb.Field_Bytes:
		return &ScalarSchema{
			Proto: schema,
			Kind:  protoreflect.BytesKind,
		}, nil

	case *schema_j5pb.Field_Any:
		return &AnyField{}, nil

	default:
		return nil, fmt.Errorf("unsupported descriptor schema type %T", st)
	}
}

var floatKinds = map[schema_j5pb.FloatField_Format]protoreflect.Kind{
	schema_j5pb.FloatField_FORMAT_FLOAT32: protoreflect.FloatKind,
	schema_j5pb.FloatField_FORMAT_FLOAT64: protoreflect.DoubleKind,
}

var intKinds = map[schema_j5pb.IntegerField_Format]protoreflect.Kind{
	schema_j5pb.IntegerField_FORMAT_INT32:  protoreflect.Int32Kind,
	schema_j5pb.IntegerField_FORMAT_INT64:  protoreflect.Int64Kind,
	schema_j5pb.IntegerField_FORMAT_UINT32: protoreflect.Uint32Kind,
	schema_j5pb.IntegerField_FORMAT_UINT64: protoreflect.Uint64Kind,
}

func (pkg *Package) objectSchemaFromDesc(sch *schema_j5pb.Object) (*ObjectSchema, error) {
	object := &ObjectSchema{
		Properties: make([]*ObjectProperty, len(sch.Properties)),
		Entity:     sch.Entity,
		rootSchema: rootSchema{
			description: sch.Description,
			name:        sch.Name,
			pkg:         pkg,
		},
	}

	for i, prop := range sch.Properties {
		var err error
		object.Properties[i], err = pkg.objectPropertyFromDesc(object, prop)
		if err != nil {
			return nil, err
		}
	}

	return object, nil
}

func (pkg *Package) oneofSchemaFromDesc(sch *schema_j5pb.Oneof) (*OneofSchema, error) {

	oneof := &OneofSchema{
		Properties: make([]*ObjectProperty, len(sch.Properties)),
		rootSchema: rootSchema{
			description: sch.Description,
			name:        sch.Name,
			pkg:         pkg,
		},
	}

	for i, prop := range sch.Properties {
		var err error
		oneof.Properties[i], err = pkg.objectPropertyFromDesc(oneof, prop)
		if err != nil {
			return nil, err
		}
	}
	return oneof, nil
}

func (pkg *Package) enumSchemaFromDesc(sch *schema_j5pb.Enum) *EnumSchema {
	return &EnumSchema{
		NamePrefix: sch.Prefix,
		rootSchema: rootSchema{
			description: sch.Description,
			name:        sch.Name,
			pkg:         pkg,
		},
		Options: sch.Options,
	}
}

func (pkg *Package) objectPropertyFromDesc(parent RootSchema, prop *schema_j5pb.ObjectProperty) (*ObjectProperty, error) {
	protoField := make([]protoreflect.FieldNumber, len(prop.ProtoField))
	for i, field := range prop.ProtoField {
		protoField[i] = protoreflect.FieldNumber(field)
	}
	propSchema, err := pkg.schemaFromDesc(prop.Schema)
	if err != nil {
		return nil, patherr.Wrap(err, "properties", prop.Name)
	}

	return &ObjectProperty{
		Parent:             parent,
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
