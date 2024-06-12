package structure

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func BuildFromImage(ctx context.Context, image *source_j5pb.SourceImage) (*schema_j5pb.API, error) {
	proseResolver := imageResolver(image.Prose)

	descriptors := &descriptorpb.FileDescriptorSet{
		File: image.File,
	}

	if len(descriptors.File) < 1 {
		panic("Expected at least one descriptor file, found none")
	}

	config := &config_j5pb.Config{
		Packages: image.Packages,
		Options:  image.Codec,
	}

	if config.Packages == nil || config.Options == nil {
		panic("Expected Packages or Configs from image, found none")
	}

	return BuildFromDescriptors(ctx, config, descriptors, proseResolver)
}

type builder struct {
	schemas      *SchemaSet
	packages     []*schema_j5pb.Package
	trimPackages []string
}

func (bb *builder) warn(ctx context.Context, format string, args ...interface{}) {
	log.WithField(ctx, "message", fmt.Sprintf(format, args...)).Warn("J5 Parser Warning")
}

func BuildFromDescriptors(ctx context.Context, config *config_j5pb.Config, descriptors *descriptorpb.FileDescriptorSet, proseResolver ProseResolver) (*schema_j5pb.API, error) {
	services := make([]protoreflect.ServiceDescriptor, 0)
	descFiles, err := protodesc.NewFiles(descriptors)
	if err != nil {
		return nil, fmt.Errorf("descriptor files: %w", err)
	}

	descFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		fileServices := file.Services()
		for ii := 0; ii < fileServices.Len(); ii++ {
			service := fileServices.Get(ii)
			services = append(services, service)
		}
		return true
	})

	if config.Options == nil {
		config.Options = &config_j5pb.CodecOptions{}
	}

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
				return nil, fmt.Errorf("prose resolver: package %s: %w", pkg.Name, err)
			}
			prose = removeMarkdownHeader(prose)
		}

		b.packages = append(b.packages, &schema_j5pb.Package{
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
			if err := b.addService(ctx, service); err != nil {
				return nil, fmt.Errorf("add service: %w", err)
			}
		} else if strings.HasSuffix(name, "Sandbox") {
			if err := b.addService(ctx, service); err != nil {
				return nil, fmt.Errorf("add sandbox: %w", err)
			}
		} else if strings.HasSuffix(name, "Events") {
			if err := b.addEvents(ctx, service); err != nil {
				return nil, fmt.Errorf("add events: %w", err)
			}
		} else if strings.HasSuffix(name, "Topic") {
		} else {
			return nil, fmt.Errorf("unsupported service name %q", name)
		}

	}

	schemas := map[string]*schema_j5pb.Schema{}
	for name, schema := range b.schemas.Schemas {
		schemas[name] = schema
	}
	bb := &schema_j5pb.API{
		Packages: b.packages,
		Schemas:  schemas,
	}

	return bb, nil
}

func (bb *builder) getPackage(file protoreflect.FileDescriptor) *schema_j5pb.Package {

	name := string(file.Package())

	for _, trimSuffix := range bb.trimPackages {
		name = strings.TrimSuffix(name, trimSuffix)
	}

	var pkg *schema_j5pb.Package
	for _, search := range bb.packages {
		if search.Name == name {
			pkg = search
			break
		}
	}

	if pkg == nil {
		pkg = &schema_j5pb.Package{
			Name: name,
		}
		bb.packages = append(bb.packages, pkg)
	}

	return pkg
}

func (bb *builder) addEvents(ctx context.Context, src protoreflect.ServiceDescriptor) error {
	methods := src.Methods()
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)

		msgFields := method.Input().Fields()

		eventMsg := msgFields.ByJSONName("event")
		if eventMsg == nil {
			return fmt.Errorf("missing event field in %s", method.Input().FullName())
		}

		eventSchema, err := bb.schemas.BuildSchemaObject(ctx, eventMsg.Message())
		if err != nil {
			return err
		}

		eventSpec := &schema_j5pb.EventSpec{
			Name:        string(method.Name()),
			EventSchema: eventSchema,
		}

		stateMsg := msgFields.ByJSONName("state")
		if stateMsg != nil {

			stateSchema, err := bb.schemas.BuildSchemaObject(ctx, stateMsg.Message())
			if err != nil {
				return err
			}
			eventSpec.StateSchema = stateSchema
		}

		pkg := bb.getPackage(method.ParentFile())

		pkg.Events = append(pkg.Events, eventSpec)

	}
	return nil

}

func (bb *builder) addService(ctx context.Context, src protoreflect.ServiceDescriptor) error {
	methods := src.Methods()
	name := string(src.FullName())
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)
		builtMethod, err := bb.buildMethod(ctx, name, method)
		if err != nil {
			return err
		}

		pkg := bb.getPackage(method.ParentFile())

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

func (bb *builder) buildMethod(ctx context.Context, serviceName string, method protoreflect.MethodDescriptor) (*schema_j5pb.Method, error) {

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

	builtMethod := &schema_j5pb.Method{
		GrpcServiceName: string(method.Parent().Name()),
		GrpcMethodName:  string(method.Name()),
		HttpMethod:      httpMethod,
		HttpPath:        converted,
		FullGrpcName:    fmt.Sprintf("/%s/%s", serviceName, method.Name()),
	}

	okResponse, err := bb.schemas.BuildSchemaObject(ctx, method.Output())
	if err != nil {
		return nil, err
	}

	builtMethod.ResponseBody = okResponse

	request, err := bb.schemas.BuildSchemaObject(ctx, method.Input())
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

		builtMethod.PathParameters = append(builtMethod.PathParameters, &schema_j5pb.Parameter{
			Name:     prop.Name,
			Required: true,
			Schema:   prop.Schema,
		})
	}

	if httpOpt.Body == "" {
		if httpMethod != "get" {
			bb.warn(ctx, "no body annotation for %s.%s which is a %s", serviceName, method.Name(), httpMethod)
		}
		// TODO: This should probably be based on the annotation setting of body
		for _, param := range requestObject.Properties {
			builtMethod.QueryParameters = append(builtMethod.QueryParameters, &schema_j5pb.Parameter{
				Name:     param.Name,
				Required: false,
				Schema:   param.Schema,
			})
		}
	} else if httpOpt.Body == "*" {
		request.Type = &schema_j5pb.Schema_ObjectItem{
			ObjectItem: requestObject,
		}
		builtMethod.RequestBody = request
	} else {
		return nil, fmt.Errorf("unsupported body type %q", httpOpt.Body)
	}

	return builtMethod, nil
}

func popProperty(obj *schema_j5pb.ObjectItem, name string) (*schema_j5pb.ObjectProperty, bool) {
	newProps := make([]*schema_j5pb.ObjectProperty, 0, len(obj.Properties)-1)
	var found *schema_j5pb.ObjectProperty
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
