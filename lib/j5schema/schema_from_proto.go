package j5schema

import (
	"fmt"
	"slices"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/protosrc"
	"github.com/pentops/j5/lib/id62"
	"github.com/pentops/j5/lib/patherr"
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

		if err := ref.check(); err != nil {
			return nil, err
		}
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

	var err error
	placeholder.To, err = schemaPackage.buildMessageSchema(src)
	if err != nil {
		return nil, err
	}
	return placeholder.To, nil
}

func (pkg *Package) schemaRootFromProto(descriptor protoreflect.Descriptor) rootSchema {
	description := commentDescription(descriptor)
	pkgName, nameInPackage := splitDescriptorName(descriptor)
	linkPackage := pkg
	if pkgName != pkg.Name {
		linkPackage = pkg.PackageSet.referencePackage(pkgName)
	}

	return rootSchema{
		name:        nameInPackage,
		pkg:         linkPackage,
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

func IsOneofWrapper(msg protoreflect.MessageDescriptor) bool {
	return inferMessageType(msg) == oneofMessage
}

type messageType int

const (
	objectMessage = iota
	oneofMessage
	polymorphMessage
)

func inferMessageType(src protoreflect.MessageDescriptor) messageType {
	options := protosrc.GetExtension[*ext_j5pb.MessageOptions](src.Options(), ext_j5pb.E_Message)
	if options != nil {

		if options.IsOneofWrapper {
			// this is deprecated.
			return oneofMessage
		}

		switch options.Type.(type) {
		case *ext_j5pb.MessageOptions_Oneof:
			return oneofMessage
		case *ext_j5pb.MessageOptions_Object:
			return objectMessage
		case *ext_j5pb.MessageOptions_Polymorph:
			return polymorphMessage
		}
	}

	oneofs := src.Oneofs()

	if oneofs.Len() != 1 {
		return objectMessage
	}

	oneof := oneofs.Get(0)
	if oneof.IsSynthetic() {
		return objectMessage
	}

	if oneof.Name() != "type" {
		return objectMessage
	}

	ext := protosrc.GetExtension[*ext_j5pb.OneofOptions](oneof.Options(), ext_j5pb.E_Oneof)
	if ext != nil {
		// even if it is marked to expose, still don't automatically make it a
		// oneof.
		return objectMessage
	}

	for ii := 0; ii < src.Fields().Len(); ii++ {
		field := src.Fields().Get(ii)
		if field.ContainingOneof() != oneof {
			return objectMessage
		}
		if field.Kind() != protoreflect.MessageKind {
			return objectMessage
		}
	}

	return oneofMessage
}

func (ss *Package) buildMessageSchema(srcMsg protoreflect.MessageDescriptor) (RootSchema, error) {
	msgOptions := protosrc.GetExtension[*ext_j5pb.MessageOptions](srcMsg.Options(), ext_j5pb.E_Message)
	switch inferMessageType(srcMsg) {
	case oneofMessage:
		return ss.buildOneofSchema(srcMsg, msgOptions.GetOneof())
	case objectMessage:
		return ss.buildObjectSchema(srcMsg, msgOptions.GetObject())
	case polymorphMessage:
		return ss.buildPolymorphSchema(srcMsg, msgOptions.GetPolymorph())
	default:
		return nil, fmt.Errorf("unknown message type %q", srcMsg.FullName())
	}
}

func (ss *Package) buildOneofSchema(srcMsg protoreflect.MessageDescriptor, _ *ext_j5pb.OneofMessageOptions) (*OneofSchema, error) {
	oneofSchema := &OneofSchema{
		rootSchema: ss.schemaRootFromProto(srcMsg),
		// TODO: Rules
	}

	properties, err := ss.messageProperties(oneofSchema, srcMsg)
	if err != nil {
		return nil, fmt.Errorf("properties of %s: %w", srcMsg.FullName(), err)
	}

	for _, prop := range properties {
		if err := prop.checkValid(); err != nil {
			return nil, fmt.Errorf("property %q: %w", prop.JSONName, err)
		}
	}

	oneofSchema.Properties = properties

	return oneofSchema, nil
}

func (ss *Package) buildObjectSchema(srcMsg protoreflect.MessageDescriptor, opts *ext_j5pb.ObjectMessageOptions) (*ObjectSchema, error) {
	objectSchema := &ObjectSchema{
		rootSchema: ss.schemaRootFromProto(srcMsg),
		// TODO: Rules
	}

	properties, err := ss.messageProperties(objectSchema, srcMsg)
	if err != nil {
		return nil, fmt.Errorf("properties of %s: %w", srcMsg.FullName(), err)
	}
	objectSchema.Properties = properties

	for _, prop := range properties {
		if err := prop.checkValid(); err != nil {
			return nil, fmt.Errorf("property %q: %w", prop.JSONName, err)
		}
	}

	entity, err := findPSMOptions(srcMsg)
	if err != nil {
		return nil, fmt.Errorf("PSM options for %s: %w", srcMsg.FullName(), err)
	}
	if entity != nil {
		objectSchema.Entity = entity
	}

	if opts != nil {
		objectSchema.PolymorphMember = opts.AnyMember
	}

	return objectSchema, nil
}

func (ss *Package) buildPolymorphSchema(srcMsg protoreflect.MessageDescriptor, opts *ext_j5pb.PolymorphMessageOptions) (*PolymorphSchema, error) {
	if opts == nil {
		return nil, fmt.Errorf("polymorph message %q has no options", srcMsg.FullName())
	}
	polymorphSchema := &PolymorphSchema{
		rootSchema: ss.schemaRootFromProto(srcMsg),
		Members:    opts.Members,
	}

	return polymorphSchema, nil
}

func findPSMOptions(srcMsg protoreflect.MessageDescriptor) (*schema_j5pb.EntityObject, error) {
	psmExt := protosrc.GetExtension[*ext_j5pb.PSMOptions](srcMsg.Options(), ext_j5pb.E_Psm)

	if psmExt == nil {
		// support 'legacy' model where only the Keys message has the extension.
		keyField := srcMsg.Fields().ByName("keys")
		if keyField == nil {
			return nil, nil
		}
		msg := keyField.Message()
		if msg == nil {
			return nil, nil
		}

		psmExt = protosrc.GetExtension[*ext_j5pb.PSMOptions](msg.Options(), ext_j5pb.E_Psm)
	}

	if psmExt == nil {
		return nil, nil
	}

	var part schema_j5pb.EntityPart

	if psmExt.EntityPart != nil {
		part = *psmExt.EntityPart
	} else {
		msgName := string(srcMsg.Name())
		if strings.HasSuffix(msgName, "Keys") {
			part = schema_j5pb.EntityPart_KEYS
		} else if strings.HasSuffix(msgName, "State") {
			part = schema_j5pb.EntityPart_STATE
		} else if strings.HasSuffix(msgName, "Event") {
			part = schema_j5pb.EntityPart_EVENT
		} else if strings.HasSuffix(msgName, "Data") {
			part = schema_j5pb.EntityPart_DATA
		} else {
			return nil, fmt.Errorf("unknown PSM type suffix for %q", msgName)
		}
	}

	return &schema_j5pb.EntityObject{
		Entity: psmExt.EntityName,
		Part:   part,
	}, nil
}

func (ss *Package) messageProperties(parent RootSchema, src protoreflect.MessageDescriptor) ([]*ObjectProperty, error) {

	properties := make([]*ObjectProperty, 0, src.Fields().Len())

	exposeOneofs := make(map[string]*OneofSchema)
	pendingOneofProps := make(map[string]*ObjectProperty)

	for idx := 0; idx < src.Oneofs().Len(); idx++ {
		oneof := src.Oneofs().Get(idx)
		if oneof.IsSynthetic() {
			continue
		}

		ext := protosrc.GetExtension[*ext_j5pb.OneofOptions](oneof.Options(), ext_j5pb.E_Oneof)
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
		if err := refPlaceholder.check(); err != nil {
			return nil, err
		}

		prop := &ObjectProperty{
			Parent:      parent,
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

		context := fieldContext{
			parent:       parent,
			nameInParent: string(field.Name()),
		}

		if field.IsList() {
			ext := getProtoFieldExtensions(field)

			arrayField := &ArrayField{
				fieldContext: context,
			}

			if ext := ext.j5.GetArray(); ext != nil {
				arrayField.Ext = &schema_j5pb.ArrayField_Ext{
					SingleForm: ext.SingleForm,
				}
			}

			childContext := fieldContext{
				parent:       arrayField,
				nameInParent: "[]",
			}

			childExt := protoFieldExtensions{}

			repeatedValidate := ext.validate.GetRepeated()
			if repeatedValidate != nil {
				arrayField.Rules = &schema_j5pb.ArrayField_Rules{
					MinItems:    repeatedValidate.MinItems,
					MaxItems:    repeatedValidate.MaxItems,
					UniqueItems: repeatedValidate.Unique,
				}
				childExt.validate = ext.validate.GetRepeated().GetItems()
			}

			fieldSchema, err := ss.buildSchema(childContext, field, childExt)
			if err != nil {
				return nil, patherr.Wrap(err, string(field.Name()))
			}

			arrayField.Schema = fieldSchema

			prop := &ObjectProperty{
				Parent:      parent,
				ProtoField:  []protoreflect.FieldNumber{field.Number()},
				JSONName:    string(field.JSONName()),
				Description: commentDescription(field),
				Schema:      arrayField,
			}

			properties = append(properties, prop)
			continue

		}

		if field.IsMap() {
			if field.MapKey().Kind() != protoreflect.StringKind {
				return nil, fmt.Errorf("map keys must be strings for J5")
			}

			ext := getProtoFieldExtensions(field)

			mapField := &MapField{
				fieldContext: context,
			}

			if ext := ext.j5.GetMap(); ext != nil {
				mapField.Ext = &schema_j5pb.MapField_Ext{
					SingleForm: ext.SingleForm,
				}
			}

			childContext := fieldContext{
				parent:       mapField,
				nameInParent: "{}",
			}

			childExt := protoFieldExtensions{}

			mapValidate := ext.validate.GetMap()
			if mapValidate != nil {
				mapField.Rules = &schema_j5pb.MapField_Rules{
					MinPairs: mapValidate.MinPairs,
					MaxPairs: mapValidate.MaxPairs,
				}
				if mapValidate.Values != nil {
					childExt.validate = mapValidate.Values
				}
			}

			valueSchema, err := ss.buildSchema(childContext, field.MapValue(), childExt)
			if err != nil {
				return nil, patherr.Wrap(err, string(field.Name()))
			}

			mapField.Schema = valueSchema

			prop := &ObjectProperty{
				ProtoField:  []protoreflect.FieldNumber{field.Number()},
				JSONName:    string(field.JSONName()),
				Description: commentDescription(field),
				Schema:      mapField,
				Parent:      parent,
			}
			properties = append(properties, prop)
			continue
		}

		prop, err := ss.buildSchemaProperty(context, field)
		if err != nil {
			return nil, patherr.Wrap(err, string(field.Name()))
		}
		prop.Parent = parent

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

		// defers adding the oneof to the property array until the first
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

type protoFieldExtensions struct {
	validate *validate.FieldConstraints
	list     *list_j5pb.FieldConstraint
	j5       *ext_j5pb.FieldOptions
}

func getProtoFieldExtensions(src protoreflect.FieldDescriptor) protoFieldExtensions {
	validateConstraint := protosrc.GetExtension[*validate.FieldConstraints](src.Options(), validate.E_Field)
	if validateConstraint != nil && validateConstraint.Ignore != nil && *validateConstraint.Ignore != validate.Ignore_IGNORE_ALWAYS {
		// constraint.IgnoreEmpty doesn't really apply

		// if the constraint is repeated, unwrap it
		repeatedConstraint, ok := validateConstraint.Type.(*validate.FieldConstraints_Repeated)
		if ok {
			validateConstraint = repeatedConstraint.Repeated.Items
		}
	}

	if validateConstraint == nil {
		validateConstraint = &validate.FieldConstraints{}
	}

	listConstraint := protosrc.GetExtension[*list_j5pb.FieldConstraint](src.Options(), list_j5pb.E_Field)

	fieldOptions := protosrc.GetExtension[*ext_j5pb.FieldOptions](src.Options(), ext_j5pb.E_Field)

	exts := protoFieldExtensions{
		validate: validateConstraint,
		list:     listConstraint,
		j5:       fieldOptions,
	}

	return exts
}

func (pkg *Package) buildSchemaProperty(context fieldContext, src protoreflect.FieldDescriptor) (*ObjectProperty, error) {
	prop := &ObjectProperty{
		JSONName:    string(src.JSONName()),
		ProtoField:  []protoreflect.FieldNumber{src.Number()},
		Description: commentDescription(src),
	}

	ext := getProtoFieldExtensions(src)
	prop.Required = (ext.validate.Required != nil && *ext.validate.Required) || (src.IsList() && ext.validate.GetRepeated().GetMinItems() > 0)

	if !prop.Required && src.HasOptionalKeyword() {
		prop.ExplicitlyOptional = true
	}

	fieldSchema, err := pkg.buildSchema(context, src, ext)
	if err != nil {
		return nil, err
	}
	prop.Schema = fieldSchema

	return prop, nil
}

func (pkg *Package) buildSchema(context fieldContext, src protoreflect.FieldDescriptor, ext protoFieldExtensions) (FieldSchema, error) {
	switch src.Kind() {
	case protoreflect.MessageKind:
		return buildMessageFieldSchema(pkg, context, src, ext)

	case protoreflect.EnumKind:
		return buildEnumFieldSchema(pkg, context, src, ext)
	}

	inner, err := buildScalarType(src, ext)
	if err != nil {
		return nil, err
	}

	schemaProto := &schema_j5pb.Field{
		Type: inner,
	}
	scalar := &ScalarSchema{
		fieldContext: context,
		Kind:         src.Kind(),
		Proto:        schemaProto,
	}

	return scalar, nil
}

func ScalarSchemaFromProto(src protoreflect.FieldDescriptor) (schema_j5pb.IsField_Type, bool, error) {
	ext := getProtoFieldExtensions(src)
	t, err := buildScalarType(src, ext)
	required := ext.validate.Required != nil && *ext.validate.Required
	return t, required, err
}

func buildScalarType(src protoreflect.FieldDescriptor, ext protoFieldExtensions) (schema_j5pb.IsField_Type, error) {

	switch src.Kind() {

	case protoreflect.StringKind:
		return buildFromStringProto(src, ext)

	case protoreflect.BoolKind:

		boolConstraint := ext.validate.GetBool()
		boolItem := &schema_j5pb.BoolField{}

		if boolConstraint != nil {
			if boolConstraint.Const != nil {
				boolItem.Rules.Const = boolConstraint.Const
			}
		}

		boolList := ext.list.GetBool()
		if boolList != nil {
			boolItem.ListRules = boolList
		}

		return &schema_j5pb.Field_Bool{
			Bool: boolItem,
		}, nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		var integerRules *schema_j5pb.IntegerField_Rules

		if constraint := ext.validate.GetInt32(); constraint != nil {
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

		return &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_INT32,
				Rules:     integerRules,
				ListRules: ext.list.GetInt32(),
			},
		}, nil

	case protoreflect.Uint32Kind:
		var integerRules *schema_j5pb.IntegerField_Rules
		uint32Constraint := ext.validate.GetUint32()
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

		return &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_UINT32,
				Rules:     integerRules,
				ListRules: ext.list.GetUint32(),
			},
		}, nil

	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		var integerRules *schema_j5pb.IntegerField_Rules
		int64Constraint := ext.validate.GetInt64()
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

		return &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_INT64,
				Rules:     integerRules,
				ListRules: ext.list.GetInt64(),
			},
		}, nil

	case protoreflect.Uint64Kind:
		var integerRules *schema_j5pb.IntegerField_Rules
		int64Constraint := ext.validate.GetUint64()
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
				case *validate.UInt64Rules_Lt:
					integerRules.Maximum = Ptr(int64(cType.Lt))
					integerRules.ExclusiveMaximum = Ptr(true)
				case *validate.UInt64Rules_Lte:
					integerRules.Maximum = Ptr(int64(cType.Lte))
				}
			}
			if int64Constraint.GreaterThan != nil {
				switch cType := int64Constraint.GreaterThan.(type) {
				case *validate.UInt64Rules_Gt:
					integerRules.Minimum = Ptr(int64(cType.Gt))
					integerRules.ExclusiveMinimum = Ptr(true)
				case *validate.UInt64Rules_Gte:
					integerRules.Minimum = Ptr(int64(cType.Gte))
				}
			}
		}

		return &schema_j5pb.Field_Integer{
			Integer: &schema_j5pb.IntegerField{
				Format:    schema_j5pb.IntegerField_FORMAT_UINT64,
				Rules:     integerRules,
				ListRules: ext.list.GetInt64(),
			},
		}, nil

	case protoreflect.FloatKind:
		var numberRules *schema_j5pb.FloatField_Rules
		floatConstraint := ext.validate.GetFloat()
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

		return &schema_j5pb.Field_Float{
			Float: &schema_j5pb.FloatField{
				Format:    schema_j5pb.FloatField_FORMAT_FLOAT32,
				Rules:     numberRules,
				ListRules: ext.list.GetFloat(),
			},
		}, nil

	case protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
		var numberRules *schema_j5pb.FloatField_Rules
		floatConstraint := ext.validate.GetDouble()
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

		return &schema_j5pb.Field_Float{
			Float: &schema_j5pb.FloatField{
				Format:    schema_j5pb.FloatField_FORMAT_FLOAT64,
				Rules:     numberRules,
				ListRules: ext.list.GetDouble(),
			},
		}, nil

	case protoreflect.BytesKind:

		return &schema_j5pb.Field_Bytes{
			Bytes: &schema_j5pb.BytesField{
				Rules: &schema_j5pb.BytesField_Rules{},
			},
		}, nil

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

	ext := protosrc.GetExtension[*ext_j5pb.EnumOptions](enumDescriptor.Options(), ext_j5pb.E_Enum)

	sourceValues := enumDescriptor.Values()
	values := make([]*EnumOption, 0, sourceValues.Len())
	for ii := 0; ii < sourceValues.Len(); ii++ {
		option := sourceValues.Get(ii)
		number := int32(option.Number())

		var info map[string]string

		optExt := protosrc.GetExtension[*ext_j5pb.EnumValueOptions](option.Options(), ext_j5pb.E_EnumValue)
		if optExt != nil && optExt.Info != nil {
			info = optExt.Info
		}

		values = append(values, &EnumOption{
			name:        string(option.Name()),
			number:      number,
			description: commentDescription(option),
			Info:        info,
		})
	}

	suffix := "UNSPECIFIED"

	unspecifiedVal := string(sourceValues.Get(0).Name())
	if !strings.HasSuffix(unspecifiedVal, suffix) {
		return nil, fmt.Errorf("enum does not have an unspecified value ending in %q", suffix)
	}
	trimPrefix := strings.TrimSuffix(unspecifiedVal, suffix)

	for ii := range values {
		values[ii].name = strings.TrimPrefix(values[ii].name, trimPrefix)
	}

	if ext != nil && ext.NoDefault {
		values = values[1:]
	}
	infoFields := []*schema_j5pb.Enum_OptionInfoField{}
	if ext != nil {
		for _, field := range ext.InfoFields {
			infoFields = append(infoFields, &schema_j5pb.Enum_OptionInfoField{
				Name:        field.Name,
				Label:       field.Label,
				Description: field.Description,
			})
		}
	}
	return &EnumSchema{
		rootSchema: pkg.schemaRootFromProto(enumDescriptor),
		Options:    values,
		NamePrefix: trimPrefix,
		InfoFields: infoFields,

		//descriptor:  enumDescriptor,
	}, nil
}

