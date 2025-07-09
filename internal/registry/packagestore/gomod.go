package packagestore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"github.com/pentops/j5/internal/gen/j5/registry/v1/registry_pb"
	"github.com/pentops/j5/internal/registry/gomodproxy"
	"github.com/pentops/sqrlx.go/sqrlx"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
	"google.golang.org/protobuf/encoding/protojson"
)

func (src *PackageStore) getGomodVersion(ctx context.Context, packageName, version string) (*registry_pb.GoModule, error) {
	pkg := &registry_pb.GoModule{}
	err := src.selectDataRow(ctx,
		sq.Select("data").
			From("go_module_version").
			Where(sq.Eq{
				"package_name": packageName,
				"version":      version,
			}), pkg)

	if errors.Is(err, sql.ErrNoRows) {
		log.WithFields(ctx, map[string]any{
			"package": packageName,
			"version": version,
		}).Info("GoMod Version not found")
		return nil, gomodproxy.VersionNotFoundError(version)
	} else if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (src *PackageStore) GoModInfo(ctx context.Context, packageName, version string) (*gomodproxy.Info, error) {
	module, err := src.getGomodVersion(ctx, packageName, version)
	if err != nil {
		return nil, err
	}

	return &gomodproxy.Info{
		Version: module.Version,
		Time:    module.CreatedAt.AsTime(),
	}, nil
}

func (src *PackageStore) GoModLatest(ctx context.Context, packageName string) (*gomodproxy.Info, error) {
	return src.GoModInfo(ctx, packageName, "main")
}

func (src *PackageStore) GoModList(ctx context.Context, packageName string) ([]string, error) {
	versions := make([]string, 0)

	// As of Go 1.13, this is useless until we do tags.

	// https://go.dev/ref/mod: Returns a list of known versions of the given module in plain text, one per line. This list should not include pseudo-versions.

	return versions, nil
}

func (src *PackageStore) GoModMod(ctx context.Context, packageName, version string) ([]byte, error) {
	module, err := src.getGomodVersion(ctx, packageName, version)
	if err != nil {
		return nil, err
	}

	body, err := src.fs.GetBytes(ctx, module.ModStorageKey)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (src *PackageStore) GoModZip(ctx context.Context, packageName, version string) (io.ReadCloser, error) {
	module, err := src.getGomodVersion(ctx, packageName, version)
	if err != nil {
		return nil, err
	}

	reader, _, err := src.fs.GetReader(ctx, module.ZipStorageKey)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (s *PackageStore) UploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, root fs.FS) error {

	gomodBytes, err := fs.ReadFile(root, "go.mod")
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

	log.WithFields(ctx, map[string]any{
		"package": packageName,
		"version": canonicalVersion,
	}).Info("uploading go module")

	files, err := listFilesInDir(root)
	if err != nil {
		return err
	}

	zipBuf := &bytes.Buffer{}
	err = zip.Create(zipBuf, module.Version{
		Path:    packageName,
		Version: canonicalVersion,
	}, files)
	if err != nil {
		return err
	}

	uploadRoot := s.fs.Join("repo", commitInfo.Owner, commitInfo.Repo, "commit", commitInfo.Hash, "gomod", packageName)

	modfileRoot := s.fs.Join(uploadRoot, fmt.Sprintf("%s.mod", canonicalVersion))
	if err := s.fs.Put(ctx, modfileRoot, strings.NewReader(string(gomodBytes)), map[string]string{
		MetadataContentType: "text/plain",
	}); err != nil {
		return err
	}

	zipRoot := s.fs.Join(uploadRoot, fmt.Sprintf("%s.zip", canonicalVersion))
	if err := s.fs.Put(ctx, zipRoot, zipBuf, map[string]string{
		MetadataContentType: "application/zip",
	},
	); err != nil {
		return err
	}

	aliases := []string{canonicalVersion, commitInfo.Hash}
	for _, alias := range commitInfo.Aliases {
		aliases = append(aliases, strings.TrimPrefix(alias, "refs/heads/"))
	}

	pkg := &registry_pb.GoModule{
		PackageName:   packageName,
		Version:       canonicalVersion,
		Aliases:       aliases,
		CreatedAt:     commitInfo.Time,
		ZipStorageKey: zipRoot,
		ModStorageKey: modfileRoot,
	}

	pkgJSON, err := protojson.Marshal(pkg)
	if err != nil {
		return err
	}

	log.WithFields(ctx, map[string]any{
		"versions": aliases,
	}).Info("Storing GoMod Version")

	if err := s.db.Transact(ctx, &sqrlx.TxOptions{
		Isolation: sql.LevelReadCommitted,
		Retryable: true,
	}, func(ctx context.Context, tx sqrlx.Transaction) error {
		for _, version := range aliases {
			_, err := tx.Insert(ctx, sqrlx.Upsert("go_module_version").
				Key("package_name", packageName).
				Key("version", version).
				Set("timestamp", commitInfo.Time.AsTime()).
				Set("data", pkgJSON))
			if err != nil {
				return err
			}

		}
		return err
	}); err != nil {
		return err
	}

	return nil

}

// Adapts a fs.DirEntry to a zip.File.
type wrappedFile struct {
	fs   fs.FS
	path string
	d    fs.DirEntry
}

func (f wrappedFile) Path() string {
	return f.path
}

func (f wrappedFile) Lstat() (os.FileInfo, error) {
	return f.d.Info()
}
func (f wrappedFile) Open() (io.ReadCloser, error) {
	return f.fs.Open(f.path)
}

// Very cut down version of zip.listFilesInDir but with fs.FS instead of os.File
func listFilesInDir(root fs.FS) ([]zip.File, error) {
	var files []zip.File
	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == "git" || name == "hg" || name == "svn" || name == "vendor" {
				return fs.SkipDir
			}

			return nil
		}

		// Skip irregular files and files in vendor directories.
		// Irregular files are ignored. They're typically symbolic links.
		if !d.Type().IsRegular() {
			return nil
		}

		files = append(files, wrappedFile{
			path: path,
			d:    d,
			fs:   root,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

/*

TODO: This is for CLI uploads
func (src *PackageStore) modFromZip(ctx context.Context, zipFileSize int64, zipFileReader io.Reader) ([]byte, error) {
	scratchDir, err := os.MkdirTemp("", "gomodproxy")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(scratchDir)

	zipPath := filepath.Join(scratchDir, "zip.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(zipFile, zipFileReader); err != nil {
		zipFile.Close()
		return nil, err
	}
	if err := zipFile.Close(); err != nil {
		return nil, err
	}

	zipReader, err := archivezip.NewReader(zipFile, zipFileSize)
	if err != nil {
		return nil, err
	}

	goMod, err := zipReader.Open("go.mod")
	if err != nil {
		return nil, err
	}

	defer goMod.Close()

	modBytes, err := io.ReadAll(goMod)
	if err != nil {
		return nil, err
	}

	return modBytes, nil
}*/
