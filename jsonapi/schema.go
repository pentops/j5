package jsonapi

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/google/uuid"
	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type SchemaSet struct {
	Options Options
	Schemas map[string]*SchemaItem

	seen map[string]bool
}

func NewSchemaSet(options Options) *SchemaSet {
	return &SchemaSet{
		Options: options,
		Schemas: make(map[string]*SchemaItem),
		seen:    make(map[string]bool),
	}
}

func walkName(src protoreflect.MessageDescriptor) string {
	goTypeName := string(src.Name())
	parent := src.Parent()
	// if parent is a message
	msg, ok := parent.(protoreflect.MessageDescriptor)
	if !ok {
		return goTypeName
	}

	return fmt.Sprintf("%s_%s", walkName(msg), goTypeName)

}
func (ss *SchemaSet) BuildSchemaObject(src protoreflect.MessageDescriptor) (*SchemaItem, error) {

	goTypeName := walkName(src)

	obj := &ObjectItem{
		FullProtoName: string(src.FullName()),
		ProtoName:     string(src.Name()),
		GoPackageName: src.ParentFile().Options().(*descriptorpb.FileOptions).GetGoPackage(),
		GRPCPackage:   string(src.ParentFile().Package()),
		GoTypeName:    goTypeName,

		Properties: make([]*ObjectProperty, 0, src.Fields().Len()),
	}

	options := proto.GetExtension(src.Options(), jsonapi_pb.E_Message).(*jsonapi_pb.MessageOptions)
	if options != nil {
		if options.IsOneofWrapper {
			obj.IsOneof = true
		}
	}

	oneofs := make(map[string]*ObjectItem)
	pendingOneofProps := make(map[string]*ObjectProperty)

	if !obj.IsOneof {
		for idx := 0; idx < src.Oneofs().Len(); idx++ {
			oneof := src.Oneofs().Get(idx)
			if oneof.IsSynthetic() {
				continue
			}
			ext := proto.GetExtension(oneof.Options(), jsonapi_pb.E_Oneof).(*jsonapi_pb.OneofOptions)

			if ext == nil || !ext.Expose {
				if !obj.IsOneof {
					fmt.Fprintf(os.Stderr, "WARN: no def for oneof %s.%s\n", src.FullName(), oneof.Name())
				}
				//continue
			}

			oneofName := string(oneof.Name())
			syntheticTypeName := fmt.Sprintf("%s_%s", src.Name(), oneofName)
			oneofObject := &ObjectItem{
				FullProtoName: string(oneof.FullName()),
				ProtoName:     oneofName,
				IsOneof:       true,
				GoTypeName:    syntheticTypeName,
				GoPackageName: obj.GoPackageName,
				GRPCPackage:   obj.GRPCPackage,
			}
			prop := &ObjectProperty{
				ProtoFieldName: string(oneof.Name()),
				Name:           string(oneof.Name()),
				Description:    commentDescription(src),
				SchemaItem: SchemaItem{
					ItemType: oneofObject,
				},
			}
			pendingOneofProps[oneofName] = prop
			oneofs[oneofName] = oneofObject

		}
	}

	for ii := 0; ii < src.Fields().Len(); ii++ {
		field := src.Fields().Get(ii)

		if field.IsList() {
			prop, err := ss.buildSchemaProperty(field)
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}
			prop.SchemaItem = SchemaItem{
				ItemType: ArrayItem{
					Items: prop.SchemaItem,
				},
			}
			obj.Properties = append(obj.Properties, prop)
			continue

		}
		if field.IsMap() {
			keyProp, err := ss.buildSchemaProperty(field.MapKey())
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}

			valueProp, err := ss.buildSchemaProperty(field.MapValue())
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}

			src := field
			prop := &ObjectProperty{
				ProtoFieldName:   string(src.Name()),
				ProtoFieldNumber: int(src.Number()),
				Name:             string(src.JSONName()),
				Description:      commentDescription(src),
			}
			prop.SchemaItem = SchemaItem{
				ItemType: MapItem{
					KeyProperty:   keyProp.SchemaItem,
					ValueProperty: valueProp.SchemaItem,
				},
			}
			obj.Properties = append(obj.Properties, prop)
			continue
		}

		prop, err := ss.buildSchemaProperty(field)
		if err != nil {
			return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
		}

		inOneof := field.ContainingOneof()
		if obj.IsOneof || inOneof == nil || inOneof.IsSynthetic() {
			obj.Properties = append(obj.Properties, prop)
			continue
		}

		name := string(inOneof.Name())

		oneof, ok := oneofs[name]
		if !ok {
			obj.Properties = append(obj.Properties, prop)
			continue
		}

		oneof.Properties = append(oneof.Properties, prop)

		// deferrs adding the oneof to the property array until the first
		// field is encountered, i.e. preserves ordering
		pending, ok := pendingOneofProps[name]
		if ok {
			obj.Properties = append(obj.Properties, pending)
			delete(pendingOneofProps, name)
		}
	}

	for _, pending := range pendingOneofProps {
		return nil, fmt.Errorf("oneof %s has not been added", pending.Name)
	}

	description := commentDescription(src)

	return &SchemaItem{
		Description: description,
		ItemType:    obj,
	}, nil
}

