package image

import (
	"context"
	"fmt"

	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"

	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/j5/internal/structure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Builder struct {
	img               *source_j5pb.SourceImage
	includedFilenames map[string]struct{}
}

func NewBuilder() *Builder {
	return &Builder{
		img:               &source_j5pb.SourceImage{},
		includedFilenames: make(map[string]struct{}),
	}
}

func NewBuilderFromImage(img *source_j5pb.SourceImage) *Builder {
	return &Builder{
		img:               img,
		includedFilenames: make(map[string]struct{}),
	}
}

func (ib *Builder) AddPackage(pkg *source_j5pb.Package) {
	ib.img.Packages = append(ib.img.Packages, pkg)
}

func (ib *Builder) AddBuilt(built *protobuild.BuiltPackage, source *ext_j5pb.J5Source) error {
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
		ib.AddProseFile(file)
	}
	return nil
}

func (ib *Builder) _addFile(file *descriptorpb.FileDescriptorProto) error {

	if _, ok := ib.includedFilenames[file.GetName()]; ok {
		for _, existingFile := range ib.img.File {
			if existingFile.GetName() != file.GetName() {
				continue
			}
			if !psrc.AssertProtoFilesAreEqual(existingFile, file) {
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

func (ib *Builder) IncludeDependencies(ctx context.Context, deps psrc.DescriptorFiles) error {

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

			if dep, ok := deps[dependencyFilename]; ok {
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

func (ib *Builder) AddProseFile(file *source_j5pb.ProseFile) {
	ib.img.Prose = append(ib.img.Prose, file)
}

func (ib *Builder) Include(ctx context.Context, img *source_j5pb.SourceImage) error {
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
		ib.AddPackage(pkg)
	}
	return nil
}

func (ib *Builder) Image() *source_j5pb.SourceImage {
	return ib.img
}
