package structure

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/schema/j5reflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type schemaRef struct {
	schema    *schema_j5pb.Schema
	inPackage string
}

func collectPackageRefs(api *j5reflect.API) (map[string]*schemaRef, error) {
	schemas := make(map[string]*schemaRef)

	var walkRefs func(j5reflect.Schema) error
	walkRefs = func(schema j5reflect.Schema) error {

		switch st := schema.(type) {
		case *j5reflect.ObjectSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk %s: %w", st.FullName(), err)
				}
			}

		case *j5reflect.ArraySchema:
			if err := walkRefs(st.Schema); err != nil {
				return fmt.Errorf("walk array: %w", err)
			}

		case *j5reflect.OneofSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk oneof: %w", err)
				}
			}

		case *j5reflect.MapSchema:
			if err := walkRefs(st.Schema); err != nil {
				return fmt.Errorf("walk map: %w", err)
			}

		case *j5reflect.RefSchema:
			stringName := st.FullName()
			if _, ok := schemas[stringName]; ok {
				return nil
			}
			if st.To == nil {
				return fmt.Errorf("unlinked schema %s", st.FullName())
			}
			asProto, err := st.To.ToJ5Proto()
			if err != nil {
				return fmt.Errorf("ref schema %q.%q: %w", st.Package, st.Schema, err)
			}
			schemas[stringName] = &schemaRef{
				schema:    asProto,
				inPackage: st.Package,
			}
			if err := walkRefs(st.To); err != nil {
				return fmt.Errorf("walk ref %q.%q: %w", st.Package, st.Schema, err)
			}
		}

		return nil
	}

	for _, pkg := range api.Packages {
		for _, service := range pkg.Services {
			for _, method := range service.Methods {
				if err := walkRefs(method.Request); err != nil {
					return nil, fmt.Errorf("request schema %q: %w", method.Request.FullName(), err)
				}

				if err := walkRefs(method.Response); err != nil {
					return nil, fmt.Errorf("response schema %q: %w", method.Response.FullName(), err)
				}
			}
		}
		for _, event := range pkg.Events {
			err := walkRefs(event.Schema)
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", event.Schema.FullName(), err)
			}
		}
	}

	return schemas, nil
}

func DescriptorFromReflection(api *j5reflect.API) (*schema_j5pb.API, error) {

	// preserves order
	packages := make([]*schema_j5pb.Package, 0, len(api.Packages))
	packageMap := map[string]*schema_j5pb.Package{}

	for _, pkg := range api.Packages {

		services := make([]*schema_j5pb.Service, 0, len(pkg.Services))
		for _, service := range pkg.Services {
			methods := make([]*schema_j5pb.Method, 0, len(service.Methods))
			for _, method := range service.Methods {
				m, err := method.ToJ5Proto()
				if err != nil {
					return nil, fmt.Errorf("method %s/%s: %w", service.Name, method.GRPCMethodName, err)
				}
				m.Name = method.GRPCMethodName
				m.FullGrpcName = fmt.Sprintf("/%s.%s/%s", pkg.Name, service.Name, method.GRPCMethodName)
				methods = append(methods, m)
			}
			services = append(services, &schema_j5pb.Service{
				Name:    service.Name,
				Methods: methods,
			})
		}

		events := make([]*schema_j5pb.EventSpec, len(pkg.Events))

		for _, event := range pkg.Events {
			asProto, err := event.Schema.ToJ5Proto()
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", event.Name, err)
			}
			events = append(events, &schema_j5pb.EventSpec{
				Name:   event.Name,
				Schema: asProto,
			})
		}

		apiPkg := &schema_j5pb.Package{
			Label:    pkg.Label,
			Name:     pkg.Name,
			Schemas:  map[string]*schema_j5pb.Schema{},
			Events:   events,
			Services: services,
		}
		packages = append(packages, apiPkg)
		packageMap[pkg.Name] = apiPkg
	}

	/*
		for _, entity := range pkg.Entities {
			eventObject, err := rootResolve(entity.Schema)
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", entity.Schema.GetRef(), err)
			}
			entity.Schema, err = eventObject.ToJ5Proto()
			if err != nil {
				return nil, err
			}
		}*/
	referencedSchemas, err := collectPackageRefs(api)
	if err != nil {
		return nil, err
	}

	for schemaName, schema := range referencedSchemas {
		apiPkg, ok := packageMap[schema.inPackage]
		if ok {
			apiPkg.Schemas[schemaName] = schema.schema
			continue
		}
		refPackage := &schema_j5pb.Package{
			Name: schema.inPackage,
			Schemas: map[string]*schema_j5pb.Schema{
				schemaName: schema.schema,
			},
		}
		packageMap[schema.inPackage] = refPackage
		packages = append(packages, refPackage)
	}
	return &schema_j5pb.API{
		Packages: packages,
		Metadata: &schema_j5pb.Metadata{
			BuiltAt: timestamppb.Now(),
		},
	}, nil
}
