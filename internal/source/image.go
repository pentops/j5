package source

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"google.golang.org/protobuf/types/descriptorpb"
)

type imageBuilder struct {
	img               *source_j5pb.SourceImage
	includedFilenames map[string]struct{}
	deps              *imageFiles
}

func newImageBuilder(deps *imageFiles) *imageBuilder {
	return &imageBuilder{
		img:               &source_j5pb.SourceImage{},
		includedFilenames: make(map[string]struct{}),
		deps:              deps,
	}
}

func (ib *imageBuilder) addPackage(pkg *source_j5pb.PackageInfo) {
	ib.img.Packages = append(ib.img.Packages, pkg)
}

func (ib *imageBuilder) addFile(file *descriptorpb.FileDescriptorProto, asSource bool) error {
	for _, dependencyFilename := range file.Dependency {
		if _, ok := ib.includedFilenames[dependencyFilename]; ok {
			continue
		}

		if dep, ok := ib.deps.primary[dependencyFilename]; ok {
			if err := ib.addFile(dep, false); err != nil {
				return fmt.Errorf("add file %s: %w", dependencyFilename, err)
			}
			continue
		}

		if dep, ok := ib.deps.dependencies[dependencyFilename]; ok {
			if err := ib.addFile(dep, false); err != nil {
				return fmt.Errorf("add file %s: %w", dependencyFilename, err)
			}
			continue
		}

		if _, ok := psrc.BuiltinFile(dependencyFilename); ok {
			// not required to add
			continue
		}

		return fmt.Errorf("file %s not found in dependencies", dependencyFilename)
	}

	ib.img.File = append(ib.img.File, file)
	ib.includedFilenames[file.GetName()] = struct{}{}
	if asSource {
		ib.img.SourceFilenames = append(ib.img.SourceFilenames, file.GetName())
	}

	return nil
}

func (ib *imageBuilder) addProseFile(file *source_j5pb.ProseFile) {
	ib.img.Prose = append(ib.img.Prose, file)
}

func (ib *imageBuilder) include(img *source_j5pb.SourceImage) {
	ib.img.Prose = append(ib.img.Prose, img.Prose...)
	ib.img.SourceFilenames = append(ib.img.SourceFilenames, img.SourceFilenames...)
	for _, file := range img.File {
		ib.addFile(file, false)
	}

	for _, pkg := range img.Packages {
		ib.addPackage(pkg)
	}
}
