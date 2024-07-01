package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

func APIFromDesc(api *schema_j5pb.API) (*API, error) {
	return buildAPI(api.Packages)
}

func buildAPI(srcPackages []*schema_j5pb.Package) (*API, error) {
	out := &API{
		Packages: make([]*Package, 0, len(srcPackages)),
	}
	packageMap := map[string]*schema_j5pb.Package{}
	builtPackageMap := map[string]*Package{}
	for _, pkg := range srcPackages {
		packageMap[pkg.Name] = pkg
		built, err := packageFromDesc(pkg)
		if err != nil {
			return nil, fmt.Errorf("package %s: %w", pkg.Name, err)
		}
		out.Packages = append(out.Packages, built)
		builtPackageMap[pkg.Name] = built
	}

	var seenSchemas = map[string]RootSchema{}

	var resolveWalk func(Schema) error
	resolveWalk = func(root Schema) error {
		if asRoot, ok := root.(RootSchema); ok {
			name := asRoot.FullName()
			if _, ok := seenSchemas[name]; ok {
				return nil
			}
			seenSchemas[name] = asRoot
		}

		switch tt := root.(type) {
		case *RefSchema:
			if tt.To != nil {
				if err := resolveWalk(tt.To); err != nil {
					return err
				}
				return nil
			}
			refPkg, ok := packageMap[tt.Package]
			if !ok {
				return fmt.Errorf("reference to unknown package %q", tt.FullName())
			}
			refSchema, ok := refPkg.Schemas[tt.FullName()]
			if !ok {
				return fmt.Errorf("reference to unknown schema %q", tt.FullName())
			}

			referencePackage := builtPackageMap[tt.Package]

			referenced, ok := seenSchemas[tt.FullName()]
			if !ok {
				var err error
				referenced, err = RootSchemaFromDesc(referencePackage, refSchema)
				if err != nil {
					return fmt.Errorf("linking schema %s.%s", tt.Package, tt.Schema)
				}
			}

			tt.To = referenced
			if err := resolveWalk(referenced); err != nil {
				return err
			}

		case *ObjectSchema:
			for _, field := range tt.Properties {
				if err := resolveWalk(field.Schema); err != nil {
					return fmt.Errorf("field %s: %w", field.JSONName, err)
				}
			}

		case *OneofSchema:
			for _, field := range tt.Properties {
				if err := resolveWalk(field.Schema); err != nil {
					return fmt.Errorf("field %s: %w", field.JSONName, err)
				}
			}

		case *MapSchema:
			if err := resolveWalk(tt.Schema); err != nil {
				return err
			}

		case *ArraySchema:
			if err := resolveWalk(tt.Schema); err != nil {
				return err
			}

		}

		return nil

	}

	for _, pkg := range out.Packages {
		for _, service := range pkg.Services {
			for _, method := range service.Methods {
				if err := resolveWalk(method.Request); err != nil {
					return nil, err
				}
				if err := resolveWalk(method.Response); err != nil {
					return nil, err
				}

			}
		}
		for _, event := range pkg.Events {
			if err := resolveWalk(event.Schema); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func packageFromDesc(pkg *schema_j5pb.Package) (*Package, error) {
	out := &Package{
		Name:     pkg.Name,
		Label:    pkg.Label,
		Events:   make([]*Event, len(pkg.Events)),
		Services: make([]*Service, len(pkg.Services)),
	}
	for serviceIdx, protoService := range pkg.Services {
		service := &Service{
			Package: out,
			Name:    protoService.Name,
			Methods: make([]*Method, len(protoService.Methods)),
		}

		for idx, protoMethod := range protoService.Methods {
			method, err := methodFromDesc(service, protoMethod)
			if err != nil {
				return nil, wrapError(err, "services", protoService.Name, protoMethod.Name)
			}
			method.Service = service
			service.Methods[idx] = method
		}
		out.Services[serviceIdx] = service
	}
	for idx, protoEvent := range pkg.Events {
		event, err := eventFromDesc(out, protoEvent)
		if err != nil {
			return nil, wrapError(err, "events", protoEvent.Name)
		}
		out.Events[idx] = event
	}
	return out, nil

}

func methodFromDesc(service *Service, protoService *schema_j5pb.Method) (*Method, error) {
	pkg := service.Package

	request, err := RootSchemaFromDesc(pkg, protoService.RequestBody)
	if err != nil {
		return nil, wrapError(err, "requestBody")
	}

	response, err := RootSchemaFromDesc(pkg, protoService.ResponseBody)
	if err != nil {
		return nil, wrapError(err, "responseBody")
	}
	out := &Method{
		GRPCMethodName: protoService.Name,
		HTTPPath:       protoService.HttpPath,
		HTTPMethod:     protoService.HttpMethod,
		HasBody:        protoService.HttpMethod != schema_j5pb.HTTPMethod_HTTP_METHOD_GET,
		Request:        request,
		Response:       response,
		Service:        service,
	}
	return out, nil
}

func eventFromDesc(pkg *Package, protoEvent *schema_j5pb.EventSpec) (*Event, error) {
	if protoEvent.Schema == nil {
		return nil, fmt.Errorf("schema is required")
	}
	schema, err := RootSchemaFromDesc(pkg, protoEvent.Schema)
	if err != nil {
		return nil, wrapError(err, "schema")
	}
	out := &Event{
		Name:    protoEvent.Name,
		Schema:  schema,
		Package: pkg,
	}
	return out, nil
}
