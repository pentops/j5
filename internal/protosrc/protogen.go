package protosrc

import (
	"fmt"
	"sort"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func SortByDependency(files []*descriptorpb.FileDescriptorProto) ([]*descriptorpb.FileDescriptorProto, error) {
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
		// doesn't matter if it can't find, we assume its in another bundle...
		return nil
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
	// TODO: Reuse the SortByDependency logic.

	out := &pluginpb.CodeGeneratorRequest{
		CompilerVersion: nil,
		FileToGenerate:  img.SourceFilenames,
	}

	includeFiles := map[string]bool{}
	for _, file := range img.SourceFilenames {
		includeFiles[file] = true
	}

	// Prepare the files for the generator.
	// From the docs on out.ProtoFile:
	// FileDescriptorProtos for all files in files_to_generate and everything
	// they import.  The files will appear in topological order, so each file
	// appears before any file that imports it.

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
