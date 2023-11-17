package structure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pentops/custom-proto-api/gen/v1/jsonapi_pb"
	"github.com/pentops/custom-proto-api/jsonapi"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Built struct {
	Packages []*Package                     `json:"packages"`
	Schemas  map[string]*jsonapi.SchemaItem `json:"schemas"`
}

type builder struct {
	schemas  *jsonapi.SchemaSet
	packages []*Package
}

type Package struct {
	Label  string `json:"label"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`

	Introduction string    `json:"introduction,omitempty"`
	Methods      []*Method `json:"methods"`
	Entities     []*Entity `json:"entities"`
}

type Method struct {
	GrpcServiceName string `json:"grpcServiceName"`
	GrpcMethodName  string `json:"grpcMethodName"`
	FullGrpcName    string `json:"fullGrpcName"`

	HTTPMethod      string              `json:"httpMethod"`
	HTTPPath        string              `json:"httpPath"`
	RequestBody     *jsonapi.SchemaItem `json:"requestBody,omitempty"`
	ResponseBody    *jsonapi.SchemaItem `json:"responseBody,omitempty"`
	QueryParameters []*Parameter        `json:"queryParameters,omitempty"`
	PathParameters  []*Parameter        `json:"pathParameters,omitempty"`
}

type Entity struct {
	StateSchema *jsonapi.SchemaItem `json:"stateSchema,omitempty"`
	EventSchema *jsonapi.SchemaItem `json:"eventSchema,omitempty"`
}

type Parameter struct {
	Name        string             `json:"name"`
	In          string             `json:"in"`
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Schema      jsonapi.SchemaItem `json:"schema"`
}

func BuildFromDescriptors(options jsonapi.Options, descriptors *descriptorpb.FileDescriptorSet) (*Built, error) {

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

func Build(options jsonapi.Options, services []protoreflect.ServiceDescriptor) (*Built, error) {
	b := builder{
		packages: make([]*Package, 0),
		schemas:  jsonapi.NewSchemaSet(options),
	}

	for _, service := range services {
		if err := b.addService(service); err != nil {
			return nil, err
		}
	}

	bb := &Built{
		Packages: b.packages,
		Schemas:  b.schemas.Schemas,
	}

	return bb, nil
}

func (bb *builder) getPackage(file protoreflect.FileDescriptor) (*Package, error) {

	name := string(file.Package())

	name = strings.TrimSuffix(name, ".service")
	name = strings.TrimSuffix(name, ".topic")

	var pkg *Package
	for _, search := range bb.packages {
		if search.Name == name {
			pkg = search
			break
		}
	}

	if pkg == nil {
		pkg = &Package{
			Name: name,
		}
		bb.packages = append(bb.packages, pkg)

	}

	packageOptions := proto.GetExtension(file.Options(), jsonapi_pb.E_Package).(*jsonapi_pb.PackageOptions)
	if packageOptions != nil {
		if packageOptions.Label != "" {
			if pkg.Label != "" && pkg.Label != packageOptions.Label {
				return nil, fmt.Errorf("package %q has conflicting labels %q and %q", name, pkg.Label, packageOptions.Label)
			}
			pkg.Label = packageOptions.Label
		}
	}

	return pkg, nil
}

func (bb *builder) addService(src protoreflect.ServiceDescriptor) error {
	methods := src.Methods()
	name := string(src.FullName())
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)
		builtMethod, err := bb.buildMethod(name, method)
		if err != nil {
			return err
		}

		pkg, err := bb.getPackage(method.ParentFile())
		if err != nil {
			return err
		}

		pkg.Methods = append(pkg.Methods, builtMethod)

	}
	return nil
}

var rePathParameter = regexp.MustCompile(`\{([^\}]+)\}`)

func convertPath(path string) (string, error) {
	parts := strings.Split(path, "/")
	for idx, part := range parts {
		if part == "" {
			continue
		}

		if part[0] == '{' && part[len(part)-1] == '}' {
			part = ":" + part[1:len(part)-1]
			parts[idx] = part
		} else if strings.ContainsAny(part, "{}*:") {
			return "", fmt.Errorf("invalid path part %q", part)
		}

	}
	return strings.Join(parts, "/"), nil
}
func (bb *builder) buildMethod(serviceName string, method protoreflect.MethodDescriptor) (*Method, error) {

	methodOptions := method.Options().(*descriptorpb.MethodOptions)
	httpOpt := proto.GetExtension(methodOptions, annotations.E_Http).(*annotations.HttpRule)

	var httpMethod string
	var httpPath string

	if httpOpt == nil {
		return nil, fmt.Errorf("missing http rule for method /%s/%s", serviceName, method.Name())
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
		return nil, fmt.Errorf("unsupported http method %T", pt)
	}

	converted, err := convertPath(httpPath)
	if err != nil {
		return nil, err
	}

	builtMethod := &Method{
		GrpcServiceName: string(method.Parent().Name()),
		GrpcMethodName:  string(method.Name()),
		HTTPMethod:      httpMethod,
		HTTPPath:        converted,
		FullGrpcName:    fmt.Sprintf("/%s/%s", serviceName, method.Name()),
	}

	okResponse, err := bb.schemas.BuildSchemaObject(method.Output())
	if err != nil {
		return nil, err
	}

	builtMethod.ResponseBody = okResponse

	request, err := bb.schemas.BuildSchemaObject(method.Input())
	if err != nil {
		return nil, err
	}

	requestObjectRaw := request.ItemType.(jsonapi.ObjectItem)
	requestObject := &requestObjectRaw

	for _, paramStr := range rePathParameter.FindAllString(httpPath, -1) {
		name := paramStr[1 : len(paramStr)-1]
		parts := strings.SplitN(name, ".", 2)
		if len(parts) > 1 {
			return nil, fmt.Errorf("path parameter %q is not a top level field", name)
		}

		prop, ok := requestObject.PopProperty(parts[0])
		if !ok {
			return nil, fmt.Errorf("path parameter %q not found in request object", name)
		}

		prop.Skip = true
		builtMethod.PathParameters = append(builtMethod.PathParameters, &Parameter{
			Name:     prop.ProtoFieldName, // Special case for Path Parameters
			Required: true,
			Schema:   prop.SchemaItem,
		})
	}

	if httpOpt.Body == "" {
		// TODO: This should probably be based on the annotation setting of body
		for _, param := range requestObject.Properties {
			builtMethod.QueryParameters = append(builtMethod.QueryParameters, &Parameter{
				Name:     param.Name,
				In:       "query",
				Required: false,
				Schema:   param.SchemaItem,
			})
		}
	} else if httpOpt.Body == "*" {
		request.ItemType = *requestObject
		builtMethod.RequestBody = request
	} else {
		return nil, fmt.Errorf("unsupported body type %q", httpOpt.Body)
	}

	return builtMethod, nil
}
