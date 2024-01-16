package swagger

import (
	"fmt"

	"github.com/pentops/jsonapi/gen/j5/v1/schema_j5pb"
)

// BuildSwagger converts the J5 Document to a Swagger Document
func BuildSwagger(b *schema_j5pb.API) (*Document, error) {
	doc := &Document{
		OpenAPI: "3.0.0",
		Components: Components{
			SecuritySchemes: make(map[string]interface{}),
		},
	}

	for _, pkg := range b.Packages {
		for _, method := range pkg.Methods {
			err := doc.addMethod(method)
			if err != nil {
				return nil, fmt.Errorf("package %s method %s: %w", pkg.Name, method.FullGrpcName, err)
			}
		}
	}

	schemas := make(map[string]*Schema)
	for key, src := range b.Schemas {
		schema, err := convertSchema(src)
		if err != nil {
			return nil, err
		}
		schemas[key] = schema
	}
	doc.Components.Schemas = schemas

	return doc, nil
}

func convertSchema(schema *schema_j5pb.Schema) (*Schema, error) {

	switch special := schema.Type.(type) {
	case *schema_j5pb.Schema_Ref:
		return &Schema{
			Ref: Ptr(special.Ref),
		}, nil
	}

	out := &Schema{
		SchemaItem: &SchemaItem{
			Description: schema.Description,
		},
	}

	var err error
	switch t := schema.Type.(type) {

	case *schema_j5pb.Schema_StringItem:
		out.SchemaItem.Type = convertStringItem(t.StringItem)

	case *schema_j5pb.Schema_IntegerItem:
		out.SchemaItem.Type = convertIntegerItem(t.IntegerItem)

	case *schema_j5pb.Schema_EnumItem:
		out.SchemaItem.Type = convertEnumItem(t.EnumItem)

	case *schema_j5pb.Schema_NumberItem:
		out.SchemaItem.Type = convertNumberItem(t.NumberItem)

	case *schema_j5pb.Schema_BooleanItem:
		out.SchemaItem.Type = convertBooleanItem(t.BooleanItem)

	case *schema_j5pb.Schema_ArrayItem:
		out.SchemaItem.Type, err = convertArrayItem(t.ArrayItem)
		if err != nil {
			return nil, err
		}

	case *schema_j5pb.Schema_ObjectItem:
		out.SchemaItem.Type, err = convertObjectItem(t.ObjectItem)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown type %T", t)
	}
	return out, nil
}

func convertStringItem(item *schema_j5pb.StringItem) *StringItem {
	out := &StringItem{
		Format:  Maybe(item.Format),
		Example: Maybe(item.Example),
	}

	if item.Rules != nil {
		out.Pattern = Maybe(item.Rules.Pattern)
		out.MinLength = Maybe(item.Rules.MinLength)
		out.MaxLength = Maybe(item.Rules.MaxLength)
	}

	return out
}

func convertIntegerItem(item *schema_j5pb.IntegerItem) *IntegerItem {
	out := &IntegerItem{
		Format: item.Format,
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

func convertNumberItem(item *schema_j5pb.NumberItem) *NumberItem {
	out := &NumberItem{
		Format: item.Format,
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

func convertBooleanItem(item *schema_j5pb.BooleanItem) *BooleanItem {
	out := &BooleanItem{}

	if item.Rules != nil {
		out.Const = Maybe(item.Rules.Const)
	}

	return out
}

func convertEnumItem(item *schema_j5pb.EnumItem) *EnumItem {
	out := &EnumItem{}

	for _, val := range item.Options {
		out.Enum = append(out.Enum, val.Name)
		out.Extended = append(out.Extended, EnumValueDescription{
			Name:        val.Name,
			Description: val.Description,
		})
	}

	return out
}

func convertArrayItem(item *schema_j5pb.ArrayItem) (*ArrayItem, error) {
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

func convertObjectItem(item *schema_j5pb.ObjectItem) (*ObjectItem, error) {
	out := &ObjectItem{
		Properties:    map[string]*ObjectProperty{},
		FullProtoName: item.ProtoFullName,
		ProtoName:     item.ProtoMessageName,
	}

	for _, prop := range item.Properties {
		schema, err := convertSchema(prop.Schema)
		if err != nil {
			return nil, fmt.Errorf("object property %s: %w", prop.Name, err)
		}
		out.Properties[prop.Name] = &ObjectProperty{
			Schema:           schema,
			ReadOnly:         prop.ReadOnly,
			WriteOnly:        prop.WriteOnly,
			Description:      prop.Description,
			ProtoFieldName:   prop.ProtoFieldName,
			ProtoFieldNumber: prop.ProtoFieldNumber,
			Optional:         prop.ExplicitlyOptional,
		}
		if prop.Required {
			out.Required = append(out.Required, prop.Name)
		}
	}

	return out, nil
}
