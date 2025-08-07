package source

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"

	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/j5/internal/structure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/descriptorpb"
)

type imageBuilder struct {
	img               *source_j5pb.SourceImage
	includedFilenames map[string]struct{}
	//deps              *imageFiles
}

func newImageBuilder() *imageBuilder {
	return &imageBuilder{
		img:               &source_j5pb.SourceImage{},
		includedFilenames: make(map[string]struct{}),
		//deps:              deps,
	}
}

func newImageBuilderFromImage(img *source_j5pb.SourceImage) *imageBuilder {
	return &imageBuilder{
		img:               img,
		includedFilenames: make(map[string]struct{}),
		/*
			deps: &imageFiles{
				primary:      make(map[string]*descriptorpb.FileDescriptorProto),
				dependencies: make(map[string]*descriptorpb.FileDescriptorProto),
			},*/
	}
}

func (ib *imageBuilder) addPackage(pkg *source_j5pb.Package) {
	ib.img.Packages = append(ib.img.Packages, pkg)
}

func (ib *imageBuilder) addBuilt(built *protobuild.BuiltPackage, source *ext_j5pb.J5Source) error {
	descriptors := make([]*descriptorpb.FileDescriptorProto, 0, len(built.Proto))
	for _, file := range built.Proto {
		descriptor := protodesc.ToFileDescriptorProto(file.Linked)
		if descriptor.Options == nil {
			descriptor.Options = &descriptorpb.FileOptions{}
		}
		proto.SetExtension(descriptor.Options, ext_j5pb.E_J5Source, source)
		descriptors = append(descriptors, descriptor)
	}

	sorted, err := structure.SortByDependency(descriptors, false)
	if err != nil {
		return fmt.Errorf("sort by dependency: %w", err)
	}
	for _, descriptor := range sorted {
		if err := ib._addFile(descriptor); err != nil {
			return fmt.Errorf("add file %s: %w", descriptor.GetName(), err)
		}
		ib.img.SourceFilenames = append(ib.img.SourceFilenames, descriptor.GetName())
	}
	for _, file := range built.Prose {
		ib.addProseFile(file)
	}
	return nil
}

func (ib *imageBuilder) _addFile(file *descriptorpb.FileDescriptorProto) error {

	if _, ok := ib.includedFilenames[file.GetName()]; ok {
		for _, existingFile := range ib.img.File {
			if existingFile.GetName() != file.GetName() {
				continue
			}
			if !assertProtoFilesAreEqual(existingFile, file) {
				return fmt.Errorf("file %s already included with different content", file.GetName())
			}
			break
		}
		return nil
	}

	ib.img.File = append(ib.img.File, file)
	ib.includedFilenames[file.GetName()] = struct{}{}

	return nil
}

func (ib *imageBuilder) includeDependencies(ctx context.Context, deps *imageFiles) error {

	var addDependency func(file *descriptorpb.FileDescriptorProto) error
	var doFile func(file *descriptorpb.FileDescriptorProto) error

	addDependency = func(dep *descriptorpb.FileDescriptorProto) error {
		if err := ib._addFile(dep); err != nil {
			return fmt.Errorf("add file %s: %w", dep.GetName(), err)
		}
		return doFile(dep)
	}
	doFile = func(file *descriptorpb.FileDescriptorProto) error {
		for _, dependencyFilename := range file.Dependency {
			if _, ok := ib.includedFilenames[dependencyFilename]; ok {
				continue
			}

			if _, ok := psrc.BuiltinFile(dependencyFilename); ok {
				// not required to add
				continue
			}

			if dep, ok := deps.primary[dependencyFilename]; ok {
				if err := addDependency(dep); err != nil {
					return err
				}
				continue
			}

			if dep, ok := deps.dependencies[dependencyFilename]; ok {
				if err := addDependency(dep); err != nil {
					return err
				}
				continue
			}

			return fmt.Errorf("file %s not found in dependencies", dependencyFilename)
		}
		return nil
	}

	for _, file := range ib.img.File {
		err := doFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ib *imageBuilder) addProseFile(file *source_j5pb.ProseFile) {
	ib.img.Prose = append(ib.img.Prose, file)
}

func (ib *imageBuilder) include(ctx context.Context, img *source_j5pb.SourceImage) error {
	ib.img.Prose = append(ib.img.Prose, img.Prose...)
	ib.img.SourceFilenames = append(ib.img.SourceFilenames, img.SourceFilenames...)
	var err error

	source := &ext_j5pb.J5Source{
		Source: img.SourceName,
	}

	if img.SourceName == "" {
		return fmt.Errorf("source name is required for included image")
	}

	for _, file := range img.File {
		if file.Options == nil {
			file.Options = &descriptorpb.FileOptions{}
		}
		proto.SetExtension(file.Options, ext_j5pb.E_J5Source, source)

		err = ib._addFile(file)
		if err != nil {
			return err
		}
	}

	for _, pkg := range img.Packages {
		ib.addPackage(pkg)
	}
	return nil
}

func combineSourceImages(images []*source_j5pb.SourceImage) (*imageFiles, error) {

	fileMap := map[string]*descriptorpb.FileDescriptorProto{}
	depMap := map[string]*descriptorpb.FileDescriptorProto{}
	fileSourceMap := map[string]*source_j5pb.SourceImage{}
	for _, img := range images {
		isSource := map[string]bool{}
		for _, file := range img.SourceFilenames {
			isSource[file] = true
		}

		for _, file := range img.File {
			if !isSource[*file.Name] {
				depMap[*file.Name] = file
				continue
			}
			existing, ok := fileMap[*file.Name]
			if !ok {
				fileMap[*file.Name] = file
				fileSourceMap[*file.Name] = img
				continue
			}

			if !assertProtoFilesAreEqual(existing, file) {
				added := fileSourceMap[*file.Name]
				aName := fmt.Sprintf("%s:%s", added.SourceName, strVal(added.Version))
				bName := fmt.Sprintf("%s:%s", img.SourceName, strVal(img.Version))
				return nil, fmt.Errorf("file %q has conflicting content in %s and %s", *file.Name, aName, bName)
			}

		}
	}

	combined := &imageFiles{
		primary:      fileMap,
		dependencies: depMap,
	}

	return combined, nil
}

func assertProtoFilesAreEqual(aSrc, bSrc *descriptorpb.FileDescriptorProto) bool {

	if proto.Equal(aSrc, bSrc) {
		return true
	}

	a := proto.Clone(aSrc).(*descriptorpb.FileDescriptorProto)
	b := proto.Clone(bSrc).(*descriptorpb.FileDescriptorProto)
	// ignore source code info for comparison
	a.SourceCodeInfo = nil
	b.SourceCodeInfo = nil
	proto.ClearExtension(a.Options, ext_j5pb.E_J5Source)
	proto.ClearExtension(b.Options, ext_j5pb.E_J5Source)

	if proto.Equal(a, b) {
		return true
	}

	diff := cmp.Diff(a, b, protocmp.Transform())
	fmt.Fprintln(os.Stderr, diff)

	return false
}
