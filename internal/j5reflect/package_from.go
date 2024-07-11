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

	var resolveWalkRoot func(RootSchema) error
	var resolveWalk func(FieldSchema) error
	resolveWalkRoot = func(root RootSchema) error {
		name := root.FullName()
		if _, ok := seenSchemas[name]; ok {
			return nil
		}
		seenSchemas[name] = root

		switch tt := root.(type) {
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

		case *EnumSchema:
		// nothing to do
		default:
			return fmt.Errorf("unexpected schema type %T", root)
		}
		return nil
	}

	resolveRef := func(tt *RefSchema) error {
		if tt.To != nil {
			if err := resolveWalkRoot(tt.To); err != nil {
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
		if err := resolveWalkRoot(referenced); err != nil {
			return err
		}
		return nil
	}

	resolveWalk = func(root FieldSchema) error {

		switch tt := root.(type) {

		case *ObjectAsFieldSchema:
			return resolveRef(tt.Ref)

		case *OneofAsFieldSchema:
			return resolveRef(tt.Ref)

		case *EnumAsFieldSchema:
			return resolveRef(tt.Ref)

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
				if err := resolveWalkRoot(method.Request); err != nil {
					return nil, err
				}
				if err := resolveWalkRoot(method.Response); err != nil {
					return nil, err
				}

			}
		}
		for _, event := range pkg.Events {
			if err := resolveWalkRoot(event.Schema); err != nil {
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

	var requestObject *schema_j5pb.Object
	if protoService.Request.Body != nil {
		requestObject = protoService.Request.Body.GetObject()
		if requestObject == nil {
			return nil, fmt.Errorf("request body must be an object")
		}
	} else {
		requestObject = &schema_j5pb.Object{
			Name:       fmt.Sprintf("%sRequest", protoService.Name),
			Properties: []*schema_j5pb.ObjectProperty{},
		}
	}

	requestObject.Properties = append(requestObject.Properties, protoService.Request.PathParameters...)
	requestObject.Properties = append(requestObject.Properties, protoService.Request.QueryParameters...)

	request, err := RootSchemaFromDesc(pkg, &schema_j5pb.RootSchema{
		Type: &schema_j5pb.RootSchema_Object{
			Object: requestObject,
		},
	})
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
