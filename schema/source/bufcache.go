package source

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"

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

type file struct {
	path    string
	content []byte
}

type bufCache struct {
	root string
}

func (bc *bufCache) tryDep(dep *BufLockFileDependency) ([]file, error) {
	contentStr := dep.Digest
	hdr, rem := contentStr[9:11], contentStr[11:]

	indexPath := filepath.Join(bc.root, dep.Owner, dep.Repository, "blobs", hdr, rem)
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("buf mod not found: %s", indexPath)
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(indexContent), "\n")
	files := make([]file, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		header, fDir, fPath, filename := line[:8], line[9:11], line[11:137], line[139:]

		if header != "shake256" {
			return nil, fmt.Errorf("invalid cache entry")
		}

		if !strings.HasSuffix(filename, ".proto") {
			continue
		}

		fileContent, err := os.ReadFile(filepath.Join(bc.root, dep.Owner, dep.Repository, "blobs", fDir, fPath))
		if err != nil {
			return nil, err
		}

		files = append(files, file{path: filename, content: fileContent})
	}

	return files, nil
}

func getBufDeps(ctx context.Context, src string) (map[string][]byte, error) {
	bufCacheDir := filepath.Join(os.Getenv("HOME"), ".cache/buf/v2/module/buf.build")
	bufCache := &bufCache{root: bufCacheDir}

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
		cached, err := bufCache.tryDep(dep)
		if err != nil {
			return nil, err
		}
		if cached != nil {
			for _, file := range cached {
				if _, ok := externalFiles[file.path]; ok {
					return nil, fmt.Errorf("duplicate file %s", file.path)
				}
				externalFiles[file.path] = file.content
			}
			continue
		}

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
