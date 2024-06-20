package j5reflect

import (
	"fmt"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/iancoleman/strcase"

	"github.com/google/uuid"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DescriptorResolver interface {
	FindDescriptorByName(protoreflect.FullName) (protoreflect.Descriptor, error)
}

type SchemaResolver struct {
	*SchemaSet
	resolver DescriptorResolver
}

func NewSchemaResolver(resolver DescriptorResolver) *SchemaResolver {
	return &SchemaResolver{
		SchemaSet: &SchemaSet{
			schemas: make(map[protoreflect.FullName]*Schema),
		},
		resolver: resolver,
	}
}

func (ss *SchemaResolver) SchemaByName(name protoreflect.FullName) (*Schema, error) {
	obj, ok := ss.schemas[name]
	if ok {
		return obj, nil
	}
	descriptor, err := ss.resolver.FindDescriptorByName(name)
	if err != nil {
		return nil, err
	}
	msg, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("descriptor %s is not a message", name)
	}
	return ss.SchemaReflect(msg)
}

type SchemaSet struct {
	schemas map[protoreflect.FullName]*Schema
}

func NewSchemaSet() *SchemaSet {
	return &SchemaSet{
		schemas: make(map[protoreflect.FullName]*Schema),
	}
}

func (ss *SchemaSet) SchemaObject(src protoreflect.MessageDescriptor) (*schema_j5pb.Schema, error) {
	val, err := ss.SchemaReflect(src)
	if err != nil {
		return nil, err
	}

	return val.ToJ5Proto()
}

func (ss *SchemaSet) SchemaReflect(src protoreflect.MessageDescriptor) (*Schema, error) {
	name := src.FullName()
	if built, ok := ss.schemas[name]; ok {
		return built, nil
	}

	resultSchema := &Schema{}
	ss.schemas[name] = resultSchema
	err := ss.buildMessageSchema(src, resultSchema)
	if err != nil {
		return nil, err
	}

	return resultSchema, nil
}

func isOneofWrapper(src protoreflect.MessageDescriptor) bool {
	options := proto.GetExtension(src.Options(), ext_j5pb.E_Message).(*ext_j5pb.MessageOptions)
	if options != nil {
		if options.IsOneofWrapper {
			return true
		}
		// TODO: Allow explicit false, allowing overriding the auto function
	}

	oneofs := src.Oneofs()

	if oneofs.Len() != 1 {
		return false
	}

	oneof := oneofs.Get(0)
	if oneof.IsSynthetic() {
		return false
	}

	if oneof.Name() != "type" {
		return false
	}

	ext := proto.GetExtension(oneof.Options(), ext_j5pb.E_Oneof).(*ext_j5pb.OneofOptions)
	if ext != nil {
		// even if it is mared to expose, still don't automatically make it a
		// oneof.
		return false
	}

	for ii := 0; ii < src.Fields().Len(); ii++ {
		field := src.Fields().Get(ii)
		if field.ContainingOneof() != oneof {
			return false
		}
		if field.Kind() != protoreflect.MessageKind {
			return false
		}
	}

	return true

}
func (ss *SchemaSet) buildMessageSchema(srcMsg protoreflect.MessageDescriptor, into *Schema) error {
	properties, err := ss.messageProperties(srcMsg)
	if err != nil {
		return err
	}
	isOneofWrapper := isOneofWrapper(srcMsg)
	description := commentDescription(srcMsg)

	if isOneofWrapper {
		into.oneofItem = &OneofSchema{
			Description:     description,
			Properties:      properties,
			ProtoMessage:    srcMsg,
			OneofDescriptor: srcMsg.Oneofs().Get(0),
			// TODO: Rules
		}
	} else {
		into.objectItem = &ObjectSchema{
			Description: description,
			Properties:  properties,
			// TODO: Rules
			ProtoMessage: srcMsg,
		}
	}

	return nil
}

