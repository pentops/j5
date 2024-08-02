package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	registry_pb "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
)

func (src *Source) GetInput(ctx context.Context, input *config_j5pb.Input) (Input, error) {

	var cachableInput *imageBundle

	switch st := input.Type.(type) {
	case *config_j5pb.Input_Local:
		bundle, ok := src.thisRepo.bundles[st.Local]
		if !ok {
			return nil, fmt.Errorf("bundle %q not found", st.Local)
		}
		return bundle, nil

	case *config_j5pb.Input_Repo_:
		var err error
		repoRoot, debugName, err := anyRoot(st.Repo.Root, st.Repo.Dir)
		if err != nil {
			return nil, fmt.Errorf("resolving repo root: %w", err)
		}
		repo, err := src.newRepo(debugName, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("input %s: %w", st.Repo.Root, err)
		}
		bundle, ok := repo.bundles[st.Repo.Bundle]
		if !ok {
			return nil, fmt.Errorf("bundle %q not found in repo %q", st.Repo.Bundle, st.Repo.Root)
		}
		return bundle, nil

	case *config_j5pb.Input_Registry_:
		bundle, err := src.registryInput(ctx, st.Registry)
		if err != nil {
			return nil, fmt.Errorf("registry input %s: %w", st.Registry.Name, err)
		}
		cachableInput = bundle

	case *config_j5pb.Input_BufRegistry_:
		bundle, err := src.bufRegistryInput(ctx, st.BufRegistry)
		if err != nil {
			return nil, fmt.Errorf("buf registry input %s: %w", st.BufRegistry.Name, err)
		}
		cachableInput = bundle

	default:
		return nil, fmt.Errorf("unsupported source type %T", input.Type)
	}

	if src.j5Cache != nil && cachableInput.version != "" {
		if err := src.j5Cache.put(ctx, cachableInput.name, cachableInput.version, cachableInput.source); err != nil {
			log.WithError(ctx, err).Error("failed to cache input")
		}
	}

	return cachableInput, nil
}

func (src *Source) UpdateLocks(ctx context.Context) error {

	if src.lockWriter == nil {
		return fmt.Errorf("lock writer not set")
	}

	lockFile := &config_j5pb.LockFile{}
	for _, bundle := range src.thisRepo.bundles {
		var lock *config_j5pb.InputLock
		cfg, err := bundle.J5Config()
		if err != nil {
			return err
		}
		for _, dep := range cfg.Dependencies {
			switch st := dep.Type.(type) {
			/*
				case *config_j5pb.Input_Registry_:
					regDep := proto.Clone(st.Registry).(*config_j5pb.Input_Registry)
					regDep.Version = nil
					bundle, err := src.registryLatest(ctx, regDep)
					if err != nil {
						return err
					}
					bundleDep = bundle
			*/
			case *config_j5pb.Input_BufRegistry_:
				regDep := proto.Clone(st.BufRegistry).(*config_j5pb.Input_BufRegistry)
				regDep.Version = nil
				bundle, err := src.bufRegistryLatest(ctx, regDep)
				if err != nil {
					return err
				}
				lock = bundle
			}
		}
		if lock == nil {
			continue
		}

		if lock.Version == "" {
			return fmt.Errorf("no version for %s", lock.Name)
		}

		lockFile.Inputs = append(lockFile.Inputs, lock)
	}

	src.locks = lockFile
	return src.lockWriter.write(src.locks)
}

func (src *Source) getInputLockVersion(name string) string {
	if src.locks == nil {
		return ""
	}
	for _, dep := range src.locks.Inputs {
		if dep.Name == name {
			return dep.Version
		}
	}
	return ""
}

func (src *Source) getCachedInput(ctx context.Context, name, version string) (*imageBundle, bool) {
	if src.j5Cache == nil {
		return nil, false
	}
	image, ok := src.j5Cache.tryGet(ctx, name, version)
	if !ok {
		return nil, false
	}
	return &imageBundle{
		source:  image,
		name:    name,
		version: version,
	}, true
}

