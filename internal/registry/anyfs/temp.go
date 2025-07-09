package anyfs

import (
	"context"
	"fmt"
	"os"
)

type TempFS struct {
	*LocalFS
}

func NewTempFS(ctx context.Context) (*TempFS, error) {
	dir, err := os.MkdirTemp("", "dst")
	if err != nil {
		return nil, fmt.Errorf("make tmp dir: %w", err)
	}

	localfs, err := NewLocalFS(dir)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		os.RemoveAll(dir)
	}()

	return &TempFS{
		LocalFS: localfs,
	}, nil
}
