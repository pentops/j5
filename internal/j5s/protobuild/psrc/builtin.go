package psrc

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	_ "github.com/pentops/j5/gen/j5/auth/v1/auth_j5pb"
	_ "github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	_ "github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	_ "github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	_ "github.com/pentops/j5/gen/j5/state/v1/psm_j5pb"
	_ "github.com/pentops/j5/j5types/any_j5t"
	_ "github.com/pentops/j5/j5types/date_j5t"
	_ "github.com/pentops/j5/j5types/decimal_j5t"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/genproto/googleapis/api/httpbody"
)

var builtinPrefixes = []string{
	"buf/validate/",
	"google/api/",
	"google/protobuf/",
	"j5/auth/v1/",
	"j5/bcl/v1/",
	"j5/client/v1/",
	"j5/ext/v1/",
	"j5/list/v1/",
	"j5/messaging/v1/",
	"j5/schema/v1/",
	"j5/source/v1/",
	"j5/sourcedef/v1/",
	"j5/state/v1/",
	"j5/types/any/v1/",
	"j5/types/date/v1/",
	"j5/types/decimal/v1/",
}

type BuiltinResolver struct {
}

func NewBuiltinResolver() *BuiltinResolver {
	return &BuiltinResolver{}
}

func isBuiltin(filename string) bool {
	for _, prefix := range builtinPrefixes {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}
	return false
}

func (br *BuiltinResolver) ListPackageFiles(pkgName string) ([]string, error) {
	root := strings.ReplaceAll(pkgName, ".", "/") + "/"
	if !isBuiltin(root) {
		return nil, errPackageNotFound
	}
	files := []string{}
	protoregistry.GlobalFiles.RangeFilesByPackage(protoreflect.FullName(pkgName), func(refl protoreflect.FileDescriptor) bool {
		files = append(files, refl.Path())
		return true
	})

	return files, nil

}

func (br *BuiltinResolver) FindFileByPath(filename string) (*File, error) {
	if !isBuiltin(filename) {
		return nil, errFileNotFound
	}

	refl, err := protoregistry.GlobalFiles.FindFileByPath(filename)
	if err != nil {
		return nil, fmt.Errorf("find builtin file %s: %w", filename, err)
	}

	ec := &errset.ErrCollector{}
	summary, err := buildSummaryFromReflect(refl, ec)
	if err != nil {
		return nil, fmt.Errorf("summary for builtin %s: %w", filename, err)
	}
	return &File{
		Filename:   filename,
		Summary:    summary,
		Refl:       refl,
		SourceType: BuiltInProtoSource,
	}, nil
}

func BuiltinFile(filename string) (*descriptorpb.FileDescriptorProto, bool) {
	if !isBuiltin(filename) {
		return nil, false
	}
	refl, err := protoregistry.GlobalFiles.FindFileByPath(filename)
	if err != nil {
		return nil, false
	}
	return protodesc.ToFileDescriptorProto(refl), true
}
