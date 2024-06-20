package gogen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/schema/j5reflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Options struct {
	TrimPackagePrefix string
	PackagePrefix     string
	AddGoPrefix       string
}

// ReferenceGoPackage returns the go package for the given proto package. It may
// be within the generated code, or a reference to an external package.
func (o Options) ReferenceGoPackage(pkg string) (string, error) {
	if pkg == "" {
		return "", fmt.Errorf("empty package")
	}

	if !strings.HasPrefix(pkg, o.PackagePrefix) {
		return "", fmt.Errorf("package %s not in prefix %s", pkg, o.PackagePrefix)
	}

	if o.TrimPackagePrefix != "" {
		pkg = strings.TrimPrefix(pkg, o.TrimPackagePrefix)
	}

	pkg = strings.TrimSuffix(pkg, ".service")
	pkg = strings.TrimSuffix(pkg, ".topic")
	pkg = strings.TrimSuffix(pkg, ".sandbox")

	parts := strings.Split(pkg, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid package: %s", pkg)
	}
	nextName := parts[len(parts)-2]
	parts = append(parts, nextName)

	pkg = strings.Join(parts, "/")

	if o.AddGoPrefix != "" {
		pkg = path.Join(o.AddGoPrefix, pkg)
	}
	return pkg, nil
}

type builder struct {
	fileSet *FileSet
	options Options
	schemas *j5reflect.SchemaResolver
}

// fileForPackage returns the file for the given package name, creating if
// required. Returns nil when the package should not be generated (i.e. outside
// of the generate prefix, a reference to externally hosted code)
func (bb *builder) fileForPackage(grpcPackageName string) (*GeneratedFile, error) {
	if !strings.HasPrefix(grpcPackageName, bb.options.PackagePrefix) {
		return nil, nil
	}
	objectPackage, err := bb.options.ReferenceGoPackage(grpcPackageName)
	if err != nil {
		return nil, fmt.Errorf("object package name '%s': %w", grpcPackageName, err)
	}
	return bb.fileSet.File(objectPackage, filepath.Base(objectPackage))
}

type FileWriter interface {
	WriteFile(name string, data []byte) error
}

type DirFileWriter string

func (fw DirFileWriter) WriteFile(relPath string, data []byte) error {
	fullPath := filepath.Join(string(fw), relPath)
	dirName := filepath.Dir(fullPath)
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return fmt.Errorf("mkdirall for %s: %w", fullPath, err)
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("writefile for %s: %w", fullPath, err)
	}
	return nil
}

func WriteGoCode(j5Package *schema_j5pb.Package, schemas *j5reflect.SchemaResolver, output FileWriter, options Options) error {

	fileSet := NewFileSet(options.AddGoPrefix)

	bb := &builder{
		fileSet: fileSet,
		options: options,
		schemas: schemas,
	}

	// Only generate packages within the prefix.
	if !strings.HasPrefix(j5Package.Name, bb.options.PackagePrefix) {
		return fmt.Errorf("package %s not in prefix %s", j5Package.Name, bb.options.PackagePrefix)
	}
	for _, operation := range j5Package.Methods {
		if err := bb.addOperation(j5Package.Name, operation); err != nil {
			return err
		}
	}

	return fileSet.WriteAll(output)
}

