package anyfs

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalFS struct {
	root string
}

func NewLocalFS(root string) (*LocalFS, error) {
	return &LocalFS{
		root: root,
	}, nil
}

func (LocalFS) Join(paths ...string) string {
	return filepath.Join(paths...)
}

func (local *LocalFS) Put(ctx context.Context, subPath string, body io.Reader, metadata map[string]string) error {
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

func (local *LocalFS) GetBytes(ctx context.Context, subPath string) ([]byte, error) {
	key := filepath.Join(local.root, subPath)
	return os.ReadFile(key)
}

func (local *LocalFS) GetReader(ctx context.Context, subPath string) (io.ReadCloser, *FileInfo, error) {
	key := filepath.Join(local.root, subPath)
	stat, err := os.Stat(key)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(key)
	if err != nil {
		return nil, nil, err
	}
	return file, &FileInfo{
		Size: stat.Size(),
	}, nil
}

func (local *LocalFS) Head(ctx context.Context, subPath string) (*FileInfo, error) {
	key := filepath.Join(local.root, subPath)
	stat, err := os.Stat(key)
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Size: stat.Size(),
	}, nil
}

func (local *LocalFS) List(ctx context.Context, subPath string) ([]ListInfo, error) {
	key := filepath.Join(local.root, subPath)
	dir, err := os.ReadDir(key)
	if err != nil {
		return nil, err
	}

	var list []ListInfo
	for _, entry := range dir {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		list = append(list, ListInfo{
			Name: entry.Name(),
			Size: info.Size(),
		})
	}
	return list, nil
}