func Ptr[T any](val T) *T {
	return &val
}

func wktSchema(src protoreflect.MessageDescriptor, ext protoFieldExtensions) (FieldSchema, bool, error) {
	fullName := src.FullName()

	switch string(fullName) {
	case "google.protobuf.Timestamp":
		var rules *schema_j5pb.TimestampField_Rules

		if constraint := ext.validate.GetTimestamp(); constraint != nil {
			rules = &schema_j5pb.TimestampField_Rules{}

			if constraint.Const != nil {
				return nil, false, fmt.Errorf("'const' not supported for Timestamp")
			}

			if constraint.Within != nil {
				return nil, false, fmt.Errorf("'within' not supported for Timestamp")
			}

			if constraint.LessThan != nil {
				switch cType := constraint.LessThan.(type) {
				case *validate.TimestampRules_Lt:
					rules.Maximum = cType.Lt
					rules.ExclusiveMaximum = Ptr(true)
				case *validate.TimestampRules_Lte:
					rules.Maximum = cType.Lte
				}
			}

			if constraint.GreaterThan != nil {
				switch cType := constraint.GreaterThan.(type) {
				case *validate.TimestampRules_Gt:
					rules.Minimum = cType.Gt
					rules.ExclusiveMinimum = Ptr(true)
				case *validate.TimestampRules_Gte:
					rules.Minimum = cType.Gte
				}
			}

		}
		return &ScalarSchema{
			WellKnownTypeName: fullName,
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Timestamp{
					Timestamp: &schema_j5pb.TimestampField{
						Rules:     rules,
						ListRules: ext.list.GetTimestamp(),
					},
				},
			},
		}, true, nil

	case "google.protobuf.Duration":
		return &ScalarSchema{
			Kind:              protoreflect.MessageKind,
			WellKnownTypeName: fullName,
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_String_{
					String_: &schema_j5pb.StringField{
						Format: Ptr("duration"),
					},
				},
			},
		}, true, nil

	case "j5.types.date.v1.Date":
		var rules *schema_j5pb.DateField_Rules

		if dateExt := ext.j5.GetDate(); dateExt != nil {
			if dateExt.Rules != nil {
				rules = &schema_j5pb.DateField_Rules{
					Minimum:          dateExt.Rules.Minimum,
					Maximum:          dateExt.Rules.Maximum,
					ExclusiveMinimum: dateExt.Rules.ExclusiveMinimum,
					ExclusiveMaximum: dateExt.Rules.ExclusiveMaximum,
				}
			}
		}

		return &ScalarSchema{
			Kind:              protoreflect.MessageKind,
			WellKnownTypeName: fullName,
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Date{
					Date: &schema_j5pb.DateField{
						Rules:     rules,
						ListRules: ext.list.GetDate(),
					},
				},
			},
		}, true, nil

	case "j5.types.decimal.v1.Decimal":
		var rules *schema_j5pb.DecimalField_Rules

		if dateExt := ext.j5.GetDate(); dateExt != nil {
			if dateExt.Rules != nil {
				rules = &schema_j5pb.DecimalField_Rules{
					Minimum:          dateExt.Rules.Minimum,
					Maximum:          dateExt.Rules.Maximum,
					ExclusiveMinimum: dateExt.Rules.ExclusiveMinimum,
					ExclusiveMaximum: dateExt.Rules.ExclusiveMaximum,
				}
			}
		}
		return &ScalarSchema{
			Kind:              protoreflect.MessageKind,
			WellKnownTypeName: fullName,
			Proto: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Decimal{
					Decimal: &schema_j5pb.DecimalField{
						Rules:     rules,
						ListRules: ext.list.GetDecimal(),
					},
				},
			},
		}, true, nil

	case "google.protobuf.Struct", "google.protobuf.FileDescriptorProto":
		return &MapField{
			Schema: &AnyField{},
		}, true, nil

	case "j5.types.any.v1.Any", "google.protobuf.Any":
		field := &AnyField{
			ListRules: ext.list.GetAny(),
		}
		/*
			if ext.j5 != nil {
				anyExt := ext.j5.GetAny()
				if anyExt != nil {
					// Nothing Here
				}
			}*/
		return field, true, nil

	}

	return nil, false, nil
}

