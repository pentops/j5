package anyfs

import (
	"context"
	"io"
	"strings"
)

type FS interface {
	Put(ctx context.Context, path string, body io.Reader, metadata map[string]string) error
	GetReader(ctx context.Context, path string) (io.ReadCloser, *FileInfo, error)
	GetBytes(ctx context.Context, path string) ([]byte, error)
	Head(ctx context.Context, path string) (*FileInfo, error)
	List(ctx context.Context, path string) ([]ListInfo, error)

	Join(path ...string) string
}

type FileInfo struct {
	Metadata map[string]string
	Size     int64
}

// NewEnvFS creates a new FS that will configure the client as needed from env
// vars.
func NewEnvFS(ctx context.Context, name string) (FS, error) {
	if strings.HasPrefix(name, "s3://") {
		return NewS3EnvFS(ctx, name)
	}

	return NewLocalFS(name)
}
