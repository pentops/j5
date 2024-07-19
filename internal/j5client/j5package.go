package j5client

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type API struct {
	Packages []*Package
	Metadata *client_j5pb.Metadata
}

type schemaSource interface {
	AnonymousObjectFromSchema(packageName string, schema *schema_j5pb.Object) (*j5reflect.ObjectSchema, error)
}

func (api *API) ToJ5Proto() (*client_j5pb.API, error) {
	// preserves order
	packages := make([]*client_j5pb.Package, 0, len(api.Packages))
	packageMap := map[string]*client_j5pb.Package{}

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
		refPackage := &client_j5pb.Package{
			Name: schema.inPackage,
			Schemas: map[string]*schema_j5pb.RootSchema{
				schemaName: schema.schema,
			},
		}
		packageMap[schema.inPackage] = refPackage
		packages = append(packages, refPackage)
	}
	return &client_j5pb.API{
		Packages: packages,
		Metadata: &client_j5pb.Metadata{
			BuiltAt: timestamppb.Now(),
		},
	}, nil
}

type Package struct {
	Name          string
	Label         string
	Services      []*Service
	StateEntities []*StateEntity
}

func (pkg *Package) ToJ5Proto() (*client_j5pb.Package, error) {

	services := make([]*client_j5pb.Service, 0, len(pkg.Services))
	for _, service := range pkg.Services {
		methods := make([]*client_j5pb.Method, 0, len(service.Methods))
		for _, method := range service.Methods {
			m, err := method.ToJ5Proto()
			if err != nil {
				return nil, fmt.Errorf("method %s/%s: %w", service.Name, method.GRPCMethodName, err)
			}
			m.Name = method.GRPCMethodName
			m.FullGrpcName = fmt.Sprintf("/%s.%s/%s", pkg.Name, service.Name, method.GRPCMethodName)
			methods = append(methods, m)
		}
		services = append(services, &client_j5pb.Service{
			Name:    service.Name,
			Methods: methods,
		})
	}

	return &client_j5pb.Package{
		Label:    pkg.Label,
		Name:     pkg.Name,
		Schemas:  map[string]*schema_j5pb.RootSchema{},
		Services: services,
	}, nil

}

type Service struct {
	Package *Package
	Name    string
	Methods []*Method
}

func (ss *Service) ToJ5Proto() (*client_j5pb.Service, error) {
	service := &client_j5pb.Service{
		Name:    ss.Name,
		Methods: make([]*client_j5pb.Method, len(ss.Methods)),
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

type SchemaLink interface {
	FullName() string
	Schema() *j5reflect.ObjectSchema
	link(schemaSource) error
}

type Method struct {
	GRPCMethodName string
	HTTPPath       string
	HTTPMethod     client_j5pb.HTTPMethod

	HasBody bool

	Request      *Request
	ResponseBody *j5reflect.ObjectSchema

	Service *Service
}

type Request struct {
	Body            *j5reflect.ObjectSchema
	PathParameters  []*j5reflect.ObjectProperty
	QueryParameters []*j5reflect.ObjectProperty

	List *client_j5pb.ListRequest
}

func (rr *Request) ToJ5Proto() *client_j5pb.Method_Request {
	pathParameters := make([]*schema_j5pb.ObjectProperty, 0, len(rr.PathParameters))
	for _, pp := range rr.PathParameters {
		pathParameters = append(pathParameters, pp.ToJ5Proto())
	}

	queryParameters := make([]*schema_j5pb.ObjectProperty, 0, len(rr.QueryParameters))
	for _, qp := range rr.QueryParameters {
		queryParameters = append(queryParameters, qp.ToJ5Proto())
	}

	var body *schema_j5pb.Object
	if rr.Body != nil {
		body = rr.Body.ToJ5Object()
	}

	return &client_j5pb.Method_Request{
		Body:            body,
		PathParameters:  pathParameters,
		QueryParameters: queryParameters,
	}
}

func (mm *Method) ToJ5Proto() (*client_j5pb.Method, error) {

	return &client_j5pb.Method{
		FullGrpcName: fmt.Sprintf("/%s.%s/%s", mm.Service.Package.Name, mm.Service.Name, mm.GRPCMethodName),
		Name:         mm.GRPCMethodName,
		HttpMethod:   mm.HTTPMethod,
		HttpPath:     mm.HTTPPath,
		ResponseBody: mm.ResponseBody.ToJ5Object(),
		Request:      mm.Request.ToJ5Proto(),
	}, nil
}

type StateEntity struct {
	Package *Package // parent

}
type StateEvent struct {
	StateEntity *StateEntity // parent
	Name        string
	Schema      *j5reflect.ObjectSchema
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

	var walkRefs func(j5reflect.FieldSchema) error
	walkRefRoot := func(schema j5reflect.RootSchema) error {
		_, ok := schemas[schema.FullName()]
		if ok {
			return nil
		}

		schemas[schema.FullName()] = &schemaRef{
			schema:    schema.ToJ5Root(),
			inPackage: schema.PackageName(),
		}
		switch st := schema.(type) {
		case *j5reflect.ObjectSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk %s: %w", st.FullName(), err)
				}
			}
		case *j5reflect.OneofSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk oneof: %w", err)
				}
			}
		case *j5reflect.EnumSchema:
		// do nothing

		default:
			return fmt.Errorf("unsupported ref type %T", st)
		}
		return nil
	}
	walkRefs = func(schema j5reflect.FieldSchema) error {

		switch st := schema.(type) {
		case *j5reflect.ObjectField:
			if err := walkRefRoot(st.Ref.To); err != nil {
				return fmt.Errorf("walk object as field: %w", err)
			}

		case *j5reflect.OneofField:
			if err := walkRefRoot(st.Ref.To); err != nil {
				return fmt.Errorf("walk oneof as field: %w", err)
			}

		case *j5reflect.EnumField:
			if err := walkRefRoot(st.Ref.To); err != nil {
				return fmt.Errorf("walk enum as field: %w", err)
			}

		case *j5reflect.ArrayField:
			if err := walkRefs(st.Schema); err != nil {
				return fmt.Errorf("walk array: %w", err)
			}

		case *j5reflect.MapField:
			if err := walkRefs(st.Schema); err != nil {
				return fmt.Errorf("walk map: %w", err)
			}
		}

		return nil
	}

	walkRootObject := func(schema *j5reflect.ObjectSchema) error {
		for _, prop := range schema.Properties {
			if err := walkRefs(prop.Schema); err != nil {
				return fmt.Errorf("walk root object: %w", err)
			}
		}
		return nil
	}

	for _, pkg := range api.Packages {
		for _, service := range pkg.Services {
			for _, method := range service.Methods {
				if method.Request.Body != nil {
					if err := walkRootObject(method.Request.Body); err != nil {
						return nil, fmt.Errorf("request schema %q: %w", method.Request.Body.FullName(), err)
					}
				}

				for _, prop := range method.Request.PathParameters {
					if err := walkRefs(prop.Schema); err != nil {
						return nil, fmt.Errorf("path parameter %q: %w", prop.JSONName, err)
					}
				}

				for _, prop := range method.Request.QueryParameters {
					if err := walkRefs(prop.Schema); err != nil {
						return nil, fmt.Errorf("path parameter %q: %w", prop.JSONName, err)
					}
				}

				if err := walkRootObject(method.ResponseBody); err != nil {
					return nil, fmt.Errorf("response schema %q: %w", method.ResponseBody.FullName(), err)
				}
			}
		}
	}

	return schemas, nil
}
