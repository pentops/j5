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

	"github.com/bufbuild/protoyaml-go"
	"github.com/go-yaml/yaml"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	registry_spb "buf.build/gen/go/bufbuild/buf/grpc/go/buf/alpha/registry/v1alpha1/registryv1alpha1grpc"
	registry_pb "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
)

var ConfigPaths = []string{
	"j5.yaml",
	"jsonapi.yaml",
	"j5.yml",
	"jsonapi.yml",
	"ext/j5/j5.yaml",
	"ext/j5/j5.yml",
}

type bufLockFile struct {
	Deps []*bufLockFileDependency `yaml:"deps"`
}

type bufLockFileDependency struct {
	Remote     string `yaml:"remote"`
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Commit     string `yaml:"commit"`
	Digest     string `yaml:"digest"`
}

func ReadImageFromSourceDir(ctx context.Context, src string) (*source_j5pb.SourceImage, error) {
	fileStat, err := os.Lstat(src)
	if err != nil {
		return nil, err
	}
	if !fileStat.IsDir() {
		return nil, fmt.Errorf("src must be a directory")
	}

	var configData []byte
	found := false
	for _, filename := range ConfigPaths {
		configData, err = os.ReadFile(filepath.Join(src, filename))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		found = true
		break
	}

	if !found {
		return nil, fmt.Errorf("no config file found")
	}

	config := &source_j5pb.Config{}
	if err := protoyaml.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	extFiles, err := getDeps(ctx, src)
	if err != nil {
		return nil, err
	}

	proseFiles := []*source_j5pb.ProseFile{}
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
			proseFiles = append(proseFiles, &source_j5pb.ProseFile{
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

	return &source_j5pb.SourceImage{
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

	bufLockFile := &bufLockFile{}
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