func (bb *builder) buildTypeName(schema *j5reflect.Schema) (*DataType, error) {

	switch schemaType := schema.Type().(type) {
	case *j5reflect.RefSchema:
		refVal := schemaType.To
		if refVal == nil {
			return nil, fmt.Errorf("Unknown ref: %s", schemaType.Name)
		}

		switch refVal := refVal.Type().(type) {
		case *j5reflect.EnumSchema:
			return &DataType{
				Name:    "string",
				Pointer: false,
			}, nil

		case *j5reflect.ObjectSchema:

			if err := bb.addObject(refVal); err != nil {
				return nil, fmt.Errorf("referencedType in %s: %w", schemaType.Name, err)
			}

			objectPackage, err := bb.options.ReferenceGoPackage(refVal.GrpcPackageName())
			if err != nil {
				return nil, fmt.Errorf("referredType in %s: %w", schemaType.Name, err)
			}

			return &DataType{
				Name:      refVal.GoTypeName(),
				GoPackage: objectPackage,
				J5Package: refVal.GrpcPackageName(),
				Pointer:   true,
			}, nil

		case *j5reflect.OneofSchema:

			if err := bb.addOneofWrapper(refVal); err != nil {
				return nil, fmt.Errorf("referencedType in %s: %w", schemaType.Name, err)
			}

			objectPackage, err := bb.options.ReferenceGoPackage(refVal.GrpcPackageName())
			if err != nil {
				return nil, fmt.Errorf("referredType in %s: %w", schemaType.Name, err)
			}

			return &DataType{
				Name:      refVal.GoTypeName(),
				GoPackage: objectPackage,
				Pointer:   true,
				J5Package: refVal.GrpcPackageName(),
			}, nil

		default:
			return nil, fmt.Errorf("Unknown ref type: %T", schemaType.Name)
		}

	case *j5reflect.ArraySchema:
		itemType, err := bb.buildTypeName(schemaType.Schema)
		if err != nil {
			return nil, err
		}

		return &DataType{
			Name:      itemType.Name,
			Pointer:   itemType.Pointer,
			J5Package: itemType.J5Package,
			GoPackage: itemType.GoPackage,
			Slice:     true,
		}, nil

	case *j5reflect.MapSchema:
		valueType, err := bb.buildTypeName(schemaType.Schema)
		if err != nil {
			return nil, fmt.Errorf("map value: %w", err)
		}

		return &DataType{
			Name:    fmt.Sprintf("map[string]%s", valueType.Name),
			Pointer: false,
		}, nil

	case *j5reflect.ObjectSchema:
		if err := bb.addObject(schemaType); err != nil {
			return nil, fmt.Errorf("directObject: %w", err)
		}

		return &DataType{
			Name:    schemaType.GoTypeName(),
			Pointer: true,
		}, nil

	case *j5reflect.OneofSchema:
		if err := bb.addOneofWrapper(schemaType); err != nil {
			return nil, fmt.Errorf("oneofWrapper: %w", err)
		}

		return &DataType{
			Name:    schemaType.GoTypeName(),
			Pointer: true,
		}, nil

	case *j5reflect.AnySchema:
		return &DataType{
			Name:    "interface{}",
			Pointer: false,
		}, nil

	case *j5reflect.EnumSchema:
		return &DataType{
			Name:    "string",
			Pointer: false,
		}, nil

	case *j5reflect.ScalarSchema:
		asProto, err := schemaType.ToJ5Proto()
		if err != nil {
			return nil, fmt.Errorf("scalarItem: %w", err)
		}

		switch schemaType := asProto.Type.(type) {
		case *schema_j5pb.Schema_StringItem:
			item := schemaType.StringItem
			if item.Format == nil {
				return &DataType{
					Name:    "string",
					Pointer: false,
				}, nil
			}

			switch *item.Format {
			case "uuid", "date", "email", "uri":
				return &DataType{
					Name:    "string",
					Pointer: false,
				}, nil
			case "date-time":
				return &DataType{
					Name:      "Time",
					Pointer:   true,
					GoPackage: "time",
				}, nil
			case "byte":
				return &DataType{
					Name:    "[]byte",
					Pointer: false,
				}, nil
			default:
				return nil, fmt.Errorf("Unknown string format: %s", *item.Format)
			}

		case *schema_j5pb.Schema_NumberItem:
			return &DataType{
				Name:    "float64",
				Pointer: false,
			}, nil

		case *schema_j5pb.Schema_IntegerItem:
			return &DataType{
				Name:    "int64",
				Pointer: false,
			}, nil

		case *schema_j5pb.Schema_BooleanItem:
			return &DataType{
				Name:    "bool",
				Pointer: false,
			}, nil

		default:
			return nil, fmt.Errorf("Unknown scalar type: %T", schemaType)
		}

	default:
		return nil, fmt.Errorf("Unknown type for Go Gen: %T\n", schema.Type)
	}

}

func (bb *builder) jsonField(property *j5reflect.ObjectProperty) (*Field, error) {

	tags := map[string]string{}

	tags["json"] = property.JSONName
	if !property.Required {
		tags["json"] += ",omitempty"
	}

	dataType, err := bb.buildTypeName(property.Schema)
	if err != nil {
		return nil, fmt.Errorf("building field %s: %w", property.JSONName, err)
	}

	if !dataType.Pointer && !dataType.Slice && property.ExplicitlyOptional {
		dataType.Pointer = true
	}

	return &Field{
		Name:     property.GoFieldName(),
		DataType: *dataType,
		Tags:     tags,
	}, nil

}

