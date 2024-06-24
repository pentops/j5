package gogen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/schema/j5reflect"
)

type Options struct {
	TrimPackagePrefix   string
	FilterPackagePrefix string
	GoPackagePrefix     string
}

// ReferenceGoPackage returns the go package for the given proto package. It may
// be within the generated code, or a reference to an external package.
func (o Options) ReferenceGoPackage(pkg string) (string, error) {
	if pkg == "" {
		return "", fmt.Errorf("empty package")
	}

	if !strings.HasPrefix(pkg, o.FilterPackagePrefix) {
		return "", fmt.Errorf("package %s not in prefix %s", pkg, o.FilterPackagePrefix)
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

	pkg = path.Join(o.GoPackagePrefix, pkg)
	return pkg, nil
}

type SchemaResolver interface {
	SchemaByRef(ref *j5reflect.RefSchema) (j5reflect.RootSchema, error)
}

// fileForPackage returns the file for the given package name, creating if
// required. Returns nil when the package should not be generated (i.e. outside
// of the generate prefix, a reference to externally hosted code)
func (bb *builder) fileForPackage(grpcPackageName string) (*GeneratedFile, error) {
	if !strings.HasPrefix(grpcPackageName, bb.options.FilterPackagePrefix) {
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

func WriteGoCode(j5Package *j5reflect.Package, output FileWriter, options Options) error {

	fileSet := NewFileSet(options.GoPackagePrefix)

	bb := &builder{
		fileSet: fileSet,
		options: options,
	}

	// Only generate packages within the prefix.
	if !strings.HasPrefix(j5Package.Name, bb.options.FilterPackagePrefix) {
		return fmt.Errorf("package %s not in prefix %s", j5Package.Name, bb.options.FilterPackagePrefix)
	}
	for _, service := range j5Package.Services {
		for _, method := range service.Methods {
			if err := bb.addMethod(j5Package.Name, method); err != nil {
				return err
			}
		}
	}

	return fileSet.WriteAll(output)
}

func (bb *builder) addMethod(grpcPackage string, operation *j5reflect.Method) error {

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

	service := gen.Service(operation.Service.Name)

	responseType := fmt.Sprintf("%sResponse", operation.GRPCMethodName)
	requestType := fmt.Sprintf("%sRequest", operation.GRPCMethodName)

	pathParameters := map[string]*j5reflect.ObjectProperty{}
	{

		requestSchema, ok := operation.Request.(*j5reflect.ObjectSchema)
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

		pathParameterSet := map[string]struct{}{}

		for _, parameter := range strings.Split(operation.HTTPPath, "/") {
			if len(parameter) == 0 || parameter[0] != ':' {
				continue
			}
			pathParameterSet[parameter[1:]] = struct{}{}
		}

		queryParameters := make([]*j5reflect.ObjectProperty, 0, len(requestSchema.Properties))
		for _, property := range requestSchema.Properties {
			field, err := bb.jsonField(property)
			if err != nil {
				return err
			}

			if _, ok := pathParameterSet[property.JSONName]; ok {
				field.Tags = map[string]string{
					"path": property.JSONName,
					"json": "-",
				}
				pathParameters[property.JSONName] = property

			} else if !operation.HasBody {
				field.Tags = map[string]string{
					"query": property.JSONName,
					"json":  "-",
				}
				queryParameters = append(queryParameters, property)
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

				goName := goFieldName(property.JSONName)

				switch property.Schema.(type) {

				case *j5reflect.ScalarSchema:
					if property.Required {
						queryMethod.P("  values.Set(\"", property.JSONName, "\", s.", goName, ")")
					} else {
						queryMethod.P("  if s.", goName, " != nil {")
						queryMethod.P("    values.Set(\"", property.JSONName, "\", *s.", goName, ")")
						queryMethod.P("  }")
					}

				case *j5reflect.ObjectSchema:
					// include as JSON
					queryMethod.P("  if s.", goName, " != nil {")
					queryMethod.P("    bb, err := ", DataType{GoPackage: "encoding/json", Name: "Marshal"}, "(s.", goName, ")")
					queryMethod.P("    if err != nil {")
					queryMethod.P("      return nil, err")
					queryMethod.P("    }")
					queryMethod.P("    values.Set(\"", property.JSONName, "\", string(bb))")
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

		responseSchema, ok := operation.Response.(*j5reflect.ObjectSchema)
		if !ok {
			return fmt.Errorf("response type %q is not an object", responseType)
		}

		sliceFields := make([]*Field, 0)
		for _, property := range responseSchema.Properties {
			field, err := bb.jsonField(property)
			if err != nil {
				return fmt.Errorf("%s.ResponseBody: %w", operation.GRPCMethodName, err)
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
		Name: operation.GRPCMethodName,
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

	pathParts := strings.Split(operation.HTTPPath, "/")
	pathParams := make([]string, 0)
	for idx, part := range pathParts {
		if len(part) == 0 {
			continue
		}
		if part[0] == ':' {
			name := part[1:]
			field, ok := pathParameters[name]
			if !ok {
				return fmt.Errorf("path parameter %q not found in request object %s", name, requestType)
			}

			pathParts[idx] = "%s"

			pathParams = append(pathParams, fmt.Sprintf("req.%s", goFieldName(field.JSONName)))
		}
	}

	requestMethod.P("  path := ", DataType{GoPackage: "fmt", Name: "Sprintf"}, "(\"", strings.Join(pathParts, "/"), "\", ")
	for _, param := range pathParams {
		requestMethod.P("   ", param, ", ")
	}
	requestMethod.P("  )")

	requestMethod.P("  resp := &", responseType, "{}")
	requestMethod.P("  err := s.Request(ctx, \"", operation.HTTPMethod.ShortString(), "\", path, req, resp)")
	requestMethod.P("  if err != nil {")
	requestMethod.P("    return nil, err")
	requestMethod.P("  }")

	requestMethod.P("  return resp, nil")

	service.Methods = append(service.Methods, requestMethod)

	return nil

}
