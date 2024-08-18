package j5lang

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5lang/ast"
	"github.com/pentops/j5/internal/j5lang/extenders"
	"github.com/pentops/j5/internal/j5lang/lexer"
	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/internal/patherr"
	"github.com/pentops/j5/schema/j5reflect"
)

const (
	VersionKey      = "version"
	FieldTag        = "field"
	ObjectTag       = "object"
	EntityTag       = "entity"
	EnumTag         = "enum"
	EnumOptionTag   = "option"
	EnumPrefixTag   = "prefix"
	OneofOptionTag  = "option"
	EntityDataTag   = "data"
	EntityKeyTag    = "key"
	EntityStatusTag = "status"
	EntityEventTag  = "event"
)

func ConvertTreeToSource(f *ast.File) (*source_j5pb.SourceFile, error) {

	schemaSet := j5schema.NewSchemaCache()
	builder := &SchemaBuilder{
		schemaSet: schemaSet,
	}

	schema := &FileBuilder{
		builder: builder,
		file:    &source_j5pb.SourceFile{},
	}
	if err := ast.ApplySchema(f, schema); err != nil {
		return nil, err
	}

	return schema.file, nil
}

type SchemaBuilder struct {
	schemaSet *j5schema.SchemaCache
}

// errUnexpected can be returned from Assign and Block, both of which will wrap
// the error as a patherr to give context.
var errUnexpected = fmt.Errorf("unexpected")

type sourceNode interface {
	Source() ast.SourceNode
}

func posErrorf(tok sourceNode, format string, args ...interface{}) error {
	source := tok.Source()
	return &lexer.PositionError{
		Position: source.Start,
		Msg:      fmt.Sprintf(format, args...),
	}
}

type FileBuilder struct {
	builder *SchemaBuilder
	file    *source_j5pb.SourceFile
	Version string
}

func (f *FileBuilder) Assign(ref ast.Reference, value ast.Value) error {
	var err error
	name := ref.String()
	switch name {
	case VersionKey:
		f.Version, err = value.AsString()
		if err != nil {
			return patherr.Wrap(err, "version")
		}
		return nil
	default:
		return fmt.Errorf("unexpected option")
	}
}

func (f *FileBuilder) Block(hdr ast.BlockHeader) (ast.Schema, error) {
	switch hdr.RootName() {
	case EnumTag:
		builder, err := newEnumBlock(hdr)
		if err != nil {
			return nil, err
		}
		f.file.Schemas = append(f.file.Schemas, builder.asRootSchema())
		return builder, nil

	case ObjectTag:
		builder, err := newObjectBlock(hdr)
		if err != nil {
			return nil, err
		}
		f.file.Schemas = append(f.file.Schemas, builder.asRootSchema())
		return builder, nil

	case EntityTag:
		builder, err := newEntityBlock(hdr)
		if err != nil {
			return nil, err
		}
		f.file.Entities = append(f.file.Entities, builder.entity)
		return builder, nil
	}
	return nil, errUnexpected
}

func (f *FileBuilder) Done() error {
	return nil
}

type entityBlock struct {
	entity *source_j5pb.Entity
}

func newEntityBlock(block ast.BlockHeader) (*entityBlock, error) {
	var name string
	err := block.ScanTags(&name)
	if err != nil {
		return nil, err
	}
	name = strcase.ToLowerCamel(name)
	entity := &source_j5pb.Entity{
		Name:        name,
		Description: block.Description,
		Keys: &schema_j5pb.Object{
			Name: strcase.ToCamel(name) + "Keys",
			Entity: &schema_j5pb.EntityObject{
				Entity: name,
				Part:   schema_j5pb.EntityPart_KEYS,
			},
		},
		Status: &schema_j5pb.Enum{
			Name:   strcase.ToCamel(name) + "Status",
			Prefix: fmt.Sprintf("%s_", strcase.ToScreamingSnake(name)),
		},
	}
	return &entityBlock{entity: entity}, nil
}

