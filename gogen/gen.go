package gogen

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pentops/custom-proto-api/jsonapi"
	"github.com/pentops/custom-proto-api/structure"
)

type Options struct {
	TrimPackagePrefix string
	AddGoPrefix       string
}

func (o Options) ToGoPackage(pkg string) string {
	if o.TrimPackagePrefix != "" {
		pkg = strings.TrimPrefix(pkg, o.TrimPackagePrefix)
	}

	pkg = strings.TrimSuffix(pkg, ".service")
	pkg = strings.TrimSuffix(pkg, ".topic")

	parts := strings.Split(pkg, ".")
	nextName := parts[len(parts)-2]
	parts = append(parts, nextName)

	pkg = strings.Join(parts, "/")

	if o.AddGoPrefix != "" {
		pkg = path.Join(o.AddGoPrefix, pkg)
	}
	return pkg
}

func scalarTypeName(item jsonapi.SchemaItem) (string, error) {
	switch item := item.ItemType.(type) {
	case jsonapi.StringItem:
		return "string", nil

	case jsonapi.NumberItem:
		return "float64", nil

	case jsonapi.IntegerItem:
		return "int64", nil

	case jsonapi.BooleanItem:
		return "bool", nil

	case jsonapi.EnumItem:

		return "string", nil

	case jsonapi.ObjectItem:
		if item.GoPackageName != "" {
			return "", fmt.Errorf("ObjectItem should not have a go package name: %s", item.GoPackageName)
		}

		return item.GoTypeName, nil

	default:
		return "", fmt.Errorf("Unknown type: %T\n", item)
	}
}

