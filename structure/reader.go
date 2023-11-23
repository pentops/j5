package structure

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protoyaml-go"
	"github.com/go-yaml/yaml"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/jsonapi/gen/v1/jsonapi_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	registry_spb "buf.build/gen/go/bufbuild/buf/grpc/go/buf/alpha/registry/v1alpha1/registryv1alpha1grpc"
	registry_pb "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
)

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

func ReadFileDescriptorSet(ctx context.Context, src string) (*descriptorpb.FileDescriptorSet, error) {
	descriptors := &descriptorpb.FileDescriptorSet{}

	if src == "-" {
		protoData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		if err := proto.Unmarshal(protoData, descriptors); err != nil {
			return nil, err
		}
		return descriptors, nil
	}

	fileStat, err := os.Lstat(src)
	if err != nil {
		return nil, err
	}

	if !fileStat.IsDir() {
		protoData, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		if err := proto.Unmarshal(protoData, descriptors); err != nil {
			return nil, err
		}

		return descriptors, nil
	}

	image, err := ReadImageFromSourceDir(ctx, src)
	if err != nil {
		return nil, err
	}

	return &descriptorpb.FileDescriptorSet{
		File: image.File,
	}, nil
}

func ReadImageFromSourceDir(ctx context.Context, src string) (*jsonapi_pb.Image, error) {
	fileStat, err := os.Lstat(src)
	if err != nil {
		return nil, err
	}
	if !fileStat.IsDir() {
		return nil, fmt.Errorf("src must be a directory")
	}

	configData, err := os.ReadFile(filepath.Join(src, "jsonapi.yaml"))
	if err != nil {
		log.Fatal(err.Error())
	}
	config := &jsonapi_pb.Config{}
	if err := protoyaml.Unmarshal(configData, config); err != nil {
		log.Fatal(err.Error())
	}

	extFiles, err := getDeps(ctx, src)
	if err != nil {
		return nil, err
	}

	proseFiles := []*jsonapi_pb.ProseFile{}
	filenames := []string{}
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
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
			filenames = append(filenames, rel)
			return nil

		case ".md":
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			proseFiles = append(proseFiles, &jsonapi_pb.ProseFile{
				Path:    rel,
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
		ImportPaths:           []string{""},
		IncludeSourceCodeInfo: true,

		Accessor: func(filename string) (io.ReadCloser, error) {
			if content, ok := extFiles[filename]; ok {
				return io.NopCloser(bytes.NewReader(content)), nil
			}
			return os.Open(filepath.Join(src, filename))
		},
	}

	customDesc, err := parser.ParseFiles(filenames...)
	if err != nil {
		return nil, err
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	return &jsonapi_pb.Image{
		File:     realDesc.File,
		Packages: config.Packages,
		Codec:    config.Options,
		Prose:    proseFiles,
		Registry: config.Registry,
	}, nil

}

func getDeps(ctx context.Context, src string) (map[string][]byte, error) {
	// TODO: Use Buf Cache if available

	lockFile, err := os.ReadFile(filepath.Join(src, "buf.lock"))
	if err != nil {
		return nil, err
	}

	bufLockFile := &BufLockFile{}
	if err := yaml.Unmarshal(lockFile, bufLockFile); err != nil {
		return nil, err
	}

	bufClient, err := grpc.Dial("buf.build:443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
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