func buildMessageFieldSchema(pkg *Package, context fieldContext, src protoreflect.FieldDescriptor, ext protoFieldExtensions) (FieldSchema, error) {
	flatten := false
	if ext.j5 != nil {
		if msgOptions := ext.j5.GetMessage(); msgOptions != nil {
			if src.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is not a message but has a message annotation", src.Name())
			}

			if msgOptions.Flatten {
				flatten = true
			}
		}
		if objOptions := ext.j5.GetObject(); objOptions != nil {
			if src.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is not a message but has a object annotation", src.Name())
			}

			if objOptions.Flatten {
				flatten = true
			}
		}
	}
	wktschema, ok, err := wktSchema(src.Message(), ext)
	if err != nil {
		return nil, err
	}
	if ok {
		return wktschema, nil
	}
	if strings.HasPrefix(string(src.Message().FullName()), "google.protobuf.") {
		return nil, fmt.Errorf("unsupported google type %s", src.Message().FullName())
	}
	msg := src.Message()

	ref, didExist := newRefPlaceholder(pkg.PackageSet, msg)
	if !didExist {
		var err error
		ref.To, err = pkg.buildMessageSchema(msg)
		if err != nil {
			return nil, err
		}

		if err := ref.check(); err != nil {
			return nil, err
		}
	}

	switch inferMessageType(msg) {
	case oneofMessage:
		return &OneofField{
			fieldContext: context,
			Ref:          ref,
		}, nil

	case objectMessage:
		return &ObjectField{
			fieldContext: context,
			Ref:          ref,
			Flatten:      flatten,
		}, nil
	case polymorphMessage:
		return &PolymorphField{
			fieldContext: context,
			Ref:          ref,
		}, nil
	default:
		return nil, fmt.Errorf("unknown schema type (ref) %T", ref.To)
	}
}