func (ss *SchemaSet) messageProperties(src protoreflect.MessageDescriptor) ([]*ObjectProperty, error) {

	properties := make([]*ObjectProperty, 0, src.Fields().Len())

	exposeOneofs := make(map[string]*OneofSchema)
	pendingOneofProps := make(map[string]*ObjectProperty)

	for idx := 0; idx < src.Oneofs().Len(); idx++ {
		oneof := src.Oneofs().Get(idx)
		if oneof.IsSynthetic() {
			continue
		}

		ext := proto.GetExtension(oneof.Options(), ext_j5pb.E_Oneof).(*ext_j5pb.OneofOptions)
		if ext == nil {
			continue
		} else if !ext.Expose {
			// By default, do not expose oneofs
			continue
		}

		oneofName := string(oneof.Name())
		oneofObject := &OneofSchema{
			OneofDescriptor: oneof,
		}
		prop := &ObjectProperty{
			JSONName:    jsonFieldName(oneofName),
			Description: commentDescription(src),
			Schema: &Schema{
				oneofItem: oneofObject,
			},
		}
		pendingOneofProps[oneofName] = prop
		exposeOneofs[oneofName] = oneofObject

	}

	for ii := 0; ii < src.Fields().Len(); ii++ {
		field := src.Fields().Get(ii)

		if field.IsList() {
			prop, err := ss.buildSchemaProperty(field)
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}
			// TODO: Rules
			prop.Schema = &Schema{
				arrayItem: &ArraySchema{
					Schema: prop.Schema,
				},
			}

			properties = append(properties, prop)
			continue

		}
		if field.IsMap() {
			// TODO: Check that the map key is a string

			valueProp, err := ss.buildSchemaProperty(field.MapValue())
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}

			src := field
			prop := &ObjectProperty{
				ProtoField:  []protoreflect.FieldDescriptor{src},
				JSONName:    string(src.JSONName()),
				Description: commentDescription(src),
				Schema: &Schema{
					mapItem: &MapSchema{
						Schema: valueProp.Schema,
					},
				},
			}
			properties = append(properties, prop)
			continue
		}

		fieldOptions := proto.GetExtension(field.Options(), ext_j5pb.E_Field).(*ext_j5pb.FieldOptions)
		if fieldOptions != nil {
			if msgOptions := fieldOptions.GetMessage(); msgOptions != nil {
				if field.Kind() != protoreflect.MessageKind {
					return nil, fmt.Errorf("field %s is not a message but has a message annotation", field.FullName())
				}

				if msgOptions.Flatten {
					// skips the schema set, this is used just to get the
					// fields.
					subMessage, err := ss.messageProperties(field.Message())
					if err != nil {
						return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
					}
					// inline the properties of the sub-message directly into
					// this message
					for _, property := range subMessage {

						property.ProtoField = append([]protoreflect.FieldDescriptor{field}, property.ProtoField...)
						properties = append(properties, property)
					}
					continue
				}
			}
		}

		prop, err := ss.buildSchemaProperty(field)
		if err != nil {
			return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
		}

		inOneof := field.ContainingOneof()
		if inOneof == nil || inOneof.IsSynthetic() {
			properties = append(properties, prop)
			continue
		}

		name := string(inOneof.Name())

		oneof, ok := exposeOneofs[name]
		if !ok {
			properties = append(properties, prop)
			continue
		}

		oneof.Properties = append(oneof.Properties, prop)

		// deferrs adding the oneof to the property array until the first
		// field is encountered, i.e. preserves ordering
		pending, ok := pendingOneofProps[name]
		if ok {
			properties = append(properties, pending)
			delete(pendingOneofProps, name)
		}
	}

	for _, pending := range pendingOneofProps {
		return nil, fmt.Errorf("oneof %s has not been added", pending.JSONName)
	}

	return properties, nil

}

