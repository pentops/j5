package psrc

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
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

func (df DescriptorFiles) AddFiles(files []*descriptorpb.FileDescriptorProto) error {
	for _, file := range files {
		filename := file.GetName()
		if existing, ok := df[filename]; ok {
			if !AssertProtoFilesAreEqual(existing, file) {
				return fmt.Errorf("file %q has conflicting content", filename)
			}
		}
		df[filename] = file
	}
	return nil
}
func AssertProtoFilesAreEqual(aSrc, bSrc *descriptorpb.FileDescriptorProto) bool {
	if proto.Equal(aSrc, bSrc) {
		return true
	}

	a := proto.Clone(aSrc).(*descriptorpb.FileDescriptorProto)
	b := proto.Clone(bSrc).(*descriptorpb.FileDescriptorProto)
	// ignore source code info for comparison
	a.SourceCodeInfo = nil
	b.SourceCodeInfo = nil
	proto.ClearExtension(a.Options, ext_j5pb.E_J5Source)
	proto.ClearExtension(b.Options, ext_j5pb.E_J5Source)

	if proto.Equal(a, b) {
		return true
	}

	diff := cmp.Diff(a, b, protocmp.Transform())
	fmt.Fprintln(os.Stderr, diff)

	return false
}
