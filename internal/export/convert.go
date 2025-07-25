package export

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/id62"
)

// BuildSwagger converts the J5 Document to a Swagger Document
func BuildSwagger(b *client_j5pb.API) (*Document, error) {
	doc := &Document{
		OpenAPI: "3.0.0",
		Components: Components{
			SecuritySchemes: make(map[string]any),
		},
	}

	for _, pkg := range b.Packages {
		for _, service := range pkg.Services {
			err := doc.addService(service)
			if err != nil {
				return nil, fmt.Errorf("package %s service %s: %w", pkg.Name, service.Name, err)
			}
		}
	}

	schemas := make(map[string]*Schema)
	for _, pkg := range b.Packages {
		for key, schema := range pkg.Schemas {
			schema, err := ConvertRootSchema(schema)
			if err != nil {
				return nil, err
			}
			fullKey := fmt.Sprintf("%s.%s", pkg.Name, key)
			schemas[fullKey] = schema
		}
	}
	doc.Components.Schemas = schemas

	return doc, nil
}

func ConvertRootSchema(schema *schema_j5pb.RootSchema) (*Schema, error) {
	switch t := schema.Type.(type) {
	case *schema_j5pb.RootSchema_Object:
		return convertObjectItem(t.Object)
	case *schema_j5pb.RootSchema_Oneof:
		return convertOneofItem(t.Oneof)
	case *schema_j5pb.RootSchema_Enum:
		return convertEnumItem(t.Enum), nil
	case *schema_j5pb.RootSchema_Polymorph:
		return convertPolymorphItem(t.Polymorph), nil
	default:
		return nil, fmt.Errorf("expected root schema, got %T", schema.Type)

	}
}

func convertSchema(schema *schema_j5pb.Field) (*Schema, error) {
	out := &Schema{
		SchemaItem: &SchemaItem{},
	}

	var err error
	switch t := schema.Type.(type) {

	case *schema_j5pb.Field_Any:
		out.Type = &AnySchemaItem{
			AdditionalProperties: true,
		}

	case *schema_j5pb.Field_String_:
		out.Type = convertStringItem(t.String_)

	case *schema_j5pb.Field_Integer:
		out.Type = convertIntegerItem(t.Integer)

	case *schema_j5pb.Field_Float:
		out.Type = convertFloatItem(t.Float)

	case *schema_j5pb.Field_Bool:
		out.Type = convertBooleanItem(t.Bool)

	case *schema_j5pb.Field_Array:
		out.Type, err = convertArrayItem(t.Array)
		if err != nil {
			return nil, err
		}

	case *schema_j5pb.Field_Key:
		out.Type, err = convertKeyItem(t.Key)
		if err != nil {
			return nil, err
		}

	case *schema_j5pb.Field_Timestamp:
		out.Type = convertTimestampItem(t.Timestamp)

	case *schema_j5pb.Field_Bytes:
		out.Type = convertBytesItem(t.Bytes)

	case *schema_j5pb.Field_Date:
		out.Type = convertDateItem(t.Date)

	case *schema_j5pb.Field_Decimal:
		out.Type = convertDecimalItem(t.Decimal)

	case *schema_j5pb.Field_Enum:
		switch t := t.Enum.Schema.(type) {
		case *schema_j5pb.EnumField_Enum:
			out.Type = convertEnumItem(t.Enum).Type

		case *schema_j5pb.EnumField_Ref:
			refStr := fmt.Sprintf("#/components/schemas/%s.%s", t.Ref.Package, t.Ref.Schema)
			out.Ref = &refStr

		default:
			return nil, fmt.Errorf("unknown schema type for swagger %T", t)
		}

	case *schema_j5pb.Field_Object:
		switch t := t.Object.Schema.(type) {
		case *schema_j5pb.ObjectField_Object:
			item, err := convertObjectItem(t.Object)
			if err != nil {
				return nil, err
			}

			out.Type = item.Type

		case *schema_j5pb.ObjectField_Ref:
			refStr := fmt.Sprintf("#/components/schemas/%s.%s", t.Ref.Package, t.Ref.Schema)
			out.Ref = &refStr

		default:
			return nil, fmt.Errorf("unknown schema type for swagger %T", t)
		}

	case *schema_j5pb.Field_Oneof:
		switch t := t.Oneof.Schema.(type) {
		case *schema_j5pb.OneofField_Oneof:
			item, err := convertOneofItem(t.Oneof)
			if err != nil {
				return nil, err
			}

			out.Type = item.Type
		case *schema_j5pb.OneofField_Ref:
			refStr := fmt.Sprintf("#/components/schemas/%s.%s", t.Ref.Package, t.Ref.Schema)
			out.Ref = &refStr

		default:
			return nil, fmt.Errorf("unknown schema type for swagger %T", t)
		}

	case *schema_j5pb.Field_Polymorph:
		switch t := t.Polymorph.Schema.(type) {
		case *schema_j5pb.PolymorphField_Polymorph:
			item := convertPolymorphItem(t.Polymorph)
			out.OneOf = item.OneOf
		case *schema_j5pb.PolymorphField_Ref:
			refStr := fmt.Sprintf("#/components/schemas/%s.%s", t.Ref.Package, t.Ref.Schema)
			out.Ref = &refStr

		default:
			return nil, fmt.Errorf("unknown schema type for swagger %T", t)
		}

	case *schema_j5pb.Field_Map:
		out.Type, err = convertMapItem(t.Map)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown schema type for swagger %T", t)
	}

	return out, nil
}