func (eb entityBlock) Assign(ref ast.Reference, value ast.Value) error {
	return fmt.Errorf("unexpected option")
}

func (eb entityBlock) Block(stmt ast.BlockHeader) (ast.Schema, error) {
	switch stmt.RootName() {
	case FieldTag:
		field, err := newFieldBlock(stmt)
		if err != nil {
			return nil, err
		}
		eb.entity.Keys.Properties = append(eb.entity.Keys.Properties, field.property)
		return field, nil

	case EntityDataTag:
		builder, err := newObjectBlock(stmt)
		if err != nil {
			return nil, err
		}
		eb.entity.Data = builder.object
		eb.entity.Data.Name = strcase.ToCamel(eb.entity.Name) + "Data"
		return builder, nil

	case EntityStatusTag:
		enumOptionBlock, err := newEnumOptionBlock(stmt)
		if err != nil {
			return nil, err
		}
		eb.entity.Status.Options = append(eb.entity.Status.Options, enumOptionBlock.option)
		return enumOptionBlock, nil

	default:
		return nil, errUnexpected
	}
}

func (eb entityBlock) Done() error {
	entity := eb.entity

	if entity.Data == nil {
		return fmt.Errorf("missing data object")
	}
	if entity.Status == nil {
		return fmt.Errorf("missing status enum")
	}
	if len(entity.Keys.Properties) == 0 {
		return fmt.Errorf("missing key fields")
	}

	autoNumber(entity.Keys)
	autoNumber(entity.Data)

	return nil
}

func autoNumber(obj *schema_j5pb.Object) {
	for idx, prop := range obj.Properties {
		prop.ProtoField = []int32{int32(idx + 1)}
	}
}

type objectBlock struct {
	object *schema_j5pb.Object
}

func newObjectBlock(block ast.BlockHeader) (*objectBlock, error) {
	var name string
	err := block.ScanTags(&name)
	if err != nil {
		return nil, err
	}

	obj := &schema_j5pb.Object{
		Name:        strcase.ToCamel(name),
		Description: block.Description,
	}
	return &objectBlock{object: obj}, nil
}

func (ob objectBlock) asRootSchema() *schema_j5pb.RootSchema {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: ob.object,
		},
	}
}

func (ob objectBlock) Assign(ref ast.Reference, value ast.Value) error {
	return fmt.Errorf("unexpected option")
}

func (ob objectBlock) Block(block ast.BlockHeader) (ast.Schema, error) {
	switch block.RootName() {
	case FieldTag:
		builder, err := newFieldBlock(block)
		if err != nil {
			return nil, err
		}
		ob.object.Properties = append(ob.object.Properties, builder.property)
		return builder, nil

	default:
		return nil, errUnexpected
	}

}

func (ob objectBlock) Done() error {
	autoNumber(ob.object)
	return nil
}

type enumBlock struct {
	enum *schema_j5pb.Enum
}

func newEnumBlock(block ast.BlockHeader) (*enumBlock, error) {
	var name string
	err := block.ScanTags(&name)
	if err != nil {
		return nil, err
	}

	enum := &schema_j5pb.Enum{
		Name:        strcase.ToCamel(name),
		Description: block.Description,
		Prefix:      fmt.Sprintf("%s_", strcase.ToScreamingSnake(name)),
	}
	enum.Options = append(enum.Options, &schema_j5pb.Enum_Value{
		Name:   "UNSPECIFIED",
		Number: 0,
	})
	return &enumBlock{enum: enum}, nil
}

func (eb enumBlock) asRootSchema() *schema_j5pb.RootSchema {
	return &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Enum{
			Enum: eb.enum,
		},
	}
}

func (eb enumBlock) Done() error {
	for idx, opt := range eb.enum.Options {
		opt.Number = int32(idx)
	}
	return nil
}

