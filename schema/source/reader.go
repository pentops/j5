package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protoyaml-go"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
)

var ConfigPaths = []string{
	"j5.yaml",
	"jsonapi.yaml",
	"j5.yml",
	"jsonapi.yml",
	"ext/j5/j5.yaml",
	"ext/j5/j5.yml",
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

	extFiles, err := getBufDeps(ctx, src)
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
