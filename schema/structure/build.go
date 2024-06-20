package structure

import (
	"context"
	"fmt"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/j5reflect"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BuildFromImage(ctx context.Context, image *source_j5pb.SourceImage) (*schema_j5pb.API, error) {
	proseResolver := imageResolver(image.Prose)

	descriptors := &descriptorpb.FileDescriptorSet{
		File: image.File,
	}

	if len(descriptors.File) < 1 {
		panic("Expected at least one descriptor file, found none")
	}

	config := &config_j5pb.Config{
		Packages: image.Packages,
		Options:  image.Codec,
	}

	if config.Packages == nil || len(config.Packages) < 1 {
		return nil, fmt.Errorf("no packages to generate")
	}

	if config.Options == nil {
		config.Options = &config_j5pb.CodecOptions{}
	}

	return BuildFromDescriptors(ctx, config, descriptors, proseResolver)
}

type builder struct {
	packages     []*schema_j5pb.Package
	trimPackages []string

	usedSchemas map[protoreflect.FullName]int
}

func BuildFromDescriptors(ctx context.Context, config *config_j5pb.Config, descriptors *descriptorpb.FileDescriptorSet, proseResolver ProseResolver) (*schema_j5pb.API, error) {

	descFiles, err := protodesc.NewFiles(descriptors)
	if err != nil {
		return nil, fmt.Errorf("descriptor files: %w", err)
	}

	packages, err := BuildPackages(config, descFiles, proseResolver)
	if err != nil {
		return nil, fmt.Errorf("packages from descriptors: %w", err)
	}

	return linkPackages(descFiles, packages)
}

func linkPackages(descFiles *protoregistry.Files, packages []*schema_j5pb.Package) (*schema_j5pb.API, error) {

	schemaSet := j5reflect.NewSchemaResolver(descFiles)

	schemas := make(map[string]*schema_j5pb.Schema)

	var walkRefs func(*j5reflect.Schema) error
	walkRefs = func(schema *j5reflect.Schema) error {

		switch st := schema.Type().(type) {
		case *j5reflect.ObjectSchema:
			for _, prop := range st.Properties {
				if err := walkRefs(prop.Schema); err != nil {
					return fmt.Errorf("walk %s: %w", st.ProtoMessage.FullName(), err)
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
			stringName := string(st.Name)
			if _, ok := schemas[stringName]; ok {
				return nil
			}
			if st.To == nil {
				return fmt.Errorf("ref schema %q has no target", st.Name)
			}
			asProto, err := st.To.ToJ5Proto()
			if err != nil {
				return fmt.Errorf("ref schema %q: %w", st.Name, err)
			}
			schemas[stringName] = asProto
			if err := walkRefs(st.To); err != nil {
				return fmt.Errorf("walk ref %s: %w", st.Name, err)
			}
		}

		return nil
	}

	usedSchemas := map[protoreflect.FullName]struct{}{}
	rootResolve := func(refItem *schema_j5pb.Schema) (*j5reflect.ObjectSchema, error) {
		name := protoreflect.FullName(refItem.GetRef())
		if _, ok := usedSchemas[name]; ok {
			return nil, fmt.Errorf("root schema %q not unique", name)
		}
		usedSchemas[name] = struct{}{}

		schema, err := schemaSet.SchemaByName(name)
		if err != nil {
			return nil, err
		}

		if err := walkRefs(schema); err != nil {
			return nil, fmt.Errorf("root schema %q: %w", name, err)
		}

		object, ok := schema.Type().(*j5reflect.ObjectSchema)
		if !ok {
			return nil, fmt.Errorf("root schema %q is not an object", name)
		}

		return object, nil
	}

	for _, pkg := range packages {
		for _, method := range pkg.Methods {
			requestObject, err := rootResolve(method.RequestBody)
			if err != nil {
				return nil, fmt.Errorf("request schema %q: %w", method.RequestBody.GetRef(), err)
			}

			responseObject, err := rootResolve(method.ResponseBody)
			if err != nil {
				return nil, fmt.Errorf("response schema %q: %w", method.ResponseBody.GetRef(), err)
			}

			method.ResponseBody, err = responseObject.ToJ5Proto()
			if err != nil {
				return nil, fmt.Errorf("response schema %q: %w", method.FullGrpcName, err)
			}

			if err := linkRequestMethod(method, requestObject); err != nil {
				return nil, err
			}

		}
		for _, event := range pkg.Events {
			eventObject, err := rootResolve(event.Schema)
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", event.Schema.GetRef(), err)
			}
			event.Schema, err = eventObject.ToJ5Proto()
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", event.Name, err)
			}

		}
		for _, entity := range pkg.Entities {
			eventObject, err := rootResolve(entity.Schema)
			if err != nil {
				return nil, fmt.Errorf("event schema %q: %w", entity.Schema.GetRef(), err)
			}
			entity.Schema, err = eventObject.ToJ5Proto()
			if err != nil {
				return nil, err
			}
		}
	}

	return &schema_j5pb.API{
		Packages: packages,
		Schemas:  schemas,
		Metadata: &schema_j5pb.Metadata{
			BuiltAt: timestamppb.Now(),
		},
	}, nil
}

func linkRequestMethod(method *schema_j5pb.Method, requestObject *j5reflect.ObjectSchema) error {
	var err error

	for _, parameter := range method.PathParameters {

		prop, ok := popProperty(requestObject, protoreflect.Name(parameter.Name))
		if !ok {
			return fmt.Errorf("path parameter %q not found in request object", parameter.Name)
		}

		propSchema, err := prop.Schema.ToJ5Proto()
		if err != nil {
			return err
		}

		parameter.Schema = propSchema
		parameter.Name = prop.JSONName
	}

	if method.HttpMethod == "get" {
		method.RequestBody = nil
		for _, prop := range requestObject.Properties {
			propSchema, err := prop.Schema.ToJ5Proto()
			if err != nil {
				return err
			}
			method.QueryParameters = append(method.QueryParameters, &schema_j5pb.Parameter{
				Name:     prop.JSONName,
				Required: false,
				Schema:   propSchema,
			})
		}
	} else {
		method.RequestBody, err = requestObject.ToJ5Proto()
		if err != nil {
			return fmt.Errorf("request schema %q: %w", method.FullGrpcName, err)
		}
	}
	return nil
}

func popProperty(obj *j5reflect.ObjectSchema, name protoreflect.Name) (*j5reflect.ObjectProperty, bool) {
	newProps := make([]*j5reflect.ObjectProperty, 0, len(obj.Properties)-1)
	var found *j5reflect.ObjectProperty
	for _, prop := range obj.Properties {
		if len(prop.ProtoField) != 1 {
			continue // TODO: Can't walk nested fields yet
		}
		fieldName := prop.ProtoField[0].Name()

		if fieldName == name {
			found = prop
			continue
		}
		newProps = append(newProps, prop)
	}
	obj.Properties = newProps
	return found, found != nil
}