func jsonFieldName(s string) string {
	return strcase.ToLowerCamel(s)
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

var lastUUID uuid.UUID

func quickUUID() string {
	if lastUUID == uuid.Nil {
		lastUUID = uuid.New()
		return lastUUID.String()
	}
	lastUUID[0]++
	lastUUID[2]++
	lastUUID[4]++
	lastUUID[6] = (lastUUID[6] & 0x0f) | 0x40 // Version 4
	lastUUID[8] = (lastUUID[8] & 0x3f) | 0x80 // Variant is 10
	return lastUUID.String()
}

func (ss *SchemaSet) buildSchemaProperty(src protoreflect.FieldDescriptor) (*ObjectProperty, error) {
	schemaProto := &schema_j5pb.Schema{
		Description: commentDescription(src),
	}

	prop := &ObjectProperty{
		JSONName:    string(src.JSONName()),
		ProtoField:  []protoreflect.FieldDescriptor{src},
		Description: commentDescription(src),
		Schema: &Schema{
			scalarItem: &ScalarSchema{
				proto: schemaProto,
				Kind:  src.Kind(),
			},
		},
	}

	// second _ prevents a panic when the exception is not set
	constraint, _ := proto.GetExtension(src.Options(), validate.E_Field).(*validate.FieldConstraints)

	if constraint != nil && constraint.Ignore != validate.Ignore_IGNORE_ALWAYS {
		if constraint.Required {
			prop.Required = true
		}

		// constraint.IgnoreEmpty doesn't really apply

		// if the constraint is repeated, unwrap it
		repeatedConstraint, ok := constraint.Type.(*validate.FieldConstraints_Repeated)
		if ok {
			constraint = repeatedConstraint.Repeated.Items
		}
	}

	if !prop.Required && src.HasOptionalKeyword() {
		prop.ExplicitlyOptional = true
	}

	// TODO: Validation / Rules
	// TODO: Map
	// TODO: Extra types (see below)

	switch src.Kind() {
	case protoreflect.BoolKind:
		boolConstraint := constraint.GetBool()
		boolItem := &schema_j5pb.BooleanItem{}

		if boolConstraint != nil {
			if boolConstraint.Const != nil {
				boolItem.Rules.Const = boolConstraint.Const
			}
		}
		schemaProto.Type = &schema_j5pb.Schema_BooleanItem{
			BooleanItem: boolItem,
		}
		prop.Required = true
		return prop, nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		var integerRules *schema_j5pb.IntegerRules
		int32Constraint := constraint.GetInt32()
		if int32Constraint != nil {
			integerRules = &schema_j5pb.IntegerRules{}
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
					integerRules.Maximum = Ptr(int64(cType.Lt))
					integerRules.ExclusiveMaximum = Ptr(true)
				case *validate.Int32Rules_Lte:
					integerRules.Maximum = Ptr(int64(cType.Lte))
				}
			}
			if int32Constraint.GreaterThan != nil {
				switch cType := int32Constraint.GreaterThan.(type) {
				case *validate.Int32Rules_Gt:
					integerRules.Minimum = Ptr(int64(cType.Gt))
					integerRules.ExclusiveMinimum = Ptr(true)
				case *validate.Int32Rules_Gte:
					integerRules.Minimum = Ptr(int64(cType.Gte))
				}
			}

		}
		schemaProto.Type = &schema_j5pb.Schema_IntegerItem{
			IntegerItem: &schema_j5pb.IntegerItem{
				Format: "int32",
				Rules:  integerRules,
			},
		}
		return prop, nil

	case protoreflect.Uint32Kind:
		var integerRules *schema_j5pb.IntegerRules
		uint32Constraint := constraint.GetUint32()
		if uint32Constraint != nil {
			integerRules = &schema_j5pb.IntegerRules{}
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
					integerRules.Maximum = Ptr(int64(cType.Lt))
					integerRules.ExclusiveMaximum = Ptr(true)
				case *validate.UInt32Rules_Lte:
					integerRules.Maximum = Ptr(int64(cType.Lte))
				}
			}
			if uint32Constraint.GreaterThan != nil {
				switch cType := uint32Constraint.GreaterThan.(type) {
				case *validate.UInt32Rules_Gt:
					integerRules.Minimum = Ptr(int64(cType.Gt))
					integerRules.ExclusiveMinimum = Ptr(true)
				case *validate.UInt32Rules_Gte:
					integerRules.Minimum = Ptr(int64(cType.Gte))
				}
			}
		}

		schemaProto.Type = &schema_j5pb.Schema_IntegerItem{
			IntegerItem: &schema_j5pb.IntegerItem{
				Format: "uint32",
				Rules:  integerRules,
			},
		}
		return prop, nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
		var integerRules *schema_j5pb.IntegerRules
		int64Constraint := constraint.GetInt64()
		if int64Constraint != nil {
			integerRules = &schema_j5pb.IntegerRules{}
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
					integerRules.Maximum = Ptr(cType.Lt)
					integerRules.ExclusiveMaximum = Ptr(true)
				case *validate.Int64Rules_Lte:
					integerRules.Maximum = Ptr(cType.Lte)
				}
			}
			if int64Constraint.GreaterThan != nil {
				switch cType := int64Constraint.GreaterThan.(type) {
				case *validate.Int64Rules_Gt:
					integerRules.Minimum = Ptr(cType.Gt)
					integerRules.ExclusiveMinimum = Ptr(true)
				case *validate.Int64Rules_Gte:
					integerRules.Minimum = Ptr(cType.Gte)
				}
			}
		}

		schemaProto.Type = &schema_j5pb.Schema_IntegerItem{
			IntegerItem: &schema_j5pb.IntegerItem{
				Format: "int64",
				Rules:  integerRules,
			},
		}
		return prop, nil

	case protoreflect.FloatKind:
		var numberRules *schema_j5pb.NumberRules
		floatConstraint := constraint.GetFloat()
		if floatConstraint != nil {
			numberRules = &schema_j5pb.NumberRules{}
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
					numberRules.Maximum = Ptr(float64(cType.Lt))
					numberRules.ExclusiveMaximum = Ptr(true)
				case *validate.FloatRules_Lte:
					numberRules.Maximum = Ptr(float64(cType.Lte))
				}
			}
			if floatConstraint.GreaterThan != nil {
				switch cType := floatConstraint.GreaterThan.(type) {
				case *validate.FloatRules_Gt:
					numberRules.Minimum = Ptr(float64(cType.Gt))
					numberRules.ExclusiveMinimum = Ptr(true)
				case *validate.FloatRules_Gte:
					numberRules.Minimum = Ptr(float64(cType.Gte))
				}
			}
		}

		schemaProto.Type = &schema_j5pb.Schema_NumberItem{
			NumberItem: &schema_j5pb.NumberItem{
				Format: "float",
				Rules:  numberRules,
			},
		}
		return prop, nil

	case protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		var numberRules *schema_j5pb.NumberRules
		floatConstraint := constraint.GetDouble()
		if floatConstraint != nil {
			numberRules = &schema_j5pb.NumberRules{}
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
					numberRules.Maximum = Ptr(float64(cType.Lt))
					numberRules.ExclusiveMaximum = Ptr(true)
				case *validate.DoubleRules_Lte:
					numberRules.Maximum = Ptr(float64(cType.Lte))
				}
			}
			if floatConstraint.GreaterThan != nil {
				switch cType := floatConstraint.GreaterThan.(type) {
				case *validate.DoubleRules_Gt:
					numberRules.Minimum = Ptr(float64(cType.Gt))
					numberRules.ExclusiveMinimum = Ptr(true)
				case *validate.DoubleRules_Gte:
					numberRules.Minimum = Ptr(float64(cType.Gte))
				}
			}
		}

		schemaProto.Type = &schema_j5pb.Schema_NumberItem{
			NumberItem: &schema_j5pb.NumberItem{
				Format: "float",
				Rules:  numberRules,
			},
		}
		return prop, nil

	case protoreflect.StringKind:
		stringItem := &schema_j5pb.StringItem{}
		if constraint != nil && constraint.Type != nil {
			stringConstraint, ok := constraint.Type.(*validate.FieldConstraints_String_)
			if !ok {
				return nil, fmt.Errorf("wrong constraint type for string: %T", constraint.Type)
			}

			stringItem.Rules = &schema_j5pb.StringRules{}
			constraint := stringConstraint.String_

			stringItem.Rules.MinLength = constraint.MinLen

			stringItem.Rules.MaxLength = constraint.MaxLen

			if constraint.Pattern != nil {
				pattern := *constraint.Pattern
				wellKnownStringPattern, ok := wellKnownStringPatterns[pattern]
				if ok {
					stringItem.Format = Ptr(wellKnownStringPattern.format)
					stringItem.Example = Ptr(wellKnownStringPattern.example)
				} else {
					stringItem.Rules.Pattern = Ptr(pattern)
				}
			}

			switch wkt := constraint.WellKnown.(type) {
			case *validate.StringRules_Uuid:
				if wkt.Uuid {
					stringItem.Format = Ptr("uuid")
					stringItem.Example = Ptr(quickUUID())
				}
			case *validate.StringRules_Email:
				if wkt.Email {
					stringItem.Format = Ptr("email")
					stringItem.Example = Ptr("test@example.com")
				}

			case *validate.StringRules_Hostname:
				if wkt.Hostname {
					stringItem.Format = Ptr("hostname")
					stringItem.Example = Ptr("example.com")
				}

			case *validate.StringRules_Ipv4:
				if wkt.Ipv4 {
					stringItem.Format = Ptr("ipv4")
					stringItem.Example = Ptr("10.10.10.10")
				}

			case *validate.StringRules_Ipv6:
				if wkt.Ipv6 {
					stringItem.Format = Ptr("ipv6")
					stringItem.Example = Ptr("2001:db8::68")
				}

			case *validate.StringRules_Uri:
				if wkt.Uri {
					stringItem.Format = Ptr("uri")
					stringItem.Example = Ptr("https://example.com")
				}

			// Other types not supported by swagger
			case nil:

			default:
				return nil, fmt.Errorf("unknown string constraint: %T", constraint.WellKnown)

			}

		}

		schemaProto.Type = &schema_j5pb.Schema_StringItem{
			StringItem: stringItem,
		}
		return prop, nil

	case protoreflect.BytesKind:
		schemaProto.Type = &schema_j5pb.Schema_StringItem{
			StringItem: &schema_j5pb.StringItem{
				Format: Ptr("byte"),
			},
		}
		return prop, nil

	case protoreflect.EnumKind:

		enumConstraint := constraint.GetEnum()
		enumSchemaItem, err := buildEnum(src.Enum(), enumConstraint)
		if err != nil {
			return nil, err
		}
		refName := src.Enum().FullName()
		ss.schemas[refName] = enumSchemaItem

		prop.Schema.scalarItem = nil
		prop.Schema.refItem = &RefSchema{
			Name: refName,
			To:   enumSchemaItem,
		}
		return prop, nil

	case protoreflect.MessageKind:
		wktschema, ok := wktSchema(src.Message())
		if ok {
			prop.Schema = wktschema
			return prop, nil
		}
		if strings.HasPrefix(string(src.Message().FullName()), "google.protobuf.") {
			return nil, fmt.Errorf("unsupported google type %s", src.Message().FullName())
		}

		msgName := src.Message().FullName()
		schema, ok := ss.schemas[msgName]
		if !ok {
			schema = &Schema{}
			ss.schemas[msgName] = schema
			if err := ss.buildMessageSchema(src.Message(), schema); err != nil {
				return nil, err
			}
		}
		prop.Schema.scalarItem = nil
		prop.Schema.refItem = &RefSchema{
			Name: msgName,
			To:   schema,
		}

		return prop, nil
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

}