func commentDescription(src protoreflect.Descriptor) string {
	sourceLocation := src.ParentFile().SourceLocations().ByDescriptor(src)
	return buildComment(sourceLocation, "")
}

func buildComment(sourceLocation protoreflect.SourceLocation, fallback string) string {
	allComments := make([]string, 0)
	if sourceLocation.LeadingComments != "" {
		allComments = append(allComments, strings.Split(sourceLocation.LeadingComments, "\n")...)
	}
	if sourceLocation.TrailingComments != "" {
		allComments = append(allComments, strings.Split(sourceLocation.TrailingComments, "\n")...)
	}

	// Trim leading whitespace
	commentsOut := make([]string, 0, len(allComments))
	for _, comment := range allComments {
		comment = strings.TrimSpace(comment)
		if comment == "" {
			continue
		}
		if strings.HasPrefix(comment, "#") {
			continue
		}
		commentsOut = append(commentsOut, comment)
	}

	if len(commentsOut) <= 0 {
		return fallback
	}
	return strings.Join(commentsOut, "\n")
}

type wellKnownStringPattern struct {
	format  string
	example string
}

var wellKnownStringPatterns = map[string]wellKnownStringPattern{
	`^\d{4}-\d{2}-\d{2}$`: {format: "date", example: "2021-01-01"},
	`^\d(.?\d)?$`:         {format: "number", example: "12.34"},
}

