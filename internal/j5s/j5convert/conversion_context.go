package j5convert

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"slices"

	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// parentContext is a file's root, or message, which can hold messages and
// enums. Implemented by FileBuilder and MessageBuilder.
type parentContext interface {
	addMessage(*MessageBuilder)
	addEnum(*descriptorpb.EnumDescriptorProto, commentSet)
	addSyntheticOneof(nameHint string) (int32, error)

	addLocalType(name string, ref *TypeRef) error
	refResolver
}

type refResolver interface {
	resolveType(ref *schema_j5pb.Ref) (*TypeRef, error)
}

type TypeResolver interface {
	ResolveType(pkg string, name string) (*TypeRef, error)
}

type fileSet struct {
	_packageName string
	mainFile     *fileContext
	files        []*fileContext
}

func newFileSet(packageName string, mainFile *fileContext) *fileSet {
	return &fileSet{
		_packageName: packageName,
		mainFile:     mainFile,
		files:        []*fileContext{mainFile},
	}
}

func subPackageFileName(sourceFilename, subPackage string) string {
	dirName, baseName := path.Split(sourceFilename)
	baseRoot := strings.TrimSuffix(baseName, ".j5s.proto")
	newBase := fmt.Sprintf("%s.p.j5s.proto", baseRoot)
	subName := path.Join(dirName, subPackage, newBase)
	return subName
}

func (rr *fileSet) subPackageFile(subPackage string) *fileContext {
	fullPackage := fmt.Sprintf("%s.%s", rr._packageName, subPackage)

	for _, search := range rr.files {
		if search.fdp.GetPackage() == fullPackage {
			return search
		}
	}
	rootName := *rr.mainFile.fdp.Name
	subName := subPackageFileName(rootName, subPackage)

	found := newFileContext(subName, rr.mainFile.refResolver)

	found.fdp.Package = &fullPackage
	rr.files = append(rr.files, found)
	return found
}

type rootResolver struct {
	importAliases *importMap
	deps          TypeResolver
}

func newRootResolver(deps TypeResolver, imports *importMap) *rootResolver {
	return &rootResolver{
		importAliases: imports,
		deps:          deps,
	}
}

func (fb *rootResolver) resolveType(refSrc *schema_j5pb.Ref) (*TypeRef, error) {
	ref := fb.importAliases.expand(refSrc)
	if ref == nil {
		return nil, &PackageNotFoundError{
			Package: refSrc.Package,
			Name:    refSrc.Schema,
		}
	}

	if ref.implicit != nil {
		return ref.implicit, nil
	}

	typeRef, err := fb.deps.ResolveType(ref.ref.Package, ref.ref.Schema)
	if err != nil {
		return nil, err
	}
	return typeRef, nil
}

type rootContext struct {
	*fileSet

	errors errpos.Errors
}

func newRootContext(files *fileSet) *rootContext {
	return &rootContext{
		fileSet: files,
	}
}

func (rr *rootContext) rootFileExtension() *ext_j5pb.PackageOptions {
	return rr.mainFile.packageExt()

}

type fileContext struct {
	fdp         *descriptorpb.FileDescriptorProto
	_packageExt *ext_j5pb.PackageOptions
	refResolver
	commentSet
}

func newFileContext(name string, parentResolver refResolver) *fileContext {
	pkgName := PackageFromFilename(name)
	fdp := &descriptorpb.FileDescriptorProto{
		Syntax:  gl.Ptr("proto3"),
		Package: gl.Ptr(pkgName),
		Name:    gl.Ptr(name),
		Options: &descriptorpb.FileOptions{},
		SourceCodeInfo: &descriptorpb.SourceCodeInfo{
			Location: []*descriptorpb.SourceCodeInfo_Location{},
		},
	}
	return &fileContext{
		refResolver: parentResolver,
		fdp:         fdp,
	}
}

func (fb *fileContext) packageExt() *ext_j5pb.PackageOptions {
	if fb._packageExt == nil {
		packageExt := &ext_j5pb.PackageOptions{}
		fb._packageExt = packageExt
		proto.SetExtension(fb.fdp.Options, ext_j5pb.E_Package, packageExt)
	}
	return fb._packageExt
}

type rootNode interface {
	NameInPackage() string
}

func (fb *fileContext) fullyQualifiedName(node rootNode) string {
	nn := node.NameInPackage()
	return fmt.Sprintf("%s.%s", *fb.fdp.Package, nn)
}