func (eb enumBlock) Assign(ref ast.Reference, value ast.Value) error {
	var err error
	name := ref.String()
	switch name {
	case EnumPrefixTag:
		eb.enum.Prefix, err = value.AsString()
		if err != nil {
			return patherr.Wrap(err, "prefix")
		}
		return nil
	default:
		return errUnexpected
	}
}

func (eb enumBlock) Block(block ast.BlockHeader) (ast.Schema, error) {
	switch block.RootName() {
	case EnumOptionTag:
		builder, err := newEnumOptionBlock(block)
		if err != nil {
			return nil, err
		}
		eb.enum.Options = append(eb.enum.Options, builder.option)
		return builder, nil
	default:
		return nil, errUnexpected
	}
}

type enumOptionBlock struct {
	option *schema_j5pb.Enum_Value
}

func newEnumOptionBlock(hdr ast.BlockHeader) (*enumOptionBlock, error) {
	var name string
	err := hdr.ScanTags(&name)
	if err != nil {
		return nil, err
	}

	option := &schema_j5pb.Enum_Value{
		Name:        name,
		Description: hdr.Description,
		Number:      0, // Must be set in Done of the parent
	}
	return &enumOptionBlock{option: option}, nil
}

func (e enumOptionBlock) Assign(ref ast.Reference, value ast.Value) error {
	return errUnexpected
}

func (e enumOptionBlock) Block(block ast.BlockHeader) (ast.Schema, error) {
	return nil, errUnexpected
}

func (e enumOptionBlock) Done() error {
	return nil
}

const (
	FieldTypeAny       = "any"
	FieldTypeOneof     = "oneof"
	FieldTypeObject    = "object"
	FieldTypeEnum      = "enum"
	FieldTypeArray     = "array"
	FieldTypeMap       = "map"
	FieldTypeString    = "string"
	FieldTypeInteger   = "integer"
	FieldTypeFloat     = "float"
	FieldTypeBoolean   = "boolean"
	FieldTypeBytes     = "bytes"
	FieldTypeDecimal   = "decimal"
	FieldTypeDate      = "date"
	FieldTypeTimestamp = "timestamp"
	FieldTypeKey       = "key"
)

type FieldExtender interface {
	FieldLevel(field *schema_j5pb.ObjectProperty, key string, value extenders.Value) (bool, error)
	Any(field *schema_j5pb.AnyField, key string, value extenders.Value) error
	Oneof(field *schema_j5pb.OneofField, key string, value extenders.Value) error
	Object(field *schema_j5pb.ObjectField, key string, value extenders.Value) error
	Enum(field *schema_j5pb.EnumField, key string, value extenders.Value) error
	Array(field *schema_j5pb.ArrayField, key string, value extenders.Value) error
	Map(field *schema_j5pb.MapField, key string, value extenders.Value) error
	String(field *schema_j5pb.StringField, key string, value extenders.Value) error
	Integer(field *schema_j5pb.IntegerField, key string, value extenders.Value) error
	Float(field *schema_j5pb.FloatField, key string, value extenders.Value) error
	Boolean(field *schema_j5pb.BooleanField, key string, value extenders.Value) error
	Bytes(field *schema_j5pb.BytesField, key string, value extenders.Value) error
	Decimal(field *schema_j5pb.DecimalField, key string, value extenders.Value) error
	Date(field *schema_j5pb.DateField, key string, value extenders.Value) error
	Timestamp(field *schema_j5pb.TimestampField, key string, value extenders.Value) error
	Key(field *schema_j5pb.KeyField, key string, value extenders.Value) error
}

var fieldExtenders = map[string]FieldExtender{
	"validate": extenders.ValidateExtender{},
}

type fieldExtension struct {
	extender FieldExtender
	root     string
	key      string
	value    extenders.Value
}

type fieldBuilder struct {
	property *schema_j5pb.ObjectProperty
	schema   j5reflect.OneofSchema
}

