package psrc

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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

type builtinResolver struct {
}

func newBuiltinResolver() *builtinResolver {
	return &builtinResolver{}
}

func (br *builtinResolver) hasRoot(filename string) bool {
	for _, prefix := range builtinPrefixes {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}
	return false
}

func (br *builtinResolver) ListPackageFiles(pkgName string) ([]string, error) {
	root := strings.ReplaceAll(pkgName, ".", "/") + "/"
	isBuiltin := br.hasRoot(root)
	if !isBuiltin {
		return nil, errPackageNotFound
	}
	files := []string{}
	protoregistry.GlobalFiles.RangeFilesByPackage(protoreflect.FullName(pkgName), func(refl protoreflect.FileDescriptor) bool {
		files = append(files, refl.Path())
		return true
	})

	return files, nil

}

func (br *builtinResolver) FindFileByPath(filename string) (*File, error) {
	if !br.hasRoot(filename) {
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