func (ss *SchemaSet) buildSchemaProperty(src protoreflect.FieldDescriptor) (*ObjectProperty, error) {

	prop := &ObjectProperty{
		ProtoFieldName:   string(src.Name()),
		ProtoFieldNumber: int(src.Number()),
		Name:             string(src.JSONName()),
		Description:      commentDescription(src),
	}

	// second _ prevents a panic when the exception is not set
	constraint, _ := proto.GetExtension(src.Options(), validate.E_Field).(*validate.FieldConstraints)

	if constraint != nil && !constraint.Skipped {
		if constraint.Required {
			prop.Required = true
		}
		// constraint.IgnoreEmpty doesn't really apply
	}

	if !prop.Required && src.HasOptionalKeyword() {
		prop.Optional = true
	}

	// TODO: Validation / Rules
	// TODO: Map
	// TODO: Extra types (see below)

	switch src.Kind() {
	case protoreflect.BoolKind:
		boolConstraint := constraint.GetBool()
		boolItem := BooleanItem{}

		if boolConstraint != nil {
			if boolConstraint.Const != nil {
				boolItem.BooleanRules.Const = Value(*boolConstraint.Const)
			}
		}
		prop.SchemaItem = SchemaItem{
			ItemType: boolItem,
		}
		prop.Required = true

	case protoreflect.EnumKind:
		enumConstraint := constraint.GetEnum()
		values, err := ss.Options.ShortEnums.EnumValues(src.Enum().Values(), enumConstraint)
		if err != nil {
			return nil, err
		}

		valueStrings := make([]string, 0, len(values))
		for _, value := range values {
			valueStrings = append(valueStrings, value.Name)
		}

		refSchemaItem := &SchemaItem{
			ItemType: EnumItem{
				EnumRules: EnumRules{
					Enum: valueStrings,
				},
				Extended: values,
			},
		}

		name := string(src.Enum().FullName())
		ss.Schemas[name] = refSchemaItem

		prop.SchemaItem = SchemaItem{
			Ref: name,
		}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind:

		var integerRules IntegerRules
		int32Constraint := constraint.GetInt32()
		if int32Constraint != nil {
			integerRules = IntegerRules{}
			if int32Constraint.Const != nil {
				return nil, fmt.Errorf("'const' not supported")
			}
			if int32Constraint.In != nil {
				return nil, fmt.Errorf("'in' not supported")
			}
			if int32Constraint.NotIn != nil {
				return nil, fmt.Errorf("'notIn' not supported")
			}

			if int32Constraint.LessThan != nil {
				switch cType := int32Constraint.LessThan.(type) {
				case *validate.Int32Rules_Lt:
					integerRules.Maximum = Value(int64(cType.Lt))
					integerRules.ExclusiveMaximum = Value(true)
				case *validate.Int32Rules_Lte:
					integerRules.Maximum = Value(int64(cType.Lte))
				}
			}
			if int32Constraint.GreaterThan != nil {
				switch cType := int32Constraint.GreaterThan.(type) {
				case *validate.Int32Rules_Gt:
					integerRules.Minimum = Value(int64(cType.Gt))
					integerRules.ExclusiveMinimum = Value(true)
				case *validate.Int32Rules_Gte:
					integerRules.Minimum = Value(int64(cType.Gte))
				}
			}

		}
		prop.SchemaItem = SchemaItem{
			ItemType: IntegerItem{
				Format:       "int32",
				IntegerRules: integerRules,
			},
		}
	case protoreflect.Uint32Kind:
		var integerRules IntegerRules
		uint32Constraint := constraint.GetUint32()
		if uint32Constraint != nil {
			integerRules = IntegerRules{}
			if uint32Constraint.Const != nil {
				return nil, fmt.Errorf("'const' not supported")
			}
			if uint32Constraint.In != nil {
				return nil, fmt.Errorf("'in' not supported")
			}
			if uint32Constraint.NotIn != nil {
				return nil, fmt.Errorf("'notIn' not supported")
			}

			if uint32Constraint.LessThan != nil {
				switch cType := uint32Constraint.LessThan.(type) {
				case *validate.UInt32Rules_Lt:
					integerRules.Maximum = Value(int64(cType.Lt))
					integerRules.ExclusiveMaximum = Value(true)
				case *validate.UInt32Rules_Lte:
					integerRules.Maximum = Value(int64(cType.Lte))
				}
			}
			if uint32Constraint.GreaterThan != nil {
				switch cType := uint32Constraint.GreaterThan.(type) {
				case *validate.UInt32Rules_Gt:
					integerRules.Minimum = Value(int64(cType.Gt))
					integerRules.ExclusiveMinimum = Value(true)
				case *validate.UInt32Rules_Gte:
					integerRules.Minimum = Value(int64(cType.Gte))
				}
			}
		}

		prop.SchemaItem = SchemaItem{
			ItemType: IntegerItem{
				Format:       "uint32", // Not an 'official' format
				IntegerRules: integerRules,
			},
		}

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
		var integerRules IntegerRules
		int64Constraint := constraint.GetInt64()
		if int64Constraint != nil {
			integerRules = IntegerRules{}
			if int64Constraint.Const != nil {
				return nil, fmt.Errorf("'const' not supported")
			}
			if int64Constraint.In != nil {
				return nil, fmt.Errorf("'in' not supported")
			}
			if int64Constraint.NotIn != nil {
				return nil, fmt.Errorf("'notIn' not supported")
			}

			if int64Constraint.LessThan != nil {
				switch cType := int64Constraint.LessThan.(type) {
				case *validate.Int64Rules_Lt:
					integerRules.Maximum = Value(cType.Lt)
					integerRules.ExclusiveMaximum = Value(true)
				case *validate.Int64Rules_Lte:
					integerRules.Maximum = Value(cType.Lte)
				}
			}
			if int64Constraint.GreaterThan != nil {
				switch cType := int64Constraint.GreaterThan.(type) {
				case *validate.Int64Rules_Gt:
					integerRules.Minimum = Value(cType.Gt)
					integerRules.ExclusiveMinimum = Value(true)
				case *validate.Int64Rules_Gte:
					integerRules.Minimum = Value(cType.Gte)
				}
			}
		}

		prop.SchemaItem = SchemaItem{
			ItemType: IntegerItem{
				Format:       "int64",
				IntegerRules: integerRules,
			},
		}

	case protoreflect.FloatKind:
		var numberRules NumberRules
		floatConstraint := constraint.GetFloat()
		if floatConstraint != nil {
			numberRules = NumberRules{}
			if floatConstraint.Const != nil {
				return nil, fmt.Errorf("'const' not supported")
			}
			if floatConstraint.In != nil {
				return nil, fmt.Errorf("'in' not supported")
			}
			if floatConstraint.NotIn != nil {
				return nil, fmt.Errorf("'notIn' not supported")
			}

			if floatConstraint.LessThan != nil {
				switch cType := floatConstraint.LessThan.(type) {
				case *validate.FloatRules_Lt:
					numberRules.Maximum = Value(float64(cType.Lt))
					numberRules.ExclusiveMaximum = Value(true)
				case *validate.FloatRules_Lte:
					numberRules.Maximum = Value(float64(cType.Lte))
				}
			}
			if floatConstraint.GreaterThan != nil {
				switch cType := floatConstraint.GreaterThan.(type) {
				case *validate.FloatRules_Gt:
					numberRules.Minimum = Value(float64(cType.Gt))
					numberRules.ExclusiveMinimum = Value(true)
				case *validate.FloatRules_Gte:
					numberRules.Minimum = Value(float64(cType.Gte))
				}
			}
		}

		prop.SchemaItem = SchemaItem{
			ItemType: NumberItem{
				Format:      "float",
				NumberRules: numberRules,
			},
		}

	case protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		var numberRules NumberRules
		floatConstraint := constraint.GetDouble()
		if floatConstraint != nil {
			numberRules = NumberRules{}
			if floatConstraint.Const != nil {
				return nil, fmt.Errorf("'const' not supported")
			}
			if floatConstraint.In != nil {
				return nil, fmt.Errorf("'in' not supported")
			}
			if floatConstraint.NotIn != nil {
				return nil, fmt.Errorf("'notIn' not supported")
			}

			if floatConstraint.LessThan != nil {
				switch cType := floatConstraint.LessThan.(type) {
				case *validate.DoubleRules_Lt:
					numberRules.Maximum = Value(float64(cType.Lt))
					numberRules.ExclusiveMaximum = Value(true)
				case *validate.DoubleRules_Lte:
					numberRules.Maximum = Value(float64(cType.Lte))
				}
			}
			if floatConstraint.GreaterThan != nil {
				switch cType := floatConstraint.GreaterThan.(type) {
				case *validate.DoubleRules_Gt:
					numberRules.Minimum = Value(float64(cType.Gt))
					numberRules.ExclusiveMinimum = Value(true)
				case *validate.DoubleRules_Gte:
					numberRules.Minimum = Value(float64(cType.Gte))
				}
			}
		}

		prop.SchemaItem = SchemaItem{
			ItemType: NumberItem{
				Format:      "float",
				NumberRules: numberRules,
			},
		}

	case protoreflect.StringKind:
		stringItem := StringItem{}
		if constraint != nil && constraint.Type != nil {
			stringConstraint, ok := constraint.Type.(*validate.FieldConstraints_String_)
			if !ok {
				return nil, fmt.Errorf("wrong constraint type for string: %T", constraint.Type)
			}

			stringItem.StringRules = StringRules{}
			constraint := stringConstraint.String_

			if constraint.MinLen != nil {
				stringItem.StringRules.MinLength = Value(*constraint.MinLen)
			}

			if constraint.MaxLen != nil {
				stringItem.StringRules.MaxLength = Value(*constraint.MaxLen)
			}

			if constraint.Pattern != nil {
				pattern := *constraint.Pattern
				wellKnownStringPattern, ok := wellKnownStringPatterns[pattern]
				if ok {
					stringItem.Format = wellKnownStringPattern.format
					stringItem.Example = wellKnownStringPattern.example
				} else {
					stringItem.Pattern = pattern
				}
			}

			switch wkt := constraint.WellKnown.(type) {
			case *validate.StringRules_Uuid:
				if wkt.Uuid {
					stringItem.Format = "uuid"
					stringItem.Example = uuid.NewString()
				}
			case *validate.StringRules_Email:
				if wkt.Email {
					stringItem.Format = "email"
					stringItem.Example = "test@example.com"
				}

			case *validate.StringRules_Hostname:
				if wkt.Hostname {
					stringItem.Format = "hostname"
					stringItem.Example = "example.com"
				}

			case *validate.StringRules_Ipv4:
				if wkt.Ipv4 {
					stringItem.Format = "ipv4"
					stringItem.Example = "10.10.10.10"
				}

			case *validate.StringRules_Ipv6:
				if wkt.Ipv6 {
					stringItem.Format = "ipv6"
					stringItem.Example = "2001:db8::68"
				}

			case *validate.StringRules_Uri:
				if wkt.Uri {
					stringItem.Format = "uri"
					stringItem.Example = "https://example.com"
				}

			// Other types not supported by swagger
			case nil:

			default:
				return nil, fmt.Errorf("unknown string constraint: %T", constraint.WellKnown)

			}

		}

		prop.SchemaItem = SchemaItem{
			ItemType: stringItem,
		}

	case protoreflect.BytesKind:
		prop.SchemaItem = SchemaItem{
			ItemType: StringItem{
				Format: "byte",
			},
		}

	case protoreflect.MessageKind:
		// When called from a field of a message, this creates a ref. When built directly from a service RPC request or create, this code is not called, they are inlined with the buildSchemaObject call directly
		if wktschema, ok := wktSchema(src.Message()); ok {
			prop.SchemaItem = *wktschema

		} else {
			prop.SchemaItem = SchemaItem{
				Ref: string(src.Message().FullName()),
			}
			if err := ss.addSchemaObject(src.Message()); err != nil {
				return nil, err
			}
		}

	default:
		/* TODO:
		Sfixed32Kind Kind = 15
		Fixed32Kind  Kind = 7
		Sfixed64Kind Kind = 16
		Fixed64Kind  Kind = 6
		GroupKind    Kind = 10
		*/
		return nil, fmt.Errorf("unsupported field type %s", src.Kind())
	}

	return prop, nil

}

