package buildwrap

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/registry/github"
)

type tmpSource struct {
	//*source.Source
	commit *source_j5pb.CommitInfo
	dir    string
}

func (ts *tmpSource) close() error {
	return os.RemoveAll(ts.dir)
}

func (bw *BuildWorker) BundleImageFromCommit(ctx context.Context, commit *source_j5pb.CommitInfo, bundleName string) (*source_j5pb.SourceImage, *config_j5pb.BundleConfigFile, error) {

	repoRoot, err := bw.tmpClone(ctx, commit)
	if err != nil {
		return nil, nil, err
	}

	defer repoRoot.close()

	img, cfg, err := bw.builder.SourceImage(ctx, repoRoot.fs(), bundleName)
	if err != nil {
		return nil, nil, fmt.Errorf("new fs input: %w", err)
	}

	return img, cfg, nil
}

func (bw *BuildWorker) tmpClone(ctx context.Context, commit *source_j5pb.CommitInfo) (*tmpSource, error) {
	workDir, err := os.MkdirTemp("", "src")
	if err != nil {
		return nil, fmt.Errorf("make workdir: %w", err)
	}

	// Clone
	ref := &github.Commit{
		Owner: commit.Owner,
		Repo:  commit.Repo,
		Sha:   commit.Hash,
	}
	if err := bw.github.GetContent(ctx, ref, workDir); err != nil {
		os.RemoveAll(workDir)
		return nil, fmt.Errorf("clone: %w", err)
	}

	return &tmpSource{
		dir:    workDir,
		commit: commit,
		//Source: buildSource,
	}, nil
}

func (src tmpSource) fs() fs.FS {
	return os.DirFS(src.dir)
}

type tmpDest struct {
	root string
	fs.FS
}

func newTmpDest() (*tmpDest, error) {
	dir, err := os.MkdirTemp("", "dst")
	if err != nil {
		return nil, fmt.Errorf("make tmp dir: %w", err)
	}

	return &tmpDest{
		root: dir,
		FS:   os.DirFS(dir),
	}, nil
}

func (local *tmpDest) PutFile(ctx context.Context, subPath string, body io.Reader) error {
	key := filepath.Join(local.root, subPath)
	err := os.MkdirAll(filepath.Dir(key), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(key)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err
	}

	return nil
}

func (td *tmpDest) Close() error {
	return os.RemoveAll(td.root)
}