func buildEnumFieldSchema(pkg *Package, context fieldContext, src protoreflect.FieldDescriptor, ext protoFieldExtensions) (*EnumField, error) {
	ref, didExist := newRefPlaceholder(pkg.PackageSet, src.Enum())
	if !didExist {
		built, err := pkg.buildEnum(src.Enum())
		if err != nil {
			return nil, err
		}
		ref.To = built
	}

	var rules *schema_j5pb.EnumField_Rules
	if vc := ext.validate.GetEnum(); vc != nil {
		enumSchema := ref.To.(*EnumSchema)
		rules = &schema_j5pb.EnumField_Rules{}
		if vc.In != nil {
			for _, num := range vc.In {
				opt := enumSchema.OptionByNumber(num)
				if opt == nil {
					return nil, fmt.Errorf("enum value %d not found", num)
				}
				rules.In = append(rules.In, opt.name)
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
				rules.NotIn = append(rules.NotIn, opt.name)
			}

		}
	}

	return &EnumField{
		fieldContext: context,
		Ref:          ref,
		ListRules:    ext.list.GetEnum(),
		Rules:        rules,
	}, nil
}

var wellKnownStringPatterns = map[string]string{
	`^\d{4}-\d{2}-\d{2}$`: dateFormat,
	`^\d(.?\d)?$`:         numberFormat,
	id62.PatternString:    id62Format,
}

