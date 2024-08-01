package source

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Source struct {
	thisRepo       *repo
	remoteRegistry string
	HTTPClient     *http.Client
}

func ReadLocalSource(ctx context.Context, root fs.FS) (*Source, error) {
	thisRepo, err := newRepo(".", root)
	if err != nil {
		return nil, err
	}

	return &Source{
		thisRepo:       thisRepo,
		remoteRegistry: os.Getenv("J5_REGISTRY"),
		HTTPClient:     &http.Client{},
	}, nil
}

func (src Source) J5Config() *config_j5pb.RepoConfigFile {
	return src.thisRepo.config
}

/*
func (src Source) CommitInfo(context.Context) (*source_j5pb.CommitInfo, error) {
	return src.thisRepo.commitInfo, nil
}*/

func (src Source) AllBundles() []Input {
	out := make([]Input, 0, len(src.thisRepo.bundles))
	for _, bundle := range src.thisRepo.bundles {
		out = append(out, bundle)
	}
	return out
}

func (src *Source) CombinedInput(ctx context.Context, inputs []*config_j5pb.Input) (Input, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no inputs")
	}
	if len(inputs) == 1 {
		return src.GetInput(ctx, inputs[0])
	}

	allFiles := map[string]string{}
	fullImage := &source_j5pb.SourceImage{
		Options: &config_j5pb.PackageOptions{},
	}
	for _, input := range inputs {
		bundle, err := src.GetInput(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("input %v: %w", input, err)
		}
		img, err := bundle.SourceImage(ctx)
		if err != nil {
			return nil, fmt.Errorf("input %v: %w", input, err)
		}

		wantFiles := map[string]struct{}{}
		for _, file := range img.SourceFilenames {
			wantFiles[file] = struct{}{}
		}

		fmt.Printf("input %q: %d files\n", bundle.Name(), len(img.File))
		for _, file := range img.File {
			hash, err := hashFile(file)
			if err != nil {
				return nil, fmt.Errorf("file %q: %w", *file.Name, err)
			}
			if existing, ok := allFiles[*file.Name]; ok {
				if existing != hash {
					return nil, fmt.Errorf("file %q has conflicting content", *file.Name)
				}
			} else {
				allFiles[*file.Name] = hash
				fullImage.File = append(fullImage.File, file)
				if _, ok := wantFiles[*file.Name]; ok {
					fullImage.SourceFilenames = append(fullImage.SourceFilenames, *file.Name)
				}
			}
		}

		if img.Options != nil {
			for _, subPkg := range img.Options.SubPackages {
				found := false
				for _, existing := range fullImage.Options.SubPackages {
					if existing.Name == subPkg.Name {
						// no config other than name for now.
						found = true
						break
					}
				}
				if !found {
					fullImage.Options.SubPackages = append(fullImage.Options.SubPackages, subPkg)
				}
			}
		}

		fullImage.Packages = append(fullImage.Packages, img.Packages...)
	}

	return &combinedBundle{
		source: fullImage,
	}, nil
}

func hashFile(file *descriptorpb.FileDescriptorProto) (string, error) {
	sh := sha256.New()
	fileContent, err := proto.Marshal(file)
	if err != nil {
		return "", err
	}
	sh.Write(fileContent)
	return base64.StdEncoding.EncodeToString(sh.Sum(nil)), nil

}

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
		repo, err := newRepo(debugName, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("input %s: %w", st.Repo.Root, err)
		}
		bundle, ok := repo.bundles[st.Repo.Bundle]
		if !ok {
			return nil, fmt.Errorf("bundle %q not found in repo %q", st.Repo.Bundle, st.Repo.Root)
		}
		return bundle, nil

	case *config_j5pb.Input_Registry_:
		return src.registryInput(ctx, st.Registry)

	default:
		return nil, fmt.Errorf("unsupported source type %T", input.Type)
	}
}

func (src *Source) registryInput(ctx context.Context, input *config_j5pb.Input_Registry) (Input, error) {
	if src.remoteRegistry == "" {
		return nil, fmt.Errorf("remote registry not set ($J5_REGISTRY)")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("registry input name not set")
	}
	if input.Organization == "" {
		return nil, fmt.Errorf("registry input organization not set")
	}
	version := "main"
	if input.Version != nil {
		version = *input.Version
	}

	imageURL := fmt.Sprintf("%s/registry/v1/%s/%s/%s/image.bin", src.remoteRegistry, input.Organization, input.Name, version)
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
		source: apiDef,
		repo: &config_j5pb.RegistryConfig{
			Organization: input.Organization,
			Name:         input.Name,
		},
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

func (src *Source) NamedInput(name string) (Input, error) {
	if name != "" {
		if bundle, ok := src.thisRepo.bundles[name]; ok {
			return bundle, nil
		}
		return nil, fmt.Errorf("bundle %q not found", name)
	}
	if len(src.thisRepo.bundles) == 0 {
		return nil, fmt.Errorf("no bundles found")
	}
	if len(src.thisRepo.bundles) > 1 {
		return nil, fmt.Errorf("multiple bundles found, must specify a name")
	}

	for _, bundle := range src.thisRepo.bundles {
		return bundle, nil
	}

	return nil, fmt.Errorf("no bundles found")

}

func (src *Source) SourceFile(ctx context.Context, filename string) ([]byte, error) {
	return fs.ReadFile(src.thisRepo.repoRoot, filename)
}