func WriteGoCode(document *structure.Built, outputDir string, options Options) error {

	fileSet := NewFileSet(options.AddGoPrefix)

	var addObject func(object jsonapi.ObjectItem) error

	jsonField := func(gen *GeneratedFile, property *jsonapi.ObjectProperty) (*Field, error) {

		tags := map[string]string{}

		tags["json"] = property.Name
		if !property.Required {
			tags["json"] += ",omitempty"
		}

		if property.Ref != "" {
			ref := strings.TrimPrefix(property.Ref, "#/components/schemas/")
			refVal, ok := document.Schemas[ref]
			if !ok {
				return nil, fmt.Errorf("Unknown ref: %s", ref)
			}

			object, ok := refVal.ItemType.(jsonapi.ObjectItem)
			if !ok {
				return nil, fmt.Errorf("Ref is not an object: %s", ref)
			}

			if err := addObject(object); err != nil {
				return nil, err
			}

			objectPackage := options.ToGoPackage(object.GRPCPackage)
			identity := ImportedName(objectPackage, object.GoTypeName)

			return &Field{
				Name:     GoName(property.Name),
				DataType: identity,
				Pointer:  true,
				Tags:     tags,
			}, nil
		}

		typeName, err := scalarTypeName(property.SchemaItem)
		if err != nil {
			return nil, fmt.Errorf("non scalar %#v: %w", property, err)
		}

		return &Field{
			Name:     GoName(property.Name),
			DataType: typeName,
			Pointer:  !property.Required,
			Tags:     tags,
		}, nil

	}

	addObject = func(object jsonapi.ObjectItem) error {
		objectPackage := options.ToGoPackage(object.GRPCPackage)
		gen, err := fileSet.File(objectPackage, filepath.Base(objectPackage))
		if err != nil {
			return err
		}

		_, ok := gen.types[object.GoTypeName]
		if ok {
			return nil
		}

		structType := &Struct{
			Name: object.GoTypeName,
		}
		gen.types[object.GoTypeName] = structType

		for _, property := range object.Properties {
			field, err := jsonField(gen, property)
			if err != nil {
				return err
			}
			structType.Fields = append(structType.Fields, field)
		}

		return nil
	}

	for _, pkg := range document.Packages {
		fullGoPackage := options.ToGoPackage(pkg.Name)
		for _, operation := range pkg.Methods {

			goPackageName := path.Base(fullGoPackage)
			gen, err := fileSet.File(fullGoPackage, goPackageName)
			if err != nil {
				return err
			}

			gen.EnsureInterface(&Interface{
				Name: "Requester",
				Methods: []*Function{{
					Name: "Request",
					Parameters: []*Parameter{
						{Name: "ctx", DataType: GoIdent{Package: "context", Name: "Context"}},
						{Name: "method", DataType: "string"},
						{Name: "path", DataType: "string"},
						{Name: "body", DataType: "interface{}"},
						{Name: "response", DataType: "interface{}"},
					},
					Returns: []*Parameter{{DataType: "error"}},
				}},
			})

			service := gen.Service(operation.GrpcServiceName)

			requestType := fmt.Sprintf("%sRequest", operation.GrpcMethodName)
			responseType := fmt.Sprintf("%sResponse", operation.GrpcMethodName)

			requestStruct := &Struct{
				Name: requestType,
			}
			gen.types[requestType] = requestStruct

			pathParameters := map[string]*Field{}

			for _, parameter := range operation.PathParameters {
				typeName, err := scalarTypeName(parameter.Schema)
				if err != nil {
					return err
				}

				field := &Field{
					Name:     GoName(parameter.Name),
					DataType: typeName,
					Pointer:  false,
					Tags: map[string]string{
						"path": parameter.Name,
						"json": "-",
					},
				}
				requestStruct.Fields = append(requestStruct.Fields, field)
				pathParameters[parameter.Name] = field
			}
			for _, parameter := range operation.QueryParameters {
				typeName, err := scalarTypeName(parameter.Schema)
				if err != nil {
					return err
				}

				requestStruct.Fields = append(requestStruct.Fields, &Field{
					Name:     GoName(parameter.Name),
					DataType: typeName,
					Pointer:  false,
					Tags: map[string]string{
						"query": parameter.Name,
						"json":  "-",
					},
				})
			}

			if operation.RequestBody != nil {
				requestSchema := operation.RequestBody.ItemType.(jsonapi.ObjectItem)
				for _, property := range requestSchema.Properties {
					field, err := jsonField(gen, property)
					if err != nil {
						return err
					}
					requestStruct.Fields = append(requestStruct.Fields, field)
				}
			}

			responseStruct := &Struct{
				Name: responseType,
			}

			gen.types[responseType] = responseStruct

			responseSchema := operation.ResponseBody.ItemType.(jsonapi.ObjectItem)
			for _, property := range responseSchema.Properties {
				field, err := jsonField(gen, property)
				if err != nil {
					return err
				}
				responseStruct.Fields = append(responseStruct.Fields, field)
			}

			requestMethod := &Function{
				Name: operation.GrpcMethodName,
				Parameters: []*Parameter{{
					Name:     "ctx",
					DataType: GoIdent{Package: "context", Name: "Context"},
				}, {
					Name:     "req",
					DataType: requestType,
					Pointer:  true,
				}},
				Returns: []*Parameter{{
					DataType: responseType,
					Pointer:  true,
				}, {
					DataType: "error",
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

					pathParts[idx] = "%q"

					pathParams = append(pathParams, fmt.Sprintf("req.%s", field.Name))
				}
			}

			requestMethod.P("  path := ", ImportedName("fmt", "Sprintf"), "(\"", strings.Join(pathParts, "/"), "\", ")
			for _, param := range pathParams {
				requestMethod.P("   ", param, ", ")
			}
			requestMethod.P("  )")

			requestMethod.P("  resp := &", responseType, "{}")
			requestMethod.P("  err := s.Request(ctx, \"", strings.ToUpper(operation.HTTPMethod), "\", path, req, resp)")
			requestMethod.P("  if err != nil {")
			requestMethod.P("    return nil, err")
			requestMethod.P("  }")

			requestMethod.P("  return resp, nil")

			service.Methods = append(service.Methods, requestMethod)

		}

	}

	return fileSet.WriteAll(outputDir)

}

var reUnsafe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// GoName exports the field name
func GoName(name string) string {
	name = reUnsafe.ReplaceAllString(name, "_")
	return strings.ToUpper(name[0:1]) + name[1:]
}