func (fb *fileContext) File() *descriptorpb.FileDescriptorProto {
	last := int32(1)
	for _, comment := range fb.commentSet {
		last += 2
		loc := &descriptorpb.SourceCodeInfo_Location{
			Span: []int32{last, 1, 1},
			Path: comment.path,
		}

		if comment.description != nil {
			loc.LeadingComments = comment.description
		}

		fb.fdp.SourceCodeInfo.Location = append(fb.fdp.SourceCodeInfo.Location, loc)
	}

	return fb.fdp
}

func (fb *fileContext) ensureImport(importPath string) {
	if importPath == "" {
		panic("empty import path")
	}
	if !strings.Contains(importPath, "/") {
		panic("invalid import path " + importPath)
	}

	if importPath == *fb.fdp.Name {
		return
	}
	if slices.Contains(fb.fdp.Dependency, importPath) {
		return
	}
	fb.fdp.Dependency = append(fb.fdp.Dependency, importPath)
	sort.Strings(fb.fdp.Dependency)
}

func (fb *fileContext) addSyntheticOneof(nameHont string) (int32, error) {
	return 0, fmt.Errorf("at file level, synthetic oneof not supported")
}

func (fb *fileContext) addMessage(message *MessageBuilder) {
	idx := int32(len(fb.fdp.MessageType))
	path := []int32{4, idx}
	fb.mergeAt(path, message.commentSet)
	fb.fdp.MessageType = append(fb.fdp.MessageType, message.descriptor)
}

func (fb *fileContext) addEnum(desc *descriptorpb.EnumDescriptorProto, comments commentSet) {
	idx := int32(len(fb.fdp.EnumType))
	path := []int32{5, idx}
	fb.mergeAt(path, comments)
	fb.fdp.EnumType = append(fb.fdp.EnumType, desc)
}

func (fb *fileContext) addService(service *serviceBuilder) {
	idx := int32(len(fb.fdp.Service))
	path := []int32{6, idx}
	fb.mergeAt(path, service.commentSet)
	fb.fdp.Service = append(fb.fdp.Service, service.desc)
}

func (fb *fileContext) addLocalType(name string, ref *TypeRef) error {
	// NOP, file level references are already resolvable
	return nil
}

type MessageBuilder struct {
	parentResolver refResolver
	localTypes     map[string]*TypeRef
	descriptor     *descriptorpb.DescriptorProto
	commentSet
}

func newMessageContext(name string, parent refResolver) *MessageBuilder {
	message := &MessageBuilder{
		parentResolver: parent,
		descriptor: &descriptorpb.DescriptorProto{
			Name:    gl.Ptr(name),
			Options: &descriptorpb.MessageOptions{},
		},
		localTypes: make(map[string]*TypeRef),
	}
	return message
}

func (msg *MessageBuilder) resolveType(ref *schema_j5pb.Ref) (*TypeRef, error) {
	if ref.Package == "" {
		if local, ok := msg.localTypes[ref.Schema]; ok {
			return local, nil
		}
	}
	return msg.parentResolver.resolveType(ref)
}

func (msg *MessageBuilder) addMessage(message *MessageBuilder) {
	msg.mergeAt([]int32{3, int32(len(msg.descriptor.NestedType))}, message.commentSet)
	msg.descriptor.NestedType = append(msg.descriptor.NestedType, message.descriptor)
}

func (msg *MessageBuilder) addEnum(desc *descriptorpb.EnumDescriptorProto, comments commentSet) {
	msg.mergeAt([]int32{4, int32(len(msg.descriptor.EnumType))}, comments)
	msg.descriptor.EnumType = append(msg.descriptor.EnumType, desc)
}

func (msg *MessageBuilder) addLocalType(name string, ref *TypeRef) error {
	if _, exists := msg.localTypes[name]; exists {
		return fmt.Errorf("local type %q already defined in message %q", name, msg.descriptor.GetName())
	}
	msg.localTypes[name] = ref
	return nil
}

func (msg *MessageBuilder) addSyntheticOneof(nameHint string) (int32, error) {
	nextIndex := len(msg.descriptor.OneofDecl)
	msg.descriptor.OneofDecl = append(msg.descriptor.OneofDecl, &descriptorpb.OneofDescriptorProto{
		Name: gl.Ptr(fmt.Sprintf("_%s", nameHint)),
	})
	return int32(nextIndex), nil
}
