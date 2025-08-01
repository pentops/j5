package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/
type Property interface {
	IsSet() bool
	Schema() *j5schema.ObjectProperty
	//CreateField() (Field, error)
	Field() (Field, error)

	IsArray() bool
	IsMap() bool

	PropertyType() PropertyType
}

type PropertyType int

const (
	MapProperty PropertyType = iota
	ArrayProperty
	ObjectProperty
	OneofProperty
	EnumProperty
	ScalarProperty
	AnyProperty
	PolymorphProperty
)

func (pt PropertyType) String() string {
	switch pt {
	case MapProperty:
		return "Map"
	case ArrayProperty:
		return "Array"
	case ObjectProperty:
		return "Object"
	case OneofProperty:
		return "Oneof"
	case EnumProperty:
		return "Enum"
	case ScalarProperty:
		return "Scalar"
	case AnyProperty:
		return "Any"
	case PolymorphProperty:
		return "Polymorph"
	default:
		return fmt.Sprintf("Unknown(%d)", pt)
	}
}

/*** Implementation ***/
type propertySchema struct {
	schema    *j5schema.ObjectProperty
	protoPath []protoreflect.FieldDescriptor
}

func (p *propertySchema) newProperty(ps *propSet) *property {
	return &property{
		propertySchema: *p,
		propSet:        ps,
	}
}

type property struct {
	propertySchema
	propSet *propSet

	hasField bool
	_field   Field
}

var _ Property = &property{}

func (p *property) IsSet() bool {
	field, err := p.Field()
	if err != nil {
		panic(err.Error())
	}
	return field.IsSet()
}

func (p *property) IsArray() bool {
	_, ok := p.schema.Schema.(*j5schema.ArrayField)
	return ok
}

func (p *property) IsMap() bool {
	_, ok := p.schema.Schema.(*j5schema.MapField)
	return ok
}

func (p *property) PropertyType() PropertyType {
	switch p.schema.Schema.(type) {
	case *j5schema.ArrayField:
		return ArrayProperty
	case *j5schema.MapField:
		return MapProperty
	case *j5schema.ObjectField:
		return ObjectProperty
	case *j5schema.OneofField:
		return OneofProperty
	case *j5schema.EnumField:
		return EnumProperty
	case *j5schema.ScalarSchema:
		return ScalarProperty
	case *j5schema.AnyField:
		return AnyProperty
	case *j5schema.PolymorphField:
		return PolymorphProperty
	default:
		return -1
	}
}

func (p *property) Schema() *j5schema.ObjectProperty {
	return p.schema
}

func (prop *property) Field() (Field, error) {
	if !prop.hasField {
		err := prop.buildField()
		if err != nil {
			return nil, fmt.Errorf("building field %s: %w", prop.schema.JSONName, err)
		}
	}
	return prop._field, nil
}

func (prop *property) buildField() error {
	var err error

	walkMessage := prop.propSet.value

	fieldContext := &propertyContext{
		schema: prop.schema,
	}

	var walkField protoreflect.FieldDescriptor
	walkPath := prop.protoPath[:]
	for len(walkPath) > 1 {
		walkField, walkPath = walkPath[0], walkPath[1:]
		fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, walkField.JSONName())
		// walking along the anonymous field, creating the mutable message on
		// the fly if it doesn't exist - this means there is no meaning in nil
		// or not-nil for anonymous messages.

		childVal, err := walkMessage.ChildField(walkField)
		if err != nil {
			return fmt.Errorf("error walking field %s: %w", walkField.FullName(), err)
		}

		childMessage, ok := childVal.AsMessage()
		if !ok {
			panic(fmt.Sprintf("expected message for field %s, got %T", walkField.FullName(), childVal))
		}

		walkMessage = childMessage
	}

	finalField := walkPath[0]
	fieldContext.walkedProtoPath = append(fieldContext.walkedProtoPath, finalField.JSONName())
	protoVal, err := walkMessage.ChildField(finalField)
	if err != nil {
		return err
	}

	built, err := buildProperty(fieldContext, prop.schema, protoVal)
	if err != nil {
		return err
	}

	prop._field = built
	prop.hasField = true

	return nil
}

func schemaIsMutable(schema j5schema.FieldSchema) (bool, error) {
	switch schema.(type) {
	case *j5schema.ObjectField, *j5schema.OneofField, *j5schema.AnyField, *j5schema.PolymorphField:
		return true, nil
	case *j5schema.ArrayField, *j5schema.MapField:
		return false, fmt.Errorf("item schema must not itself be an array or map")
	case *j5schema.EnumField, *j5schema.ScalarSchema:
		return false, nil
	default:
		return false, fmt.Errorf("unknown schema type %T", schema)
	}
}

