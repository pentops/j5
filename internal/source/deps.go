package source

import (
	"fmt"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/types/descriptorpb"
)

type DependencySet interface {
	GetDependencyFile(filename string) (*descriptorpb.FileDescriptorProto, error)
	ListDependencyFiles(prefix string) []string
	AllDependencyFiles() ([]*descriptorpb.FileDescriptorProto, []string)
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type imageFiles struct {
	primary      map[string]*descriptorpb.FileDescriptorProto
	dependencies map[string]*descriptorpb.FileDescriptorProto
}

func (ii *imageFiles) FindFileByPath(filename string) (protocompile.SearchResult, error) {
	depFile, err := ii.GetDependencyFile(filename)
	if err == nil {
		return protocompile.SearchResult{
			Proto: depFile,
		}, nil
	}
	return protocompile.SearchResult{}, fmt.Errorf("FindPackageByPath: file %s not found", filename)
}

func (ii *imageFiles) GetDependencyFile(filename string) (*descriptorpb.FileDescriptorProto, error) {
	if file, ok := ii.primary[filename]; ok {
		return file, nil
	}
	if file, ok := ii.dependencies[filename]; ok {
		return file, nil
	}
	return nil, fmt.Errorf("could not find file %q", filename)
}

func (ii *imageFiles) ListDependencyFiles(prefix string) []string {

	files := make([]string, 0, len(ii.primary))
	for _, file := range ii.primary {
		name := file.GetName()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		files = append(files, name)
	}
	return files
}

func (ii *imageFiles) AllDependencyFiles() ([]*descriptorpb.FileDescriptorProto, []string) {

	files := make([]*descriptorpb.FileDescriptorProto, 0, len(ii.primary)+len(ii.dependencies))
	filenames := make([]string, 0, len(ii.primary))

	for filename, file := range ii.dependencies {
		if _, ok := ii.primary[filename]; ok {
			continue
		}
		files = append(files, file)
	}
	for _, file := range ii.primary {
		files = append(files, file)
		filenames = append(filenames, file.GetName())
	}
	return files, filenames
}