func convertStringItem(item *schema_j5pb.StringField) *StringItem {
	out := &StringItem{
		Format:  Maybe(item.Format),
		Example: Maybe(stringExample(item.Format)),
	}

	if item.Rules != nil {
		out.Pattern = Maybe(item.Rules.Pattern)
		out.MinLength = Maybe(item.Rules.MinLength)
		out.MaxLength = Maybe(item.Rules.MaxLength)
	}

	return out
}

var integerFormats map[schema_j5pb.IntegerField_Format]string = map[schema_j5pb.IntegerField_Format]string{
	schema_j5pb.IntegerField_FORMAT_INT32:  "int32",
	schema_j5pb.IntegerField_FORMAT_INT64:  "int64",
	schema_j5pb.IntegerField_FORMAT_UINT32: "uint32",
	schema_j5pb.IntegerField_FORMAT_UINT64: "uint64",
}

func convertIntegerItem(item *schema_j5pb.IntegerField) *IntegerItem {
	out := &IntegerItem{
		Format: integerFormats[item.Format], // may result in empty string if not set, should be pre-validated
	}

	if item.Rules != nil {
		out.Minimum = Maybe(item.Rules.Minimum)
		out.Maximum = Maybe(item.Rules.Maximum)
		out.ExclusiveMinimum = Maybe(item.Rules.ExclusiveMinimum)
		out.ExclusiveMaximum = Maybe(item.Rules.ExclusiveMaximum)
		out.MultipleOf = Maybe(item.Rules.MultipleOf)
	}

	return out
}

var floatFormats map[schema_j5pb.FloatField_Format]string = map[schema_j5pb.FloatField_Format]string{
	schema_j5pb.FloatField_FORMAT_FLOAT32: "float",
	schema_j5pb.FloatField_FORMAT_FLOAT64: "double",
}

func convertFloatItem(item *schema_j5pb.FloatField) *FloatItem {
	out := &FloatItem{
		Format: floatFormats[item.Format], // may result in empty string if not set, should be pre-validated
	}

	if item.Rules != nil {
		out.Minimum = Maybe(item.Rules.Minimum)
		out.Maximum = Maybe(item.Rules.Maximum)
		out.ExclusiveMinimum = Maybe(item.Rules.ExclusiveMinimum)
		out.ExclusiveMaximum = Maybe(item.Rules.ExclusiveMaximum)
		out.MultipleOf = Maybe(item.Rules.MultipleOf)
	}

	return out
}

func convertBooleanItem(item *schema_j5pb.BoolField) *BoolItem {
	out := &BoolItem{}

	if item.Rules != nil {
		out.Const = Maybe(item.Rules.Const)
	}

	return out
}

func convertEnumItem(item *schema_j5pb.Enum) *Schema {
	out := &EnumItem{}

	for _, val := range item.Options {
		out.Enum = append(out.Enum, val.Name)
		out.Extended = append(out.Extended, EnumValueDescription{
			Name:        val.Name,
			Description: val.Description,
		})
	}

	return &Schema{
		SchemaItem: &SchemaItem{
			Type: out,
		},
	}
}

