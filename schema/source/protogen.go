package source

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v2"

	registry_spb "buf.build/gen/go/bufbuild/buf/grpc/go/buf/alpha/registry/v1alpha1/registryv1alpha1grpc"
	registry_pb "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
)

func CodeGeneratorRequestFromSource(ctx context.Context, src string) (*pluginpb.CodeGeneratorRequest, error) {

	out := &pluginpb.CodeGeneratorRequest{
		CompilerVersion: &pluginpb.Version{
			Major: ptr(0),
			Minor: ptr(0),
			Patch: ptr(0),
		},
	}

	includeFiles := map[string]bool{}

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		switch ext {
		case ".proto":
			out.FileToGenerate = append(out.FileToGenerate, rel)
			includeFiles[rel] = true
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	extFiles, err := getBufDeps(ctx, src)
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
			return os.Open(filepath.Join(src, filename))
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

func ptr(i int32) *int32 {
	return &i
}

type BufLockFile struct {
	Deps []*BufLockFileDependency `yaml:"deps"`
}

type BufLockFileDependency struct {
	Remote     string `yaml:"remote"`
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Commit     string `yaml:"commit"`
	Digest     string `yaml:"digest"`
}

func getBufDeps(ctx context.Context, src string) (map[string][]byte, error) {
	// TODO: Use Buf Cache if available

	lockFile, err := os.ReadFile(filepath.Join(src, "buf.lock"))
	if err != nil {
		return nil, err
	}

	bufLockFile := &BufLockFile{}
	if err := yaml.Unmarshal(lockFile, bufLockFile); err != nil {
		return nil, err
	}

	bufClient, err := grpc.NewClient("buf.build:443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}
	registryClient := registry_spb.NewDownloadServiceClient(bufClient)

	externalFiles := map[string][]byte{}
	for _, dep := range bufLockFile.Deps {
		downloadRes, err := registryClient.Download(ctx, &registry_pb.DownloadRequest{
			Owner:      dep.Owner,
			Repository: dep.Repository,
			Reference:  dep.Commit,
		})
		if err != nil {
			return nil, err
		}

		for _, file := range downloadRes.Module.Files {
			if _, ok := externalFiles[file.Path]; ok {
				return nil, fmt.Errorf("duplicate file %s", file.Path)
			}

			externalFiles[file.Path] = file.Content
		}
	}

	return externalFiles, nil

}

/*
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
}*/
