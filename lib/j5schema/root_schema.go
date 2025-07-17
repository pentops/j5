package j5schema

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/proto"
)

type RootSchema interface {
	AsRef() *RefSchema
	FullName() string
	Name() string
	PackageName() string
	Package() *Package
	ToJ5Root() *schema_j5pb.RootSchema
	ToJ5ClientRoot() *schema_j5pb.RootSchema
	Description() string
}

type RefSchema struct {
	Package *Package
	Schema  string
	To      RootSchema
}

func (ref *RefSchema) check() error {
	if ref.To.FullName() != ref.FullName() {
		//panic(fmt.Sprintf("placeholder %q links to %q", ref.FullName(), ref.To.FullName()))
		return fmt.Errorf("schema %q has wrong name %q", ref.FullName(), ref.To.FullName())
	}
	return nil
}

func (s *RefSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.Package.Name, s.Schema)
}

type rootSchema struct {
	description string
	pkg         *Package
	name        string
}

func (s *rootSchema) FullName() string {
	return fmt.Sprintf("%s.%s", s.pkg.Name, s.name)
}

func (s *rootSchema) Name() string {
	if s == nil {
		return "<nil>"
	}
	return s.name
}

func (s *rootSchema) PackageName() string {
	return s.pkg.Name
}

func (s *rootSchema) Package() *Package {
	return s.pkg
}

func (s *rootSchema) AsRef() *RefSchema {
	return &RefSchema{
		Package: s.pkg,
		Schema:  s.name,
	}
}

func (s *rootSchema) Description() string {
	return s.description
}

type EnumOption struct {
	name        string
	number      int32
	description string
	Info        map[string]string
}

func (eo *EnumOption) Name() string {
	return eo.name
}

func (eo *EnumOption) Number() int32 {
	return eo.number
}

func (eo *EnumOption) Description() string {
	return eo.description
}

func (eo *EnumOption) ToJ5EnumValue() *schema_j5pb.Enum_Option {
	return &schema_j5pb.Enum_Option{
		Name:        eo.name,
		Number:      eo.number,
		Description: eo.description,
		Info:        eo.Info,
	}

}

type EnumSchema struct {
	rootSchema

	NamePrefix string
	Options    []*EnumOption //schema_j5pb.Enum_Value

	InfoFields []*schema_j5pb.Enum_OptionInfoField
}

var _ RootSchema = (*EnumSchema)(nil)

func (s *EnumSchema) OptionByName(name string) *EnumOption {
	shortName := strings.TrimPrefix(name, s.NamePrefix)
	for _, opt := range s.Options {
		if opt.name == shortName {
			return opt
		}
	}
	return nil
}

func (s *EnumSchema) OptionsList() []string {
	options := make([]string, len(s.Options))
	for idx, opt := range s.Options {
		options[idx] = opt.name
	}
	return options
}

func (s *EnumSchema) OptionByNumber(num int32) *EnumOption {
	for _, opt := range s.Options {
		if opt.number == num {
			return opt
		}
	}
	return nil
}

func (s *EnumSchema) ToJ5Root() *schema_j5pb.RootSchema {
	options := make([]*schema_j5pb.Enum_Option, len(s.Options))
	for idx, opt := range s.Options {
		options[idx] = opt.ToJ5EnumValue()
	}
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Enum{
			Enum: &schema_j5pb.Enum{
				Name:        s.name,
				Description: s.description,
				Options:     options,
				Prefix:      s.NamePrefix,
				Info:        s.InfoFields,
			},
		},
	}
}

func (s *EnumSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	return s.ToJ5Root()
}

type ObjectSchema struct {
	rootSchema
	Entity          *schema_j5pb.EntityObject
	PolymorphMember []string
	Properties      PropertySet
	BCL             *bcl_j5pb.Block
	ListRequest     *list_j5pb.ListRequestMessage
}