func wktSchema(src protoreflect.MessageDescriptor) (*SchemaItem, bool) {

	switch string(src.FullName()) {
	case "google.protobuf.Timestamp":
		return &SchemaItem{
			ItemType: StringItem{
				Format: "date-time",
			},
		}, true
	case "google.protobuf.Duration":
		return &SchemaItem{
			ItemType: StringItem{
				Format: "duration",
			},
		}, true

	case "google.protobuf.Struct":
		return &SchemaItem{
			ItemType: &MapItem{
				KeyProperty: SchemaItem{
					ItemType: StringItem{},
				},
				ValueProperty: SchemaItem{
					ItemType: EmptySchemaItem{},
				},
			},
		}, true

	}

	return nil, false

}

func (ss *SchemaSet) addSchemaObject(src protoreflect.MessageDescriptor) error {

	if _, ok := ss.Schemas[string(src.FullName())]; ok {
		return nil
	}

	if strings.HasPrefix(string(src.FullName()), "google.protobuf.") {
		return fmt.Errorf("unknown google.protobuf type %s", src.FullName())
	}

	// Prevents recursion errors
	ss.Schemas[string(src.FullName())] = &SchemaItem{}

	schema, err := ss.BuildSchemaObject(src)
	if err != nil {
		return err
	}

	ss.Schemas[string(src.FullName())] = schema

	return nil

}

