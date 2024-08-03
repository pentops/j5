package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
		return src.cacheDance(ctx, cacheSpec{
			repoType:  "registry",
			owner:     st.Registry.Owner,
			repoName:  st.Registry.Name,
			version:   st.Registry.Version,
			reference: coalesce(st.Registry.Reference, ptr("main")),
		}, src.regClient.input)
		//func(ctx context.Context, version string) (*imageBundle, error) {
	//		return src.regClient.input(ctx, st.Registry.Owner, st.Registry.Name, version)
	//	})

	case *config_j5pb.Input_BufRegistry_:
		return src.cacheDance(ctx, cacheSpec{
			repoType:  "buf",
			owner:     st.BufRegistry.Owner,
			repoName:  st.BufRegistry.Name,
			version:   st.BufRegistry.Version,
			reference: st.BufRegistry.Reference,
		}, src.bufRegistryInput)

	default:
		return nil, fmt.Errorf("unsupported source type %T", input.Type)
	}

}

type cacheSpec struct {
	repoType  string
	owner     string
	repoName  string
	version   *string
	reference *string
}

func (src *Source) cacheDance(ctx context.Context, spec cacheSpec, callback func(ctx context.Context, owner string, name string, version string) (*imageBundle, error)) (*imageBundle, error) {

	fullName := fmt.Sprintf("%s/%s/%s", spec.repoType, spec.owner, spec.repoName)
	ctx = log.WithField(ctx, "bundle", fullName)
	var version *string
	if spec.version != nil {
		version = ptr(*spec.version)
	} else if lockVersion := src.getInputLockVersion(fullName); lockVersion != nil {
		log.WithField(ctx, "lockVersion", *lockVersion).Debug("using lock version")
		version = ptr(*lockVersion)
	}

	ctx = log.WithField(ctx, "version", version)
	// only use cache if version is explicit, otherwise needs to pull latest
	if version != nil {
		if cached, ok := src.getCachedInput(ctx, fullName, *version); ok {
			log.Debug(ctx, "using cached input")
			return cached, nil
		}
	}
	if version == nil {
		if spec.reference != nil {
			version = ptr(*spec.reference)
		} else {
			version = ptr("main")
		}
	}

	ctx = log.WithField(ctx, "depVersion", *version)
	log.Debug(ctx, "cache miss")

	bundle, err := callback(ctx, spec.owner, spec.repoName, *version)
	if err != nil {
		return nil, err
	}

	if src.j5Cache != nil && bundle.version != "" {
		if err := src.j5Cache.put(ctx, fullName, bundle.version, bundle.source); err != nil {
			log.WithError(ctx, err).Error("failed to cache input")
		}
	}

	return bundle, nil
}

func (src *Source) UpdateLocks(ctx context.Context) error {

	if src.lockWriter == nil {
		return fmt.Errorf("lock writer not set")
	}

	lockFile := &config_j5pb.LockFile{}
	for _, bundle := range src.thisRepo.bundles {

		cfg, err := bundle.J5Config()
		if err != nil {
			return err
		}
		for _, dep := range cfg.Dependencies {

			var lock *config_j5pb.InputLock

			switch st := dep.Type.(type) {
			case *config_j5pb.Input_Registry_:
				regDep := proto.Clone(st.Registry).(*config_j5pb.Input_Registry)
				regDep.Version = nil
				bundle, err := src.registryLatest(ctx, regDep)
				if err != nil {
					return err
				}
				lock = bundle
			case *config_j5pb.Input_BufRegistry_:
				regDep := proto.Clone(st.BufRegistry).(*config_j5pb.Input_BufRegistry)
				regDep.Version = nil
				bundle, err := src.bufRegistryLatest(ctx, regDep)
				if err != nil {
					return err
				}
				lock = bundle
			}

			if lock == nil {
				continue
			}

			ctx = log.WithFields(ctx, map[string]interface{}{
				"bundle":      bundle.Name,
				"dependency":  lock.Name,
				"lockVersion": lock.Version,
			})

			if lock.Version == "" {
				return fmt.Errorf("no version for %s", lock.Name)
			}

			log.Info(ctx, "adding lock")

			lockFile.Inputs = append(lockFile.Inputs, lock)
		}
	}

	src.locks = lockFile
	return src.lockWriter.write(src.locks)
}

func (src *Source) getInputLockVersion(name string) *string {
	if src.locks == nil {
		return nil
	}
	for _, dep := range src.locks.Inputs {
		if dep.Name == name {
			return ptr(dep.Version)
		}
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}

func coalesce[T any](vals ...*T) *T {
	for _, val := range vals {
		if val != nil {
			return val
		}
	}
	return nil
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

func (src *Source) registryLatest(ctx context.Context, input *config_j5pb.Input_Registry) (*config_j5pb.InputLock, error) {

	version := "main"
	if input.Reference != nil {
		version = *input.Reference
	}
	bundle, err := src.regClient.input(ctx, input.Owner, input.Name, version)
	if err != nil {
		return nil, err
	}

	if src.j5Cache != nil && bundle.version != "" {
		if err := src.j5Cache.put(ctx, bundle.name, bundle.version, bundle.source); err != nil {
			log.WithError(ctx, err).Error("failed to cache input")
		}
	}

	return &config_j5pb.InputLock{
		Name:    bundle.name,
		Version: bundle.version,
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

func (src *Source) bufRegistryInput(ctx context.Context, owner, name, version string) (*imageBundle, error) {

	downloadRes, err := src.bufCache.client.Download(ctx, &registry_pb.DownloadRequest{
		Owner:      owner,
		Repository: name,
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

		return nil, fmt.Errorf("parsing buf input %q/%q: %w", owner, name, err)
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	img := &source_j5pb.SourceImage{
		File:            realDesc.File,
		SourceFilenames: filenames,
	}

	return &imageBundle{
		source:  img,
		name:    fmt.Sprintf("buf/%s/%s", owner, name),
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
