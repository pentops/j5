package j5reflect

import "google.golang.org/protobuf/reflect/protoreflect"

type DescriptorResolver interface {
	FindDescriptorByName(protoreflect.FullName) (protoreflect.Descriptor, error)
}

type SchemaResolver struct {
	*PackageSet
	resolver DescriptorResolver
}

func NewSchemaResolver(resolver DescriptorResolver) *SchemaResolver {
	return &SchemaResolver{
		PackageSet: NewPackageSet(),
		resolver:   resolver,
	}
}

/*
func (ss *SchemaResolver) SchemaByName(name protoreflect.FullName) (RootSchema, error) {
	obj, ok := ss.refs[name]
	if ok {
		if obj.To == nil {
			return nil, fmt.Errorf("unlinked ref: %s", name)
		}
		return obj.To, nil
	}
	descriptor, err := ss.resolver.FindDescriptorByName(name)
	if err != nil {
		return nil, err
	}
	msg, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("descriptor %s is not a message", name)
	}
	return ss.SchemaReflect(msg)
}

func (ss *Package) SchemaObject(src protoreflect.MessageDescriptor) (*schema_j5pb.RootSchema, error) {
	val, err := ss.SchemaReflect(src)
	if err != nil {
		return nil, err
	}

	return val.ToJ5Root(), nil
}*/