func (bb *builder) addObject(object *j5reflect.ObjectSchema) error {
	gen, err := bb.fileForPackage(object.GrpcPackageName())
	if err != nil {
		return err
	}
	if gen == nil {
		return nil
	}

	_, ok := gen.types[object.GoTypeName()]
	if ok {
		return nil
	}

	structType := &Struct{
		Name: object.GoTypeName(),
		Comment: fmt.Sprintf(
			"Proto: %s",
			object.ProtoMessage.FullName(),
		),
	}
	gen.types[object.GoTypeName()] = structType

	for _, property := range object.Properties {
		field, err := bb.jsonField(property)
		if err != nil {
			return fmt.Errorf("object %s: %w", object.ProtoMessage.FullName(), err)
		}
		structType.Fields = append(structType.Fields, field)
	}

	return nil
}

func (bb *builder) addOneofWrapper(wrapper *j5reflect.OneofSchema) error {
	gen, err := bb.fileForPackage(wrapper.GrpcPackageName())
	if err != nil {
		return err
	}
	if gen == nil {
		return nil
	}

	_, ok := gen.types[wrapper.GoTypeName()]
	if ok {
		return nil
	}

	var comment string
	if wrapper.ProtoMessage != nil {
		comment = fmt.Sprintf(
			"Proto Message: %s", wrapper.ProtoMessage.FullName(),
		)
	} else if wrapper.OneofDescriptor != nil {
		comment = fmt.Sprintf(
			"Oneof Descriptor: %s", wrapper.OneofDescriptor.FullName(),
		)
	}

	structType := &Struct{
		Name:    wrapper.GoTypeName(),
		Comment: comment,
	}
	gen.types[wrapper.GoTypeName()] = structType

	keyMethod := &Function{
		Name: "OneofKey",
		Returns: []*Parameter{{
			DataType: DataType{
				Name:    "string",
				Pointer: false,
			}},
		},
		StringGen: gen.ChildGen(),
	}

	valueMethod := &Function{
		Name: "Type",
		Returns: []*Parameter{{
			DataType: DataType{
				Name:    "interface{}",
				Pointer: false,
			}},
		},
		StringGen: gen.ChildGen(),
	}

	for _, property := range wrapper.Properties {
		field, err := bb.jsonField(property)
		if err != nil {
			return fmt.Errorf("object %s: %w", wrapper.GoTypeName(), err)
		}
		field.DataType.Pointer = true
		structType.Fields = append(structType.Fields, field)
		keyMethod.P("if s.", field.Name, " != nil {")
		keyMethod.P("  return \"", property.JSONName, "\"")
		keyMethod.P("}")
		valueMethod.P("if s.", field.Name, " != nil {")
		valueMethod.P("  return s.", field.Name)
		valueMethod.P("}")
	}
	keyMethod.P("return \"\"")
	valueMethod.P("return nil")

	structType.Methods = append(structType.Methods, keyMethod, valueMethod)

	return nil
}

