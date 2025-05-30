package psrc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"google.golang.org/protobuf/types/descriptorpb"
)

type DescriptorFiles map[string]*descriptorpb.FileDescriptorProto

func (df DescriptorFiles) FindFileByPath(filename string) (*File, error) {
	file, ok := df[filename]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	ec := errset.NewCollector()

	summary, err := SummaryFromDescriptor(file, ec)
	if err != nil {
		return nil, fmt.Errorf("summary for dependency %s: %w", file, err)
	}
	return &File{
		Summary:    summary,
		Desc:       file,
		SourceType: ExternalProtoSource,
	}, nil
}

func (df DescriptorFiles) ListPackageFiles(pkgName string) ([]string, error) {
	prefix := strings.ReplaceAll(pkgName, ".", "/")
	files := make([]string, 0, len(df))
	for filename := range df {
		if strings.HasPrefix(filename, prefix) {
			files = append(files, filename)
		}
	}
	sort.Strings(files)
	return files, nil
}