type ItemType interface {
	TypeName() string
}

type SchemaItem struct {
	Ref   string
	OneOf []SchemaItem
	AnyOf []SchemaItem

	ItemType
	Description string `json:"description,omitempty"`
}

func (si SchemaItem) fieldMap() (map[string]json.RawMessage, error) {
	if si.Ref != "" {
		if len(si.OneOf) > 0 || len(si.AnyOf) > 0 || si.ItemType != nil {
			return nil, fmt.Errorf("schema item has both a ref and other properties")
		}
		return toJsonFieldMap(map[string]interface{}{
			"$ref": si.Ref,
		})
	}

	if len(si.OneOf) > 0 {
		if len(si.AnyOf) > 0 || si.ItemType != nil {
			return nil, fmt.Errorf("schema item has both oneOf and other properties")
		}
		return toJsonFieldMap(map[string]interface{}{
			"oneOf": si.OneOf,
		})
	}

	if len(si.AnyOf) > 0 {
		if len(si.OneOf) > 0 || si.ItemType != nil {
			return nil, fmt.Errorf("schema item has both a anyOf and other properties")
		}
		return toJsonFieldMap(map[string]interface{}{
			"anyOf": si.AnyOf,
		})
	}

	if si.ItemType == nil {
		return nil, fmt.Errorf("schema item has no type")
	}

	propOut := map[string]json.RawMessage{}
	if err := jsonFieldMap(si.ItemType, propOut); err != nil {
		return nil, fmt.Errorf("fieldMap item: %w", err)
	}
	propOut["type"], _ = json.Marshal(si.TypeName())
	if si.Description != "" {
		propOut["description"], _ = json.Marshal(si.Description)
	}

	return propOut, nil
}

