package structure

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func BuildPackages(config *config_j5pb.Config, descFiles *protoregistry.Files, proseResolver ProseResolver) ([]*schema_j5pb.Package, error) {
	services := make([]protoreflect.ServiceDescriptor, 0)

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
		trimPackages: trimSuffixes,
		usedSchemas:  map[protoreflect.FullName]int{},
	}

	wantPackages := make(map[string]bool)
	for _, pkg := range config.Packages {
		wantPackages[pkg.Name] = true

		var prose string

		if pkg.Prose != "" && proseResolver != nil {
			resolved, err := proseResolver.ResolveProse(pkg.Prose)
			if err != nil {
				return nil, fmt.Errorf("prose resolver: package %s: %w", pkg.Name, err)
			}
			prose = removeMarkdownHeader(resolved)
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
			if err := b.addService(service); err != nil {
				return nil, fmt.Errorf("add service: %w", err)
			}
		} else if strings.HasSuffix(name, "Sandbox") {
			if err := b.addService(service); err != nil {
				return nil, fmt.Errorf("add sandbox: %w", err)
			}
		} else if strings.HasSuffix(name, "Events") {
			if err := b.addEvents(service); err != nil {
				return nil, fmt.Errorf("add events: %w", err)
			}
		} else if strings.HasSuffix(name, "Topic") {
		} else {
			return nil, fmt.Errorf("unsupported service name %q", name)
		}

	}

	return b.packages, nil
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

func (bb *builder) addEvents(src protoreflect.ServiceDescriptor) error {
	methods := src.Methods()
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)

		msgFields := method.Input().Fields()

		eventMsg := msgFields.ByJSONName("event")
		if eventMsg == nil {
			return fmt.Errorf("missing event field in %s", method.Input().FullName())
		}

		eventSpec := &schema_j5pb.EventSpec{
			Name:   string(method.Name()),
			Schema: refTo(method.Input()),
		}

		pkg := bb.getPackage(method.ParentFile())

		pkg.Events = append(pkg.Events, eventSpec)

	}
	return nil

}

func refTo(msg protoreflect.MessageDescriptor) *schema_j5pb.Schema {
	return &schema_j5pb.Schema{
		Type: &schema_j5pb.Schema_Ref{
			Ref: string(msg.FullName()),
		},
	}
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

		pkg := bb.getPackage(method.ParentFile())

		pkg.Methods = append(pkg.Methods, builtMethod)

	}
	return nil
}

func (bb *builder) buildMethod(serviceName string, method protoreflect.MethodDescriptor) (*schema_j5pb.Method, error) {

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

	requestFields := method.Input().Fields()

	pathParameters := make([]*schema_j5pb.Parameter, 0)
	pathParts := strings.Split(httpPath, "/")
	for idx, part := range pathParts {
		if part == "" {
			continue
		}

		if part[0] == '{' && part[len(part)-1] == '}' {
			fieldName := part[1 : len(part)-1]
			field := requestFields.ByName(protoreflect.Name(fieldName))
			if field == nil {
				return nil, fmt.Errorf("path parameter %q not found in request object", fieldName)
			}
			pathParameters = append(pathParameters, &schema_j5pb.Parameter{
				Name:     string(field.Name()),
				Required: true,
				// Schema will be resolved at linking time based on the name
			})

			pathParts[idx] = ":" + field.JSONName()
		} else if strings.ContainsAny(part, "{}*:") {
			return nil, fmt.Errorf("invalid path part %q", part)
		}

	}
	newPath := strings.Join(pathParts, "/")

	builtMethod := &schema_j5pb.Method{
		GrpcServiceName: string(method.Parent().Name()),
		GrpcMethodName:  string(method.Name()),
		HttpMethod:      httpMethod,
		HttpPath:        newPath,
		FullGrpcName:    fmt.Sprintf("/%s/%s", serviceName, method.Name()),

		RequestBody:    refTo(method.Input()),
		ResponseBody:   refTo(method.Output()),
		PathParameters: pathParameters,
	}

	return builtMethod, nil
}
