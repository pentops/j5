package j5client

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
)

func APIFromDesc(api *client_j5pb.API) (*API, error) {
	schemaSet, err := j5reflect.PackageSetFromClientAPI(api)
	if err != nil {
		return nil, fmt.Errorf("package set from api: %w", err)
	}

	apiBase, err := apiBaseFromDesc(api)
	if err != nil {
		return nil, fmt.Errorf("api base from desc: %w", err)
	}

	if err := apiBase.linkSchemas(schemaSet); err != nil {
		return nil, fmt.Errorf("link schemas: %w", err)
	}

	return apiBase, nil
}

func apiBaseFromDesc(api *client_j5pb.API) (*API, error) {
	apiPkg := &API{
		Packages: []*Package{},
		Metadata: api.Metadata,
	}

	for _, pkgSource := range api.Packages {
		pkg := &Package{
			Name:  pkgSource.Name,
			Label: pkgSource.Label,
		}
		apiPkg.Packages = append(apiPkg.Packages, pkg)

		for idx, serivceSrc := range pkgSource.Services {
			pkg.Services[idx] = serviceFromDesc(pkg, serivceSrc)
		}
	}

	return apiPkg, nil
}

func serviceFromDesc(pkg *Package, src *client_j5pb.Service) *Service {
	service := &Service{
		Package: pkg,
		Name:    src.Name,
		Methods: make([]*Method, len(src.Methods)),
	}

	for idx, methodSrc := range src.Methods {
		service.Methods[idx] = methodFromDesc(service, methodSrc)
	}

	return service
}

func methodFromDesc(service *Service, src *client_j5pb.Method) *Method {
	return &Method{
		Service:        service,
		GRPCMethodName: src.Name,
		HTTPPath:       src.HttpPath,
		HTTPMethod:     src.HttpMethod,
		HasBody:        src.HttpMethod != client_j5pb.HTTPMethod_GET,
		Request: &schemaPlaceholder{
			source:      src.Request.Body,
			packageName: service.Package.Name,
		},
		Response: &schemaPlaceholder{
			source:      src.ResponseBody,
			packageName: service.Package.Name,
		},
	}
}

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
}
