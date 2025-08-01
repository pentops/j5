package structure

import (
	"fmt"
	"sort"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func SortByDependency(files []*descriptorpb.FileDescriptorProto, resolveMissing bool) ([]*descriptorpb.FileDescriptorProto, error) {
	// Sort to topological order, so each file
	// appears before any file that imports it.
	// For consistent ordering, files are alphabetical first.

	sort.Slice(files, func(i, j int) bool {
		if *files[i].Name == *files[j].Name {
			return false
		}
		return *files[i].Name < *files[j].Name
	})

	workingOn := make(map[string]bool)
	hasFile := make(map[string]bool)
	out := make([]*descriptorpb.FileDescriptorProto, 0, len(files))

	var addFile func(file *descriptorpb.FileDescriptorProto) error

	requireFile := func(name string) error {
		for _, f := range files {
			if *f.Name == name {
				return addFile(f)
			}
		}
		if !resolveMissing {
			// doesn't matter if it can't find, we assume its in another bundle...
			return nil
		}
		file, ok := psrc.BuiltinFile(name)
		if !ok {
			return fmt.Errorf("file %s not found", name)
		}
		return addFile(file)
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

		out = append(out, file)
		delete(workingOn, *file.Name)
		hasFile[*file.Name] = true
		return nil
	}

	for _, file := range files {
		if err := addFile(file); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func CodeGeneratorRequestFromImage(img *source_j5pb.SourceImage) (*pluginpb.CodeGeneratorRequest, error) {

	out := &pluginpb.CodeGeneratorRequest{
		CompilerVersion: nil,
		FileToGenerate:  img.SourceFilenames,
	}

	includeFiles := map[string]bool{}
	for _, file := range img.SourceFilenames {
		includeFiles[file] = true
	}

	filesInOrder, err := SortByDependency(img.File, true)
	if err != nil {
		return nil, fmt.Errorf("sorting files: %w", err)
	}
	out.ProtoFile = filesInOrder

	for _, file := range filesInOrder {
		if includeFiles[*file.Name] {
			out.SourceFileDescriptors = append(out.SourceFileDescriptors, file)
		}
	}

	return out, nil
}
