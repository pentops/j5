package j5convert

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ConvertJ5File(deps TypeResolver, source *sourcedef_j5pb.SourceFile) ([]*descriptorpb.FileDescriptorProto, error) {

	importMap, err := j5Imports(source)
	if err != nil {
		return nil, err
	}

	file := newFileContext(source.Path + ".proto")
	root := newRootContext(deps, importMap, file)

	walker := &conversionVisitor{
		root:          root,
		file:          file,
		parentContext: file,
	}

	fileNode := sourcewalk.NewRoot(source)

	if err := walker.visitFileNode(fileNode); err != nil {
		// Errors returned here from the sourcewalk code, not the visitors
		return nil, fmt.Errorf("schema error: %w", err)
	}

	if len(root.errors) > 0 {
		return nil, errors.Join(root.errors...)
	}

	descriptors := []*descriptorpb.FileDescriptorProto{}
	for _, extra := range root.files {
		descriptors = append(descriptors, extra.File())
	}

	return descriptors, nil

}
func PackageFromFilename(filename string) string {
	dirName, _ := path.Split(filename)
	dirName = strings.TrimSuffix(dirName, "/")
	pathPackage := strings.Join(strings.Split(dirName, "/"), ".")
	return pathPackage
}

var reVersion = regexp.MustCompile(`^v\d+$`)

func SplitPackageFromFilename(filename string) (string, string, error) {
	pkg := PackageFromFilename(filename)
	parts := strings.Split(pkg, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid package %q for file %q", pkg, filename)
	}

	// foo.v1 -> foo, v1
	// foo.v1.service -> foo.v1, service
	// foo.bar.v1.service -> foo.bar.v1, service

	if reVersion.MatchString(parts[len(parts)-1]) {
		return pkg, "", nil
	}
	if reVersion.MatchString(parts[len(parts)-2]) {
		upToVersion := parts[:len(parts)-1]
		return strings.Join(upToVersion, "."), parts[len(parts)-1], nil
	}
	return pkg, "", fmt.Errorf("no version in package %q", pkg)
}