func buildProperty(context fieldContext, schema *j5schema.ObjectProperty, value protoval.Value) (Field, error) {

	switch st := schema.Schema.(type) {

	case *j5schema.ArrayField:

		listVal, ok := value.AsList()
		if !ok {
			return nil, fmt.Errorf("ArrayField is not a proto list")
		}

		schemaIsMutable, err := schemaIsMutable(st.ItemSchema)
		if err != nil {
			return nil, fmt.Errorf("item schema for Array %s: %w", context.FullTypeName(), err)
		}
		if schemaIsMutable {
			msgDesc, ok := listVal.ItemMessageDescriptor()
			if !ok {
				return nil, fmt.Errorf("ArrayField item is not a proto message")
			}

			ff, err := newMessageFieldFactory(st.ItemSchema, msgDesc)
			if err != nil {
				return nil, err
			}

			field, err := newMessageArrayField(context, st, listVal, ff)
			if err != nil {
				return nil, err
			}

			return field, nil
		}

		ff, err := newFieldFactory(st.ItemSchema, listVal.ItemFieldDescriptor())
		if err != nil {
			return nil, err
		}

		field, err := newLeafArrayField(context, st, listVal, ff)
		if err != nil {
			return nil, err
		}

		return field, nil

	case *j5schema.MapField:

		mapVal, ok := value.AsMap()
		if !ok {
			return nil, fmt.Errorf("MapField is not a proto map")
		}

		schemaIsMutable, err := schemaIsMutable(st.ItemSchema)
		if err != nil {
			return nil, fmt.Errorf("item schema for Map %s: %w", context.FullTypeName(), err)
		}
		if schemaIsMutable {
			msgDesc, ok := mapVal.ItemMessageDescriptor()
			if !ok {
				return nil, fmt.Errorf("MapField item is not a proto message")
			}
			ff, err := newMessageFieldFactory(st.ItemSchema, msgDesc)
			if err != nil {
				return nil, err
			}
			field, err := newMessageMapField(context, st, mapVal, ff)
			if err != nil {
				return nil, err
			}
			return field, nil
		}

		ff, err := newFieldFactory(st.ItemSchema, mapVal.ItemFieldDescriptor())
		if err != nil {
			return nil, err
		}

		field, err := newLeafMapField(context, st, mapVal, ff)
		if err != nil {
			return nil, err
		}
		return field, nil

	case *j5schema.ObjectField, *j5schema.OneofField, *j5schema.AnyField, *j5schema.PolymorphField:

		msgVal, ok := value.AsMessage()
		if !ok {
			return nil, fmt.Errorf("ObjectField is not a proto message")
		}
		ff, err := newMessageFieldFactory(schema.Schema, msgVal.MessageDescriptor())
		if err != nil {
			return nil, err
		}
		field := ff.buildField(context, msgVal)
		return field, nil
	}

	fieldDesc, ok := value.FieldDescriptor()
	if !ok {
		return nil, fmt.Errorf("value %s has no field descriptor, for property", value)
	}
	ff, err := newFieldFactory(schema.Schema, fieldDesc)
	if err != nil {
		return nil, fmt.Errorf("field factoy for scalar: %w", err)
	}

	field := ff.buildField(context, value)
	return field, nil

}

type fieldFactory interface {
	buildField(schema fieldContext, value protoval.Value) Field
}

func newMessageFieldFactory(schema j5schema.FieldSchema, desc protoreflect.MessageDescriptor) (fieldFactory, error) {

	switch st := schema.(type) {
	case *j5schema.ObjectField:
		propSetFactory, err := newPropSetFactory(st.ObjectSchema(), desc)
		if err != nil {
			return nil, err
		}
		return &objectFieldFactory{
			schema:  st,
			propSet: propSetFactory,
		}, nil

	case *j5schema.OneofField:
		propSetFactory, err := newPropSetFactory(st.OneofSchema(), desc)
		if err != nil {
			return nil, err
		}
		return &oneofFieldFactory{
			schema:  st,
			propSet: propSetFactory,
		}, nil

	case *j5schema.AnyField:
		return &anyFieldFactory{schema: st}, nil

	case *j5schema.PolymorphField:
		return &polymorphFieldFactory{schema: st}, nil

	default:
		panic(fmt.Sprintf("invalid schema for message field: %T", schema))
	}
}

func newFieldFactory(schema j5schema.FieldSchema, field protoreflect.FieldDescriptor) (fieldFactory, error) {
	switch st := schema.(type) {
	case *j5schema.EnumField:
		if field.Kind() != protoreflect.EnumKind {
			return nil, fmt.Errorf("EnumField is kind %s", field.Kind())
		}
		return &enumFieldFactory{schema: st}, nil

	case *j5schema.ScalarSchema:
		if st.WellKnownTypeName != "" {
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("ScalarField is proto kind %s, want message for %T", field.Kind(), st.Proto.Type)
			}
			if string(field.Message().FullName()) != string(st.WellKnownTypeName) {
				return nil, fmt.Errorf("ScalarField message is %s, want %s for %T", field.Message().FullName(), st.WellKnownTypeName, st.Proto.Type)
			}
		} else if field.Kind() != st.Kind {
			return nil, fmt.Errorf("ScalarField is proto kind %s, want schema %q for %T", field.Kind(), st.Kind, st.Proto.Type)
		}
		return &scalarFieldFactory{schema: st}, nil

	default:
		panic(fmt.Sprintf("invalid schema for leaf field: %T", schema))
	}
}
