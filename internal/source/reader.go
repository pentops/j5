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

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/bufbuild/protoyaml-go"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/types/descriptorpb"
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

func readLockFile(root fs.FS, filename string) (*config_j5pb.LockFile, error) {
	data, err := fs.ReadFile(root, filename)
	if err != nil {
		return nil, err
	}
	lockFile := &config_j5pb.LockFile{}
	if err := protoyaml.Unmarshal(data, lockFile); err != nil {
		return nil, err
	}
	return lockFile, nil
}

func (src *Source) readImageFromDir(ctx context.Context, bundle *bundleSource) (*source_j5pb.SourceImage, error) {
	_, err := bundle.J5Config()
	if err != nil {
		return nil, err
	}

	bundleRoot, err := bundle.fs()
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

	realDesc, err := src.bundleProtoparse(ctx, bundle, filenames)
	if err != nil {
		return nil, err
	}

	return &source_j5pb.SourceImage{
		File:            realDesc.File,
		Packages:        bundle.config.Packages,
		Options:         bundle.config.Options,
		Prose:           proseFiles,
		SourceFilenames: filenames,
	}, nil

}

type FileSource interface {
	GetFile(filename string) (io.ReadCloser, error)
	Name() string
}

type fsSource struct {
	fs   fs.FS
	name string
}

func (f fsSource) Name() string {
	return f.name
}

func (f fsSource) GetFile(filename string) (io.ReadCloser, error) {
	return f.fs.Open(filename)
}

type mapSource struct {
	files map[string][]byte
	name  string
}

func (f mapSource) Name() string {
	return f.name
}

func (f mapSource) GetFile(filename string) (io.ReadCloser, error) {
	content, ok := f.files[filename]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func (src *Source) bundleProtoparse(ctx context.Context, rootBundle *bundleSource, files []string) (*descriptorpb.FileDescriptorSet, error) {
	walkRoot, err := rootBundle.fs()
	if err != nil {
		return nil, err
	}

	allSources := []FileSource{
		fsSource{
			fs:   walkRoot,
			name: rootBundle.debugName,
		},
	}
	fileMap := map[string]*descriptorpb.FileDescriptorProto{}

	j5Config, err := rootBundle.J5Config()
	if err != nil {
		return nil, err
	}

	for _, dep := range j5Config.Dependencies {
		depInput, err := src.GetInput(ctx, dep)
		if err != nil {
			return nil, err
		}

		img, err := depInput.SourceImage(ctx)
		if err != nil {
			return nil, err
		}

		for _, file := range img.File {
			fileMap[*file.Name] = file
		}

	}

	parser := protoparse.Parser{
		ImportPaths:                     []string{""},
		IncludeSourceCodeInfo:           true,
		InterpretOptionsInUnlinkedFiles: true,

		WarningReporter: func(err reporter.ErrorWithPos) {
			log.WithFields(ctx, map[string]interface{}{
				"error": err.Error(),
			}).Warn("protoparse warning")
		},

		LookupImportProto: func(filename string) (*descriptorpb.FileDescriptorProto, error) {
			file, ok := fileMap[filename]
			if !ok {
				return nil, fmt.Errorf("could not find file %q", filename)
			}
			return file, nil
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

	customDesc, err := parser.ParseFiles(files...)
	if err != nil {
		panicErr := protocompile.PanicError{}
		if errors.As(err, &panicErr) {
			fmt.Printf("PANIC: %s\n", panicErr.Stack)
		}

		return nil, err
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)
	return realDesc, nil
}
