package structure

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ReflectFromSource(image *source_j5pb.SourceImage) (*j5reflect.API, error) {

	descFiles, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: image.File,
	})
	if err != nil {
		return nil, err
	}

	services := make([]protoreflect.ServiceDescriptor, 0)

	descFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		fileServices := file.Services()
		for ii := 0; ii < fileServices.Len(); ii++ {
			service := fileServices.Get(ii)
			services = append(services, service)
		}
		return true
	})

	refs := NewSchemaResolver(descFiles)

	if image.Options == nil {
		image.Options = &config_j5pb.CodecOptions{}
	}

	trimSuffixes := make([]string, len(image.Options.TrimSubPackages))
	for idx, suffix := range image.Options.TrimSubPackages {
		trimSuffixes[idx] = "." + suffix
	}

	b := packageSet{
		trimPackages: trimSuffixes,
	}

	wantPackages := make(map[string]bool)
	for _, pkg := range image.Packages {
		wantPackages[pkg.Name] = true

		b.packages = append(b.packages, &j5reflect.Package{
			Name:  pkg.Name,
			Label: pkg.Label,
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

		pkg := b.getPackage(service.ParentFile())

		if strings.HasSuffix(name, "Service") || strings.HasSuffix(name, "Sandbox") {
			built, err := buildService(refs, service)
			if err != nil {
				return nil, fmt.Errorf("add service %s: %w", name, err)
			}
			built.Package = pkg
			pkg.Services = append(pkg.Services, built)
		} else if strings.HasSuffix(name, "Events") {
			events, err := buildEvents(refs, service)
			if err != nil {
				return nil, fmt.Errorf("add events: %w", err)
			}
			for _, evt := range events {
				evt.Package = pkg
			}
			pkg.Events = append(pkg.Events, events...)
		} else if strings.HasSuffix(name, "Topic") {
			// ignore.
		} else {
			return nil, fmt.Errorf("unsupported service name %q", name)
		}

	}

	return &j5reflect.API{
		Packages: b.packages,
	}, nil
}

type packageSet struct {
	packages     []*j5reflect.Package
	trimPackages []string
}

func (bb *packageSet) getPackage(file protoreflect.FileDescriptor) *j5reflect.Package {

	name := string(file.Package())

	for _, trimSuffix := range bb.trimPackages {
		name = strings.TrimSuffix(name, trimSuffix)
	}

	var pkg *j5reflect.Package
	for _, search := range bb.packages {
		if search.Name == name {
			pkg = search
			break
		}
	}

	if pkg == nil {
		pkg = &j5reflect.Package{
			Name: name,
		}
		bb.packages = append(bb.packages, pkg)
	}

	return pkg
}

func buildEvents(refs *SchemaResolver, src protoreflect.ServiceDescriptor) ([]*j5reflect.Event, error) {
	events := make([]*j5reflect.Event, 0)
	methods := src.Methods()
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)

		schema, err := refs.SchemaReflect(method.Input())
		if err != nil {
			return nil, fmt.Errorf("method %s: %w", method.FullName(), err)
		}

		eventSpec := &j5reflect.Event{
			Name:   string(method.Name()),
			Schema: schema,
		}

		events = append(events, eventSpec)
	}
	return events, nil
}

func buildService(refs *SchemaResolver, src protoreflect.ServiceDescriptor) (*j5reflect.Service, error) {
	methods := src.Methods()
	service := &j5reflect.Service{
		Name:    string(src.Name()),
		Methods: make([]*j5reflect.Method, 0, methods.Len()),
	}
	for ii := 0; ii < methods.Len(); ii++ {
		method := methods.Get(ii)
		builtMethod, err := buildMethod(refs, method)
		if err != nil {
			return nil, fmt.Errorf("build method %s: %w", method.FullName(), err)
		}
		service.Methods = append(service.Methods, builtMethod)
		builtMethod.Service = service

	}
	return service, nil
}

func buildMethod(refs *SchemaResolver, method protoreflect.MethodDescriptor) (*j5reflect.Method, error) {

	methodOptions := method.Options().(*descriptorpb.MethodOptions)
	httpOpt := proto.GetExtension(methodOptions, annotations.E_Http).(*annotations.HttpRule)

	if httpOpt == nil {
		return nil, fmt.Errorf("missing http rule")
	}

	request, err := refs.SchemaReflect(method.Input())
	if err != nil {
		return nil, err
	}

	response, err := refs.SchemaReflect(method.Output())
	if err != nil {
		return nil, err
	}

	builtMethod := &j5reflect.Method{
		GRPCMethodName: string(method.Name()),
		Request:        request,
		Response:       response,
	}

	switch pt := httpOpt.Pattern.(type) {
	case *annotations.HttpRule_Get:
		builtMethod.HTTPMethod = schema_j5pb.HTTPMethod_HTTP_METHOD_GET
		builtMethod.HTTPPath = pt.Get
		builtMethod.HasBody = false

	case *annotations.HttpRule_Post:
		builtMethod.HTTPMethod = schema_j5pb.HTTPMethod_HTTP_METHOD_POST
		builtMethod.HTTPPath = pt.Post
		builtMethod.HasBody = true

	case *annotations.HttpRule_Put:
		builtMethod.HTTPMethod = schema_j5pb.HTTPMethod_HTTP_METHOD_PUT
		builtMethod.HTTPPath = pt.Put
		builtMethod.HasBody = true

	case *annotations.HttpRule_Delete:
		builtMethod.HTTPMethod = schema_j5pb.HTTPMethod_HTTP_METHOD_DELETE
		builtMethod.HTTPPath = pt.Delete
		builtMethod.HasBody = true

	case *annotations.HttpRule_Patch:
		builtMethod.HTTPMethod = schema_j5pb.HTTPMethod_HTTP_METHOD_PATCH
		builtMethod.HTTPPath = pt.Patch
		builtMethod.HasBody = true

	default:
		return nil, fmt.Errorf("unsupported http method %T", pt)
	}

	pathParts := strings.Split(builtMethod.HTTPPath, "/")
	for idx, part := range pathParts {
		if part == "" {
			continue
		}
		if part[0] == '{' && part[len(part)-1] == '}' {
			fieldName := part[1 : len(part)-1]
			inputField := method.Input().Fields().ByName(protoreflect.Name(fieldName))
			if inputField == nil {
				return nil, fmt.Errorf("path field %q not found in input", fieldName)
			}
			jsonName := inputField.JSONName()
			pathParts[idx] = ":" + jsonName

		} else if strings.ContainsAny(part, "{}*:") {
			return nil, fmt.Errorf("invalid path part %q", part)
		}

	}
	builtMethod.HTTPPath = strings.Join(pathParts, "/")

	return builtMethod, nil
}
