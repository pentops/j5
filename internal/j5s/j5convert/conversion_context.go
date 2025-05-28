package j5convert

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"slices"

	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/internal/bcl/errpos"
	"google.golang.org/protobuf/types/descriptorpb"
)

// parentContext is a file's root, or message, which can hold messages and
// enums. Implemented by FileBuilder and MessageBuilder.
type parentContext interface {
	addMessage(*MessageBuilder)
	addEnum(*descriptorpb.EnumDescriptorProto, commentSet)
	addSyntheticOneof(nameHint string) (int32, error)
}

type rootContext struct {
	packageName string
	deps        TypeResolver
	//source      sourceLink
	errors errpos.Errors

	importAliases *importMap

	mainFile *fileContext
	files    []*fileContext
}

func newRootContext(deps TypeResolver, imports *importMap, file *fileContext) *rootContext {
	return &rootContext{
		packageName:   file.fdp.GetPackage(),
		deps:          deps,
		mainFile:      file,
		importAliases: imports,
		files:         []*fileContext{file},
	}
}

func subPackageFileName(sourceFilename, subPackage string) string {
	dirName, baseName := path.Split(sourceFilename)
	baseRoot := strings.TrimSuffix(baseName, ".j5s.proto")
	newBase := fmt.Sprintf("%s.p.j5s.proto", baseRoot)
	subName := path.Join(dirName, subPackage, newBase)
	return subName
}

func (rr *rootContext) subPackageFile(subPackage string) *fileContext {
	fullPackage := fmt.Sprintf("%s.%s", rr.packageName, subPackage)

	for _, search := range rr.files {
		if search.fdp.GetPackage() == fullPackage {
			return search
		}
	}
	rootName := *rr.mainFile.fdp.Name
	subName := subPackageFileName(rootName, subPackage)

	found := newFileContext(subName)

	found.fdp.Package = &fullPackage
	rr.files = append(rr.files, found)
	return found
}

type fileContext struct {
	fdp *descriptorpb.FileDescriptorProto
	commentSet
}

func newFileContext(name string) *fileContext {
	pkgName := PackageFromFilename(name)
	return &fileContext{
		fdp: &descriptorpb.FileDescriptorProto{
			Syntax:  gl.Ptr("proto3"),
			Package: gl.Ptr(pkgName),
			Name:    gl.Ptr(name),
			Options: &descriptorpb.FileOptions{},
			SourceCodeInfo: &descriptorpb.SourceCodeInfo{
				Location: []*descriptorpb.SourceCodeInfo_Location{},
			},
		},
	}
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
		panic("empty alias")
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

type MessageBuilder struct {
	descriptor *descriptorpb.DescriptorProto
	commentSet
}

func blankMessage(name string) *MessageBuilder {
	message := &MessageBuilder{
		descriptor: &descriptorpb.DescriptorProto{
			Name:    gl.Ptr(name),
			Options: &descriptorpb.MessageOptions{},
		},
	}
	return message
}

func (msg *MessageBuilder) addMessage(message *MessageBuilder) {
	msg.mergeAt([]int32{3, int32(len(msg.descriptor.NestedType))}, message.commentSet)
	msg.descriptor.NestedType = append(msg.descriptor.NestedType, message.descriptor)
}

func (msg *MessageBuilder) addEnum(desc *descriptorpb.EnumDescriptorProto, comments commentSet) {
	msg.mergeAt([]int32{4, int32(len(msg.descriptor.EnumType))}, comments)
	msg.descriptor.EnumType = append(msg.descriptor.EnumType, desc)
}

func (msg *MessageBuilder) addSyntheticOneof(nameHint string) (int32, error) {
	nextIndex := len(msg.descriptor.OneofDecl)
	msg.descriptor.OneofDecl = append(msg.descriptor.OneofDecl, &descriptorpb.OneofDescriptorProto{
		Name: gl.Ptr(fmt.Sprintf("_%s", nameHint)),
	})
	return int32(nextIndex), nil
}
