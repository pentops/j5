package gogen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
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

func WriteGoCode(api *schema_j5pb.API, output FileWriter, options Options) error {

	reflect, err := j5reflect.APIFromDesc(api)
	if err != nil {
		return err
	}
	fileSet := NewFileSet(options.GoPackagePrefix)

	bb := &builder{
		fileSet: fileSet,
		options: options,
	}

	for _, j5Package := range reflect.Packages {
		for _, service := range j5Package.Services {
			for _, method := range service.Methods {
				if err := bb.addMethod(j5Package.Name, method); err != nil {
					return fmt.Errorf("%s.%s/%s: %w", service.Package.Name, service.Name, method.GRPCMethodName, err)
				}
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

	req, err := bb.prepareRequestObject(operation)
	if err != nil {
		return err
	}
	{

		if _, ok := gen.types[req.Request.Name]; ok {
			return fmt.Errorf("request type %q already exists", req.Request.Name)
		}
		gen.types[req.Request.Name] = req.Request

		if req.pageRequestField != nil {
			if err := bb.addPaginationMethod(gen, req); err != nil {
				return err
			}
		}

		if !operation.HasBody {
			if err := bb.addQueryMethod(gen, req); err != nil {
				return err
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
					Name:    req.Request.Name,
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

		requestMethod.P("  path := ", DataType{GoPackage: "fmt", Name: "Sprintf"}, "(\"", req.pathFmtString, "\", ")
		for _, param := range req.pathFields {
			requestMethod.P("   req.", param.Name, ", ")
		}
		requestMethod.P("  )")

		requestMethod.P("  resp := &", responseType, "{}")
		requestMethod.P("  err := s.Request(ctx, \"", operation.HTTPMethod.ShortString(), "\", path, req, resp)")
		requestMethod.P("  if err != nil {")
		requestMethod.P("    return nil, err")
		requestMethod.P("  }")

		requestMethod.P("  return resp, nil")

		service.Methods = append(service.Methods, requestMethod)
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

	return nil

}

type builtRequest struct {
	Request          *Struct
	pageRequestField *Field

	pathFmtString string
	pathFields    []*Field
}

func (bb *builder) prepareRequestObject(operation *j5reflect.Method) (*builtRequest, error) {
	requestType := fmt.Sprintf("%sRequest", operation.GRPCMethodName)

	requestSchema, ok := operation.Request.(*j5reflect.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("request type %q is not an object", requestType)
	}

	requestStruct := &Struct{
		Name: requestType,
	}

	req := &builtRequest{
		Request: requestStruct,
	}

	fields := map[string]*Field{}

	for _, property := range requestSchema.Properties {
		field, err := bb.jsonField(property)
		if err != nil {
			return nil, err
		}
		fields[property.JSONName] = field
		requestStruct.Fields = append(requestStruct.Fields, field)

		if field.DataType.J5Package == "psm.list.v1" && field.DataType.Name == "PageRequest" {
			req.pageRequestField = field
		}
	}

	pathParameterSet := map[string]struct{}{}

	pathParts := strings.Split(operation.HTTPPath, "/")
	for idx, part := range pathParts {
		if len(part) == 0 || part[0] != ':' {
			continue
		}
		name := part[1:]

		field, ok := fields[name]
		if !ok {
			return nil, fmt.Errorf("path parameter %q not found in request object %s", name, requestType)
		}

		field.Tags["json"] = "-"
		field.Tags["path"] = field.Property.JSONName

		pathParameterSet[field.Name] = struct{}{}
		pathParts[idx] = "%s"
		req.pathFields = append(req.pathFields, field)
	}
	req.pathFmtString = strings.Join(pathParts, "/")

	if !operation.HasBody {
		for _, field := range requestStruct.Fields {
			if _, ok := pathParameterSet[field.Name]; ok {
				continue
			}
			field.Tags["json"] = "-"
			field.Tags["query"] = field.Property.JSONName
		}
	}

	return req, nil
}

func (bb *builder) addPaginationMethod(gen *GeneratedFile, req *builtRequest) error {

	field := req.pageRequestField
	if field.DataType.J5Package != "psm.list.v1" || field.DataType.Name != "PageRequest" {
		return fmt.Errorf("invalid page request field %q %q", field.DataType.J5Package, field.DataType.Name)
	}

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

	req.Request.Methods = append(req.Request.Methods, setter)

	return nil
}

func (bb *builder) addQueryMethod(gen *GeneratedFile, req *builtRequest) error {
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

	for _, field := range req.Request.Fields {
		if _, ok := field.Tags["query"]; !ok {
			continue
		}

		switch fieldType := field.Property.Schema.(type) {

		case *j5reflect.ScalarSchema:
			accessor := "s." + field.Name
			if !field.Property.Required {
				accessor = "*s." + field.Name
			}

			switch fieldType.Proto.Type.(type) {
			case *schema_j5pb.Schema_String_:
				// pass through
			case *schema_j5pb.Schema_Boolean:
				accessor = "fmt.Sprintf(\"%v\", " + accessor + ")"

			default:
				return fmt.Errorf("unsupported scalar type %T", fieldType.Proto.Type)

			}

			if field.Property.Required {
				queryMethod.P("  values.Set(\"", field.Property.JSONName, "\", ", accessor, ")")
			} else {
				queryMethod.P("  if s.", field.Name, " != nil {")
				queryMethod.P("    values.Set(\"", field.Property.JSONName, "\", ", accessor, ")")
				queryMethod.P("  }")
			}

		case *j5reflect.ObjectSchema:
			// include as JSON
			queryMethod.P("  if s.", field.Name, " != nil {")
			queryMethod.P("    bb, err := ", DataType{GoPackage: "encoding/json", Name: "Marshal"}, "(s.", field.Name, ")")
			queryMethod.P("    if err != nil {")
			queryMethod.P("      return nil, err")
			queryMethod.P("    }")
			queryMethod.P("    values.Set(\"", field.Name, "\", string(bb))")
			queryMethod.P("  }")

		default:
			queryMethod.P(" // Skipping query parameter ", field.Property.JSONName)
			//queryMethod.P("    values.Set(\"", parameter.Name, "\", fmt.Sprintf(\"%v\", *s.", GoName(parameter.Name), "))")
		}
	}

	queryMethod.P("  return values, nil")

	req.Request.Methods = append(req.Request.Methods, queryMethod)
	return nil
}