func (si SchemaItem) MarshalJSON() ([]byte, error) {
	propOut, err := si.fieldMap()
	if err != nil {
		return nil, fmt.Errorf("si MarshalJson fieldMap: %w", err)
	}

	return json.Marshal(propOut)
}

type EmptySchemaItem struct{}

func (ri EmptySchemaItem) TypeName() string {
	return "object"
}

type StringItem struct {
	Format  string `json:"format,omitempty"`
	Example string `json:"example,omitempty"`
	StringRules
}

func (ri StringItem) TypeName() string {
	return "string"
}

type StringRules struct {
	Pattern   string           `json:"pattern,omitempty"`
	MinLength Optional[uint64] `json:"minLength,omitempty"`
	MaxLength Optional[uint64] `json:"maxLength,omitempty"`
}

// EnumItem represents a PROTO enum in Swagger, so can only be a string
type EnumItem struct {
	EnumRules
	Extended []EnumValueDescription `json:"x-enum"`
}

func (ri EnumItem) TypeName() string {
	return "string"
}

type EnumRules struct {
	Enum []string `json:"enum,omitempty"`
}

type NumberItem struct {
	Format string `json:"format,omitempty"`
	NumberRules
}

func (ri NumberItem) TypeName() string {
	return "number"
}

type NumberRules struct {
	ExclusiveMaximum Optional[bool]    `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum Optional[bool]    `json:"exclusiveMinimum,omitempty"`
	Minimum          Optional[float64] `json:"minimum,omitempty"`
	Maximum          Optional[float64] `json:"maximum,omitempty"`
	MultipleOf       Optional[float64] `json:"multipleOf,omitempty"`
}

type IntegerItem struct {
	Format string `json:"format,omitempty"`
	IntegerRules
}

func (ri IntegerItem) TypeName() string {
	return "integer"
}

type IntegerRules struct {
	ExclusiveMaximum Optional[bool]  `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum Optional[bool]  `json:"exclusiveMinimum,omitempty"`
	Minimum          Optional[int64] `json:"minimum,omitempty"`
	Maximum          Optional[int64] `json:"maximum,omitempty"`
	MultipleOf       Optional[int64] `json:"multipleOf,omitempty"`
}

type BooleanItem struct {
	BooleanRules
}