func (s *ObjectSchema) Clone() *ObjectSchema {
	properties := make(PropertySet, len(s.Properties))
	copy(properties, s.Properties)
	return &ObjectSchema{
		rootSchema: rootSchema{
			description: s.description,
			pkg:         s.pkg,
			name:        s.name,
		},
		Properties: properties,
		BCL:        proto.Clone(s.BCL).(*bcl_j5pb.Block),
		Entity:     proto.Clone(s.Entity).(*schema_j5pb.EntityObject),
	}
}

func (s *ObjectSchema) ToJ5Object() *schema_j5pb.Object {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property := prop.ToJ5Proto()
		properties = append(properties, property)
	}
	return &schema_j5pb.Object{
		Description:     s.description,
		Name:            s.name,
		Properties:      properties,
		Entity:          s.Entity,
		PolymorphMember: s.PolymorphMember,
		Bcl:             s.BCL,
	}
}

func (s *ObjectSchema) AllProperties() PropertySet {
	return s.Properties
}

func (s *ObjectSchema) ClientProperties() PropertySet { //[]*ObjectProperty {
	properties := make([]*ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		switch propType := prop.Schema.(type) {
		case *ObjectField:
			if propType.Flatten {
				children := propType.ObjectSchema().ClientProperties()
				for _, child := range children {
					child := child.nestedClone() //prop.ProtoField)
					properties = append(properties, child)
				}

				continue
			}

		}
		properties = append(properties, prop)
	}
	return properties
}

func (s *ObjectSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: s.ToJ5ClientObject(),
		},
	}
}

func (s *ObjectSchema) ToJ5ClientObject() *schema_j5pb.Object {
	clientProperties := s.ClientProperties()
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(clientProperties))
	for _, prop := range clientProperties {
		property := prop.ToJ5Proto()
		properties = append(properties, property)
	}
	return &schema_j5pb.Object{
		Description:     s.description,
		Name:            s.name,
		Properties:      properties,
		Entity:          s.Entity,
		PolymorphMember: s.PolymorphMember,
	}
}

func (s *ObjectSchema) ToJ5Root() *schema_j5pb.RootSchema {
	built := s.ToJ5Object()
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: built,
		},
	}
}

type PolymorphSchema struct {
	rootSchema
	Members []string
}

func (s *PolymorphSchema) ToJ5Root() *schema_j5pb.RootSchema {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Polymorph{
			Polymorph: &schema_j5pb.Polymorph{
				Description: s.description,
				Name:        s.name,
				Members:     s.Members,
			},
		},
	}
}

func (s *PolymorphSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	return s.ToJ5Root()
}

type OneofSchema struct {
	rootSchema
	Properties PropertySet
	BCL        *bcl_j5pb.Block
}

func (s *OneofSchema) ToJ5Root() *schema_j5pb.RootSchema {
	properties := make([]*schema_j5pb.ObjectProperty, 0, len(s.Properties))
	for _, prop := range s.Properties {
		property := prop.ToJ5Proto()
		properties = append(properties, property)
	}

	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Oneof{
			Oneof: &schema_j5pb.Oneof{
				Description: s.description,
				Name:        s.name,
				Properties:  properties,
				Bcl:         s.BCL,
			},
		},
	}
}

func (s *OneofSchema) ToJ5ClientRoot() *schema_j5pb.RootSchema {
	return s.ToJ5Root()
}

func (s *OneofSchema) ClientProperties() PropertySet {
	return s.Properties
}

func (s *OneofSchema) AllProperties() PropertySet {
	return s.Properties
}

type PropertySet []*ObjectProperty

var _ Container = PropertySet{}

func (ps PropertySet) ByJSONName(name string) *ObjectProperty {
	for _, prop := range ps {
		if prop.JSONName == name {
			return prop
		}
	}
	return nil
}

