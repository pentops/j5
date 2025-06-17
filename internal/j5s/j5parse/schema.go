package j5parse

import (
	"path"
	"strings"

	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func FileStub(sourceFilename string) protoreflect.Message {
	dirName, _ := path.Split(sourceFilename)
	dirName = strings.TrimSuffix(dirName, "/")

	pathPackage := strings.Join(strings.Split(dirName, "/"), ".")
	file := &sourcedef_j5pb.SourceFile{
		Path: sourceFilename,
		Package: &sourcedef_j5pb.Package{
			Name: pathPackage,
		},
		SourceLocations: &bcl_j5pb.SourceLocation{},
	}
	refl := file.ProtoReflect()

	return refl
}
