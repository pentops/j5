package j5client

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
	"github.com/pentops/j5/internal/patherr"
)

func APIFromSource(api *source_j5pb.API) (*client_j5pb.API, error) {
	schemaSet, err := j5reflect.PackageSetFromSourceAPI(api)
	if err != nil {
		return nil, fmt.Errorf("package set from api: %w", err)
	}

	sb := &sourceBuilder{
		schemas: schemaSet,
	}

	apiBase, err := sb.apiBaseFromSource(api)
	if err != nil {
		return nil, fmt.Errorf("api base from desc: %w", err)
	}

	return apiBase.ToJ5Proto()
}

type sourceBuilder struct {
	schemas *j5reflect.SchemaSet
}

func (sb *sourceBuilder) apiBaseFromSource(api *source_j5pb.API) (*API, error) {
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

		for _, subPkg := range pkgSource.SubPackages {
			for _, serivceSrc := range subPkg.Services {
				service, err := sb.serviceFromSource(pkg, serivceSrc)
				if err != nil {
					return nil, patherr.Wrap(err, "package", pkg.Name, "service", serivceSrc.Name)
				}
				pkg.Services = append(pkg.Services, service)
			}
		}
	}

	return apiPkg, nil
}

func (sb *sourceBuilder) serviceFromSource(pkg *Package, src *source_j5pb.Service) (*Service, error) {

	service := &Service{
		Package: pkg,
		Name:    src.Name,
		Methods: make([]*Method, len(src.Methods)),
	}

	for idx, src := range src.Methods {
		method, err := sb.methodFromSource(service, src)
		if err != nil {
			return nil, patherr.Wrap(err, "method", src.Name)
		}
		service.Methods[idx] = method
	}

	return service, nil
}

func (sb *sourceBuilder) methodFromSource(service *Service, src *source_j5pb.Method) (*Method, error) {

	requestSchema, err := sb.schemas.SchemaByName(service.Package.Name, src.RequestSchema)
	if err != nil {
		return nil, patherr.Wrap(err, "request")
	}
	requestObject, ok := requestSchema.(*j5reflect.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("request schema is not an object")
	}

	response, err := sb.schemas.SchemaByName(service.Package.Name, src.ResponseSchema)
	if err != nil {
		return nil, patherr.Wrap(err, "response")
	}
	responseObject, ok := response.(*j5reflect.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("response schema is not an object")
	}

	method := &Method{
		Service:        service,
		GRPCMethodName: src.Name,
		HTTPPath:       src.HttpPath,
		HTTPMethod:     src.HttpMethod,
		HasBody:        src.HttpMethod != client_j5pb.HTTPMethod_GET,
		ResponseBody:   responseObject,
	}

	if err := method.fillRequest(requestObject); err != nil {
		return nil, fmt.Errorf("fill request: %w", err)
	}

	return method, nil
}

func (mm *Method) fillRequest(requestObject *j5reflect.ObjectSchema) error {

	pathParameterNames := map[string]struct{}{}
	pathParts := strings.Split(mm.HTTPPath, "/")
	for _, part := range pathParts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		fieldName := strings.TrimPrefix(part, ":")
		pathParameterNames[fieldName] = struct{}{}
	}

	pathProperties := make([]*j5reflect.ObjectProperty, 0)
	bodyProperties := make([]*j5reflect.ObjectProperty, 0)

	isQueryRequest := false

	for _, prop := range requestObject.Properties {
		if prop.JSONName == "query" {
			if propObj, ok := prop.Schema.(*j5reflect.ObjectField); ok {
				ref := propObj.Schema().AsRef()
				if ref != nil {
					if ref.Package.Name == "j5.list.v1" && ref.Schema == "Query" {
						isQueryRequest = true
					}
				}
			}
		}
		_, isPath := pathParameterNames[prop.JSONName]
		if isPath {
			pathProperties = append(pathProperties, prop)
		} else {
			bodyProperties = append(bodyProperties, prop)
		}
	}

	request := &Request{
		PathParameters: pathProperties,
	}

	if mm.HasBody {
		request.Body = requestObject.Clone()
		request.Body.Properties = bodyProperties
	} else {
		request.QueryParameters = bodyProperties
	}

	responseSchema := mm.ResponseBody

	if isQueryRequest {
		listRequest, err := buildListRequest(responseSchema)
		if err != nil {
			return err
		}
		request.List = listRequest
	}

	mm.Request = request

	return nil
}
