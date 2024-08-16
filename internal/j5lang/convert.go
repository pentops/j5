package j5lang

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5lang/extenders"
	"github.com/pentops/j5/internal/j5lang/lexer"
	"github.com/pentops/j5/internal/patherr"
)

const (
	VersionKey = "version"
)

func ConvertFile(f *File) (*source_j5pb.SourceFile, error) {
	file := &source_j5pb.SourceFile{}

	for _, decl := range f.Body.Decls {
		switch d := decl.(type) {
		case SpecialDecl:
			switch d.Key {
			case lexer.PACKAGE:
				file.Package = d.Value.ToString()

			default:
				return nil, fmt.Errorf("unexpected special decl in file root %s", d.Key)
			}

		case ValueAssign:
			if len(d.Key) == 0 {
				return nil, fmt.Errorf("empty value for option %s", d.Key)
			} else if len(d.Key) == 1 {
				switch d.Key[0] {
				case VersionKey:
					file.Version = d.Value.AsString()
				default:
					return nil, fmt.Errorf("unexpected option %s", d.Key)
				}
			} else {
				return nil, fmt.Errorf("custom option not supported %s", d.Key)
			}

		case EnumDecl:
			converted, err := ConvertEnum(d)
			if err != nil {
				return nil, fmt.Errorf("error converting enum: %w", err)
			}
			wrapped := &schema_j5pb.RootSchema{
				Type: &schema_j5pb.RootSchema_Enum{
					Enum: converted,
				},
			}

			file.Schemas = append(file.Schemas, wrapped)

		case ObjectDecl:
			converted, err := ConvertObject(d)
			if err != nil {
				return nil, patherr.Wrap(err, "object", string(d.Name))
			}
			wrapped := &schema_j5pb.RootSchema{
				Type: &schema_j5pb.RootSchema_Object{
					Object: converted,
				},
			}

			file.Schemas = append(file.Schemas, wrapped)

		case EntityDecl:
			converted, err := ConvertEntity(d)
			if err != nil {
				return nil, patherr.Wrap(err, "entity", string(d.Name))
			}

			file.Entities = append(file.Entities, converted)

		default:
			return nil, fmt.Errorf("unexpected type at root %T", d)
		}
	}

	return file, nil
}

func unexpectedDeclError(d Decl) error {
	return fmt.Errorf("unexpected decl %T at %d:%d", d, d.Start().Line, d.Start().Column)
}

func ConvertEntity(src EntityDecl) (*source_j5pb.Entity, error) {
	entity := &source_j5pb.Entity{
		Name:        strcase.ToLowerCamel(string(src.Name)),
		Description: src.Description,
		Keys: &schema_j5pb.Object{
			Name: strcase.ToCamel(string(src.Name)) + "Keys",
			Entity: &schema_j5pb.EntityObject{
				Entity: string(src.Name),
				Part:   schema_j5pb.EntityPart_KEYS,
			},
		},
	}

	for _, decl := range src.Body.Decls {
		switch d := decl.(type) {
		case FieldDecl:
			converted, err := ConvertField(d)
			if err != nil {
				return nil, patherr.Wrap(err, string(d.Name))
			}
			entity.Keys.Properties = append(entity.Keys.Properties, converted)

		case ObjectDecl:

			converted, err := ConvertObject(d)
			if err != nil {
				return nil, patherr.Wrap(err, string(d.Name))
			}

			switch d.Name {
			case "data":
				if entity.Data != nil {
					return nil, fmt.Errorf("duplicate data object")
				}
				converted.Name = strcase.ToCamel(string(src.Name)) + "Data"
				entity.Data = converted

			default:
				return nil, fmt.Errorf("unexpected object %s", d.Name)
			}

		//case OneofDecl:

		case EnumDecl:
			if d.Name != "status" {
				return nil, fmt.Errorf("unexpected enum %s", d.Name)
			}
			if entity.Status != nil {
				return nil, fmt.Errorf("duplicate status enum")
			}
			d.Name = src.Name + "_status"
			converted, err := ConvertEnum(d)
			if err != nil {
				return nil, fmt.Errorf("error converting status enum: %w", err)
			}
			entity.Status = converted

		default:
			return nil, unexpectedDeclError(d)
		}
	}

	if entity.Data == nil {
		return nil, fmt.Errorf("missing data object")
	}
	if entity.Status == nil {
		return nil, fmt.Errorf("missing status enum")
	}
	if len(entity.Keys.Properties) == 0 {
		return nil, fmt.Errorf("missing key fields")
	}

	autoNumber(entity.Keys)
	autoNumber(entity.Data)

	return entity, nil
}

func autoNumber(obj *schema_j5pb.Object) {
	for idx, prop := range obj.Properties {
		prop.ProtoField = []int32{int32(idx + 1)}
	}
}

func ConvertObject(o ObjectDecl) (*schema_j5pb.Object, error) {
	obj := &schema_j5pb.Object{
		Name:        strcase.ToCamel(string(o.Name)),
		Description: o.Description,
	}

	for _, decl := range o.Body.Decls {
		switch d := decl.(type) {
		case FieldDecl:
			converted, err := ConvertField(d)
			if err != nil {
				return nil, patherr.Wrap(err, string(d.Name))
			}
			obj.Properties = append(obj.Properties, converted)

		default:
			return nil, unexpectedDeclError(d)
		}
	}

	autoNumber(obj)
	return obj, nil
}

func ConvertEnum(e EnumDecl) (*schema_j5pb.Enum, error) {
	enum := &schema_j5pb.Enum{
		Name:        strcase.ToCamel(string(e.Name)),
		Description: e.Description,
		Prefix:      fmt.Sprintf("%s_", strcase.ToScreamingSnake(string(e.Name))),
	}

	enum.Options = append(enum.Options, &schema_j5pb.Enum_Value{
		Name:   "UNSPECIFIED",
		Number: 0,
	})
	for idx, v := range e.Options {
		enum.Options = append(enum.Options, &schema_j5pb.Enum_Value{
			Name:        string(v.Name),
			Description: v.Description,
			Number:      int32(idx + 1),
		})
	}

	return enum, nil
}

const (
	FieldTypeAny       = Ident("any")
	FieldTypeOneof     = Ident("oneof")
	FieldTypeObject    = Ident("object")
	FieldTypeEnum      = Ident("enum")
	FieldTypeArray     = Ident("array")
	FieldTypeMap       = Ident("map")
	FieldTypeString    = Ident("string")
	FieldTypeInteger   = Ident("integer")
	FieldTypeFloat     = Ident("float")
	FieldTypeBoolean   = Ident("boolean")
	FieldTypeBytes     = Ident("bytes")
	FieldTypeDecimal   = Ident("decimal")
	FieldTypeDate      = Ident("date")
	FieldTypeTimestamp = Ident("timestamp")
	FieldTypeKey       = Ident("key")
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