func (ri BooleanItem) TypeName() string {
	return "boolean"
}

type BooleanRules struct {
	Const Optional[bool] `json:"const,omitempty"`
}

type ArrayItem struct {
	ArrayRules
	Items SchemaItem `json:"items,omitempty"`
}

func (ri ArrayItem) TypeName() string {
	return "array"
}

type ArrayRules struct {
	MinItems    Optional[uint64] `json:"minItems,omitempty"`
	MaxItems    Optional[uint64] `json:"maxItems,omitempty"`
	UniqueItems Optional[bool]   `json:"uniqueItems,omitempty"`
}

type MapItem struct {
	ValueProperty SchemaItem `json:"additionalProperties,omitempty"`
	KeyProperty   SchemaItem `json:"x-key-property,omitempty"` // Only used for maps
}

func (mi MapItem) TypeName() string {
	return "object"
}

type ObjectItem struct {
	ObjectRules
	Properties []*ObjectProperty `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	debug      string

	FullProtoName string `json:"x-proto-full-name"`
	ProtoName     string `json:"x-proto-name"`
	IsOneof       bool   `json:"x-is-oneof,omitempty"`

	GoPackageName string `json:"-"`
	GoTypeName    string `json:"-"`
	GRPCPackage   string `json:"-"`
}

func (ri *ObjectItem) TypeName() string {
	return "object"
}

func (op *ObjectItem) PopProperty(name string) (*ObjectProperty, bool) {
	newProps := make([]*ObjectProperty, 0, len(op.Properties))
	var found *ObjectProperty
	for _, prop := range op.Properties {
		if prop.ProtoFieldName == name {
			found = prop
		} else {
			newProps = append(newProps, prop)
		}
	}
	op.Properties = newProps
	if found != nil {
		return found, true
	}
	return nil, false
}

func (op ObjectItem) GetProperty(name string) (*ObjectProperty, bool) {
	for _, prop := range op.Properties {
		if prop.ProtoFieldName == name {
			return prop, true
		}
	}
	return nil, false
}

func (op ObjectItem) jsonFieldMap(out map[string]json.RawMessage) error {
	if err := jsonFieldMapFromStructFields(op, out); err != nil {
		return err
	}

	properties := map[string]map[string]json.RawMessage{}
	required := []string{}
	for _, prop := range op.Properties {
		if prop.Skip {
			continue
		}

		propMap := map[string]json.RawMessage{}
		err := prop.jsonFieldMap(propMap)

		if err != nil {
			return fmt.Errorf("property %s: %w", prop.Name, err)
		}
		properties[prop.Name] = propMap

		if prop.Required {
			required = append(required, prop.Name)
		}
	}
	out["properties"], _ = json.Marshal(properties)

	if len(required) > 0 {
		out["required"], _ = json.Marshal(required)
	}

	if op.IsOneof {
		out["x-is-oneof"], _ = json.Marshal(true)
	}

	return nil
}

type ObjectProperty struct {
	SchemaItem       `json:"-"`
	Skip             bool   `json:"-"`
	Name             string `json:"-"`
	Required         bool   `json:"-"` // this bubbles up to the required array of the object
	ReadOnly         bool   `json:"readOnly,omitempty"`
	WriteOnly        bool   `json:"writeOnly,omitempty"`
	Description      string `json:"description,omitempty"`
	ProtoFieldName   string `json:"x-proto-name,omitempty"`
	ProtoFieldNumber int    `json:"x-proto-number,omitempty"`
	Optional         bool   `json:"x-proto-optional"` // The proto field is marked as optional, go code etc should use a pointer
}

func (op ObjectProperty) jsonFieldMap(out map[string]json.RawMessage) error {
	propOut, err := op.SchemaItem.fieldMap()
	if err != nil {
		return err
	}

	for k, v := range propOut {
		out[k] = v
	}

	if err := jsonFieldMapFromStructFields(op, out); err != nil {
		return err
	}

	return nil
}

type ObjectRules struct {
	MinProperties Optional[uint64] `json:"minProperties,omitempty"`
	MaxProperties Optional[uint64] `json:"maxProperties,omitempty"`
}
