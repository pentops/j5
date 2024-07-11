package source

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pentops/log.go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"

	registry_spb "buf.build/gen/go/bufbuild/buf/grpc/go/buf/alpha/registry/v1alpha1/registryv1alpha1grpc"
	registry_pb "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
)

type BufLockFile struct {
	Version string                   `yaml:"version"`
	Deps    []*BufLockFileDependency `yaml:"deps"`
}

type BufLockFileDependency struct {
	Remote     string `yaml:"remote"`
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Commit     string `yaml:"commit"`
	Digest     string `yaml:"digest"`
	Name       string `yaml:"name"`
}

type BufCache struct {
	root        string
	client      registry_spb.DownloadServiceClient
	memoryCache map[string]FileSource
}

func NewBufCache() (*BufCache, error) {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache")
	specified := os.Getenv("BUF_CACHE_DIR")
	if specified != "" {
		cacheDir = specified
	}
	root := filepath.Join(cacheDir, "buf")
	bufClient, err := grpc.NewClient("buf.build:443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}
	registryClient := registry_spb.NewDownloadServiceClient(bufClient)

	return &BufCache{
		root:        root,
		client:      registryClient,
		memoryCache: map[string]FileSource{},
	}, nil
}

func (bc *BufCache) getDep(ctx context.Context, depSpec *BufLockFileDependency) (FileSource, error) {
	key := fmt.Sprintf("%s/%s:%s", depSpec.Owner, depSpec.Repository, depSpec.Commit)
	if cached, ok := bc.memoryCache[key]; ok {
		return cached, nil
	}
	files, err := bc.fetchDep(ctx, depSpec)
	if err != nil {
		return nil, err
	}
	bc.memoryCache[key] = files
	return files, nil
}

func (bc *BufCache) fetchDep(ctx context.Context, dep *BufLockFileDependency) (FileSource, error) {

	if fsCached, err := bc.tryV3FSDep(ctx, dep); err == nil {
		return fsCached, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if fsCached, err := bc.tryV2FSDep(ctx, dep); err == nil {
		return fsCached, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	externalFiles := map[string][]byte{}
	downloadRes, err := bc.client.Download(ctx, &registry_pb.DownloadRequest{
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

	return mapSource{
		files: externalFiles,
		name:  fmt.Sprintf("bufRemote:%s/%s:%s", dep.Owner, dep.Repository, dep.Commit),
	}, nil
}

func (bc *BufCache) GetDeps(ctx context.Context, root fs.FS, subDir string) ([]FileSource, error) {

	var lockFileData []byte

	searchPath := subDir
	for {
		lockFile, err := fs.ReadFile(root, path.Join(searchPath, "buf.lock"))
		if err == nil {
			lockFileData = lockFile

			log.WithFields(ctx, map[string]interface{}{
				"lockFile": path.Join(searchPath, "buf.lock"),
			}).Debug("found lock file")
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if searchPath == "." {
			break
		}
		searchPath = filepath.Dir(searchPath)
	}

	if lockFileData == nil {
		return nil, fmt.Errorf("buf.lock not found")
	}

	bufLockFile := &BufLockFile{}
	if err := yaml.Unmarshal(lockFileData, bufLockFile); err != nil {
		return nil, err
	}

	switch bufLockFile.Version {
	case "", "v1":

	case "v2":
		for _, dep := range bufLockFile.Deps {
			parts := strings.Split(dep.Name, "/")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid remote %s", dep.Remote)
			}

			if parts[0] != "buf.build" {
				return nil, fmt.Errorf("unsupported remote %s", parts[0])
			}
			dep.Owner = parts[1]
			dep.Repository = parts[2]
		}

	default:
		return nil, fmt.Errorf("unsupported buf.lock version %s", bufLockFile.Version)

	}

	allDeps := make([]FileSource, 0, len(bufLockFile.Deps))
	for _, dep := range bufLockFile.Deps {
		files, err := bc.getDep(ctx, dep)
		if err != nil {
			return nil, err
		}
		log.WithField(ctx, "dep", fmt.Sprintf("buf.build/%s/%s", dep.Owner, dep.Repository)).Debug("including buf dep")
		allDeps = append(allDeps, files)
	}

	return allDeps, nil

}

func (bc *BufCache) tryV3FSDep(ctx context.Context, dep *BufLockFileDependency) (FileSource, error) {
	ctx = log.WithFields(ctx, map[string]interface{}{
		"owner":      dep.Owner,
		"repository": dep.Repository,
		"commit":     dep.Commit,
	})

	v3Dep := filepath.Join(bc.root, "v3", "modules", "shake256", "buf.build", dep.Owner, dep.Repository, dep.Commit, "files")

	if _, err := os.Stat(v3Dep); err != nil {
		log.WithField(ctx, "v3Path", v3Dep).Debug("No v3 found, falling back to v2")
		return nil, err
	}

	log.WithField(ctx, "v3Path", v3Dep).Debug("found v3 dep")
	return fsSource{
		fs:   os.DirFS(v3Dep),
		name: fmt.Sprintf("bufv3:%s/%s/%s", dep.Owner, dep.Repository, dep.Commit),
	}, nil
}

type bufv2Dep struct {
	root  string
	files map[string]string
}

func (bd bufv2Dep) GetFile(filename string) (io.ReadCloser, error) {
	if f, ok := bd.files[filename]; ok {
		return os.Open(filepath.Join(bd.root, f))
	}
	return nil, os.ErrNotExist
}

func (bd bufv2Dep) Name() string {
	return "bufv2:" + bd.root
}

func (bc *BufCache) tryV2FSDep(ctx context.Context, dep *BufLockFileDependency) (FileSource, error) {

	contentStr := dep.Digest
	hdr, rem := contentStr[9:11], contentStr[11:]

	indexPath := filepath.Join("v2", "module", "buf.build", bc.root, dep.Owner, dep.Repository, "blobs", hdr, rem)
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(indexContent), "\n")
	files := map[string]string{}

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

		if _, ok := files[filename]; ok {
			return nil, fmt.Errorf("duplicate file %s", filename)
		}
		files[filename] = filepath.Join(fDir, fPath)

	}

	return bufv2Dep{
		root:  filepath.Join(bc.root, dep.Owner, dep.Repository, "blobs"),
		files: files,
	}, nil
}
