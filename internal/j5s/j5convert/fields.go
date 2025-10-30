package j5convert

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/iancoleman/strcase"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
	"github.com/pentops/j5/internal/protosrc"
	"github.com/pentops/j5/lib/id62"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// mirros the buf check for 'isMap', which in turn mirrors the algorithm in protoc:
// https://github.com/bufbuild/protocompile/blob/a1712a89e0b94bbc102f376be995692c56435195/internal/util.go#L29
// https://github.com/protocolbuffers/protobuf/blob/v21.3/src/google/protobuf/descriptor.cc#L95
// It is necessary for this to match exactly for a message to be interpreted as
// a map entry.
func mapName(name string) string {
	var js []rune
	nextUpper := true
	for _, r := range name {
		if r == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			nextUpper = false
			js = append(js, unicode.ToUpper(r))
		} else {
			js = append(js, r)
		}
	}
	return string(js) + "Entry"
}

func buildProperty(ww *conversionVisitor, node *sourcewalk.PropertyNode) (*descriptorpb.FieldDescriptorProto, error) {
	if node.Schema.Schema == nil {
		return nil, fmt.Errorf("missing schema")
	}

	var fieldDesc *descriptorpb.FieldDescriptorProto
	var err error
	protoFieldName := strcase.ToSnake(node.Schema.Name)
	jsonFieldName := strcase.ToLowerCamel(protoFieldName)

	switch st := node.Field.Schema.(type) {
	case *schema_j5pb.Field_Map:
		if st.Map.ItemSchema == nil {
			return nil, errors.New("missing map item schema")
		}

		itemDesc, err := buildField(ww, *node.Field.Items)
		if err != nil {
			return nil, err
		}

		keyDesc := &descriptorpb.FieldDescriptorProto{
			Name:   gl.Ptr("key"),
			Number: gl.Ptr(int32(1)),
			Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		}

		itemDesc.Number = gl.Ptr(int32(2))
		itemDesc.Name = gl.Ptr("value")

		entryName := mapName(protoFieldName)

		mb := newMessageContext(entryName, ww.file)
		mb.descriptor.Field = []*descriptorpb.FieldDescriptorProto{
			keyDesc,
			itemDesc,
		}
		mb.descriptor.Options.MapEntry = gl.Ptr(true)

		ww.parentContext.addMessage(mb)

		fieldDesc = &descriptorpb.FieldDescriptorProto{
			Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
			TypeName: &entryName,
		}

	case *schema_j5pb.Field_Array:
		if st.Array.Items == nil {
			return nil, errors.New("missing array items")
		}

		fieldDesc, err = buildField(ww, *node.Field.Items)
		if err != nil {
			return nil, err
		}

		fieldDesc.Label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()

		if st.Array.Ext != nil {
			proto.SetExtension(fieldDesc.Options, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
				Type: &ext_j5pb.FieldOptions_Array{
					Array: st.Array.Ext,
				},
			})
		}

		ww.setJ5Ext(node.Source, fieldDesc.Options, "array", st.Array.Ext)

		validateExt := protosrc.GetExtension[*validate.FieldRules](fieldDesc.Options, validate.E_Field)

		// Add validation rules based on the type of the array regardless of array
		// rules being specified. This is specifically to cover cases where types
		// are created from other primitives, like id62 having a string validation.
		if validateExt != nil {
			repeated := &validate.RepeatedRules{
				Items: validateExt,
			}

			if st.Array.Rules != nil {
				repeated.MinItems = st.Array.Rules.MinItems
				repeated.MaxItems = st.Array.Rules.MaxItems
				repeated.Unique = st.Array.Rules.UniqueItems
			}

			rules := &validate.FieldRules{
				Type: &validate.FieldRules_Repeated{
					Repeated: repeated,
				},
			}

			proto.SetExtension(fieldDesc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

	default:
		fieldDesc, err = buildField(ww, node.Field)
		if err != nil {
			return nil, err
		}
	}

	required := node.Schema.Required
	if node.Schema.EntityKey != nil && node.Schema.EntityKey.Primary {
		required = true
	}

	if required {
		ext := protosrc.GetExtension[*validate.FieldRules](fieldDesc.Options, validate.E_Field)
		if ext == nil {
			ext = &validate.FieldRules{}
		}
		ww.file.ensureImport(bufValidateImport)
		ext.Required = gl.Ptr(true)
		proto.SetExtension(fieldDesc.Options, validate.E_Field, ext)
		ww.file.ensureImport(j5ExtImport)
	}

	if node.Schema.ExplicitlyOptional {
		if required {
			return nil, fmt.Errorf("cannot be both required and optional")
		}
		if fieldDesc.Label == nil {
			// Caller must create a 'synthetic' proto2 oneof in the parent
			// message to match the optional field
			fieldDesc.Proto3Optional = gl.Ptr(true)
			index, err := ww.parentContext.addSyntheticOneof(protoFieldName)
			if err != nil {
				return nil, err
			}
			fieldDesc.OneofIndex = gl.Ptr(index)
		}
	}

	if node.Schema.EntityKey != nil {
		proto.SetExtension(fieldDesc.Options, ext_j5pb.E_Key, node.Schema.EntityKey)
	}

	fieldDesc.Name = gl.Ptr(protoFieldName)
	fieldDesc.JsonName = gl.Ptr(jsonFieldName)
	fieldDesc.Number = gl.Ptr(node.Number)
	return fieldDesc, nil
}

func buildField(ww *conversionVisitor, node sourcewalk.FieldNode) (*descriptorpb.FieldDescriptorProto, error) {
	desc := &descriptorpb.FieldDescriptorProto{
		Options: &descriptorpb.FieldOptions{},
	}

	switch st := node.Schema.(type) {
	case *schema_j5pb.Field_Object:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()

		typeRef, err := ww.resolveType(node.Ref) //.Package, ref.Schema)
		if err != nil {
			return nil, err
		}

		desc.TypeName = typeRef.protoTypeName()
		if typeRef.Object == nil {
			return nil, fmt.Errorf("%s is not an object", typeRef.debugName())
		}

		ext := ww.setJ5Ext(node.Source, desc.Options, "object", st.Object.Ext)

		if st.Object.Flatten {
			ext.Type.(*ext_j5pb.FieldOptions_Object).Object.Flatten = true
		}

		if st.Object.Rules != nil {
			rules := &validate.FieldRules{}
			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		return desc, nil

	case *schema_j5pb.Field_Oneof:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()

		typeRef, err := ww.resolveType(node.Ref)
		if err != nil {
			return nil, err
		}
		desc.TypeName = typeRef.protoTypeName()
		if typeRef.Oneof == nil {
			return nil, fmt.Errorf("%s is not a oneof", typeRef.debugName())
		}

		ww.setJ5Ext(node.Source, desc.Options, "oneof", st.Oneof.Ext)

		if st.Oneof.Rules != nil {
			rules := &validate.FieldRules{}
			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		if st.Oneof.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Oneof{
					Oneof: st.Oneof.ListRules,
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Polymorph:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		typeRef, err := ww.resolveType(node.Ref) //.Package, ref.Schema)
		if err != nil {
			return nil, err
		}

		desc.TypeName = typeRef.protoTypeName()
		if typeRef.Polymorph == nil {
			return nil, fmt.Errorf("%s is not a polymorph", typeRef.debugName())
		}

		ww.setJ5Ext(node.Source, desc.Options, "polymorph", st.Polymorph.Ext)

		return desc, nil

	case *schema_j5pb.Field_Enum:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum()
		var enumRef *EnumRef

		typeRef, err := ww.resolveType(node.Ref)
		if err != nil {
			return nil, err
		}
		desc.TypeName = typeRef.protoTypeName()
		if typeRef.Enum == nil {
			return nil, fmt.Errorf("%s is not an enum", typeRef.debugName())
		}
		enumRef = typeRef.Enum
		ww.setJ5Ext(node.Source, desc.Options, "enum", st.Enum.Ext)

		enumRules := &validate.EnumRules{
			DefinedOnly: gl.Ptr(true),
		}

		if st.Enum.Rules != nil {
			enumRules.In, err = enumRef.mapValues(st.Enum.Rules.In)
			if err != nil {
				return nil, err
			}
			enumRules.NotIn, err = enumRef.mapValues(st.Enum.Rules.NotIn)
			if err != nil {
				return nil, err
			}
		}

		rules := &validate.FieldRules{
			Type: &validate.FieldRules_Enum{
				Enum: enumRules,
			},
		}
		proto.SetExtension(desc.Options, validate.E_Field, rules)
		ww.file.ensureImport(bufValidateImport)

		if st.Enum.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Enum{
					Enum: st.Enum.ListRules,
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Bool:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum()

		ww.setJ5Ext(node.Source, desc.Options, "bool", st.Bool.Ext)

		if st.Bool.Rules != nil {
			rules := &validate.FieldRules{
				Type: &validate.FieldRules_Bool{
					Bool: &validate.BoolRules{
						Const: st.Bool.Rules.Const,
					},
				},
			}
			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		if st.Bool.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Bool{
					Bool: st.Bool.ListRules,
				},
			})
		}
		return desc, nil

	case *schema_j5pb.Field_Bytes:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum()

		ww.setJ5Ext(node.Source, desc.Options, "bytes", st.Bytes.Ext)

		if st.Bytes.Rules != nil {
			rules := &validate.FieldRules{
				Type: &validate.FieldRules_Bytes{
					Bytes: &validate.BytesRules{
						MinLen: st.Bytes.Rules.MinLength,
						MaxLen: st.Bytes.Rules.MaxLength,
					},
				},
			}
			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		return desc, nil

	case *schema_j5pb.Field_Date:
		ww.file.ensureImport(j5DateImport)
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		desc.TypeName = gl.Ptr(".j5.types.date.v1.Date")

		if st.Date.Rules != nil {
			if st.Date.Rules.ExclusiveMinimum != nil && !(*st.Date.Rules.ExclusiveMinimum) && st.Date.Rules.Minimum == nil {
				return nil, fmt.Errorf("date rules: exclusive minimum requires minimum to be set")
			}

			if st.Date.Rules.ExclusiveMaximum != nil && !(*st.Date.Rules.ExclusiveMaximum) && st.Date.Rules.Maximum == nil {
				return nil, fmt.Errorf("date rules: exclusive maximum requires maximum to be set")
			}

			proto.SetExtension(desc.Options, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
				Type: &ext_j5pb.FieldOptions_Date{
					Date: &ext_j5pb.DateField{
						Rules: &ext_j5pb.DateField_Rules{
							Minimum:          st.Date.Rules.Minimum,
							Maximum:          st.Date.Rules.Maximum,
							ExclusiveMinimum: st.Date.Rules.ExclusiveMinimum,
							ExclusiveMaximum: st.Date.Rules.ExclusiveMaximum,
						},
					},
				},
			})
		}

		if st.Date.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Date{
					Date: st.Date.ListRules,
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Decimal:
		ww.file.ensureImport(j5DecimalImport)
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		desc.TypeName = gl.Ptr(".j5.types.decimal.v1.Decimal")

		if st.Decimal.Rules != nil {
			if st.Decimal.Rules.ExclusiveMinimum != nil && !(*st.Decimal.Rules.ExclusiveMinimum) && st.Decimal.Rules.Minimum == nil {
				return nil, fmt.Errorf("decimal rules: exclusive minimum requires minimum to be set")
			}

			if st.Decimal.Rules.ExclusiveMaximum != nil && !(*st.Decimal.Rules.ExclusiveMaximum) && st.Decimal.Rules.Maximum == nil {
				return nil, fmt.Errorf("decimal rules: exclusive maximum requires maximum to be set")
			}

			proto.SetExtension(desc.Options, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
				Type: &ext_j5pb.FieldOptions_Decimal{
					Decimal: &ext_j5pb.DecimalField{
						Rules: &ext_j5pb.DecimalField_Rules{
							Minimum:          st.Decimal.Rules.Minimum,
							Maximum:          st.Decimal.Rules.Maximum,
							ExclusiveMinimum: st.Decimal.Rules.ExclusiveMinimum,
							ExclusiveMaximum: st.Decimal.Rules.ExclusiveMaximum,
						},
					},
				},
			})
		}

		if st.Decimal.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Decimal{
					Decimal: st.Decimal.ListRules,
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Float:
		switch st.Float.Format {
		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			desc.Type = descriptorpb.FieldDescriptorProto_TYPE_FLOAT.Enum()

		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			desc.Type = descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum()

		case schema_j5pb.FloatField_FORMAT_UNSPECIFIED:
			return nil, fmt.Errorf("float format unspecified")

		default:
			return nil, fmt.Errorf("unknown float format %v", st.Float.Format)
		}

		ww.setJ5Ext(node.Source, desc.Options, "float", st.Float.Ext)

		if st.Float.Rules != nil {
			if st.Float.Rules.ExclusiveMinimum != nil && !(*st.Float.Rules.ExclusiveMinimum) && st.Float.Rules.Minimum == nil {
				return nil, fmt.Errorf("float rules: exclusive minimum requires minimum to be set")
			}

			if st.Float.Rules.ExclusiveMaximum != nil && !(*st.Float.Rules.ExclusiveMaximum) && st.Float.Rules.Maximum == nil {
				return nil, fmt.Errorf("float rules: exclusive maximum requires maximum to be set")
			}

			rules := &validate.FieldRules{}

			switch st.Float.Format {
			case schema_j5pb.FloatField_FORMAT_FLOAT32:
				rules.Type = &validate.FieldRules_Float{
					Float: &validate.FloatRules{},
				}

				if st.Float.Rules.Maximum != nil {
					if st.Float.Rules.ExclusiveMaximum != nil {
						rules.GetFloat().LessThan = &validate.FloatRules_Lt{
							Lt: float32(*st.Float.Rules.Maximum),
						}
					} else {
						rules.GetFloat().LessThan = &validate.FloatRules_Lte{
							Lte: float32(*st.Float.Rules.Maximum),
						}
					}
				}

				if st.Float.Rules.Minimum != nil {
					if st.Float.Rules.ExclusiveMinimum != nil {
						rules.GetFloat().GreaterThan = &validate.FloatRules_Gt{
							Gt: float32(*st.Float.Rules.Minimum),
						}
					} else {
						rules.GetFloat().GreaterThan = &validate.FloatRules_Gte{
							Gte: float32(*st.Float.Rules.Minimum),
						}
					}
				}

			case schema_j5pb.FloatField_FORMAT_FLOAT64:
				rules.Type = &validate.FieldRules_Double{
					Double: &validate.DoubleRules{},
				}

				if st.Float.Rules.Maximum != nil {
					if st.Float.Rules.ExclusiveMaximum != nil {
						rules.GetDouble().LessThan = &validate.DoubleRules_Lt{
							Lt: *st.Float.Rules.Maximum,
						}
					} else {
						rules.GetDouble().LessThan = &validate.DoubleRules_Lte{
							Lte: *st.Float.Rules.Maximum,
						}
					}
				}

				if st.Float.Rules.Minimum != nil {
					if st.Float.Rules.ExclusiveMinimum != nil {
						rules.GetDouble().GreaterThan = &validate.DoubleRules_Gt{
							Gt: *st.Float.Rules.Minimum,
						}
					} else {
						rules.GetDouble().GreaterThan = &validate.DoubleRules_Gte{
							Gte: *st.Float.Rules.Minimum,
						}
					}
				}

			default:
				return nil, fmt.Errorf("rules: float integer format %v", st.Float.Format)
			}

			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		if st.Float.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Float{
					Float: st.Float.ListRules,
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Integer:
		switch st.Integer.Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			desc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()

		case schema_j5pb.IntegerField_FORMAT_INT64:
			desc.Type = descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()

		case schema_j5pb.IntegerField_FORMAT_UINT32:
			desc.Type = descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum()

		case schema_j5pb.IntegerField_FORMAT_UINT64:
			desc.Type = descriptorpb.FieldDescriptorProto_TYPE_UINT64.Enum()

		default:
			return nil, fmt.Errorf("unknown integer format %v", st.Integer.Format)
		}

		ww.setJ5Ext(node.Source, desc.Options, "integer", st.Integer.Ext)

		if st.Integer.Rules != nil {
			if st.Integer.Rules.ExclusiveMinimum != nil && !(*st.Integer.Rules.ExclusiveMinimum) && st.Integer.Rules.Minimum == nil {
				return nil, fmt.Errorf("integer rules: exclusive minimum requires minimum to be set")
			}

			if st.Integer.Rules.ExclusiveMaximum != nil && !(*st.Integer.Rules.ExclusiveMaximum) && st.Integer.Rules.Maximum == nil {
				return nil, fmt.Errorf("integer rules: exclusive maximum requires maximum to be set")
			}

			rules := &validate.FieldRules{}

			switch st.Integer.Format {
			case schema_j5pb.IntegerField_FORMAT_INT32:
				rules.Type = &validate.FieldRules_Int32{
					Int32: &validate.Int32Rules{},
				}

				if st.Integer.Rules.Maximum != nil {
					if st.Integer.Rules.ExclusiveMaximum != nil {
						rules.GetInt32().LessThan = &validate.Int32Rules_Lt{
							Lt: int32(*st.Integer.Rules.Maximum),
						}
					} else {
						rules.GetInt32().LessThan = &validate.Int32Rules_Lte{
							Lte: int32(*st.Integer.Rules.Maximum),
						}
					}
				}

				if st.Integer.Rules.Minimum != nil {
					if st.Integer.Rules.ExclusiveMinimum != nil {
						rules.GetInt32().GreaterThan = &validate.Int32Rules_Gt{
							Gt: int32(*st.Integer.Rules.Minimum),
						}
					} else {
						rules.GetInt32().GreaterThan = &validate.Int32Rules_Gte{
							Gte: int32(*st.Integer.Rules.Minimum),
						}
					}
				}

			case schema_j5pb.IntegerField_FORMAT_INT64:
				rules.Type = &validate.FieldRules_Int64{
					Int64: &validate.Int64Rules{},
				}

				if st.Integer.Rules.Maximum != nil {
					if st.Integer.Rules.ExclusiveMaximum != nil {
						rules.GetInt64().LessThan = &validate.Int64Rules_Lt{
							Lt: *st.Integer.Rules.Maximum,
						}
					} else {
						rules.GetInt64().LessThan = &validate.Int64Rules_Lte{
							Lte: *st.Integer.Rules.Maximum,
						}
					}
				}

				if st.Integer.Rules.Minimum != nil {
					if st.Integer.Rules.ExclusiveMinimum != nil {
						rules.GetInt64().GreaterThan = &validate.Int64Rules_Gt{
							Gt: *st.Integer.Rules.Minimum,
						}
					} else {
						rules.GetInt64().GreaterThan = &validate.Int64Rules_Gte{
							Gte: *st.Integer.Rules.Minimum,
						}
					}
				}

			case schema_j5pb.IntegerField_FORMAT_UINT32:
				rules.Type = &validate.FieldRules_Uint32{
					Uint32: &validate.UInt32Rules{},
				}

				if st.Integer.Rules.Maximum != nil {
					if st.Integer.Rules.ExclusiveMaximum != nil {
						rules.GetUint32().LessThan = &validate.UInt32Rules_Lt{
							Lt: uint32(*st.Integer.Rules.Maximum),
						}
					} else {
						rules.GetUint32().LessThan = &validate.UInt32Rules_Lte{
							Lte: uint32(*st.Integer.Rules.Maximum),
						}
					}
				}

				if st.Integer.Rules.Minimum != nil {
					if st.Integer.Rules.ExclusiveMinimum != nil {
						rules.GetUint32().GreaterThan = &validate.UInt32Rules_Gt{
							Gt: uint32(*st.Integer.Rules.Minimum),
						}
					} else {
						rules.GetUint32().GreaterThan = &validate.UInt32Rules_Gte{
							Gte: uint32(*st.Integer.Rules.Minimum),
						}
					}
				}

			case schema_j5pb.IntegerField_FORMAT_UINT64:
				rules.Type = &validate.FieldRules_Uint64{
					Uint64: &validate.UInt64Rules{},
				}

				if st.Integer.Rules.Maximum != nil {
					if st.Integer.Rules.ExclusiveMaximum != nil {
						rules.GetUint64().LessThan = &validate.UInt64Rules_Lt{
							Lt: uint64(*st.Integer.Rules.Maximum),
						}
					} else {
						rules.GetUint64().LessThan = &validate.UInt64Rules_Lte{
							Lte: uint64(*st.Integer.Rules.Maximum),
						}
					}
				}

				if st.Integer.Rules.Minimum != nil {
					if st.Integer.Rules.ExclusiveMinimum != nil {
						rules.GetUint64().GreaterThan = &validate.UInt64Rules_Gt{
							Gt: uint64(*st.Integer.Rules.Minimum),
						}
					} else {
						rules.GetUint64().GreaterThan = &validate.UInt64Rules_Gte{
							Gte: uint64(*st.Integer.Rules.Minimum),
						}
					}
				}

			default:
				return nil, fmt.Errorf("rules: unknown integer format %v", st.Integer.Format)
			}

			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		if st.Integer.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			switch st.Integer.Format {
			case schema_j5pb.IntegerField_FORMAT_INT32:
				proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
					Type: &list_j5pb.FieldConstraint_Int32{
						Int32: st.Integer.ListRules,
					},
				})
			case schema_j5pb.IntegerField_FORMAT_INT64:
				proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
					Type: &list_j5pb.FieldConstraint_Int64{
						Int64: st.Integer.ListRules,
					},
				})
			case schema_j5pb.IntegerField_FORMAT_UINT32:
				proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
					Type: &list_j5pb.FieldConstraint_Uint32{
						Uint32: st.Integer.ListRules,
					},
				})
			case schema_j5pb.IntegerField_FORMAT_UINT64:
				proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
					Type: &list_j5pb.FieldConstraint_Uint64{
						Uint64: st.Integer.ListRules,
					},
				})
			default:
				return nil, fmt.Errorf("unknown integer format %v", st.Integer.Format)
			}
		}

		return desc, nil

	case *schema_j5pb.Field_Key:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		ww.file.ensureImport(j5ExtImport)

		keyExt := &ext_j5pb.KeyField{}
		if st.Key.Ext != nil {
			if st.Key.Ext.Foreign != nil {
				keyExt.Foreign = st.Key.Ext.Foreign
			}
		}
		proto.SetExtension(desc.Options, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
			Type: &ext_j5pb.FieldOptions_Key{
				Key: keyExt,
			},
		})

		if st.Key.ListRules != nil {
			var fkt list_j5pb.IsForeignKeyRules_Type

			if st.Key.Format == nil {
				fkt = &list_j5pb.ForeignKeyRules_UniqueString{
					UniqueString: st.Key.ListRules,
				}
			} else {
				switch st.Key.Format.Type.(type) {
				case *schema_j5pb.KeyFormat_Id62:
					fkt = &list_j5pb.ForeignKeyRules_Id62{
						Id62: st.Key.ListRules,
					}
				case *schema_j5pb.KeyFormat_Uuid:
					fkt = &list_j5pb.ForeignKeyRules_Uuid{
						Uuid: st.Key.ListRules,
					}
				case *schema_j5pb.KeyFormat_Custom_, *schema_j5pb.KeyFormat_Informal_:
					fkt = &list_j5pb.ForeignKeyRules_UniqueString{
						UniqueString: st.Key.ListRules,
					}
				default:
					return nil, fmt.Errorf("unknown key format %T", st.Key.Format.Type)
				}
			}

			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_String_{
					String_: &list_j5pb.StringRules{
						WellKnown: &list_j5pb.StringRules_ForeignKey{
							ForeignKey: &list_j5pb.ForeignKeyRules{
								Type: fkt,
							},
						},
					},
				},
			})
		}

		stringRules := &validate.StringRules{}

		if st.Key.Format != nil {
			switch ff := st.Key.Format.Type.(type) {
			case *schema_j5pb.KeyFormat_Uuid:
				stringRules.WellKnown = &validate.StringRules_Uuid{
					Uuid: true,
				}
				keyExt.Type = &ext_j5pb.KeyField_Format_{
					Format: ext_j5pb.KeyField_FORMAT_UUID,
				}

			case *schema_j5pb.KeyFormat_Id62:
				stringRules.Pattern = gl.Ptr(id62.PatternString)
				keyExt.Type = &ext_j5pb.KeyField_Format_{
					Format: ext_j5pb.KeyField_FORMAT_ID62,
				}

			case *schema_j5pb.KeyFormat_Custom_:
				stringRules.Pattern = &ff.Custom.Pattern
				keyExt.Type = &ext_j5pb.KeyField_Pattern{
					Pattern: ff.Custom.Pattern,
				}

			case *schema_j5pb.KeyFormat_Informal_:

			default:
				return nil, fmt.Errorf("unknown key format %T", st.Key.Format.Type)
			}
			proto.SetExtension(desc.Options, validate.E_Field, &validate.FieldRules{
				Type: &validate.FieldRules_String_{
					String_: stringRules,
				},
			})
			ww.file.ensureImport(bufValidateImport)

		}
		return desc, nil

	case *schema_j5pb.Field_String_:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()

		ww.setJ5Ext(node.Source, desc.Options, "string", st.String_.Ext)

		var rules *validate.StringRules
		if st.String_.Rules != nil {
			rules = &validate.StringRules{
				MinLen:  st.String_.Rules.MinLength,
				MaxLen:  st.String_.Rules.MaxLength,
				Pattern: st.String_.Rules.Pattern,
			}
		}
		if st.String_.Format != nil {
			if rules == nil {
				rules = &validate.StringRules{}
			}
			ref := stringToRef(*st.String_.Format)

			typeRef, err := ww.resolveType(&sourcewalk.RefNode{
				Source: node.Source,
				Ref:    ref,
			})
			if err != nil {
				return nil, fmt.Errorf("string format %s: %w", *st.String_.Format, err)
			}

			if typeRef.StringFormat == nil {
				return nil, fmt.Errorf("%s refers to a %s, not a string format", *st.String_.Format, typeRef.debugName())
			}

			rules.Pattern = gl.Ptr(typeRef.StringFormat.Regex)
		}
		if rules != nil {
			fieldRules := &validate.FieldRules{
				Type: &validate.FieldRules_String_{
					String_: rules,
				},
			}
			proto.SetExtension(desc.Options, validate.E_Field, fieldRules)
			ww.file.ensureImport(bufValidateImport)
		}

		if st.String_.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_String_{
					String_: &list_j5pb.StringRules{
						WellKnown: &list_j5pb.StringRules_OpenText{
							OpenText: st.String_.ListRules,
						},
					},
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Timestamp:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		desc.TypeName = gl.Ptr(".google.protobuf.Timestamp")
		ww.file.ensureImport(pbTimestamp)

		ww.setJ5Ext(node.Source, desc.Options, "timestamp", st.Timestamp.Ext)

		if st.Timestamp.Rules != nil {
			rules := &validate.FieldRules{
				Type: &validate.FieldRules_Timestamp{
					Timestamp: &validate.TimestampRules{},
				},
			}

			if st.Timestamp.Rules.Maximum != nil {
				if st.Timestamp.Rules.ExclusiveMaximum != nil {
					rules.GetTimestamp().LessThan = &validate.TimestampRules_Lt{
						Lt: st.Timestamp.Rules.Maximum,
					}
				} else {
					rules.GetTimestamp().LessThan = &validate.TimestampRules_Lte{
						Lte: st.Timestamp.Rules.Maximum,
					}
				}
			}

			if st.Timestamp.Rules.Minimum != nil {
				if st.Timestamp.Rules.ExclusiveMinimum != nil {
					rules.GetTimestamp().GreaterThan = &validate.TimestampRules_Gt{
						Gt: st.Timestamp.Rules.Minimum,
					}
				} else {
					rules.GetTimestamp().GreaterThan = &validate.TimestampRules_Gte{
						Gte: st.Timestamp.Rules.Minimum,
					}
				}
			}

			proto.SetExtension(desc.Options, validate.E_Field, rules)
			ww.file.ensureImport(bufValidateImport)
		}

		if st.Timestamp.ListRules != nil {
			ww.file.ensureImport(j5ListAnnotationsImport)
			proto.SetExtension(desc.Options, list_j5pb.E_Field, &list_j5pb.FieldConstraint{
				Type: &list_j5pb.FieldConstraint_Timestamp{
					Timestamp: st.Timestamp.ListRules,
				},
			})
		}

		return desc, nil

	case *schema_j5pb.Field_Any:
		desc.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		desc.TypeName = gl.Ptr(".j5.types.any.v1.Any")
		ww.file.ensureImport(j5AnyImport)

		proto.SetExtension(desc.Options, ext_j5pb.E_Field, &ext_j5pb.FieldOptions{
			Type: &ext_j5pb.FieldOptions_Any{
				Any: &ext_j5pb.AnyField{},
			},
		})

		return desc, nil

	default:
		return nil, fmt.Errorf("unknown schema type %T", st)
	}

}