func (ps PropertySet) ByClientJSONName(name string) (*ObjectProperty, []string) {
	for _, prop := range ps {
		if prop.JSONName == name {
			return prop, []string{name}
		}
		switch propType := prop.Schema.(type) {
		case *ObjectField:
			if propType.Flatten {
				childProp, pathToChild := propType.ObjectSchema().Properties.ByClientJSONName(name)
				if childProp != nil {
					return childProp, append([]string{prop.JSONName}, pathToChild...)
				}
			}

		}
	}
	return nil, nil

}

func (ps PropertySet) PropertyField(name string) FieldSchema {
	val := ps.ByJSONName(name)
	if val == nil {
		return nil
	}
	return val.Schema
}

func (ps PropertySet) WalkToProperty(name ...string) (FieldSchema, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("empty property path")
	}
	prop := ps.ByJSONName(name[0])
	if prop == nil {
		return nil, fmt.Errorf("property %q not found", name[0])
	}
	if len(name) == 1 {
		return prop.Schema, nil
	}
	propContainer, ok := prop.Schema.AsContainer()
	if !ok {
		return nil, fmt.Errorf("property %q is not a container", name[0])
	}
	return propContainer.WalkToProperty(name[1:]...)
}

type ObjectProperty struct {
	Parent RootSchema
	Schema FieldSchema
	Entity *schema_j5pb.EntityKey

	//ProtoField []protoreflect.FieldNumber
	//ProtoName  []protoreflect.Name

	JSONName string

	Required           bool
	ReadOnly           bool
	WriteOnly          bool
	ExplicitlyOptional bool

	Description string
}

func (prop *ObjectProperty) checkValid() error {
	if prop.Parent == nil {
		return fmt.Errorf("property %q has no parent", prop.FullName())
	}
	if prop.JSONName == "" {
		return fmt.Errorf("property %q has no JSON name", prop.FullName())
	}
	if prop.Schema == nil {
		return fmt.Errorf("property %q has no schema", prop.FullName())
	}
	return nil
}

func (prop *ObjectProperty) FullName() string {
	if prop.Parent == nil {
		return "<root>." + prop.JSONName
	}
	return prop.Parent.FullName() + "." + prop.JSONName
}

func (prop *ObjectProperty) ToJ5Proto() *schema_j5pb.ObjectProperty {

	propSchema := prop.Schema.ToJ5Field()

	switch propSchema.Type.(type) {
	case *schema_j5pb.Field_Key:
		// add deprecated key fields
		propSchema = proto.Clone(propSchema).(*schema_j5pb.Field)
		key := propSchema.GetKey()

		if prop.Entity != nil && prop.Entity.Primary {
			key.Entity = &schema_j5pb.KeyField_DeprecatedEntityKey{
				Type: &schema_j5pb.KeyField_DeprecatedEntityKey_PrimaryKey{
					PrimaryKey: true,
				},
			}
		} else if key.Ext != nil && key.Ext.Foreign != nil {
			key.Entity = &schema_j5pb.KeyField_DeprecatedEntityKey{
				Type: &schema_j5pb.KeyField_DeprecatedEntityKey_ForeignKey{
					ForeignKey: key.Ext.Foreign,
				},
			}
		}

	}

	return &schema_j5pb.ObjectProperty{
		Schema:             propSchema,
		EntityKey:          prop.Entity,
		Name:               prop.JSONName,
		Required:           prop.Required,
		ExplicitlyOptional: prop.ExplicitlyOptional,
		Description:        prop.Description,
		//	ProtoField:         int32(prop.ProtoField),
	}

}

func (prop *ObjectProperty) nestedClone() *ObjectProperty {
	//protoField := append(inParent, prop.ProtoField...)
	return &ObjectProperty{
		Parent: prop.Parent,
		Schema: prop.Schema,
		Entity: prop.Entity,
		//ProtoField:         protoField,
		JSONName:           prop.JSONName,
		Required:           prop.Required,
		ReadOnly:           prop.ReadOnly,
		WriteOnly:          prop.WriteOnly,
		ExplicitlyOptional: prop.ExplicitlyOptional,
		Description:        prop.Description,
	}
}
