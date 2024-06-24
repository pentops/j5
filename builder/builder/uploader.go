package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/jdef"
	"github.com/pentops/j5/schema/swagger"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"google.golang.org/protobuf/proto"
)

type FS interface {
	Put(ctx context.Context, path string, body io.Reader) error
}

type RawUploader struct {
	ProtoGenOutputs map[string]string
	J5Output        string
}

func NewRawUploader() *RawUploader {
	return &RawUploader{
		ProtoGenOutputs: map[string]string{},
	}
}

type J5Upload struct {
	Image   *source_j5pb.SourceImage
	J5API   *schema_j5pb.API
	Swagger *swagger.Document
}

func (uu *RawUploader) UploadJsonAPI(ctx context.Context, info FullInfo, data J5Upload) error {
	if uu.J5Output == "" {
		return nil
	}

	image, err := proto.Marshal(data.Image)
	if err != nil {
		return err
	}

	jDefJSON, err := jdef.FromProto(data.J5API)
	if err != nil {
		return err
	}

	jDefJSONBytes, err := json.Marshal(jDefJSON)
	if err != nil {
		return err
	}

	swaggerJSONBytes, err := json.Marshal(data.Swagger)
	if err != nil {
		return err
	}

	p := uu.J5Output

	if err := os.WriteFile(filepath.Join(p, "image.bin"), image, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(p, "jdef.json"), jDefJSONBytes, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(p, "swagger.json"), swaggerJSONBytes, 0644); err != nil {
		return err
	}

	return nil
}

type FSUploader struct {
	fs          FS
	GomodPrefix string
	JsonPrefix  string
}

func NewFSUploader(fs FS) *FSUploader {
	return &FSUploader{
		fs:          fs,
		GomodPrefix: "gomod",
		JsonPrefix:  "japi",
	}
}

type FullInfo struct {
	Version string
	Package string
	Commit  *source_j5pb.CommitInfo
}

func (uu *FSUploader) UploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, goModData []byte, packageRoot string) error {

	gomodBytes, err := os.ReadFile(filepath.Join(packageRoot, "go.mod"))
	if err != nil {
		return err
	}

	parsedGoMod, err := modfile.Parse("go.mod", gomodBytes, nil)
	if err != nil {
		return err
	}

	if parsedGoMod.Module == nil {
		return fmt.Errorf("no module found in go.mod")
	}

	packageName := parsedGoMod.Module.Mod.Path

	commitHashPrefix := commitInfo.Hash
	if len(commitHashPrefix) > 12 {
		commitHashPrefix = commitHashPrefix[:12]
	}

	canonicalVersion := module.PseudoVersion("", "", commitInfo.Time.AsTime(), commitHashPrefix)

	version := FullInfo{
		Version: canonicalVersion,
		Package: packageName,
		Commit:  commitInfo,
	}

	zipBuf := &bytes.Buffer{}

	err = zip.CreateFromDir(zipBuf, module.Version{
		Path:    version.Package,
		Version: version.Version,
	}, packageRoot)
	if err != nil {
		return err
	}

	// TODO: This is a stub
	return fmt.Errorf("not implemented")
}
