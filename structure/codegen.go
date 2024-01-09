package structure

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type CodeGenOptions struct {
	PackagePrefix string
}

func CodeGeneratorRequestFromDescriptors(opts CodeGenOptions, descriptors *descriptorpb.FileDescriptorSet) (*pluginpb.CodeGeneratorRequest, error) {

	out := &pluginpb.CodeGeneratorRequest{
		CompilerVersion: &pluginpb.Version{
			Major: ptr(0),
			Minor: ptr(0),
			Patch: ptr(0),
		},
	}

	workingOn := make(map[string]bool)
	hasFile := make(map[string]bool)

	var addFile func(file *descriptorpb.FileDescriptorProto) error

	requireFile := func(name string) error {
		for _, f := range descriptors.File {
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

		delete(workingOn, *file.Name)
		hasFile[*file.Name] = true

		return nil
	}

	for _, file := range descriptors.File {
		if strings.HasPrefix(*file.Name, opts.PackagePrefix) {
			out.FileToGenerate = append(out.FileToGenerate, *file.Name)
			out.SourceFileDescriptors = append(out.SourceFileDescriptors, file)
			if err := addFile(file); err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

func ptr(i int32) *int32 {
	return &i
}
