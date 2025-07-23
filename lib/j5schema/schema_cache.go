package j5schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var Global = NewSchemaCache()

// SchemaCache acts like PackageSet, but builds schemas on demand from reflection.
type SchemaCache struct {
	*packageSet
}

func NewSchemaCache() *SchemaCache {
	return &SchemaCache{
		packageSet: newPackageSet(),
	}
}

// Schema returns the J5 schema for the given message descriptor.
func (sc *SchemaCache) Schema(src protoreflect.MessageDescriptor) (RootSchema, error) {
	packageName, nameInPackage := splitDescriptorName(src)
	schemaPackage := sc.referencePackage(packageName)

	built, didExist := schemaPackage.Schemas.getOrCreate(nameInPackage, func() *RefSchema {
		return &RefSchema{
			Package: schemaPackage,
			Schema:  nameInPackage,
		}
	})

	if didExist {
		if built.To == nil {
			// When building from reflection, the 'to' should be linked by the
			// caller which created the ref.
			return nil, fmt.Errorf("unlinked ref: %s/%s", packageName, nameInPackage)
		}
		return built.To, nil
	}

	var err error
	built.To, err = schemaPackage.buildMessageSchema(src)
	if err != nil {
		return nil, err
	}
	if built.To.FullName() != built.FullName() {
		return nil, fmt.Errorf("schema %q has wrong name %q", built.FullName(), built.To.FullName())
	}
	return built.To, nil
}

func (sc *SchemaCache) ObjectSchema(src protoreflect.MessageDescriptor) (*ObjectSchema, error) {
	schema, err := sc.Schema(src)
	if err != nil {
		return nil, err
	}
	if objSchema, ok := schema.(*ObjectSchema); ok {
		return objSchema, nil
	}
	return nil, fmt.Errorf("expected object schema for %s/%s, got %T", src.Parent().FullName(), src.Name(), schema)
}

func MustObjectSchema(src protoreflect.MessageDescriptor) *ObjectSchema {
	schema, err := Global.ObjectSchema(src)
	if err != nil {
		panic(fmt.Sprintf("j5schema: %v", err))
	}
	return schema
}
