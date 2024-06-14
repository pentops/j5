package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile/reporter"
	"github.com/bufbuild/protoyaml-go"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/prototools/protosrc"
)

var ConfigPaths = []string{
	"j5.yaml",
	"jsonapi.yaml",
	"ext/j5/j5.yaml",
}

func readDirConfigs(root fs.FS) (*config_j5pb.Config, error) {
	var config *config_j5pb.Config
	var err error
	for _, filename := range ConfigPaths {
		config, err = readConfigFile(root, filename)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		break
	}

	if config == nil {
		return nil, fmt.Errorf("no J5 config found")
	}

	return config, nil
}

func readConfigFile(root fs.FS, filename string) (*config_j5pb.Config, error) {
	data, err := fs.ReadFile(root, filename)
	if err != nil {
		return nil, err
	}
	config := &config_j5pb.Config{}
	if err := protoyaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func resolveBundle(repoRoot fs.FS, dir string) (*config_j5pb.Config, fs.FS, error) {

	if dir == "" {
		config, err := readDirConfigs(repoRoot)
		if err != nil {
			return nil, nil, err
		}
		if len(config.Packages) > 0 {
			return config, repoRoot, nil
		}
		if len(config.Bundles) == 0 {
			return nil, nil, fmt.Errorf("no packages or bundles in root config. Either specify a bundle or add a bundle config to the root config")
		}
		if len(config.Bundles) > 1 {
			return nil, nil, fmt.Errorf("multiple bundles in root config. Specify a bundle")
		}
		dir = config.Bundles[0].Dir
	}

	subPath := path.Join(dir, "j5.yaml")
	newConfig, err := readConfigFile(repoRoot, subPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading bundle config %s: %w", subPath, err)
	}
	subRoot, err := fs.Sub(repoRoot, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("subbing bundle root %s: %w", dir, err)
	}

	return newConfig, subRoot, nil
}

func readImageFromDir(ctx context.Context, repoRoot fs.FS, dir string) (*source_j5pb.SourceImage, error) {

	config, bundleRoot, err := resolveBundle(repoRoot, dir)
	if err != nil {
		return nil, err
	}

	bufCache := protosrc.NewBufCache()
	extFiles, err := bufCache.GetDeps(ctx, repoRoot, dir)
	if err != nil {
		return nil, fmt.Errorf("reading buf deps: %w", err)
	}

	proseFiles := []*source_j5pb.ProseFile{}
	filenames := []string{}
	err = fs.WalkDir(bundleRoot, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".proto":
			filenames = append(filenames, path)
			return nil

		case ".md":
			data, err := fs.ReadFile(bundleRoot, path)
			if err != nil {
				return err
			}
			proseFiles = append(proseFiles, &source_j5pb.ProseFile{
				Path:    path,
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
		WarningReporter: func(err reporter.ErrorWithPos) {
			log.WithField(ctx, "error", err).Error("protoparse warning")
		},

		Accessor: func(filename string) (io.ReadCloser, error) {
			if content, ok := extFiles[filename]; ok {
				return io.NopCloser(bytes.NewReader(content)), nil
			}
			return bundleRoot.Open(filename)
		},
	}

	customDesc, err := parser.ParseFiles(filenames...)
	if err != nil {
		return nil, fmt.Errorf("protoparse: %w", err)
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	return &source_j5pb.SourceImage{
		File:            realDesc.File,
		Packages:        config.Packages,
		Codec:           config.Options,
		Prose:           proseFiles,
		Registry:        config.Registry,
		SourceFilenames: filenames,
	}, nil

}
