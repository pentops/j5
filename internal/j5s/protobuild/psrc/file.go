package psrc

import (
	"fmt"

	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/parser"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type SourceType int

const (
	LocalJ5Source SourceType = iota
	LocalProtoSource
	BuiltInProtoSource
	ExternalProtoSource
)

var sourceTypeNames = map[SourceType]string{
	LocalJ5Source:       "Local J5",
	LocalProtoSource:    "Local Proto",
	BuiltInProtoSource:  "Built-in Proto",
	ExternalProtoSource: "External Proto",
}

func (st SourceType) String() string {
	if name, ok := sourceTypeNames[st]; ok {
		return name
	}
	return fmt.Sprintf("Unknown SourceType %d", st)
}

type File struct {
	Filename string
	Summary  *j5convert.FileSummary

	SourceType SourceType

	Linked        linker.File
	LinkerVisited bool
	Dependencies  []*File

	// Oneof
	Refl        protoreflect.FileDescriptor
	Desc        *descriptorpb.FileDescriptorProto
	ParseResult *parser.Result
}

func (sr *File) ListDependencies() ([]string, error) {
	if sr.Refl != nil {
		imports := sr.Refl.Imports()
		deps := make([]string, 0, imports.Len())
		for i := range imports.Len() {
			deps = append(deps, imports.Get(i).Path())
		}
		return deps, nil
	}
	if sr.Desc != nil {
		return sr.Desc.Dependency, nil
	}
	if sr.ParseResult != nil {
		return (*sr.ParseResult).FileDescriptorProto().Dependency, nil
	}
	return nil, fmt.Errorf("no dependencies available in File")

}
