package j5reflect

import (
	"fmt"
	"slices"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/patherr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func SchemaSetFromFiles(descFiles *protoregistry.Files, include func(protoreflect.FileDescriptor) bool) (*SchemaSet, error) {
	messages := make([]protoreflect.MessageDescriptor, 0)
	enums := make([]protoreflect.EnumDescriptor, 0)

	descFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if !include(file) {
			return true
		}
		fileMessages := file.Messages()
		for ii := 0; ii < fileMessages.Len(); ii++ {
			message := fileMessages.Get(ii)
			messages = append(messages, message)
		}

		fileEnums := file.Enums()
		for ii := 0; ii < fileEnums.Len(); ii++ {
			enum := fileEnums.Get(ii)
			enums = append(enums, enum)
		}
		return true
	})

	pkgSet := newSchemaSet()
	for _, message := range messages {
		_, err := pkgSet.messageSchema(message)
		if err != nil {
			return nil, fmt.Errorf("package from reflect: %w", err)
		}
	}

	for _, enum := range enums {
		ref, didExist := newRefPlaceholder(pkgSet, enum)
		if didExist {
			continue // was referenced by an earlier message
		}
		packageName, _ := splitDescriptorName(enum)
		pkg := pkgSet.Package(packageName)
		built, err := pkg.buildEnum(enum)
		if err != nil {
			return nil, err
		}
		ref.To = built
	}

	return pkgSet, nil
}

func (ps *SchemaSet) messageSchema(src protoreflect.MessageDescriptor) (RootSchema, error) {
	packageName, nameInPackage := splitDescriptorName(src)
	schemaPackage := ps.Package(packageName)
	if built, ok := schemaPackage.Schemas[nameInPackage]; ok {
		if built.To == nil {
			// When building from reflection, the 'to' should be linked by the
			// caller which created the ref.
			return nil, fmt.Errorf("unlinked ref: %s/%s", packageName, nameInPackage)
		}
		return built.To, nil
	}

	placeholder := &RefSchema{
		Package: schemaPackage,
		Schema:  nameInPackage,
	}
	schemaPackage.Schemas[nameInPackage] = placeholder

	isOneofWrapper := isOneofWrapper(src)
	var err error
	if isOneofWrapper {
		placeholder.To, err = schemaPackage.buildOneofSchema(src)
	} else {
		placeholder.To, err = schemaPackage.buildObjectSchema(src)
	}
	if err != nil {
		return nil, err
	}
	return placeholder.To, nil
}

func (pkg *Package) schemaRootFromProto(descriptor protoreflect.Descriptor) rootSchema {
	description := commentDescription(descriptor)
	_, nameInPackage := splitDescriptorName(descriptor)
	return rootSchema{
		name:        nameInPackage,
		pkg:         pkg,
		description: description,
	}
}

func newRefPlaceholder(ss RootSet, descriptor protoreflect.Descriptor) (*RefSchema, bool) {
	packageName, nameInPackage := splitDescriptorName(descriptor)
	return ss.refTo(packageName, nameInPackage)
}

func splitDescriptorName(descriptor protoreflect.Descriptor) (string, string) {
	path := []string{}
	current := descriptor
	for {
		path = append(path, string(current.Name()))
		parent := current.Parent()
		parentFile, ok := parent.(protoreflect.FileDescriptor)
		if ok {
			slices.Reverse(path)
			return string(parentFile.Package()), strings.Join(path, "_")

		}
		current = current.Parent()
	}
}

