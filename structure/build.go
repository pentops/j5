package structure

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type builder struct {
	schemas  *SchemaSet
	packages []*jsonapi_pb.Package

	trimPackages []string
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

type mapResolver map[string]string

func (mr mapResolver) ResolveProse(filename string) (string, error) {
	data, ok := mr[filename]
	if !ok {
		return "", fmt.Errorf("prose file %q not found", filename)
	}
	return data, nil
}

func removeMarkdownHeader(data string) string {
	// only look at the first 5 lines, that should be well enough to deal with
	// both title formats (# or \n===), and a few trailing empty lines

	lines := strings.SplitN(data, "\n", 5)
	if len(lines) == 0 {
		return ""
	}
	if strings.HasPrefix(lines[0], "# ") {
		lines = lines[1:]
	} else if strings.HasPrefix(lines[1], "==") {
		lines = lines[2:]
	}

	// Remove any leading empty lines
	for len(lines) > 1 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	return strings.Join(lines, "\n")
}

func imageResolver(proseFiles []*jsonapi_pb.ProseFile) ProseResolver {
	mr := make(mapResolver)
	for _, proseFile := range proseFiles {
		mr[proseFile.Path] = string(proseFile.Content)
	}
	return mr
}

func BuildFromImage(image *jsonapi_pb.Image) (*jsonapi_pb.API, error) {
	proseResolver := imageResolver(image.Prose)

	descriptors := &descriptorpb.FileDescriptorSet{
		File: image.File,
	}

	config := &config_j5pb.Config{
		Packages: image.Packages,
		Options:  image.Codec,
	}

	return BuildFromDescriptors(config, descriptors, proseResolver)
}

func BuildFromDescriptors(config *config_j5pb.Config, descriptors *descriptorpb.FileDescriptorSet, proseResolver ProseResolver) (*jsonapi_pb.API, error) {

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
		schemas:      NewSchemaSet(config.Options),
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
			prose = removeMarkdownHeader(prose)
		}

		b.packages = append(b.packages, &jsonapi_pb.Package{
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
		} else if strings.HasSuffix(name, "Sandbox") {
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

	schemas := map[string]*jsonapi_pb.Schema{}
	for name, schema := range b.schemas.Schemas {
		schemas[name] = schema
	}
	bb := &jsonapi_pb.API{
		Packages: b.packages,
		Schemas:  schemas,
	}

	return bb, nil
}

func (bb *builder) getPackage(file protoreflect.FileDescriptor) (*jsonapi_pb.Package, error) {

	name := string(file.Package())

	for _, trimSuffix := range bb.trimPackages {
		name = strings.TrimSuffix(name, trimSuffix)
	}

	var pkg *jsonapi_pb.Package
	for _, search := range bb.packages {
		if search.Name == name {
			pkg = search
			break
		}
	}

	if pkg == nil {
		pkg = &jsonapi_pb.Package{
			Name: name,
		}
		bb.packages = append(bb.packages, pkg)
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

		eventSpec := &jsonapi_pb.EventSpec{
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

func (bb *builder) buildMethod(serviceName string, method protoreflect.MethodDescriptor) (*jsonapi_pb.Method, error) {

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

	builtMethod := &jsonapi_pb.Method{
		GrpcServiceName: string(method.Parent().Name()),
		GrpcMethodName:  string(method.Name()),
		HttpMethod:      httpMethod,
		HttpPath:        converted,
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

	requestObject := request.GetObjectItem()

	for _, paramStr := range rePathParameter.FindAllString(httpPath, -1) {
		name := paramStr[1 : len(paramStr)-1]
		parts := strings.SplitN(name, ".", 2)
		if len(parts) > 1 {
			return nil, fmt.Errorf("path parameter %q is not a top level field", name)
		}

		prop, ok := popProperty(requestObject, parts[0])
		if !ok {
			return nil, fmt.Errorf("path parameter %q not found in request object", name)
		}

		builtMethod.PathParameters = append(builtMethod.PathParameters, &jsonapi_pb.Parameter{
			Name:     prop.Name,
			Required: true,
			Schema:   prop.Schema,
		})
	}

	if httpOpt.Body == "" {
		// TODO: This should probably be based on the annotation setting of body
		for _, param := range requestObject.Properties {
			builtMethod.QueryParameters = append(builtMethod.QueryParameters, &jsonapi_pb.Parameter{
				Name:     param.Name,
				Required: false,
				Schema:   param.Schema,
			})
		}
	} else if httpOpt.Body == "*" {
		request.Type = &jsonapi_pb.Schema_ObjectItem{
			ObjectItem: requestObject,
		}
		builtMethod.RequestBody = request
	} else {
		return nil, fmt.Errorf("unsupported body type %q", httpOpt.Body)
	}

	return builtMethod, nil
}

func popProperty(obj *jsonapi_pb.ObjectItem, name string) (*jsonapi_pb.ObjectProperty, bool) {
	newProps := make([]*jsonapi_pb.ObjectProperty, 0, len(obj.Properties)-1)
	var found *jsonapi_pb.ObjectProperty
	for _, prop := range obj.Properties {
		if prop.ProtoFieldName == name {
			found = prop
			continue
		}
		newProps = append(newProps, prop)
	}
	obj.Properties = newProps
	return found, found != nil
}