func (bb *builder) addOperation(grpcPackage string, operation *schema_j5pb.Method) error {

	gen, err := bb.fileForPackage(grpcPackage)
	if err != nil {
		return err
	}
	if gen == nil {
		return nil
	}

	gen.EnsureInterface(&Interface{
		Name: "Requester",
		Methods: []*Function{{
			Name: "Request",
			Parameters: []*Parameter{{
				Name: "ctx",
				DataType: DataType{
					GoPackage: "context",
					Name:      "Context",
				},
			}, {
				Name:     "method",
				DataType: DataType{Name: "string"},
			}, {
				Name:     "path",
				DataType: DataType{Name: "string"},
			}, {
				Name:     "body",
				DataType: DataType{Name: "interface{}"},
			}, {
				Name:     "response",
				DataType: DataType{Name: "interface{}"},
			}},
			Returns: []*Parameter{{
				DataType: DataType{
					Name: "error",
				}},
			},
		}},
	})

	service := gen.Service(operation.GrpcServiceName)

	responseType := fmt.Sprintf("%sResponse", operation.GrpcMethodName)
	requestType := fmt.Sprintf("%sRequest", operation.GrpcMethodName)

	pathParameters := map[protoreflect.Name]*j5reflect.ObjectProperty{}
	{

		requestSchemaAny, err := bb.schemas.SchemaByName(protoreflect.FullName(operation.RequestBody.GetRef()))
		if err != nil {
			return fmt.Errorf("request type %q: %w", requestType, err)
		}
		requestSchema, ok := requestSchemaAny.Type().(*j5reflect.ObjectSchema)
		if !ok {
			return fmt.Errorf("request type %q is not an object", requestType)
		}

		requestStruct := &Struct{
			Name: requestType,
		}

		if _, ok := gen.types[requestType]; ok {
			return fmt.Errorf("request type %q already exists", requestType)
		}
		gen.types[requestType] = requestStruct

		pathParameterSet := map[protoreflect.Name]struct{}{}
		queryParameterSet := map[protoreflect.Name]struct{}{}

		for _, parameter := range operation.PathParameters {
			pathParameterSet[protoreflect.Name(parameter.Name)] = struct{}{}
		}
		for _, parameter := range operation.QueryParameters {
			queryParameterSet[protoreflect.Name(parameter.Name)] = struct{}{}
		}

		queryParameters := make([]*j5reflect.ObjectProperty, 0, len(operation.QueryParameters))
		for _, property := range requestSchema.Properties {
			field, err := bb.jsonField(property)
			if err != nil {
				return err
			}

			if len(property.ProtoField) == 1 {
				fieldName := property.ProtoField[0].Name()
				if _, ok := pathParameterSet[fieldName]; ok {
					field.Tags = map[string]string{
						"path": property.JSONName,
						"json": "-",
					}
					pathParameters[fieldName] = property

				} else if _, ok := queryParameterSet[fieldName]; ok {
					field.Tags = map[string]string{
						"query": property.JSONName,
						"json":  "-",
					}
					queryParameters = append(queryParameters, property)
				}
			}
			requestStruct.Fields = append(requestStruct.Fields, field)
		}

		if len(queryParameters) > 0 {
			queryMethod := &Function{
				Name:       "QueryParameters",
				Parameters: []*Parameter{},
				Returns: []*Parameter{{
					DataType: DataType{
						Name:      "Values",
						GoPackage: "net/url",
					},
				}, {
					DataType: DataType{
						Name: "error",
					},
				}},
				StringGen: gen.ChildGen(),
			}

			queryMethod.P("  values := ", DataType{GoPackage: "net/url", Name: "Values"}, "{}")

			for _, property := range queryParameters {
				protoName, err := property.ProtoName()
				if err != nil {
					return fmt.Errorf("query parameter cannot be converted to a proto name %s: %w", property.JSONName, err)
				}

				goName := property.GoFieldName()

				switch property.Schema.Type().(type) {

				case *j5reflect.ScalarSchema:
					if property.Required {
						queryMethod.P("  values.Set(\"", protoName, "\", s.", goName, ")")
					} else {
						queryMethod.P("  if s.", goName, " != nil {")
						queryMethod.P("    values.Set(\"", protoName, "\", *s.", goName, ")")
						queryMethod.P("  }")
					}

				case *j5reflect.ObjectSchema:
					// include as JSON
					queryMethod.P("  if s.", goName, " != nil {")
					queryMethod.P("    bb, err := ", DataType{GoPackage: "encoding/json", Name: "Marshal"}, "(s.", goName, ")")
					queryMethod.P("    if err != nil {")
					queryMethod.P("      return nil, err")
					queryMethod.P("    }")
					queryMethod.P("    values.Set(\"", protoName, "\", string(bb))")
					queryMethod.P("  }")

				default:
					queryMethod.P(" // Skipping query parameter ", property.JSONName)
					//queryMethod.P("    values.Set(\"", parameter.Name, "\", fmt.Sprintf(\"%v\", *s.", GoName(parameter.Name), "))")
				}
			}

			queryMethod.P("  return values, nil")

			requestStruct.Methods = append(requestStruct.Methods, queryMethod)
		}

		for _, field := range requestStruct.Fields {
			if field.DataType.J5Package == "psm.list.v1" && field.DataType.Name == "PageRequest" {
				setter := &Function{
					Name:     "SetPageToken",
					TakesPtr: true,
					Parameters: []*Parameter{{
						Name: "pageToken",
						DataType: DataType{
							Name:    "string",
							Pointer: false,
						}},
					},
					StringGen: gen.ChildGen(),
				}
				setter.P("if s.", field.Name, " == nil {")
				setter.P("  s.", field.Name, " = ", field.DataType.Addr(), "{}")
				setter.P("}")
				setter.P("s.", field.Name, ".Token = &pageToken")

				requestStruct.Methods = append(requestStruct.Methods, setter)
			}
		}
	}

	{

		responseStruct := &Struct{
			Name: responseType,
		}

		if _, ok := gen.types[responseType]; ok {
			return fmt.Errorf("response type %q already exists", responseType)
		}

		gen.types[responseType] = responseStruct

		var pageResponseField *Field

		responseSchemaAny, err := bb.schemas.SchemaByName(protoreflect.FullName(operation.ResponseBody.GetRef()))
		if err != nil {
			return fmt.Errorf("response type %q: %w", responseType, err)
		}
		responseSchema, ok := responseSchemaAny.Type().(*j5reflect.ObjectSchema)
		if !ok {
			return fmt.Errorf("response type %q is not an object", responseType)
		}

		sliceFields := make([]*Field, 0)
		for _, property := range responseSchema.Properties {
			field, err := bb.jsonField(property)
			if err != nil {
				return fmt.Errorf("%s.ResponseBody: %w", operation.FullGrpcName, err)
			}
			responseStruct.Fields = append(responseStruct.Fields, field)
			if field.DataType.J5Package == "psm.list.v1" && field.DataType.Name == "PageResponse" {
				pageResponseField = field
			} else if field.DataType.Slice {
				sliceFields = append(sliceFields, field)
			}
		}

		if pageResponseField != nil {
			setter := &Function{
				Name: "GetPageToken",
				Returns: []*Parameter{{
					DataType: DataType{
						Name:    "string",
						Pointer: true,
					}},
				},
				StringGen: gen.ChildGen(),
			}
			setter.P("if s.", pageResponseField.Name, " == nil {")
			setter.P("  return nil")
			setter.P("}")
			setter.P("return s.", pageResponseField.Name, ".NextToken")
			responseStruct.Methods = append(responseStruct.Methods, setter)

			// Special case for list responses
			if len(sliceFields) == 1 {
				field := sliceFields[0]
				setter := &Function{
					Name: "GetItems",
					Returns: []*Parameter{{
						DataType: field.DataType.AsSlice(),
					}},
					StringGen: gen.ChildGen(),
				}
				setter.P("return s.", field.Name)
				responseStruct.Methods = append(responseStruct.Methods, setter)
			}
		}
	}

	requestMethod := &Function{
		Name: operation.GrpcMethodName,
		Parameters: []*Parameter{{
			Name: "ctx",
			DataType: DataType{
				GoPackage: "context",
				Name:      "Context",
			},
		}, {
			Name: "req",
			DataType: DataType{
				Name:    requestType,
				Pointer: true,
			},
		}},
		Returns: []*Parameter{{
			DataType: DataType{
				Name:    responseType,
				Pointer: true,
			},
		}, {
			DataType: DataType{
				Name: "error",
			},
		}},
		StringGen: gen.ChildGen(),
	}

	pathParts := strings.Split(operation.HttpPath, "/")
	pathParams := make([]string, 0)
	for idx, part := range pathParts {
		if len(part) == 0 {
			continue
		}
		if part[0] == ':' {
			name := part[1:]
			field, ok := pathParameters[protoreflect.Name(name)]
			if !ok {
				return fmt.Errorf("path parameter %q not found in request object %s", name, requestType)
			}

			pathParts[idx] = "%s"

			pathParams = append(pathParams, fmt.Sprintf("req.%s", field.GoFieldName()))
		}
	}

	requestMethod.P("  path := ", DataType{GoPackage: "fmt", Name: "Sprintf"}, "(\"", strings.Join(pathParts, "/"), "\", ")
	for _, param := range pathParams {
		requestMethod.P("   ", param, ", ")
	}
	requestMethod.P("  )")

	requestMethod.P("  resp := &", responseType, "{}")
	requestMethod.P("  err := s.Request(ctx, \"", strings.ToUpper(operation.HttpMethod), "\", path, req, resp)")
	requestMethod.P("  if err != nil {")
	requestMethod.P("    return nil, err")
	requestMethod.P("  }")

	requestMethod.P("  return resp, nil")

	service.Methods = append(service.Methods, requestMethod)

	return nil

}