func buildEnum(enumDescriptor protoreflect.EnumDescriptor, constraint *validate.EnumRules) (*Schema, error) {

	specMap := map[int32]struct{}{}
	var notIn bool
	var isIn bool

	if constraint != nil {
		if constraint.NotIn != nil {
			for _, notIn := range constraint.NotIn {
				specMap[notIn] = struct{}{}
			}
			notIn = true

		} else if constraint.In != nil {
			for _, in := range constraint.In {
				specMap[in] = struct{}{}
			}
			isIn = true
		}
	}

	if notIn && isIn {
		return nil, fmt.Errorf("enum cannot have both in and not_in constraints")
	}

	sourceValues := enumDescriptor.Values()
	values := make([]*schema_j5pb.EnumItem_Value, 0, sourceValues.Len())
	for ii := 0; ii < sourceValues.Len(); ii++ {
		option := sourceValues.Get(ii)
		number := int32(option.Number())

		if notIn {
			_, exclude := specMap[number]
			if exclude {
				continue
			}
		} else if isIn {
			_, include := specMap[number]
			if !include {
				continue
			}
		}

		values = append(values, &schema_j5pb.EnumItem_Value{
			Name:        string(option.Name()),
			Number:      number,
			Description: commentDescription(option),
		})
	}

	suffix := "UNSPECIFIED"

	unspecifiedVal := string(sourceValues.Get(0).Name())
	if !strings.HasSuffix(unspecifiedVal, suffix) {
		return nil, fmt.Errorf("enum does not have an unspecified value ending in %q", suffix)
	}
	trimPrefix := strings.TrimSuffix(unspecifiedVal, suffix)

	for ii := range values {
		values[ii].Name = strings.TrimPrefix(values[ii].Name, trimPrefix)
	}

	enumSchemaItem := &Schema{
		enumItem: &EnumSchema{
			Description: commentDescription(enumDescriptor),
			Options:     values,
			NamePrefix:  trimPrefix,
			Descriptor:  enumDescriptor,
		},
	}
	return enumSchemaItem, nil
}

