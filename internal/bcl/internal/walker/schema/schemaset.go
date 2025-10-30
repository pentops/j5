package schema

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/proto"
)

type specSource string

const (
	specSourceAuto       specSource = "reflect"
	specSourceSchema     specSource = "schema"
	specSourceAnnotation specSource = "annotation"
)

func (bs *BlockSpec) ErrName() string {
	if bs.DebugName != "" {
		return fmt.Sprintf("%s from %s as %q", bs.schema, bs.source, bs.DebugName)
	}
	return fmt.Sprintf("%s from %s", bs.schema, bs.source)
}

type SchemaSet struct {
	cachedSpecs map[string]*BlockSpec
}

func convertBlock(src *bcl_j5pb.Block) (*BlockSpec, error) {
	aliases := map[string]PathSpec{}
	for _, alias := range src.Alias {
		aliases[alias.Name] = PathSpec(alias.Path)
	}

	block := &BlockSpec{
		Name:        convertTag(src.Name),
		TypeSelect:  convertTag(src.TypeSelect),
		Qualifier:   convertTag(src.Qualifier),
		OnlyDefined: src.OnlyExplicit,
		Aliases:     aliases,
	}
	if src.DescriptionField != nil {
		block.Description = src.DescriptionField
	}

	if src.ScalarSplit != nil {
		ss := src.ScalarSplit
		block.ScalarSplit = &ScalarSplit{
			Delimiter:   ss.Delimiter,
			RightToLeft: ss.RightToLeft,
		}
		for _, required := range ss.RequiredFields {
			block.ScalarSplit.Required = append(block.ScalarSplit.Required, PathSpec(required.Path))
		}
		for _, optional := range ss.OptionalFields {
			block.ScalarSplit.Optional = append(block.ScalarSplit.Optional, PathSpec(optional.Path))
		}
		if ss.RemainderField != nil {
			ps := PathSpec(ss.RemainderField.Path)
			block.ScalarSplit.Remainder = &ps
		}
	}

	if err := block.Validate(); err != nil {
		return nil, fmt.Errorf("invalid block spec for %s: %s", src.SchemaName, err)
	}
	return block, nil
}

func NewSchemaSet() (*SchemaSet, error) {
	return &SchemaSet{
		cachedSpecs: map[string]*BlockSpec{},
	}, nil
}

func (ss *SchemaSet) lookupBlockSpec(node j5PropSet) (*BlockSpec, error) {
	var err error
	schemaName := node.SchemaName()

	blockSpec := &BlockSpec{}

	var bclBlock *bcl_j5pb.Block
	root, ok := node.RootSchema()
	if ok {
		switch nt := root.(type) {
		case *j5schema.ObjectSchema:
			bclBlock = nt.BCL
		case *j5schema.OneofSchema:
			bclBlock = nt.BCL
		}
	}

	if bclBlock == nil {
		blockSpec.source = specSourceAuto
	} else {
		blockSpec, err = convertBlock(bclBlock)
		if err != nil {
			return nil, err
		}
		blockSpec.source = specSourceAnnotation
	}

	blockSpec.schema = schemaName
	return blockSpec, nil
}

func (ss *SchemaSet) _buildSpec(node j5PropSet) (*BlockSpec, error) {
	blockSpec, err := ss.lookupBlockSpec(node)
	if err != nil {
		return nil, err
	}

	if blockSpec.OnlyDefined {
		return blockSpec, nil
	}

	if blockSpec.Aliases == nil {
		blockSpec.Aliases = map[string]PathSpec{}
	}

	newAliases := map[string]PathSpec{}

	err = node.RangePropertySchemas(func(name string, required bool, schema *schema_j5pb.Field) error {

		switch field := schema.Type.(type) {
		case *schema_j5pb.Field_Object:

		case *schema_j5pb.Field_Oneof:

		case *schema_j5pb.Field_String_:
			if name == "name" && blockSpec.Name == nil {
				blockSpec.Name = &Tag{
					FieldName: "name",
				}
				blockSpec.Name.IsOptional = !required
			}
			if name == "description" && blockSpec.Description == nil {
				blockSpec.Description = proto.String("description")
			}

		case *schema_j5pb.Field_Bool,
			*schema_j5pb.Field_Integer,
			*schema_j5pb.Field_Float,
			*schema_j5pb.Field_Key,
			*schema_j5pb.Field_Enum,
			*schema_j5pb.Field_Bytes,
			*schema_j5pb.Field_Date,
			*schema_j5pb.Field_Timestamp,
			*schema_j5pb.Field_Decimal:

		case *schema_j5pb.Field_Array:
			if field.Array != nil && field.Array.Ext != nil && field.Array.Ext.SingleForm != nil {
				singleForm := *field.Array.Ext.SingleForm
				newAliases[singleForm] = PathSpec{name}
			} else {
				items := field.Array.Items
				switch itemSchema := items.Type.(type) {
				case *schema_j5pb.Field_Object:
					newAliases[arrayName(itemSchema.Object)] = PathSpec{name}
				}
			}
		case *schema_j5pb.Field_Map:
			if field.Map != nil && field.Map.Ext != nil && field.Map.Ext.SingleForm != nil {
				singleForm := *field.Map.Ext.SingleForm
				newAliases[singleForm] = PathSpec{name}
			}

		default:
			return fmt.Errorf("unimplemented schema type: %T", field)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for alias, path := range newAliases {
		if _, ok := blockSpec.Aliases[alias]; !ok {
			blockSpec.Aliases[alias] = path
		}
	}

	return blockSpec, nil
}

func (ss *SchemaSet) wrapContainer(node j5PropSet, path []string, loc *bcl_j5pb.SourceLocation) (*containerField, error) {
	spec, err := ss.blockSpec(node)
	if err != nil {
		return nil, err
	}

	return &containerField{
		schemaName: node.SchemaName(),
		container:  node,
		spec:       *spec,
		path:       path,
		location:   loc,
	}, nil

}

func (ss *SchemaSet) blockSpec(node j5PropSet) (*BlockSpec, error) {
	schemaName := node.SchemaName()

	var err error
	spec, ok := ss.cachedSpecs[schemaName]
	if !ok {
		spec, err = ss._buildSpec(node)
		if err != nil {
			return nil, err
		}
		ss.cachedSpecs[schemaName] = spec
	}

	return spec, nil
}

func arrayName(obj *schema_j5pb.ObjectField) string {
	var name string
	if ref := obj.GetRef(); ref != nil {
		name = ref.Schema
	} else if inline := obj.GetObject(); obj != nil {
		name = inline.Name
	}
	return strcase.ToLowerCamel(name)
}