func newFieldBlock(block ast.BlockHeader) (*fieldBuilder, error) {
	var name, typeName string
	err := block.ScanTags(&name, &typeName)
	if err != nil {
		return nil, err
	}

	property := &schema_j5pb.ObjectProperty{
		Name:        strcase.ToCamel(name),
		Description: block.Description,
		Schema:      &schema_j5pb.Field{},
	}

	refl := property.Schema.ProtoReflect()

	return &fieldBuilder{
		property: property,
		schema:   refl,
	}, nil
}

func (fb fieldBuilder) Assign(ref ast.Reference, value ast.Value) error {
	return errUnexpected
}

func (fb fieldBuilder) Block(block ast.BlockHeader) (ast.Schema, error) {
	return nil, errUnexpected
}

func (fb fieldBuilder) Done() error {
	return nil
}

/*

func ConvertField(f FieldDecl) (*schema_j5pb.ObjectProperty, error) {
	field := &schema_j5pb.ObjectProperty{
		Name:        string(f.Name),
		Description: f.Description,
	}
	var err error

	valueAssigns := map[string]ValueAssign{}
	specialDecls := map[lexer.TokenType]SpecialDecl{}
	exts := []fieldExtension{}

	for _, decl := range f.Body.Decls {
		switch d := decl.(type) {
		case SpecialDecl:
			specialDecls[d.Key] = d
		case ValueAssign:
			if len(d.Key) == 0 {
				return nil, fmt.Errorf("empty value for option %s", d.Key)
			} else if len(d.Key) == 1 {
				valueAssigns[string(d.Key[0])] = d
			} else {
				root := d.Key[0 : len(d.Key)-1].ToString()
				key := string(d.Key[len(d.Key)-1])

				fieldExtender, ok := fieldExtenders[root]
				if !ok {
					return nil, fmt.Errorf("unexpected field extender %s", root)
				}
				exts = append(exts, fieldExtension{
					extender: fieldExtender,
					root:     root,
					key:      key,
					value:    d.Value,
				})
			}
		default:
			return nil, fmt.Errorf("unexpected decl %T", d)
		}
	}

	popSpecial := func(key lexer.TokenType) (SpecialDecl, bool) {
		decl, ok := specialDecls[key]
		if ok {
			delete(specialDecls, key)
		}
		return decl, ok
	}

	popAssign := func(key string) (ValueAssign, bool) {
		assign, ok := valueAssigns[key]
		if ok {
			delete(valueAssigns, key)
		}
		return assign, ok
	}

	popRef := func() *schema_j5pb.Ref {
		ref, ok := popSpecial(lexer.REF)
		if !ok {
			return nil
		}
		return &schema_j5pb.Ref{
			Package: ref.Value[0 : len(ref.Value)-1].ToString(),
			Schema:  string(ref.Value[len(ref.Value)-1]),
		}
	}

	switch f.Type {
	case FieldTypeAny:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Any{
				Any: &schema_j5pb.AnyField{},
			},
		}

	case FieldTypeArray:
		itemSchema := &schema_j5pb.Field{}
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Array{
				Array: &schema_j5pb.ArrayField{
					Rules: &schema_j5pb.ArrayField_Rules{},
					Items: itemSchema,
				},
			},
		}

	case FieldTypeMap:
		itemSchema := &schema_j5pb.Field{}
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Map{
				Map: &schema_j5pb.MapField{
					Rules:      &schema_j5pb.MapField_Rules{},
					ItemSchema: itemSchema,
					KeySchema: &schema_j5pb.Field{
						Type: &schema_j5pb.Field_String_{
							String_: &schema_j5pb.StringField{
								Rules: &schema_j5pb.StringField_Rules{},
							},
						},
					},
				},
			},
		}

	case FieldTypeInteger:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Integer{
				Integer: &schema_j5pb.IntegerField{
					Rules: &schema_j5pb.IntegerField_Rules{},
				},
			},
		}

	case FieldTypeFloat:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Float{
				Float: &schema_j5pb.FloatField{
					Rules: &schema_j5pb.FloatField_Rules{},
				},
			},
		}

	case FieldTypeBoolean:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Boolean{
				Boolean: &schema_j5pb.BooleanField{
					Rules: &schema_j5pb.BooleanField_Rules{},
				},
			},
		}

	case FieldTypeBytes:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Bytes{
				Bytes: &schema_j5pb.BytesField{
					Rules: &schema_j5pb.BytesField_Rules{},
				},
			},
		}

	case FieldTypeDecimal:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Decimal{
				Decimal: &schema_j5pb.DecimalField{
					Rules: &schema_j5pb.DecimalField_Rules{},
				},
			},
		}

	case FieldTypeDate:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Date{
				Date: &schema_j5pb.DateField{
					Rules: &schema_j5pb.DateField_Rules{},
				},
			},
		}

	case FieldTypeTimestamp:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Timestamp{
				Timestamp: &schema_j5pb.TimestampField{
					Rules: &schema_j5pb.TimestampField_Rules{},
				},
			},
		}

	case FieldTypeKey:
		keyField := &schema_j5pb.KeyField{
			Rules: &schema_j5pb.KeyField_Rules{},
		}

		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Key{

				Key: keyField,
			},
		}

		format, ok := popAssign("format")
		if ok {
			switch format.Value.AsString() {
			case "uuid":
				keyField.Format = schema_j5pb.KeyFormat_UUID
			case "natural":
				keyField.Format = schema_j5pb.KeyFormat_NATURAL_KEY
			default:
				return nil, fmt.Errorf("unexpected key format %s", format.Value.AsString())
			}
		}

		primary, ok := popAssign("primary")
		if ok {
			keyField.Primary, err = primary.Value.AsBoolean()
			if err != nil {
				return nil, fmt.Errorf("error parsing primary %w", err)
			}
		}

		popAssign("references") // TODO: implement references

	case FieldTypeString:
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_String_{
				String_: &schema_j5pb.StringField{
					Rules: &schema_j5pb.StringField_Rules{},
				},
			},
		}

	case FieldTypeOneof:
		ref := popRef()
		if ref == nil {
			return nil, fmt.Errorf("missing ref for oneof field")
		}
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Oneof{
				Oneof: &schema_j5pb.OneofField{
					Schema: &schema_j5pb.OneofField_Ref{
						Ref: ref,
					},
				},
			},
		}

	case FieldTypeEnum:
		ref := popRef()
		if ref == nil {
			return nil, fmt.Errorf("missing ref for enum field")
		}
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Enum{
				Enum: &schema_j5pb.EnumField{
					Schema: &schema_j5pb.EnumField_Ref{
						Ref: ref,
					},
				},
			},
		}

	case FieldTypeObject:
		ref := popRef()
		if ref == nil {
			return nil, fmt.Errorf("missing ref for object field")
		}
		obj := &schema_j5pb.ObjectField{
			Schema: &schema_j5pb.ObjectField_Ref{
				Ref: ref,
			},
		}
		field.Schema = &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Object{
				Object: obj,
			},
		}

	default:
		return nil, fmt.Errorf("unsupported field type %s", f.Type)
	}

	for key := range specialDecls {
		return nil, fmt.Errorf("unexpected special %s", key.String())
	}

	for _, va := range valueAssigns {
		return nil, fmt.Errorf("unexpected value assign %s", va.Key)
	}

	for _, ext := range exts {
		didRoot, err := ext.extender.FieldLevel(field, ext.key, ext.value)
		if err != nil {
			return nil, fmt.Errorf("error running field level %w", err)
		}
		if didRoot {
			continue
		}

		switch f := field.Schema.Type.(type) {
		case *schema_j5pb.Field_String_:
			if err := ext.extender.String(f.String_, ext.key, ext.value); err != nil {
				return nil, fmt.Errorf("error running string %w", err)
			}

		default:
			return nil, fmt.Errorf("unexpected field type %T", f)
		}
	}

	return field, nil
}
*/