func Ptr[T any](val T) *T {
	return &val
}

func wktSchema(src protoreflect.MessageDescriptor) (*Schema, bool) {
	switch string(src.FullName()) {
	case "google.protobuf.Timestamp":
		return &Schema{
			scalarItem: &ScalarSchema{
				WellKnownTypeName: src.FullName(),
				proto: &schema_j5pb.Schema{
					Type: &schema_j5pb.Schema_StringItem{
						StringItem: &schema_j5pb.StringItem{
							Format: Ptr("date-time"),
						},
					},
				},
			},
		}, true

	case "google.protobuf.Duration":
		return &Schema{
			scalarItem: &ScalarSchema{
				WellKnownTypeName: src.FullName(),
				proto: &schema_j5pb.Schema{
					Type: &schema_j5pb.Schema_StringItem{
						StringItem: &schema_j5pb.StringItem{
							Format: Ptr("duration"),
						},
					},
				},
			},
		}, true

	case "j5.types.date.v1.Date":
		return &Schema{
			scalarItem: &ScalarSchema{
				WellKnownTypeName: src.FullName(),
				proto: &schema_j5pb.Schema{
					Type: &schema_j5pb.Schema_StringItem{
						StringItem: &schema_j5pb.StringItem{
							Format: Ptr("date"),
						},
					},
				},
			},
		}, true

	case "google.protobuf.Struct":
		return &Schema{
			mapItem: &MapSchema{
				Schema: &Schema{
					anyItem: &AnySchema{},
				},
			},
		}, true

	case "google.protobuf.Any":
		return &Schema{
			anyItem: &AnySchema{},
		}, true
	}

	return nil, false
}
