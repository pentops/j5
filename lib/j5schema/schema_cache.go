package j5schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

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
