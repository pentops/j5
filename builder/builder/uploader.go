package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/schema/jdef"
	"github.com/pentops/log.go/log"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"google.golang.org/protobuf/proto"
)

const (
	// TODO: from gomodproxy, for S3 uploads so should be removed
	S3MetadataAlias      = "x-gomod-alias"
	S3MetadataCommitHash = "x-gomod-commit-hash"
	S3MetadataCommitTime = "x-gomod-commit-time"
)

type FS interface {
	Put(ctx context.Context, path string, body io.Reader, metadata map[string]string) error
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

func (uu *RawUploader) BuildGoModule(ctx context.Context, commitInfo *builder_j5pb.CommitInfo, label string, callback BuilderCallback) error {
	output, ok := uu.ProtoGenOutputs[label]
	if !ok {
		tmpDir, err := os.MkdirTemp("", "j5-workdir")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)
		output = tmpDir
	}

	return callback(ctx, output, commitInfo)
}

func (uu *RawUploader) UploadJsonAPI(ctx context.Context, info FullInfo, data J5Upload) error {
	if uu.J5Output == "" {
		return nil
	}

	image, err := proto.Marshal(data.Image)
	if err != nil {
		return err
	}

	jDefJSON, err := jdef.FromProto(data.JDef)
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
	Commit  *builder_j5pb.CommitInfo
}

type BuilderCallback func(ctx context.Context, workingDir string, commitInfo *builder_j5pb.CommitInfo) error

func (uu *FSUploader) BuildGoModule(ctx context.Context, commitInfo *builder_j5pb.CommitInfo, label string, callback BuilderCallback) error {

	dest, err := os.MkdirTemp("", "j5-workdir")
	if err != nil {
		return err
	}
	packageRoot := filepath.Join(dest, "package")

	defer os.RemoveAll(dest)

	if err := callback(ctx, packageRoot, commitInfo); err != nil {
		return err
	}

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

	info := FullInfo{
		Version: canonicalVersion,
		Package: packageName,
		Commit:  commitInfo,
	}

	return uu.UploadGoModule(ctx, info, gomodBytes, packageRoot)

}

func (uu *FSUploader) UploadGoModule(ctx context.Context, version FullInfo, goModData []byte, packageRoot string) error {

	zipBuf := &bytes.Buffer{}

	err := zip.CreateFromDir(zipBuf, module.Version{
		Path:    version.Package,
		Version: version.Version,
	}, packageRoot)
	if err != nil {
		return err
	}

	log.WithFields(ctx, map[string]interface{}{
		"package": version.Package,
		"version": version.Version,
	}).Info("uploading go module")

	metadata := map[string]string{
		S3MetadataCommitTime: version.Commit.Time.AsTime().Format(time.RFC3339),
		S3MetadataCommitHash: version.Commit.Hash,
	}

	if err := uu.fs.Put(ctx,
		path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.mod", version.Version)),
		strings.NewReader(string(goModData)),
		metadata,
	); err != nil {
		return err
	}

	if err := uu.fs.Put(ctx,
		path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.zip", version.Version)),
		zipBuf,
		metadata,
	); err != nil {
		return err
	}

	aliasMetadata := map[string]string{}
	for k, v := range metadata {
		aliasMetadata[k] = v
	}
	aliasMetadata[S3MetadataAlias] = version.Version
	for _, alias := range version.Commit.Aliases {
		if err := uu.fs.Put(ctx,
			path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.zip", alias)),
			bytes.NewReader([]byte(version.Version)),
			aliasMetadata,
		); err != nil {
			return err
		}
	}

	if err := uu.fs.Put(ctx,
		path.Join(uu.GomodPrefix, version.Package, fmt.Sprintf("%s.zip", version.Commit.Hash)),
		bytes.NewReader([]byte(version.Version)),
		aliasMetadata,
	); err != nil {
		return err
	}

	return nil
}

func (uu *FSUploader) UploadJsonAPI(ctx context.Context, info FullInfo, data J5Upload) error {

	log.WithFields(ctx, map[string]interface{}{
		"package": info.Package,
		"version": info.Version,
		"aliases": info.Commit.Aliases,
	}).Info("uploading jsonapi")

	image, err := proto.Marshal(data.Image)
	if err != nil {
		return err
	}

	jDefJSON, err := json.Marshal(data.JDef)
	if err != nil {
		return err
	}

	swaggerJSON, err := json.Marshal(data.Swagger)
	if err != nil {
		return err
	}

	versionDests := make([]string, 0, len(info.Commit.Aliases)+1)
	versionDests = append(versionDests, info.Commit.Hash)
	versionDests = append(versionDests, info.Commit.Aliases...)
	for _, version := range versionDests {
		p := path.Join(uu.JsonPrefix, info.Package, version)
		log.WithField(ctx, "path", p).Info("uploading image")

		if err := uu.fs.Put(ctx, path.Join(p, "image.bin"), bytes.NewReader(image), map[string]string{
			"Content-Type": "application/octet-stream",
		}); err != nil {
			return err
		}
		if err := uu.fs.Put(ctx, path.Join(p, "jdef.json"), bytes.NewReader(jDefJSON), map[string]string{
			"Content-Type": "application/json",
		}); err != nil {
			return err
		}
		if err := uu.fs.Put(ctx, path.Join(p, "swagger.json"), bytes.NewReader(swaggerJSON), map[string]string{
			"Content-Type": "application/json",
		}); err != nil {
			return err
		}
	}

	return nil
}