func (src *Source) registryInput(ctx context.Context, input *config_j5pb.Input_Registry) (*imageBundle, error) {

	if src.remoteRegistry == "" {
		return nil, fmt.Errorf("remote registry not set ($J5_REGISTRY)")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("registry input name not set")
	}
	if input.Organization == "" {
		return nil, fmt.Errorf("registry input organization not set")
	}

	name := fmt.Sprintf("registry/v1/%s/%s", input.Organization, input.Name)
	ctx = log.WithField(ctx, "bundle", name)
	var version string
	if input.Version == nil {
		version = src.getInputLockVersion(name)
		log.WithField(ctx, "version", version).Debug("using lock version")
	} else {
		version = *input.Version
	}

	ctx = log.WithField(ctx, "version", version)
	if version != "" {
		if cached, ok := src.getCachedInput(ctx, name, version); ok {
			log.Debug(ctx, "using cached input")
			return cached, nil
		}
	}
	if version == "" {
		version = "latest"
	}
	log.Debug(ctx, "cache miss")

	imageURL := fmt.Sprintf("%s/%s/%s/image.bin", src.remoteRegistry, name, version)
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating registry input request: %w", err)
	}
	req = req.WithContext(ctx)

	res, err := src.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching registry input: %q %w", imageURL, err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading registry input: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching registry input: %q %s %q", imageURL, res.Status, string(data))
	}

	apiDef := &source_j5pb.SourceImage{}
	if err := proto.Unmarshal(data, apiDef); err != nil {
		return nil, fmt.Errorf("unmarshalling registry input %s: %w", imageURL, err)
	}

	return &imageBundle{
		name:    fmt.Sprintf("%s/%s", input.Organization, input.Name),
		version: version,
		source:  apiDef,
	}, nil
}

func (src *Source) bufRegistryLatest(ctx context.Context, input *config_j5pb.Input_BufRegistry) (*config_j5pb.InputLock, error) {

	reference := ""
	if input.Reference != nil {
		reference = *input.Reference
	}
	res, err := src.bufCache.versionClient.GetRepositoryCommitByReference(ctx, &registry_pb.GetRepositoryCommitByReferenceRequest{
		RepositoryOwner: input.Owner,
		RepositoryName:  input.Name,
		Reference:       reference,
	})
	if err != nil {
		return nil, err
	}

	return &config_j5pb.InputLock{
		Name:    fmt.Sprintf("buf/%s/%s", input.Owner, input.Name),
		Version: res.RepositoryCommit.Name,
	}, nil

}

func (src *Source) bufRegistryInput(ctx context.Context, input *config_j5pb.Input_BufRegistry) (*imageBundle, error) {

	name := fmt.Sprintf("buf/%s/%s", input.Owner, input.Name)
	ctx = log.WithField(ctx, "bundle", name)
	var version string
	if input.Version != nil {
		version = *input.Version
	} else {
		version = src.getInputLockVersion(name)
		log.WithField(ctx, "version", version).Debug("using lock version")
		if version == "" && input.Reference != nil {
			version = *input.Reference
		}
	}
	ctx = log.WithField(ctx, "version", version)

	if version != "" {
		if cached, ok := src.getCachedInput(ctx, name, version); ok {
			log.Debug(ctx, "using cached input")
			return cached, nil
		}
	}
	log.Debug(ctx, "cache miss")

	downloadRes, err := src.bufCache.client.Download(ctx, &registry_pb.DownloadRequest{
		Owner:      input.Owner,
		Repository: input.Name,
		Reference:  version,
	})
	if err != nil {
		return nil, err
	}

	allSources := []FileSource{}

	filenames := []string{}
	fileMap := mapSource{
		files: map[string][]byte{},
	}
	for _, file := range downloadRes.Module.Files {
		if _, ok := fileMap.files[file.Path]; ok {
			return nil, fmt.Errorf("duplicate file %s", file.Path)
		}
		fileMap.files[file.Path] = file.Content
		filenames = append(filenames, file.Path)
	}
	allSources = append(allSources, fileMap)

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

	customDesc, err := parser.ParseFiles(filenames...)
	if err != nil {
		panicErr := protocompile.PanicError{}
		if errors.As(err, &panicErr) {
			fmt.Printf("PANIC: %s\n", panicErr.Stack)
		}

		return nil, fmt.Errorf("parsing buf input %q: %w", name, err)
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	img := &source_j5pb.SourceImage{
		File:            realDesc.File,
		SourceFilenames: filenames,
	}

	return &imageBundle{
		source:  img,
		name:    name,
		version: version,
	}, nil

}

func anyRoot(name, subdir string) (fs.FS, string, error) {

	if strings.HasPrefix(name, "file://") {
		fullPath := strings.TrimPrefix(name, "file://")
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			return nil, "", fmt.Errorf("resolving absolute path: %w", err)
		}

		if subdir != "" {
			absPath = filepath.Join(absPath, subdir)
		}

		workDir, err := os.Getwd()
		if err == nil {
			relPath, err := filepath.Rel(workDir, absPath)
			if err == nil {
				absPath = relPath
			}
		}
		return os.DirFS(absPath), absPath, nil
	}

	return nil, "", fmt.Errorf("unsupported scheme %q", name)
}
