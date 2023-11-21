package structure

import (
	"fmt"
	"os"
	"path/filepath"
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

	trimPackages []string
}

type Package struct {
	Label  string `json:"label"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`

	Introduction string       `json:"introduction,omitempty"`
	Methods      []*Method    `json:"methods"`
	Events       []*EventSpec `json:"events"`
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

type EventSpec struct {
	Name        string              `json:"name"`
	StateSchema *jsonapi.SchemaItem `json:"stateSchema,omitempty"`
	EventSchema *jsonapi.SchemaItem `json:"eventSchema,omitempty"`
}

type Parameter struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Schema      jsonapi.SchemaItem `json:"schema"`
}

type ProseResolver interface {
	ResolveProse(filename string) (string, error)
}

type DirResolver string

func (dr DirResolver) ResolveProse(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(string(dr), filename))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func BuildFromDescriptors(config *jsonapi_pb.Config, descriptors *descriptorpb.FileDescriptorSet, proseResolver ProseResolver) (*Built, error) {

	codecOptions := jsonapi.Options{
		ShortEnums: &jsonapi.ShortEnumsOption{
			UnspecifiedSuffix: "UNSPECIFIED",
			StrictUnmarshal:   true,
		},
		WrapOneof: config.Options.WrapOneof,
	}

	if config.Options.ShortEnums != nil {
		codecOptions.ShortEnums = &jsonapi.ShortEnumsOption{
			UnspecifiedSuffix: config.Options.ShortEnums.UnspecifiedSuffix,
			StrictUnmarshal:   config.Options.ShortEnums.StrictUnmarshal,
		}
	}

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

	trimSuffixes := make([]string, len(config.Options.TrimSubPackages))
	for idx, suffix := range config.Options.TrimSubPackages {
		trimSuffixes[idx] = "." + suffix
	}

	b := builder{
		schemas:      jsonapi.NewSchemaSet(codecOptions),
		trimPackages: trimSuffixes,
	}

	wantPackages := make(map[string]bool)
	for _, pkg := range config.Packages {
		wantPackages[pkg.Name] = true

		var prose string

		if pkg.Prose != "" && proseResolver != nil {
			prose, err = proseResolver.ResolveProse(pkg.Prose)
			if err != nil {
				return nil, fmt.Errorf("package %s: %w", pkg.Name, err)
			}
		}

		b.packages = append(b.packages, &Package{
			Name:         pkg.Name,
			Label:        pkg.Label,
			Introduction: prose,
		})
	}

	for _, service := range services {
		name := string(service.FullName())
		packageName := string(service.ParentFile().Package())

		for _, suffix := range b.trimPackages {
			packageName = strings.TrimSuffix(packageName, suffix)
		}

		if !wantPackages[packageName] {
			continue
		}

		if strings.HasSuffix(name, "Service") {
			if err := b.addService(service); err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(name, "Events") {
			if err := b.addEvents(service); err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(name, "Topic") {
		} else {
			return nil, fmt.Errorf("unsupported service name %q", name)
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

	for _, trimSuffix := range bb.trimPackages {
		name = strings.TrimSuffix(name, trimSuffix)
	}

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

func (bb *builder) addEvents(src protoreflect.ServiceDescriptor) error {
	methods := src.Methods()
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)

		msgFields := method.Input().Fields()

		eventMsg := msgFields.ByJSONName("event")
		if eventMsg == nil {
			return fmt.Errorf("missing event field in %s", method.Input().FullName())
		}

		eventSchema, err := bb.schemas.BuildSchemaObject(eventMsg.Message())
		if err != nil {
			return err
		}

		eventSpec := &EventSpec{
			Name:        string(method.Name()),
			EventSchema: eventSchema,
		}

		stateMsg := msgFields.ByJSONName("state")
		if stateMsg != nil {

			stateSchema, err := bb.schemas.BuildSchemaObject(stateMsg.Message())
			if err != nil {
				return err
			}
			eventSpec.StateSchema = stateSchema
		}

		pkg, err := bb.getPackage(method.ParentFile())
		if err != nil {
			return err
		}

		pkg.Events = append(pkg.Events, eventSpec)

	}
	return nil

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

func convertPath(path string, requestObject protoreflect.MessageDescriptor) (string, error) {
	parts := strings.Split(path, "/")
	requestFields := requestObject.Fields()
	for idx, part := range parts {
		if part == "" {
			continue
		}

		if part[0] == '{' && part[len(part)-1] == '}' {
			fieldName := part[1 : len(part)-1]
			field := requestFields.ByName(protoreflect.Name(fieldName))
			if field == nil {
				return "", fmt.Errorf("path parameter %q not found in request object", fieldName)
			}

			parts[idx] = ":" + field.JSONName()
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

	converted, err := convertPath(httpPath, method.Input())
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

	requestObject := request.ItemType.(*jsonapi.ObjectItem)

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
			Name:     prop.Name,
			Required: true,
			Schema:   prop.SchemaItem,
		})
	}

	if httpOpt.Body == "" {
		// TODO: This should probably be based on the annotation setting of body
		for _, param := range requestObject.Properties {
			builtMethod.QueryParameters = append(builtMethod.QueryParameters, &Parameter{
				Name:     param.Name,
				Required: false,
				Schema:   param.SchemaItem,
			})
		}
	} else if httpOpt.Body == "*" {
		request.ItemType = requestObject
		builtMethod.RequestBody = request
	} else {
		return nil, fmt.Errorf("unsupported body type %q", httpOpt.Body)
	}

	return builtMethod, nil
}
