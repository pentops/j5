package j5schema

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SchemaCache acts like PackageSet, but builds schemas on demand from reflection.
type SchemaCache struct {
	packages map[string]*Package
}

func NewSchemaCache() *SchemaCache {
	return &SchemaCache{
		packages: map[string]*Package{},
	}
}

// Schema returns the J5 schema for the given message descriptor.
func (sc *SchemaCache) Schema(src protoreflect.MessageDescriptor) (RootSchema, error) {
	packageName, nameInPackage := splitDescriptorName(src)
	schemaPackage := sc.referencePackage(packageName)
	if built, ok := schemaPackage.Schemas[nameInPackage]; ok {
		if built.To == nil {
			// When building from reflection, the 'to' should be linked by the
			// caller which created the ref.
			return nil, fmt.Errorf("unlinked ref: %s/%s", packageName, nameInPackage)
		}
		return built.To, nil
	}

	placeholder := &RefSchema{
		Package: schemaPackage,
		Schema:  nameInPackage,
	}
	schemaPackage.Schemas[nameInPackage] = placeholder

	msgOptions := proto.GetExtension(src.Options(), ext_j5pb.E_Message).(*ext_j5pb.MessageOptions)
	isOneofWrapper := isOneofWrapper(src, msgOptions)
	var err error
	if isOneofWrapper {
		placeholder.To, err = schemaPackage.buildOneofSchema(src, msgOptions.GetOneof())
	} else {
		placeholder.To, err = schemaPackage.buildObjectSchema(src, msgOptions.GetObject())
	}
	if err != nil {
		return nil, err
	}
	if placeholder.To.FullName() != placeholder.FullName() {
		return nil, fmt.Errorf("schema %q has wrong name %q", placeholder.FullName(), placeholder.To.FullName())
	}
	return placeholder.To, nil
}

func (sc *SchemaCache) refTo(pkg, schema string) (*RefSchema, bool) {
	refPackage := sc.referencePackage(pkg)
	if existing, ok := refPackage.Schemas[schema]; ok {
		return existing, true
	}

	refSchema := &RefSchema{
		Package: refPackage,
		Schema:  schema,
	}
	refPackage.Schemas[schema] = refSchema

	return refSchema, false
}

func (sc *SchemaCache) referencePackage(name string) *Package {
	if existing, ok := sc.packages[name]; ok {
		return existing
	}
	pkg := &Package{
		Name:       name,
		Schemas:    map[string]*RefSchema{},
		PackageSet: sc,
	}
	sc.packages[name] = pkg
	return pkg
}