const (
	id62Format   = "id62"
	dateFormat   = "date"
	numberFormat = "number"
)

func buildFromStringProto(src protoreflect.FieldDescriptor, ext protoFieldExtensions) (schema_j5pb.IsField_Type, error) {
	stringItem := &schema_j5pb.StringField{}

	looksLikeKey := false

	if ext.validate != nil && ext.validate.Type != nil {
		stringItem.Rules = &schema_j5pb.StringField_Rules{}
		constraint := ext.validate.GetString()
		if constraint == nil {
			return nil, fmt.Errorf("constraint for string is %T", ext.validate.Type)
		}

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
			}
			looksLikeKey = true

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

	listRules := ext.list.GetString_()

	var fkRules *list_j5pb.KeyRules

	if fk := listRules.GetForeignKey(); fk != nil {
		switch fkt := fk.Type.(type) {
		case *list_j5pb.ForeignKeyRules_UniqueString:
			looksLikeKey = true

			if stringItem.Format != nil {
				// no formats should be set for these, there is no validate equivalent.
				return nil, fmt.Errorf("string format %q is not compatible with list.unique_string", *stringItem.Format)
			}
			fkRules = fkt.UniqueString
			if stringItem.Format == nil {
				stringItem.Format = Ptr("natural_key")
			}

		case *list_j5pb.ForeignKeyRules_Id62:
			looksLikeKey = true
			fkRules = fkt.Id62
			if stringItem.Format != nil && *stringItem.Format != "id62" {
				return nil, fmt.Errorf("string format %q is not compatible with list.id62", *stringItem.Format)
			}
			if stringItem.Format == nil {
				stringItem.Format = Ptr("id62")
			}

		case *list_j5pb.ForeignKeyRules_Uuid:
			looksLikeKey = true
			fkRules = fkt.Uuid
			if stringItem.Format != nil && *stringItem.Format != "uuid" {
				return nil, fmt.Errorf("string format %q is not compatible with list.uuid", *stringItem.Format)
			}
			if stringItem.Format == nil {
				stringItem.Format = Ptr("uuid")
			}
		}
	}

	psmKeyExt := protosrc.GetExtension[*ext_j5pb.PSMKeyFieldOptions](src.Options(), ext_j5pb.E_Key)

	if openText := listRules.GetOpenText(); openText != nil {
		if stringItem.Format != nil {
			return nil, fmt.Errorf("open_text and format %q do not match", *stringItem.Format)
		}

		if psmKeyExt != nil {
			return nil, fmt.Errorf("open_text and key constraint do not match")
		}
		stringItem.ListRules = openText
	}

	if fkRules != nil {
		looksLikeKey = true
	} else if psmKeyExt != nil {
		looksLikeKey = true
	} else if stringItem.Format != nil && *stringItem.Format == id62Format {
		looksLikeKey = true
	}

	keyFieldOpt := ext.j5.GetKey()
	if keyFieldOpt != nil {
		looksLikeKey = true
	}

	if !looksLikeKey {
		return &schema_j5pb.Field_String_{
			String_: stringItem,
		}, nil
	}

	keyField := &schema_j5pb.KeyField{
		ListRules: fkRules,
	}

	if keyFieldOpt != nil {
		if keyFieldOpt.Type != nil {
			switch keyType := keyFieldOpt.Type.(type) {
			case *ext_j5pb.KeyField_Pattern:
				keyField.Format = &schema_j5pb.KeyFormat{
					Type: &schema_j5pb.KeyFormat_Custom_{
						Custom: &schema_j5pb.KeyFormat_Custom{
							Pattern: keyType.Pattern,
						},
					},
				}
			case *ext_j5pb.KeyField_Format_:
				switch keyType.Format {
				case ext_j5pb.KeyField_FORMAT_ID62:
					keyField.Format = &schema_j5pb.KeyFormat{
						Type: &schema_j5pb.KeyFormat_Id62{
							Id62: &schema_j5pb.KeyFormat_ID62{},
						},
					}
				case ext_j5pb.KeyField_FORMAT_UUID:
					keyField.Format = &schema_j5pb.KeyFormat{
						Type: &schema_j5pb.KeyFormat_Uuid{
							Uuid: &schema_j5pb.KeyFormat_UUID{},
						},
					}
				default:
					return nil, fmt.Errorf("unknown key format %q", keyType.Format)
				}
			default:
				return nil, fmt.Errorf("unknown key type %T", keyFieldOpt.Type)

			}

		}

	}
	if psmKeyExt != nil {
		ee := &schema_j5pb.EntityKey{}
		if psmKeyExt.PrimaryKey {
			ee.Type = &schema_j5pb.EntityKey_PrimaryKey{
				PrimaryKey: true,
			}
		} else if psmKeyExt.ForeignKey != nil {
			ee.Type = &schema_j5pb.EntityKey_ForeignKey{
				ForeignKey: psmKeyExt.ForeignKey,
			}
		}
		keyField.Entity = ee
	}

	if stringItem.Format != nil {
		switch *stringItem.Format {
		case "uuid":
			keyField.Format = &schema_j5pb.KeyFormat{
				Type: &schema_j5pb.KeyFormat_Uuid{
					Uuid: &schema_j5pb.KeyFormat_UUID{},
				},
			}

		case "id62":
			keyField.Format = &schema_j5pb.KeyFormat{
				Type: &schema_j5pb.KeyFormat_Id62{
					Id62: &schema_j5pb.KeyFormat_ID62{},
				},
			}

		case "natural_key":
			keyField.Format = &schema_j5pb.KeyFormat{
				Type: &schema_j5pb.KeyFormat_Informal_{
					Informal: &schema_j5pb.KeyFormat_Informal{},
				},
			}
		}
	}

	return &schema_j5pb.Field_Key{
		Key: keyField,
	}, nil

}
