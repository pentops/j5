package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/prototools/protosrc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func codeGeneratorRequestFromSource(ctx context.Context, bundle *bundle) (*pluginpb.CodeGeneratorRequest, error) {

	out := &pluginpb.CodeGeneratorRequest{
		CompilerVersion: nil,
	}

	includeFiles := map[string]bool{}

	walkRoot, err := bundle.fs()
	if err != nil {
		return nil, err
	}

	err = fs.WalkDir(walkRoot, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".proto":
			out.FileToGenerate = append(out.FileToGenerate, path)
			includeFiles[path] = true
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bufCache := protosrc.NewBufCache()
	extFiles, err := bufCache.GetDeps(ctx, bundle.repo.repoRoot, bundle.dirInRepo)
	if err != nil {
		return nil, err
	}

	parser := protoparse.Parser{
		ImportPaths:           []string{""},
		IncludeSourceCodeInfo: true,

		Accessor: func(filename string) (io.ReadCloser, error) {
			if content, ok := extFiles[filename]; ok {
				return io.NopCloser(bytes.NewReader(content)), nil
			}
			return walkRoot.Open(filename)
		},
	}

	customDesc, err := parser.ParseFiles(out.FileToGenerate...)
	if err != nil {
		return nil, err
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

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
		for _, f := range realDesc.File {
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

	for _, file := range realDesc.File {
		if err := addFile(file); err != nil {
			return nil, err
		}
	}

	return out, nil
}
