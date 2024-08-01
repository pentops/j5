package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type FileSource interface {
	GetFile(filename string) (io.ReadCloser, error)
	Name() string
}

type fsSource struct {
	fs   fs.FS
	name string
}

func (f fsSource) Name() string {
	return f.name
}

func (f fsSource) GetFile(filename string) (io.ReadCloser, error) {
	return f.fs.Open(filename)
}

type mapSource struct {
	files map[string][]byte
	name  string
}

func (f mapSource) Name() string {
	return f.name
}

func (f mapSource) GetFile(filename string) (io.ReadCloser, error) {
	content, ok := f.files[filename]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func bundleProtoparse(ctx context.Context, rootBundle *bundle, files []string) (*descriptorpb.FileDescriptorSet, error) {
	walkRoot, err := rootBundle.fs()
	if err != nil {
		return nil, err
	}

	allSources := []FileSource{
		fsSource{
			fs:   walkRoot,
			name: rootBundle.debugName,
		},
	}

	depBundles := make([]*bundle, 0, len(rootBundle.refConfig.Deps))

	for _, localDep := range rootBundle.refConfig.Deps {
		depBundle, ok := rootBundle.repo.bundles[localDep]
		if !ok {
			return nil, fmt.Errorf("unknown local dependency %q", localDep)
		}
		depBundles = append(depBundles, depBundle)

		bundleFS, err := depBundle.fs()
		if err != nil {
			return nil, err
		}

		log.WithField(ctx, "dep", localDep).Debug("adding local dep")
		allSources = append(allSources, fsSource{
			fs:   bundleFS,
			name: depBundle.debugName,
		})
	}

	bufCache, err := NewBufCache()
	if err != nil {
		return nil, err
	}

	bufDeps, err := bufCache.GetDeps(ctx, rootBundle.repo.repoRoot, rootBundle.dirInRepo)
	if err != nil {
		return nil, err
	}
	allSources = append(allSources, bufDeps...)

	for _, depBundle := range depBundles {

		bufDeps, err := bufCache.GetDeps(ctx, depBundle.repo.repoRoot, depBundle.dirInRepo)
		if err != nil {
			return nil, err
		}

		allSources = append(allSources, bufDeps...)
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

		Accessor: func(filename string) (io.ReadCloser, error) {
			for _, src := range allSources {
				content, err := src.GetFile(filename)
				if err == nil {
					return content, nil
				}
			}
			return nil, fs.ErrNotExist
		},
	}

	customDesc, err := parser.ParseFiles(files...)
	if err != nil {
		panicErr := protocompile.PanicError{}
		if errors.As(err, &panicErr) {
			fmt.Printf("PANIC: %s\n", panicErr.Stack)
		}

		return nil, err
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)
	return realDesc, nil
}

func codeGeneratorRequestFromSource(ctx context.Context, bundle *bundle) (*pluginpb.CodeGeneratorRequest, error) {

	img, err := readImageFromDir(ctx, bundle)
	if err != nil {
		return nil, err
	}

	return codeGeneratorRequestFromImage(img)

}

func codeGeneratorRequestFromImage(img *source_j5pb.SourceImage) (*pluginpb.CodeGeneratorRequest, error) {

	out := &pluginpb.CodeGeneratorRequest{
		CompilerVersion: nil,
		FileToGenerate:  img.SourceFilenames,
	}

	includeFiles := map[string]bool{}
	for _, file := range img.File {
		includeFiles[*file.Name] = true
	}

	// Prepare the files for the generator.
	// From the docs on out.ProtoFile:
	// FileDescriptorProtos for all files in files_to_generate and everything
	// they import.  The files will appear in topological order, so each file
	// appears before any file that imports it.

	// TODO: For now we are only including files that are in the FileToGenerate list, we should include the dependencies as well

	workingOn := make(map[string]bool)
	hasFile := make(map[string]bool)

	var addFile func(file *descriptorpb.FileDescriptorProto) error

	requireFile := func(name string) error {
		for _, f := range img.File {
			if *f.Name == name {
				return addFile(f)
			}
		}
		return fmt.Errorf("could not find file %q", name)
	}

	addFile = func(file *descriptorpb.FileDescriptorProto) error {
		if hasFile[*file.Name] {
			return nil
		}

		if workingOn[*file.Name] {
			return fmt.Errorf("circular dependency detected: %s", *file.Name)
		}
		workingOn[*file.Name] = true

		for _, dep := range file.Dependency {
			if err := requireFile(dep); err != nil {
				return fmt.Errorf("resolving dep %s for %s: %w", dep, *file.Name, err)
			}
		}

		out.ProtoFile = append(out.ProtoFile, file)
		if includeFiles[*file.Name] {
			out.SourceFileDescriptors = append(out.SourceFileDescriptors, file)
		}

		delete(workingOn, *file.Name)
		hasFile[*file.Name] = true

		return nil
	}

	for _, file := range img.File {
		if err := addFile(file); err != nil {
			return nil, err
		}
	}

	return out, nil
}
