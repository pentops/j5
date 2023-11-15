package swagger

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pentops/custom-proto-api/jsonapi"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type builder struct {
	document *Document
	paths    map[string]*PathItem
	schemas  *jsonapi.SchemaSet
}

func BuildFromDescriptors(options jsonapi.Options, descriptors *descriptorpb.FileDescriptorSet) (*Document, error) {

	services := make([]protoreflect.ServiceDescriptor, 0)
	descFiles, err := protodesc.NewFiles(descriptors)
	if err != nil {
		return nil, err
	}

	descFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		fileServices := file.Services()
		for ii := 0; ii < fileServices.Len(); ii++ {
			service := fileServices.Get(ii)
			services = append(services, service)
		}
		return true
	})

	filteredServices := make([]protoreflect.ServiceDescriptor, 0)
	for _, service := range services {
		name := service.FullName()
		if !strings.HasSuffix(string(name), "Service") {
			continue
		}

		filteredServices = append(filteredServices, service)
	}

	return Build(options, filteredServices)
}

func Build(options jsonapi.Options, services []protoreflect.ServiceDescriptor) (*Document, error) {
	b := builder{
		document: &Document{
			OpenAPI: "3.0.0",
			Components: Components{
				SecuritySchemes: make(map[string]interface{}),
			},
		},
		paths:   make(map[string]*PathItem),
		schemas: jsonapi.NewSchemaSet(options),
	}

	for _, service := range services {
		if err := b.addService(service); err != nil {
			return nil, err
		}
	}

	b.document.Components.Schemas = b.schemas.Schemas
	return b.document, nil

}

func (bb *builder) addService(src protoreflect.ServiceDescriptor) error {
	methods := src.Methods()
	name := string(src.FullName())
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)
		if err := bb.registerMethod(name, method); err != nil {
			return err
		}
	}
	return nil
}

var rePathParameter = regexp.MustCompile(`\{([^\}]+)\}`)

func (bb *builder) registerMethod(serviceName string, method protoreflect.MethodDescriptor) error {

	methodOptions := method.Options().(*descriptorpb.MethodOptions)
	httpOpt := proto.GetExtension(methodOptions, annotations.E_Http).(*annotations.HttpRule)

	var httpMethod string
	var httpPath string

	if httpOpt == nil {
		return fmt.Errorf("missing http rule for method /%s/%s", serviceName, method.Name())
	}
	switch pt := httpOpt.Pattern.(type) {
	case *annotations.HttpRule_Get:
		httpMethod = "get"
		httpPath = pt.Get
	case *annotations.HttpRule_Post:
		httpMethod = "post"
		httpPath = pt.Post
	case *annotations.HttpRule_Put:
		httpMethod = "put"
		httpPath = pt.Put
	case *annotations.HttpRule_Delete:
		httpMethod = "delete"
		httpPath = pt.Delete
	case *annotations.HttpRule_Patch:
		httpMethod = "patch"
		httpPath = pt.Patch

	default:
		return fmt.Errorf("unsupported http method %T", pt)
	}

	operation := &Operation{
		OperationHeader: OperationHeader{
			Method:      httpMethod,
			Path:        httpPath,
			OperationID: fmt.Sprintf("/%s/%s", serviceName, method.Name()),
		},
	}

	okResponse, err := bb.schemas.BuildSchemaObject(method.Output())
	if err != nil {
		return err
	}

	operation.Responses = &ResponseSet{{
		Code:        200,
		Description: "OK",
		Content: OperationContent{
			JSON: &OperationSchema{
				Schema: *okResponse,
			},
		},
	}}

	request, err := bb.schemas.BuildSchemaObject(method.Input())
	if err != nil {
		return err
	}

	requestObject := request.ItemType.(jsonapi.ObjectItem)

	for _, paramStr := range rePathParameter.FindAllString(httpPath, -1) {
		name := paramStr[1 : len(paramStr)-1]
		parts := strings.SplitN(name, ".", 2)
		if len(parts) > 1 {
			return fmt.Errorf("path parameter %q is not a top level field", name)
		}

		prop, ok := requestObject.GetProperty(parts[0])
		if !ok {
			return fmt.Errorf("path parameter %q not found in request object", name)
		}

		prop.Skip = true
		operation.Parameters = append(operation.Parameters, Parameter{
			Name:     name,
			In:       "path",
			Required: true,
			Schema:   prop.SchemaItem,
		})
	}

	if httpOpt.Body == "" {
		// TODO: This should probably be based on the annotation setting of body
		for _, param := range requestObject.Properties {
			operation.Parameters = append(operation.Parameters, Parameter{
				Name:     param.Name,
				In:       "query",
				Required: false,
				Schema:   param.SchemaItem,
			})
		}
	} else if httpOpt.Body == "*" {
		operation.RequestBody = &RequestBody{
			Required: true,
			Content: OperationContent{
				JSON: &OperationSchema{
					Schema: *request,
				},
			},
		}
	} else {
		return fmt.Errorf("unsupported body type %q", httpOpt.Body)
	}

	path, ok := bb.paths[httpPath]
	if !ok {
		path = &PathItem{}
		bb.paths[httpPath] = path
		bb.document.Paths = append(bb.document.Paths, path)
	}
	path.AddOperation(operation)

	return nil
}
