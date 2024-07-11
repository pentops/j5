package j5reflect

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type API struct {
	Packages []*Package
	Metadata *schema_j5pb.Metadata
}

func (api *API) ToJ5Proto() (*schema_j5pb.API, error) {
	// preserves order
	packages := make([]*schema_j5pb.Package, 0, len(api.Packages))
	packageMap := map[string]*schema_j5pb.Package{}

	for _, pkg := range api.Packages {
		apiPkg, err := pkg.ToJ5Proto()
		if err != nil {
			return nil, fmt.Errorf("package %q: %w", pkg.Name, err)
		}

		packages = append(packages, apiPkg)
		packageMap[pkg.Name] = apiPkg
	}

	referencedSchemas, err := collectPackageRefs(api)
	if err != nil {
		return nil, fmt.Errorf("collecting package refs: %w", err)
	}

	for schemaName, schema := range referencedSchemas {
		apiPkg, ok := packageMap[schema.inPackage]
		if ok {
			apiPkg.Schemas[schemaName] = schema.schema
			continue
		}
		refPackage := &schema_j5pb.Package{
			Name: schema.inPackage,
			Schemas: map[string]*schema_j5pb.RootSchema{
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

type Package struct {
	Name     string
	Label    string
	Services []*Service
	Events   []*Event
}

func (pkg *Package) ToJ5Proto() (*schema_j5pb.Package, error) {

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

	events := make([]*schema_j5pb.EventSpec, 0, len(pkg.Events))

	for _, event := range pkg.Events {
		asProto, err := event.Schema.ToJ5Root()
		if err != nil {
			return nil, fmt.Errorf("event schema %q: %w", event.Name, err)
		}
		events = append(events, &schema_j5pb.EventSpec{
			Name:   event.Name,
			Schema: asProto,
		})
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
	return &schema_j5pb.Package{
		Label:    pkg.Label,
		Name:     pkg.Name,
		Schemas:  map[string]*schema_j5pb.RootSchema{},
		Events:   events,
		Services: services,
	}, nil

}

type Event struct {
	Package *Package
	Name    string
	Schema  RootSchema
}

type Service struct {
	Package *Package
	Name    string
	Methods []*Method
}

func (ss *Service) ToJ5Proto() (*schema_j5pb.Service, error) {
	service := &schema_j5pb.Service{
		Name:    ss.Name,
		Methods: make([]*schema_j5pb.Method, len(ss.Methods)),
	}

	for i, method := range ss.Methods {
		m, err := method.ToJ5Proto()
		if err != nil {
			return nil, err
		}
		service.Methods[i] = m
	}

	return service, nil
}

type Method struct {
	GRPCMethodName string
	HTTPPath       string
	HTTPMethod     schema_j5pb.HTTPMethod

	HasBody bool

	Request  RootSchema
	Response RootSchema

	Service *Service
}

func (mm *Method) ToJ5Proto() (*schema_j5pb.Method, error) {

	requestSchema, ok := mm.Request.(*ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("request should be an object, got %T", mm.Request)
	}

	requestBody, err := requestSchema.ToJ5Root()
	if err != nil {
		return nil, err
	}

	requestBodyObject := requestBody.GetObject()
	if requestBodyObject == nil {
		return nil, fmt.Errorf("request body should be an object, got %T", requestBody)
	}

	pathParameterNames := map[string]struct{}{}
	pathParts := strings.Split(mm.HTTPPath, "/")
	for _, part := range pathParts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		fieldName := strings.TrimPrefix(part, ":")
		pathParameterNames[fieldName] = struct{}{}
	}

	pathProperties := make([]*schema_j5pb.ObjectProperty, 0)
	bodyProperties := make([]*schema_j5pb.ObjectProperty, 0)

	for _, prop := range requestBodyObject.Properties {
		_, isPath := pathParameterNames[prop.Name]
		if isPath {
			pathProperties = append(pathProperties, prop)
		} else {
			bodyProperties = append(bodyProperties, prop)
		}
	}

	request := &schema_j5pb.Method_Request{
		PathParameters: pathProperties,
	}

	if mm.HasBody {
		request.Body = &schema_j5pb.RootSchema{
			Type: &schema_j5pb.RootSchema_Object{
				Object: &schema_j5pb.Object{
					Properties:  bodyProperties,
					Name:        requestBodyObject.Name,
					Description: requestBodyObject.Description,
				},
			},
		}
	} else {
		request.QueryParameters = bodyProperties
	}

	responseSchema, ok := mm.Response.(*ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("response schema was not an object: %T", mm.Response)
	}

	responseBody, err := responseSchema.ToJ5Root()
	if err != nil {
		return nil, err
	}

	return &schema_j5pb.Method{
		FullGrpcName: fmt.Sprintf("/%s.%s/%s", mm.Service.Package.Name, mm.Service.Name, mm.GRPCMethodName),
		Name:         mm.GRPCMethodName,
		HttpMethod:   mm.HTTPMethod,
		HttpPath:     mm.HTTPPath,
		ResponseBody: responseBody,
		Request:      request,
	}, nil
}

type schemaRef struct {
	schema    *schema_j5pb.RootSchema
	inPackage string
}

// collectPackageRefs walks the entire API, returning all schemas which are
// accessible via a method, event etc.
func collectPackageRefs(api *API) (map[string]*schemaRef, error) {
	// map[
	schemas := make(map[string]*schemaRef)

	var walkRefRoot func(RootSchema) error

	var walkRefs func(FieldSchema) error
	walkRefRoot = func(schema RootSchema) error {
		_, ok := schemas[schema.FullName()]
		if ok {
			return nil
		}

		schemaProto, err := schema.ToJ5Root()
		if err != nil {
			return fmt.Errorf("schema %s to j5 root: %w", schema.FullName(), err)
		}
		schemas[schema.FullName()] = &schemaRef{
			schema:    schemaProto,
			inPackage: schema.PackageName(),
		}
		switch st := schema.(type) {
		case *ObjectSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk %s: %w", st.FullName(), err)
				}
			}
		case *OneofSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk oneof: %w", err)
				}
			}
		case *EnumSchema:
		// do nothing

		default:
			return fmt.Errorf("unsupported ref type %T", st)
		}
		return nil
	}
	walkRefs = func(schema FieldSchema) error {

		switch st := schema.(type) {
		case *ObjectFieldSchema:
			if err := walkRefRoot(st.Ref.To); err != nil {
				return fmt.Errorf("walk object as field: %w", err)
			}

		case *OneofFieldSchema:
			if err := walkRefRoot(st.Ref.To); err != nil {
				return fmt.Errorf("walk oneof as field: %w", err)
			}

		case *EnumFieldSchema:
			if err := walkRefRoot(st.Ref.To); err != nil {
				return fmt.Errorf("walk enum as field: %w", err)
			}

		case *ArraySchema:
			if err := walkRefs(st.Schema); err != nil {
				return fmt.Errorf("walk array: %w", err)
			}

		case *MapSchema:
			if err := walkRefs(st.Schema); err != nil {
				return fmt.Errorf("walk map: %w", err)
			}
		}

		return nil
	}

	walkRootObject := func(schema RootSchema) error {
		obj, ok := schema.(*ObjectSchema)
		if !ok {
			return fmt.Errorf("expected object schema, got %T", schema)
		}

		for _, prop := range obj.Properties {
			if err := walkRefs(prop.Schema); err != nil {
				return fmt.Errorf("walk root object: %w", err)
			}
		}
		return nil

	}

	for _, pkg := range api.Packages {
		for _, service := range pkg.Services {
			for _, method := range service.Methods {
				if err := walkRootObject(method.Request); err != nil {
					return nil, fmt.Errorf("request schema %q: %w", method.Request.FullName(), err)
				}

				if err := walkRootObject(method.Response); err != nil {
					return nil, fmt.Errorf("response schema %q: %w", method.Response.FullName(), err)
				}
			}
		}
		for _, event := range pkg.Events {
			err := walkRefRoot(event.Schema)
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", event.Schema.FullName(), err)
			}
		}
	}

	return schemas, nil
}