func convertArrayItem(item *schema_j5pb.ArrayField) (*ArrayItem, error) {
	items, err := convertSchema(item.Items)
	if err != nil {
		return nil, err
	}

	out := &ArrayItem{
		Items: items,
	}

	if item.Rules != nil {
		out.MinItems = Maybe(item.Rules.MinItems)
		out.MaxItems = Maybe(item.Rules.MaxItems)
		out.UniqueItems = Maybe(item.Rules.UniqueItems)
	}

	return out, nil
}

func convertKeyItem(item *schema_j5pb.KeyField) (*StringItem, error) {
	out := &StringItem{}

	switch item.Format.Type.(type) {
	case *schema_j5pb.KeyFormat_Uuid:
		out.Format = Some("uuid")
		out.Pattern = Some(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	case *schema_j5pb.KeyFormat_Id62:
		out.Format = Some("id62")
		out.Pattern = Some(id62.PatternString)
	case *schema_j5pb.KeyFormat_Informal_:
		// Informal keys don't have a specific format or pattern
	case *schema_j5pb.KeyFormat_Custom_:
		out.Format = Some("custom")
		out.Pattern = Some(item.Format.GetCustom().Pattern)
	default:
		return nil, fmt.Errorf("unknown key format %T", item.Format.Type)
	}

	out.Example = Maybe(stringExample(&out.Format.Value))

	return out, nil
}

func convertTimestampItem(_ *schema_j5pb.TimestampField) *StringItem {
	return &StringItem{
		Format:  Some("date-time"),
		Example: Maybe(stringExample(Ptr("date-time"))),
	}
}

func convertBytesItem(item *schema_j5pb.BytesField) *StringItem {
	return &StringItem{
		Format:    Some("bytes"),
		MinLength: Maybe(item.Rules.MinLength),
		MaxLength: Maybe(item.Rules.MaxLength),
	}
}

func convertDateItem(_ *schema_j5pb.DateField) *StringItem {
	return &StringItem{
		Format:  Some("date"),
		Pattern: Some(`^\d{4}-\d{2}-\d{2}$`),
		Example: Maybe(stringExample(Ptr("date"))),
	}
}

func convertDecimalItem(_ *schema_j5pb.DecimalField) *StringItem {
	return &StringItem{
		Format:  Some("decimal"),
		Pattern: Some(`^\d*\.?\d+$`),
		Example: Maybe(stringExample(Ptr("number"))),
	}
}

func convertObjectItem(item *schema_j5pb.Object) (*Schema, error) {
	out := &ObjectItem{
		Properties:  map[string]*ObjectProperty{},
		Name:        item.Name,
		Description: item.Description,
	}

	for _, prop := range item.Properties {
		schema, err := convertSchema(prop.Schema)
		if err != nil {
			return nil, fmt.Errorf("object property '%s': %w", prop.Name, err)
		}
		out.Properties[prop.Name] = &ObjectProperty{
			Schema:      schema,
			Description: prop.Description,
			Optional:    prop.ExplicitlyOptional,
		}
		if prop.Required {
			out.Required = append(out.Required, prop.Name)
		}
	}

	return &Schema{
		SchemaItem: &SchemaItem{
			Type: out,
		},
	}, nil
}

func convertOneofItem(item *schema_j5pb.Oneof) (*Schema, error) {
	out := &ObjectItem{
		Properties: map[string]*ObjectProperty{},
		Name:       item.Name,
		IsOneof:    true,
	}

	for _, prop := range item.Properties {
		schema, err := convertSchema(prop.Schema)
		if err != nil {
			return nil, fmt.Errorf("oneof property '%s': %w", prop.Name, err)
		}
		out.Properties[prop.Name] = &ObjectProperty{
			Schema:      schema,
			Description: prop.Description,
			Optional:    prop.ExplicitlyOptional,
		}
	}

	return &Schema{
		SchemaItem: &SchemaItem{
			Type: out,
		},
	}, nil
}

func convertPolymorphItem(item *schema_j5pb.Polymorph) *Schema {
	out := &Schema{}

	for _, member := range item.Members {
		out.OneOf = append(out.OneOf, &Schema{
			Ref: Ptr(fmt.Sprintf("#/components/schemas/%s", member)),
		})
	}

	return out
}

func convertMapItem(item *schema_j5pb.MapField) (*MapSchemaItem, error) {
	schema, err := convertSchema(item.ItemSchema)
	if err != nil {
		return nil, err
	}

	out := &MapSchemaItem{
		ValueProperty: schema,
		KeyProperty: &Schema{
			SchemaItem: &SchemaItem{
				Type: &StringItem{},
			},
		},
	}
	return out, nil

}
