package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5build/internal/builtin"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (bundle *bundleSource) GetDependencies(ctx context.Context, resolver InputSource) (DependencySet, error) {
	j5Config, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}
	dependencies := make([]*source_j5pb.SourceImage, 0, len(j5Config.Dependencies))
	for _, dep := range j5Config.Dependencies {
		img, err := resolver.GetSourceImage(ctx, dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, img)
	}
	for _, dep := range j5Config.Includes {
		img, err := resolver.GetSourceImage(ctx, dep.Input)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, img)
	}
	return combineSourceImages(dependencies)

}
func (bundle *bundleSource) getDependencies(ctx context.Context, resolver InputSource) ([]*source_j5pb.SourceImage, error) {
	j5Config, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}
	dependencies := make([]*source_j5pb.SourceImage, len(j5Config.Dependencies))
	for idx, dep := range j5Config.Dependencies {
		img, err := resolver.GetSourceImage(ctx, dep)
		if err != nil {
			return nil, err
		}
		dependencies[idx] = img
	}
	return dependencies, nil
}

// getIncludes returns the images corresponding to the inputs. The returned
// slice will have the same indexes as the input.
func (bundle *bundleSource) getIncludes(ctx context.Context, resolver InputSource) ([]*source_j5pb.SourceImage, error) {
	j5Config, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}
	dependencies := make([]*source_j5pb.SourceImage, len(j5Config.Includes))
	for idx, spec := range j5Config.Includes {
		img, err := resolver.GetSourceImage(ctx, spec.Input)
		if err != nil {
			return nil, err
		}

		dependencies[idx] = img
	}

	return dependencies, nil
}

func (bundle *bundleSource) readImageFromDir(ctx context.Context, resolver InputSource) (*source_j5pb.SourceImage, error) {

	dependencyImages, err := bundle.getDependencies(ctx, resolver)
	if err != nil {
		return nil, err
	}

	includeImages, err := bundle.getIncludes(ctx, resolver)
	if err != nil {
		return nil, err
	}

	combinedDeps, err := combineSourceImages(append(dependencyImages, includeImages...))
	if err != nil {
		return nil, err
	}

	includedFilenames := make([]string, 0)
	for _, included := range includeImages {
		includedFilenames = append(includedFilenames, included.SourceFilenames...)
	}

	img, err := readImageFromDir(ctx, bundle.fs, includedFilenames, combinedDeps)
	if err != nil {
		return nil, err
	}

	img.Packages = make([]*source_j5pb.PackageInfo, len(bundle.config.Packages))
	for idx, pkg := range bundle.config.Packages {
		img.Packages[idx] = &source_j5pb.PackageInfo{
			Name:  pkg.Name,
			Prose: pkg.Prose,
			Label: pkg.Label,
		}
	}

	for _, included := range includeImages {
		img.Prose = append(img.Prose, included.Prose...)
		img.Packages = append(img.Packages, included.Packages...)
	}
	return img, nil
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type imageFiles struct {
	primary      map[string]*descriptorpb.FileDescriptorProto
	dependencies map[string]*descriptorpb.FileDescriptorProto
}

func (ii *imageFiles) GetDependencyFile(filename string) (*descriptorpb.FileDescriptorProto, error) {
	if file, ok := ii.primary[filename]; ok {
		return file, nil
	}
	if file, ok := ii.dependencies[filename]; ok {
		return file, nil
	}
	return nil, fmt.Errorf("could not find file %q", filename)
}

func (ii *imageFiles) ListDependencyFiles(prefix string) []string {

	files := make([]string, 0, len(ii.primary))
	for _, file := range ii.primary {
		name := file.GetName()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		files = append(files, name)
	}
	return files
}

func (ii *imageFiles) AllDependencyFiles() ([]*descriptorpb.FileDescriptorProto, []string) {

	files := make([]*descriptorpb.FileDescriptorProto, 0, len(ii.primary)+len(ii.dependencies))
	filenames := make([]string, 0, len(ii.primary))

	for _, file := range ii.primary {
		files = append(files, file)
		filenames = append(filenames, file.GetName())
	}
	for filename, file := range ii.dependencies {
		if _, ok := ii.primary[filename]; ok {
			continue
		}
		files = append(files, file)
	}
	return files, filenames
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

			if proto.Equal(existing, file) {
				continue
			}

			a := fileSourceMap[*file.Name]
			aName := fmt.Sprintf("%s:%s", a.SourceName, strVal(a.Version))
			bName := fmt.Sprintf("%s:%s", img.SourceName, strVal(img.Version))

			return nil, fmt.Errorf("file %q has conflicting content in %s and %s", *file.Name, aName, bName)
		}
	}

	combined := &imageFiles{
		primary:      fileMap,
		dependencies: depMap,
	}

	return combined, nil
}

type DependencySet interface {
	GetDependencyFile(filename string) (*descriptorpb.FileDescriptorProto, error)
	ListDependencyFiles(prefix string) []string
	AllDependencyFiles() ([]*descriptorpb.FileDescriptorProto, []string)
}

func readImageFromDir(ctx context.Context, bundleRoot fs.FS, includeFilenames []string, dependencies DependencySet) (*source_j5pb.SourceImage, error) {

	proseFiles := []*source_j5pb.ProseFile{}
	filenames := includeFilenames
	err := fs.WalkDir(bundleRoot, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".proto":
			filenames = append(filenames, path)
			return nil

		case ".md":
			data, err := fs.ReadFile(bundleRoot, path)
			if err != nil {
				return err
			}
			proseFiles = append(proseFiles, &source_j5pb.ProseFile{
				Path:    path,
				Content: data,
			})
			return nil

		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	parser := protoparse.Parser{
		ImportPaths:                     []string{""},
		IncludeSourceCodeInfo:           true,
		InterpretOptionsInUnlinkedFiles: true,

		WarningReporter: func(err reporter.ErrorWithPos) {
			log.WithFields(ctx, map[string]interface{}{
				"error": err.Error(),
			}).Warn("protoparse warning")
		},
		LookupImport: func(filename string) (*desc.FileDescriptor, error) {
			if builtin.IsBuiltInProto(filename) {
				ff, err := desc.LoadFileDescriptor(filename)
				if err != nil {
					return nil, fmt.Errorf("loading pre-loaded file %q: %w", filename, err)
				}
				return ff, nil
			}
			return nil, fmt.Errorf("could not find file %q", filename)

		},
		LookupImportProto: dependencies.GetDependencyFile,

		Accessor: func(filename string) (io.ReadCloser, error) {
			return bundleRoot.Open(filename)
		},
	}

	customDesc, err := parser.ParseFiles(filenames...)
	if err != nil {
		panicErr := protocompile.PanicError{}
		if errors.As(err, &panicErr) {
			fmt.Printf("PANIC: %s\n", panicErr.Stack)
		}

		return nil, err
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	return &source_j5pb.SourceImage{
		File:            realDesc.File,
		Prose:           proseFiles,
		SourceFilenames: filenames,
	}, nil
}