func jsonFieldName(s protoreflect.Name) string {
	return strcase.ToLowerCamel(string(s))
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
		// even if it is marked to expose, still don't automatically make it a
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

func (ss *Package) buildOneofSchema(srcMsg protoreflect.MessageDescriptor) (*OneofSchema, error) {
	properties, err := ss.messageProperties(srcMsg)
	if err != nil {
		return nil, fmt.Errorf("properties of %s: %w", srcMsg.FullName(), err)
	}
	return &OneofSchema{
		rootSchema: ss.schemaRootFromProto(srcMsg),
		Properties: properties,
		// TODO: Rules
	}, nil
}

func (ss *Package) buildObjectSchema(srcMsg protoreflect.MessageDescriptor) (*ObjectSchema, error) {
	properties, err := ss.messageProperties(srcMsg)
	if err != nil {
		return nil, fmt.Errorf("properties of %s: %w", srcMsg.FullName(), err)
	}
	objectSchema := &ObjectSchema{
		rootSchema: ss.schemaRootFromProto(srcMsg),
		Properties: properties,
		// TODO: Rules
	}

	if psmExt, ok := proto.GetExtension(srcMsg.Options(), ext_j5pb.E_Psm).(*ext_j5pb.PSMOptions); ok && psmExt != nil {
		var part schema_j5pb.EntityPart
		msgName := string(srcMsg.Name())
		if strings.HasSuffix(msgName, "Keys") {
			part = schema_j5pb.EntityPart_KEYS
		} else if strings.HasSuffix(msgName, "State") {
			part = schema_j5pb.EntityPart_STATE
		} else if strings.HasSuffix(msgName, "Event") {
			part = schema_j5pb.EntityPart_EVENT
		} else {
			return nil, fmt.Errorf("unknown PSM type suffix for %q", msgName)
		}

		objectSchema.Entity = &schema_j5pb.EntityObject{
			Entity: psmExt.EntityName,
			Part:   part,
		}

	}

	return objectSchema, nil

}

func (ss *Package) messageProperties(src protoreflect.MessageDescriptor) ([]*ObjectProperty, error) {

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
			rootSchema: ss.schemaRootFromProto(oneof),
			//oneofDescriptor: oneof,
		}
		refPlaceholder, didExist := newRefPlaceholder(ss.PackageSet, oneof)
		if didExist {
			return nil, fmt.Errorf("placeholder already exists for oneof wrapper %q", oneofName)
		}
		refPlaceholder.To = oneofObject
		prop := &ObjectProperty{
			JSONName:    jsonFieldName(oneof.Name()),
			Description: commentDescription(src),
			Schema: &OneofField{
				Ref: refPlaceholder,
				// TODO: Oneof Rules
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
				return nil, patherr.Wrap(err, string(field.Name()))
			}
			// TODO: Rules
			prop.Schema = &ArrayField{
				Schema: prop.Schema,
			}

			properties = append(properties, prop)
			continue

		}
		if field.IsMap() {
			// TODO: Check that the map key is a string

			valueProp, err := ss.buildSchemaProperty(field.MapValue())
			if err != nil {
				return nil, patherr.Wrap(err, string(field.Name()))
			}

			src := field
			prop := &ObjectProperty{
				ProtoField:  []protoreflect.FieldNumber{src.Number()},
				JSONName:    string(src.JSONName()),
				Description: commentDescription(src),
				Schema: &MapField{
					Schema: valueProp.Schema,
				},
			}
			properties = append(properties, prop)
			continue
		}

		fieldOptions := proto.GetExtension(field.Options(), ext_j5pb.E_Field).(*ext_j5pb.FieldOptions)
		if fieldOptions != nil {
			if msgOptions := fieldOptions.GetMessage(); msgOptions != nil {
				if field.Kind() != protoreflect.MessageKind {
					return nil, fmt.Errorf("field %s is not a message but has a message annotation", field.Name())
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

						property.ProtoField = append([]protoreflect.FieldNumber{field.Number()}, property.ProtoField...)
						properties = append(properties, property)
					}
					continue
				}
			}
		}

		prop, err := ss.buildSchemaProperty(field)
		if err != nil {
			return nil, patherr.Wrap(err, string(field.Name()))
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

var wellKnownStringPatterns = map[string]string{
	`^\d{4}-\d{2}-\d{2}$`: "date",
	`^\d(.?\d)?$`:         "number",
}

func (pkg *Package) buildSchemaProperty(src protoreflect.FieldDescriptor) (*ObjectProperty, error) {
	schemaProto := &schema_j5pb.Field{}

	prop := &ObjectProperty{
		JSONName:    string(src.JSONName()),
		ProtoField:  []protoreflect.FieldNumber{src.Number()},
		Description: commentDescription(src),
		Schema: &ScalarSchema{
			Proto: schemaProto,
			Kind:  src.Kind(),
		},
	}

	validateConstraint := proto.GetExtension(src.Options(), validate.E_Field).(*validate.FieldConstraints)
	if validateConstraint != nil && validateConstraint.Ignore != validate.Ignore_IGNORE_ALWAYS {
		if validateConstraint.Required {
			prop.Required = true
		}

		// constraint.IgnoreEmpty doesn't really apply

		// if the constraint is repeated, unwrap it
		repeatedConstraint, ok := validateConstraint.Type.(*validate.FieldConstraints_Repeated)
		if ok {
			validateConstraint = repeatedConstraint.Repeated.Items
		}
	}

	listConstraint := proto.GetExtension(src.Options(), list_j5pb.E_Field).(*list_j5pb.FieldConstraint)

	if !prop.Required && src.HasOptionalKeyword() {
		prop.ExplicitlyOptional = true
	}

	// TODO: Validation / Rules
	// TODO: Map
	// TODO: Extra types (see below)

	switch src.Kind() {
	case protoreflect.BoolKind:
		boolConstraint := validateConstraint.GetBool()
		boolItem := &schema_j5pb.BooleanField{}

		if boolConstraint != nil {
			if boolConstraint.Const != nil {
				boolItem.Rules.Const = boolConstraint.Const
			}
		}
		schemaProto.Type = &schema_j5pb.Field_Boolean{
			Boolean: boolItem,
		}
		prop.Required = true
		return prop, nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		var integerRules *schema_j5pb.IntegerField_Rules

		if constraint := validateConstraint.GetInt32(); constraint != nil {
			integerRules = &schema_j5pb.IntegerField_Rules{}
			if constraint.Const != nil {
				return nil, fmt.Errorf("'const' not supported")
			}
			if constraint.In != nil {
				return nil, fmt.Errorf("'in' not supported")
			}
			if constraint.NotIn != nil {
				return nil, fmt.Errorf("'notIn' not supported")
			}

			if constraint.LessThan != nil {
				switch cType := constraint.LessThan.(type) {
				case *validate.Int32Rules_Lt:
					integerRules.Maximum = Ptr(int64(cType.Lt))
					integerRules.ExclusiveMaximum = Ptr(true)
				case *validate.Int32Rules_Lte:
					integerRules.Maximum = Ptr(int64(cType.Lte))
				}
			}
			if constraint.GreaterThan != nil {
				switch cType := constraint.GreaterThan.(type) {
				case *validate.Int32Rules_Gt:
					integerRules.Minimum = Ptr(int64(cType.Gt))
					integerRules.ExclusiveMinimum = Ptr(true)
				case *validate.Int32Rules_Gte:
					integerRules.Minimum = Ptr(int64(cType.Gte))
				}
			}
		}

		schemaProto.Type = &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_INT32,
				Rules:     integerRules,
				ListRules: listConstraint.GetInt32(),
			},
		}
		return prop, nil

	case protoreflect.Uint32Kind:
		var integerRules *schema_j5pb.IntegerField_Rules
		uint32Constraint := validateConstraint.GetUint32()
		if uint32Constraint != nil {
			integerRules = &schema_j5pb.IntegerField_Rules{}
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

		schemaProto.Type = &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_UINT32,
				Rules:     integerRules,
				ListRules: listConstraint.GetUint32(),
			},
		}
		return prop, nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
		var integerRules *schema_j5pb.IntegerField_Rules
		int64Constraint := validateConstraint.GetInt64()
		if int64Constraint != nil {
			integerRules = &schema_j5pb.IntegerField_Rules{}
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

		schemaProto.Type = &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_INT64,
				Rules:     integerRules,
				ListRules: listConstraint.GetInt64(),
			},
		}
		return prop, nil

	case protoreflect.FloatKind:
		var numberRules *schema_j5pb.FloatField_Rules
		floatConstraint := validateConstraint.GetFloat()
		if floatConstraint != nil {
			numberRules = &schema_j5pb.FloatField_Rules{}
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

		schemaProto.Type = &schema_j5pb.Field_Float{
			Float: &schema_j5pb.FloatField{
				Format:    schema_j5pb.FloatField_FORMAT_FLOAT32,
				Rules:     numberRules,
				ListRules: listConstraint.GetFloat(),
			},
		}
		return prop, nil

	case protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		var numberRules *schema_j5pb.FloatField_Rules
		floatConstraint := validateConstraint.GetDouble()
		if floatConstraint != nil {
			numberRules = &schema_j5pb.FloatField_Rules{}
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

		schemaProto.Type = &schema_j5pb.Field_Float{
			Float: &schema_j5pb.FloatField{
				Format:    schema_j5pb.FloatField_FORMAT_FLOAT64,
				Rules:     numberRules,
				ListRules: listConstraint.GetDouble(),
			},
		}
		return prop, nil

	case protoreflect.StringKind:
		stringItem := &schema_j5pb.StringField{}

		var typeHint string

		if validateConstraint != nil && validateConstraint.Type != nil {
			stringConstraint, ok := validateConstraint.Type.(*validate.FieldConstraints_String_)
			if !ok {
				return nil, fmt.Errorf("wrong constraint type for string: %T", validateConstraint.Type)
			}

			stringItem.Rules = &schema_j5pb.StringField_Rules{}
			constraint := stringConstraint.String_

			stringItem.Rules.MinLength = constraint.MinLen

			stringItem.Rules.MaxLength = constraint.MaxLen

			if constraint.Pattern != nil {
				pattern := *constraint.Pattern
				wellKnownStringPattern, ok := wellKnownStringPatterns[pattern]
				if ok {
					stringItem.Format = Ptr(wellKnownStringPattern)
				} else {
					stringItem.Rules.Pattern = Ptr(pattern)
				}
			}

			switch wkt := constraint.WellKnown.(type) {
			case *validate.StringRules_Uuid:
				if wkt.Uuid {
					stringItem.Format = Ptr("uuid")
					typeHint = "uuid"
				}
			case *validate.StringRules_Email:
				if wkt.Email {
					stringItem.Format = Ptr("email")
				}

			case *validate.StringRules_Hostname:
				if wkt.Hostname {
					stringItem.Format = Ptr("hostname")
				}

			case *validate.StringRules_Ipv4:
				if wkt.Ipv4 {
					stringItem.Format = Ptr("ipv4")
				}

			case *validate.StringRules_Ipv6:
				if wkt.Ipv6 {
					stringItem.Format = Ptr("ipv6")
				}

			case *validate.StringRules_Uri:
				if wkt.Uri {
					stringItem.Format = Ptr("uri")
				}

			// Other types not supported by swagger
			case nil:

			default:
				return nil, fmt.Errorf("unknown string constraint: %T", constraint.WellKnown)

			}
		}

		listRules := listConstraint.GetString_()

		var fkRules *list_j5pb.KeyRules

		if fk := listRules.GetForeignKey(); fk != nil {
			switch fkt := fk.Type.(type) {
			case *list_j5pb.ForeignKeyRules_UniqueString:
				fkRules = fkt.UniqueString
				if typeHint == "" {
					typeHint = "natural_key"
				} else if typeHint != "natural_key" {
					return nil, fmt.Errorf("rules (%s) and list constraints (natural_key) do not match", typeHint)
				}

			case *list_j5pb.ForeignKeyRules_Uuid:
				fkRules = fkt.Uuid
				if typeHint == "" {
					typeHint = "uuid"
				} else if typeHint != "uuid" {
					return nil, fmt.Errorf("rules (%s) and list constraints (uuid) do not match", typeHint)
				}
			}
		}

		if openText := listRules.GetOpenText(); openText != nil {
			if typeHint != "" {
				return nil, fmt.Errorf("open_text and rules (%s) do not match", typeHint)
			}
			typeHint = "open_text"
			stringItem.ListRules = openText
		}

		isPrimary := false

		if psmKeyExt, ok := proto.GetExtension(src.Options(), ext_j5pb.E_Key).(*ext_j5pb.KeyFieldOptions); ok && psmKeyExt != nil {
			if psmKeyExt.PrimaryKey {
				if typeHint == "" {
					typeHint = "uuid"
				} else if typeHint != "uuid" {
					return nil, fmt.Errorf("rules (%s) and key constraint do not match", typeHint)
				}

				isPrimary = true
			}

		}

		switch typeHint {
		case "uuid":
			schemaProto.Type = &schema_j5pb.Field_Key{
				Key: &schema_j5pb.KeyField{
					Format:    schema_j5pb.KeyFormat_UUID,
					ListRules: fkRules,
					Primary:   isPrimary,
				},
			}

		case "natural_key":
			schemaProto.Type = &schema_j5pb.Field_Key{
				Key: &schema_j5pb.KeyField{
					Format:    schema_j5pb.KeyFormat_UNSPECIFIED,
					ListRules: fkRules,
				},
			}

		default:

			schemaProto.Type = &schema_j5pb.Field_String_{
				String_: stringItem,
			}
		}
		return prop, nil

	case protoreflect.BytesKind:
		schemaProto.Type = &schema_j5pb.Field_Bytes{
			Bytes: &schema_j5pb.BytesField{
				Rules: &schema_j5pb.BytesField_Rules{},
			},
		}
		return prop, nil

	case protoreflect.EnumKind:
		ref, didExist := newRefPlaceholder(pkg.PackageSet, src.Enum())
		if !didExist {
			built, err := pkg.buildEnum(src.Enum())
			if err != nil {
				return nil, err
			}
			ref.To = built
		}

		var rules *schema_j5pb.EnumField_Rules
		if vc := validateConstraint.GetEnum(); vc != nil {
			enumSchema := ref.To.(*EnumSchema)
			rules = &schema_j5pb.EnumField_Rules{}
			if vc.In != nil {
				for _, num := range vc.In {
					opt := enumSchema.OptionByNumber(num)
					if opt == nil {
						return nil, fmt.Errorf("enum value %d not found", num)
					}
					rules.In = append(rules.In, opt.Name)
				}
			}
			if vc.NotIn != nil {
				for _, num := range vc.NotIn {
					opt := enumSchema.OptionByNumber(num)
					if opt == nil {
						if num == 0 {
							continue // _UNSPECIFIED is being excluded already
						}
						return nil, fmt.Errorf("enum value %d not found", num)
					}
					rules.NotIn = append(rules.NotIn, opt.Name)
				}

			}
		}
		prop.Schema = &EnumField{
			Ref:       ref,
			ListRules: listConstraint.GetEnum(),
			Rules:     rules,
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
		msg := src.Message()
		isOneofWrapper := isOneofWrapper(msg)

		ref, didExist := newRefPlaceholder(pkg.PackageSet, msg)
		if !didExist {
			var err error
			if isOneofWrapper {
				ref.To, err = pkg.buildOneofSchema(msg)
			} else {
				ref.To, err = pkg.buildObjectSchema(msg)
			}
			if err != nil {
				return nil, err
			}
		}
		if isOneofWrapper {
			prop.Schema = &OneofField{
				Ref: ref,
			}
		} else {
			prop.Schema = &ObjectField{
				Ref: ref,
			}
		}
		return prop, nil
	default:
		/* Not Supported in J5:
		Sfixed32Kind Kind = 15
		Fixed32Kind  Kind = 7
		Sfixed64Kind Kind = 16
		Fixed64Kind  Kind = 6
		GroupKind    Kind = 10
		*/
		return nil, fmt.Errorf("unsupported field type %s", src.Kind())
	}

}

func (pkg *Package) buildEnum(enumDescriptor protoreflect.EnumDescriptor) (*EnumSchema, error) {

	ext := proto.GetExtension(enumDescriptor.Options(), ext_j5pb.E_Enum).(*ext_j5pb.EnumOptions)

	sourceValues := enumDescriptor.Values()
	values := make([]*schema_j5pb.Enum_Value, 0, sourceValues.Len())
	for ii := 0; ii < sourceValues.Len(); ii++ {
		option := sourceValues.Get(ii)
		number := int32(option.Number())

		values = append(values, &schema_j5pb.Enum_Value{
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

	if ext != nil && ext.NoDefault {
		values = values[1:]
	}
	return &EnumSchema{
		rootSchema: pkg.schemaRootFromProto(enumDescriptor),
		Options:    values,
		NamePrefix: trimPrefix,
		//descriptor:  enumDescriptor,
	}, nil
}

func Ptr[T any](val T) *T {
	return &val
}

func wktSchema(src protoreflect.MessageDescriptor) (FieldSchema, bool) {
	switch string(src.FullName()) {
	case "google.protobuf.Timestamp":
		return &ScalarSchema{
			WellKnownTypeName: src.FullName(),
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_String_{
					String_: &schema_j5pb.StringField{
						Format: Ptr("date-time"),
					},
				},
			},
		}, true

	case "google.protobuf.Duration":
		return &ScalarSchema{
			WellKnownTypeName: src.FullName(),
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_String_{
					String_: &schema_j5pb.StringField{
						Format: Ptr("duration"),
					},
				},
			},
		}, true

	case "j5.types.date.v1.Date":
		return &ScalarSchema{
			WellKnownTypeName: src.FullName(),
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_String_{
					String_: &schema_j5pb.StringField{
						Format: Ptr("date"),
					},
				},
			},
		}, true

	case "google.protobuf.Struct":
		return &MapField{
			Schema: &AnyField{},
		}, true

	case "google.protobuf.Any":
		return &AnyField{}, true
	}

	return nil, false
}
