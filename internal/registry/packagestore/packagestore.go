package packagestore

import (
	"context"
	"database/sql"
	"io"

	"github.com/pentops/j5/internal/registry/anyfs"
	"github.com/pentops/j5/lib/j5codec"
	"github.com/pentops/sqrlx.go/sqrlx"
	"google.golang.org/protobuf/proto"
)

const (
	MetadataAlias       = "x-gomod-alias"
	MetadataCommitHash  = "x-gomod-commit-hash"
	MetadataCommitTime  = "x-gomod-commit-time"
	MetadataContentType = "Content-Type"
)

type FS interface {
	Put(ctx context.Context, path string, body io.Reader, metadata map[string]string) error
	GetBytes(ctx context.Context, path string) ([]byte, error)
	GetReader(ctx context.Context, path string) (io.ReadCloser, *anyfs.FileInfo, error)

	Join(elem ...string) string
}

type PackageStore struct {
	db sqrlx.Transactor
	fs FS
}

func NewPackageStore(db sqrlx.Transactor, fs FS) (*PackageStore, error) {
	return &PackageStore{
		db: db,
		fs: fs,
	}, nil
}

func (s *PackageStore) selectDataRow(ctx context.Context, query sqrlx.Sqlizer, dest proto.Message) error {
	var bytesOut []byte
	if err := s.db.Transact(ctx, &sqrlx.TxOptions{
		ReadOnly:  true,
		Retryable: true,
		Isolation: sql.LevelReadCommitted,
	}, func(ctx context.Context, tx sqrlx.Transaction) error {

		if err := tx.QueryRow(ctx, query).Scan(&bytesOut); err != nil {
			return err
		}

		return nil

	}); err != nil {
		return err
	}

	return j5codec.Global.JSONToProto(bytesOut, dest.ProtoReflect())
}
