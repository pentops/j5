package j5client

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
	"github.com/pentops/j5/internal/patherr"
)

func APIFromDesc(api *client_j5pb.API) (*client_j5pb.API, error) {
	schemaSet, err := j5reflect.PackageSetFromClientAPI(api)
	if err != nil {
		return nil, fmt.Errorf("package set from api: %w", err)
	}

	db := &descBuilder{
		schemas: schemaSet,
	}

	apiBase, err := db.apiBaseFromDesc(api)
	if err != nil {
		return nil, fmt.Errorf("api base from desc: %w", err)
	}

	return apiBase.ToJ5Proto()
}

type descBuilder struct {
	schemas *j5reflect.SchemaSet
}

func (db *descBuilder) apiBaseFromDesc(api *client_j5pb.API) (*API, error) {
	apiPkg := &API{
		Packages: []*Package{},
		Metadata: &client_j5pb.Metadata{},
	}

	for _, pkgSource := range api.Packages {
		pkg := &Package{
			Name:  pkgSource.Name,
			Label: pkgSource.Label,
		}
		apiPkg.Packages = append(apiPkg.Packages, pkg)

		for _, serivceSrc := range pkgSource.Services {
			service, err := db.serviceFromDesc(pkg, serivceSrc)
			if err != nil {
				return nil, patherr.Wrap(err, "package", pkg.Name, "service", serivceSrc.Name)
			}
			pkg.Services = append(pkg.Services, service)
		}
	}

	return apiPkg, nil
}

func (db *descBuilder) serviceFromDesc(pkg *Package, src *client_j5pb.Service) (*Service, error) {
	service := &Service{
		Package: pkg,
		Name:    src.Name,
		Methods: make([]*Method, len(src.Methods)),
	}

	for idx, methodSrc := range src.Methods {
		method, err := db.methodFromDesc(service, methodSrc)
		if err != nil {
			return nil, patherr.Wrap(err, "method", methodSrc.Name)
		}

		service.Methods[idx] = method
	}

	return service, nil
}

func (db *descBuilder) methodFromDesc(service *Service, src *client_j5pb.Method) (*Method, error) {

	request := &Request{}

	if src.Request.Body != nil {
		requestBody, err := db.schemas.AnonymousObjectFromSchema(service.Package.Name, src.Request.Body)
		if err != nil {
			return nil, patherr.Wrap(err, "request")
		}
		request.Body = requestBody
	}

	if len(src.Request.PathParameters) > 0 {
		pathParams, err := db.schemas.AnonymousObjectFromSchema(service.Package.Name, &schema_j5pb.Object{
			Properties: src.Request.PathParameters,
		})
		if err != nil {
			return nil, patherr.Wrap(err, "request")
		}
		request.PathParameters = pathParams.Properties
	}

	if len(src.Request.QueryParameters) > 0 {
		queryParams, err := db.schemas.AnonymousObjectFromSchema(service.Package.Name, &schema_j5pb.Object{
			Properties: src.Request.QueryParameters,
		})
		if err != nil {
			return nil, patherr.Wrap(err, "request")
		}
		request.QueryParameters = queryParams.Properties
	}

	response, err := db.schemas.AnonymousObjectFromSchema(service.Package.Name, src.ResponseBody)
	if err != nil {
		return nil, patherr.Wrap(err, "response")
	}

	return &Method{
		Service:        service,
		GRPCMethodName: src.Name,
		HTTPPath:       src.HttpPath,
		HTTPMethod:     src.HttpMethod,
		HasBody:        src.HttpMethod != client_j5pb.HTTPMethod_GET,
		Request:        request,
		ResponseBody:   response,
	}, nil
}

/*
type schemaPlaceholder struct {
	source      *schema_j5pb.Object
	packageName string
	linked      *j5reflect.ObjectSchema
}

func (sr *schemaPlaceholder) FullName() string {
	return fmt.Sprintf("%s/%s", sr.packageName, sr.source.Name)
}

func (sr *schemaPlaceholder) link(linker schemaSource) error {
	if sr.linked != nil {
		return nil
	}

	linked, err := linker.AnonymousObjectFromSchema(sr.packageName, sr.source)
	if err != nil {
		return fmt.Errorf("link %q: %w", sr.FullName(), err)
	}

	sr.linked = linked
	return nil
}

func (sr *schemaPlaceholder) Schema() *j5reflect.ObjectSchema {
	return sr.linked
}*/
