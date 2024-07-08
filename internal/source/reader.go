package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

var configPaths = []string{
	"j5.yaml",
	"ext/j5/j5.yaml",
}

func readDirConfigs(root fs.FS) (*config_j5pb.RepoConfigFile, error) {
	var config *config_j5pb.RepoConfigFile
	var err error
	for _, filename := range configPaths {
		config, err = readConfigFile(root, filename)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("reading file %s: %w", filename, err)
		}
		break
	}

	if config == nil {
		return nil, fmt.Errorf("no J5 config found")
	}

	return config, nil
}

func readConfigFile(root fs.FS, filename string) (*config_j5pb.RepoConfigFile, error) {
	data, err := fs.ReadFile(root, filename)
	if err != nil {
		return nil, err
	}
	config := &config_j5pb.RepoConfigFile{}
	if err := protoyaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func readBundleConfigFile(root fs.FS, filename string) (*config_j5pb.BundleConfigFile, error) {
	data, err := fs.ReadFile(root, filename)
	if err != nil {
		return nil, err
	}
	config := &config_j5pb.BundleConfigFile{}
	if err := protoyaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func readImageFromDir(ctx context.Context, src *bundle) (*source_j5pb.SourceImage, error) {
	_, err := src.J5Config()
	if err != nil {
		return nil, err
	}

	bufCache := protosrc.NewBufCache()
	extFiles, err := bufCache.GetDeps(ctx, src.repo.repoRoot, src.dirInRepo)
	if err != nil {
		return nil, fmt.Errorf("reading buf deps: %w", err)
	}

	var searchFS []fs.FS

	for _, dep := range src.refConfig.Deps {
		localBundle, ok := src.repo.bundles[dep]
		if !ok {
			return nil, fmt.Errorf("unknown dependency %q", dep)
		}

		fs, err := localBundle.fs()
		if err != nil {
			return nil, err
		}

		searchFS = append(searchFS, fs)

	}

	bundleRoot, err := src.fs()
	if err != nil {
		return nil, err
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
			log.WithFields(ctx, map[string]interface{}{
				"error": err.Error(),
			}).Warn("protoparse warning")
		},

		Accessor: func(filename string) (io.ReadCloser, error) {
			if content, ok := extFiles[filename]; ok {
				return io.NopCloser(bytes.NewReader(content)), nil
			}
			if reader, err := bundleRoot.Open(filename); err == nil {
				return reader, nil
			}
			for _, fs := range searchFS {
				if reader, err := fs.Open(filename); err == nil {
					return reader, nil
				}
			}
			return nil, fmt.Errorf("file not found: %s", filename)

		},
	}

	customDesc, err := parser.ParseFiles(filenames...)
	if err != nil {
		return nil, fmt.Errorf("protoparse: %w", err)
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	return &source_j5pb.SourceImage{
		File:            realDesc.File,
		Packages:        src.config.Packages,
		Options:         src.config.Options,
		Prose:           proseFiles,
		SourceFilenames: filenames,
	}, nil

}