// Copies the J5 extension object to the equivalent protoreflect extension type
// by field names.
func (ww *conversionVisitor) setJ5Ext(node sourcewalk.SourceNode, dest *descriptorpb.FieldOptions, fieldType protoreflect.Name, j5Ext proto.Message) *ext_j5pb.FieldOptions {
	// Options in the *proto* representation.
	extOptions := &ext_j5pb.FieldOptions{}
	extOptionsRefl := extOptions.ProtoReflect()

	// The proto extension is a oneof to each field type, which should match the
	// specified type.

	typeField := extOptionsRefl.Descriptor().Fields().ByName(fieldType)
	if typeField == nil {
		ww.addErrorf(node, "Field %s does not have a type field", fieldType)
		return nil
	}

	extTypedRefl := extOptionsRefl.Mutable(typeField).Message()
	if extTypedRefl == nil {
		ww.addErrorf(node, "Field %s type field is not a message", fieldType)
		return nil
	}

	// The J5 extension should already be typed. It should have the same fields
	// as the Proto extension.
	j5ExtRefl := j5Ext.ProtoReflect()
	if j5ExtRefl.IsValid() {
		j5ExtFields := extTypedRefl.Descriptor().Fields()

		// Copy each field from the J5 extension to the Proto extension.
		err := RangeField(j5ExtRefl, func(fd protoreflect.FieldDescriptor, v protoreflect.Value) error {
			destField := j5ExtFields.ByName(fd.Name())
			if destField == nil {
				return fmt.Errorf("no equivalent for %s in %s", fd.FullName(), j5ExtRefl.Descriptor().FullName())
			}

			if destField.Kind() != fd.Kind() {
				return fmt.Errorf("field %s has different kind in %s", fd.FullName(), j5ExtRefl.Descriptor().FullName())
			}

			extTypedRefl.Set(fd, j5ExtRefl.Get(destField))
			return nil
		})
		if err != nil {
			ww.addErrorf(node, "Error copying J5 extension to Proto extension: %v", err)
			return nil
		}
	}

	ww.file.ensureImport(j5ExtImport)
	// Set the extension, even if no fields were set, as this indicates the J5
	// type.
	proto.SetExtension(dest, ext_j5pb.E_Field, extOptions)

	return extOptions
}

func RangeField(pt protoreflect.Message, f func(protoreflect.FieldDescriptor, protoreflect.Value) error) error {
	var err error
	pt.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		err = f(fd, v)
		return err == nil
	})
	return err
}

func stringToRef(st string) *schema_j5pb.Ref {
	ref := &schema_j5pb.Ref{}
	parts := strings.Split(st, ".")
	if len(parts) == 1 {
		ref.Schema = parts[0]
	} else {
		head, tail := parts[0:len(parts)-1], parts[len(parts)-1]
		ref.Package = strings.Join(head, ".")
		ref.Schema = tail
	}
	return ref
}
