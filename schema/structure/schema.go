package structure

import (
	"context"
	"fmt"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/google/uuid"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/log.go/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type SchemaSet struct {
	trimSubPackages []string
	Schemas         map[string]*schema_j5pb.Schema

	seen map[string]bool
}

func NewSchemaSet(options *config_j5pb.CodecOptions) *SchemaSet {
	return &SchemaSet{
		trimSubPackages: options.TrimSubPackages,

		Schemas: make(map[string]*schema_j5pb.Schema),
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
func (ss *SchemaSet) BuildSchemaObject(ctx context.Context, src protoreflect.MessageDescriptor) (*schema_j5pb.Schema, error) {
	goTypeName := walkName(src)

	properties := make([]*schema_j5pb.ObjectProperty, 0, src.Fields().Len())

	// Has no effect on the generated output, but validates that the oneof rules
	// are met:
	// Exactly one oneof which all fields belong to and is named 'type'

	isOneofWrapper := false
	options := proto.GetExtension(src.Options(), ext_j5pb.E_Message).(*ext_j5pb.MessageOptions)
	if options != nil {
		if options.IsOneofWrapper {
			isOneofWrapper = true
		}
	}

	exposeOneofs := make(map[string]*schema_j5pb.OneofWrapperItem)
	pendingOneofProps := make(map[string]*schema_j5pb.ObjectProperty)

	for idx := 0; idx < src.Oneofs().Len(); idx++ {
		oneof := src.Oneofs().Get(idx)
		if oneof.IsSynthetic() {
			continue
		}

		ext := proto.GetExtension(oneof.Options(), ext_j5pb.E_Oneof).(*ext_j5pb.OneofOptions)
		if ext == nil {
			if !isOneofWrapper {
				log.WithFields(ctx, map[string]interface{}{
					"message": src.FullName(),
					"oneof":   oneof.Name(),
				}).Warn("Unexposed Oneof")
			}
			continue
		} else if !ext.Expose {
			// By default, do not expose oneofs
			continue
		} else if isOneofWrapper {
			return nil, fmt.Errorf("oneof wrapper cannot contain exposed oneofs")
		}

		oneofName := string(oneof.Name())
		syntheticTypeName := fmt.Sprintf("%s_%s", src.Name(), oneofName)
		oneofObject := &schema_j5pb.OneofWrapperItem{
			ProtoFullName:    string(oneof.FullName()),
			ProtoMessageName: oneofName,
			GoTypeName:       syntheticTypeName,
			GoPackageName:    src.ParentFile().Options().(*descriptorpb.FileOptions).GetGoPackage(),
			GrpcPackageName:  string(src.ParentFile().Package()),
		}
		prop := &schema_j5pb.ObjectProperty{
			ProtoFieldName: string(oneof.Name()),
			Name:           camelCase(oneofName),
			Description:    commentDescription(src),
			Schema: &schema_j5pb.Schema{
				Type: &schema_j5pb.Schema_OneofWrapper{
					OneofWrapper: oneofObject,
				},
			},
		}
		pendingOneofProps[oneofName] = prop
		exposeOneofs[oneofName] = oneofObject

	}

	for ii := 0; ii < src.Fields().Len(); ii++ {
		field := src.Fields().Get(ii)

		if field.IsList() {
			prop, err := ss.buildSchemaProperty(ctx, field)
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}
			prop.Schema = &schema_j5pb.Schema{
				Type: &schema_j5pb.Schema_ArrayItem{
					ArrayItem: &schema_j5pb.ArrayItem{
						Items: prop.Schema,
					},
				},
			}
			properties = append(properties, prop)
			continue

		}
		if field.IsMap() {
			// TODO: Check that the map key is a string

			valueProp, err := ss.buildSchemaProperty(ctx, field.MapValue())
			if err != nil {
				return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
			}

			src := field
			prop := &schema_j5pb.ObjectProperty{
				ProtoFieldName:   string(src.Name()),
				ProtoFieldNumber: int32(src.Number()),
				Name:             string(src.JSONName()),
				Description:      commentDescription(src),
				Schema: &schema_j5pb.Schema{
					Type: &schema_j5pb.Schema_MapItem{
						MapItem: &schema_j5pb.MapItem{
							ItemSchema: valueProp.Schema,
						},
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
					subMessage, err := ss.BuildSchemaObject(ctx, field.Message())
					if err != nil {
						return nil, fmt.Errorf("building field %s: %w", field.FullName(), err)
					}
					// inline the properties of the sub-message directly into
					// this message
					properties = append(properties, subMessage.GetObjectItem().Properties...)
					continue
				}
			}
		}

		prop, err := ss.buildSchemaProperty(ctx, field)
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
		return nil, fmt.Errorf("oneof %s has not been added", pending.Name)
	}
	description := commentDescription(src)

	if isOneofWrapper {
		return &schema_j5pb.Schema{
			Description: description,
			Type: &schema_j5pb.Schema_OneofWrapper{
				OneofWrapper: &schema_j5pb.OneofWrapperItem{
					ProtoFullName:    string(src.FullName()),
					ProtoMessageName: string(src.Name()),
					GoPackageName:    src.ParentFile().Options().(*descriptorpb.FileOptions).GetGoPackage(),
					GrpcPackageName:  string(src.ParentFile().Package()),
					GoTypeName:       goTypeName,
					Properties:       properties,
				},
			},
		}, nil

	}

	return &schema_j5pb.Schema{
		Description: description,
		Type: &schema_j5pb.Schema_ObjectItem{
			ObjectItem: &schema_j5pb.ObjectItem{
				ProtoFullName:    string(src.FullName()),
				ProtoMessageName: string(src.Name()),
				GoPackageName:    src.ParentFile().Options().(*descriptorpb.FileOptions).GetGoPackage(),
				GrpcPackageName:  string(src.ParentFile().Package()),
				GoTypeName:       goTypeName,
				Properties:       properties,
			},
		},
	}, nil
}

func camelCase(s string) string {
	caser := cases.Title(language.English)
	s = caser.String(strings.ReplaceAll(s, "_", " "))
	return strings.ReplaceAll(strings.ToLower(s[:1])+s[1:], " ", "")
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

func (ss *SchemaSet) buildSchemaProperty(ctx context.Context, src protoreflect.FieldDescriptor) (*schema_j5pb.ObjectProperty, error) {
	prop := &schema_j5pb.ObjectProperty{
		ProtoFieldName:   string(src.Name()),
		ProtoFieldNumber: int32(src.Number()),
		Name:             string(src.JSONName()),
		Description:      commentDescription(src),
		Schema: &schema_j5pb.Schema{
			Description: commentDescription(src),
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
		prop.Schema.Type = &schema_j5pb.Schema_BooleanItem{
			BooleanItem: boolItem,
		}
		prop.Required = true

	case protoreflect.EnumKind:
		enumConstraint := constraint.GetEnum()
		values, err := EnumValues(src.Enum().Values(), enumConstraint)
		if err != nil {
			return nil, err
		}

		protoValues := make([]*schema_j5pb.EnumItem_Value, 0, len(values))
		for _, value := range values {
			protoValues = append(protoValues, &schema_j5pb.EnumItem_Value{
				Name:        value.Name,
				Description: value.Description,
			})
		}

		refSchemaItem := &schema_j5pb.Schema{
			Description: commentDescription(src),
			Type: &schema_j5pb.Schema_EnumItem{
				EnumItem: &schema_j5pb.EnumItem{
					Options: protoValues,
				},
			},
		}

		refName := string(src.Enum().FullName())
		ss.Schemas[refName] = refSchemaItem

		prop.Schema.Type = &schema_j5pb.Schema_Ref{
			Ref: refName,
		}

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
		prop.Schema.Type = &schema_j5pb.Schema_IntegerItem{
			IntegerItem: &schema_j5pb.IntegerItem{
				Format: "int32",
				Rules:  integerRules,
			},
		}

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

		prop.Schema.Type = &schema_j5pb.Schema_IntegerItem{
			IntegerItem: &schema_j5pb.IntegerItem{
				Format: "uint32",
				Rules:  integerRules,
			},
		}

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

		prop.Schema.Type = &schema_j5pb.Schema_IntegerItem{
			IntegerItem: &schema_j5pb.IntegerItem{
				Format: "int64",
				Rules:  integerRules,
			},
		}

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

		prop.Schema.Type = &schema_j5pb.Schema_NumberItem{
			NumberItem: &schema_j5pb.NumberItem{
				Format: "float",
				Rules:  numberRules,
			},
		}

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

		prop.Schema.Type = &schema_j5pb.Schema_NumberItem{
			NumberItem: &schema_j5pb.NumberItem{
				Format: "float",
				Rules:  numberRules,
			},
		}

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

		prop.Schema.Type = &schema_j5pb.Schema_StringItem{
			StringItem: stringItem,
		}

	case protoreflect.BytesKind:
		prop.Schema.Type = &schema_j5pb.Schema_StringItem{
			StringItem: &schema_j5pb.StringItem{
				Format: Ptr("byte"),
			},
		}

	case protoreflect.MessageKind:
		// When called from a field of a message, this creates a ref. When built directly from a service RPC request or create, this code is not called, they are inlined with the buildSchemaObject call directly
		if wktschema, ok := wktSchema(src.Message()); ok {
			prop.Schema = wktschema

		} else {
			prop.Schema.Type = &schema_j5pb.Schema_Ref{
				Ref: string(src.Message().FullName()),
			}
			if err := ss.addSchemaObject(ctx, src.Message()); err != nil {
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

func EnumValues(src protoreflect.EnumValueDescriptors, constraint *validate.EnumRules) ([]*schema_j5pb.EnumItem_Value, error) {
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

	values := make([]*schema_j5pb.EnumItem_Value, 0, src.Len())
	for ii := 0; ii < src.Len(); ii++ {
		option := src.Get(ii)
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

	unspecifiedVal := string(src.Get(0).Name())
	if !strings.HasSuffix(unspecifiedVal, suffix) {
		return nil, fmt.Errorf("enum does not have an unspecified value ending in %q", suffix)
	}
	trimPrefix := strings.TrimSuffix(unspecifiedVal, suffix)

	for ii := range values {
		values[ii].Name = strings.TrimPrefix(values[ii].Name, trimPrefix)
	}
	return values, nil
}

func Ptr[T any](val T) *T {
	return &val
}

func wktSchema(src protoreflect.MessageDescriptor) (*schema_j5pb.Schema, bool) {
	switch string(src.FullName()) {
	case "google.protobuf.Timestamp":
		return &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_StringItem{
				StringItem: &schema_j5pb.StringItem{
					Format: Ptr("date-time"),
				},
			},
		}, true

	case "google.protobuf.Duration":
		return &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_StringItem{
				StringItem: &schema_j5pb.StringItem{
					Format: Ptr("duration"),
				},
			},
		}, true

	case "google.protobuf.Struct":
		return &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_MapItem{
				MapItem: &schema_j5pb.MapItem{
					ItemSchema: &schema_j5pb.Schema{
						Type: &schema_j5pb.Schema_Any{
							Any: &schema_j5pb.AnySchemmaItem{},
						},
					},
				},
			},
		}, true

	case "google.protobuf.Any":
		return &schema_j5pb.Schema{
			Type: &schema_j5pb.Schema_Any{
				Any: &schema_j5pb.AnySchemmaItem{},
			},
		}, true
	}

	return nil, false
}

func (ss *SchemaSet) addSchemaObject(ctx context.Context, src protoreflect.MessageDescriptor) error {
	if _, ok := ss.Schemas[string(src.FullName())]; ok {
		return nil
	}

	if strings.HasPrefix(string(src.FullName()), "google.protobuf.") {
		return fmt.Errorf("unknown google.protobuf type %s", src.FullName())
	}

	// Prevents recursion errors
	ss.Schemas[string(src.FullName())] = &schema_j5pb.Schema{}

	schema, err := ss.BuildSchemaObject(ctx, src)
	if err != nil {
		return err
	}

	ss.Schemas[string(src.FullName())] = schema

	return nil
}
