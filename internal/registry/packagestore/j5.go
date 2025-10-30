package packagestore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"strings"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/gen/j5/registry/v1/registry_pb"
	"github.com/pentops/j5/internal/source"
	"github.com/pentops/j5/lib/j5codec"
	"github.com/pentops/log.go/log"
	"github.com/pentops/sqrlx.go/sqrlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *PackageStore) GetRemoteDependency(ctx context.Context, input *config_j5pb.Input, locks *config_j5pb.LockFile) (*source_j5pb.SourceImage, error) {
	regInput := input.GetRegistry()
	if regInput == nil {
		return nil, status.Errorf(codes.InvalidArgument, "input is not a registry input: %v", input)
	}
	var version string
	if regInput.Version != nil {
		version = *regInput.Version
	} else if regInput.Reference != nil {
		version = *regInput.Reference
	} else {
		version = "main"
	}

	return s.GetJ5Image(ctx, regInput.Owner, regInput.Name, version)
}

func (s *PackageStore) LatestLocks(ctx context.Context, deps []*config_j5pb.Input) (*config_j5pb.LockFile, error) {
	return nil, status.Errorf(codes.Unimplemented, "LatestLocks not implemented")
}

func (s *PackageStore) GetJ5Image(ctx context.Context, orgName, imageName, version string) (*source_j5pb.SourceImage, error) {
	ctx = log.WithFields(ctx, map[string]any{
		"org":     orgName,
		"image":   imageName,
		"version": version,
	})
	log.Debug(ctx, "Getting J5 Image")

	pkg := &registry_pb.J5Package{}
	err := s.selectDataRow(ctx,
		sq.Select("data").
			From("j5_version").
			Where(sq.Eq{
				"owner":   orgName,
				"repo":    imageName,
				"version": version,
			}), pkg)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "image %s/%s:%s not found", orgName, imageName, version)
	} else if err != nil {
		return nil, err
	}

	data, err := s.fs.GetBytes(ctx, pkg.StorageKey)
	if err != nil {
		return nil, err
	}

	img := &source_j5pb.SourceImage{}
	if err := protojson.Unmarshal(data, img); err != nil { // nolint:forbidigo
		return nil, err
	}

	// package version is the 'canonical' version (commit hash)
	img.Version = &pkg.Version

	resolved, err := source.ResolveIncludes(ctx, s, img, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to resolve includes: %v", err)
	}

	return resolved, nil
}

func (s *PackageStore) UploadJ5Image(ctx context.Context, commitInfo *source_j5pb.CommitInfo, img *source_j5pb.SourceImage, registry *config_j5pb.RegistryConfig) error {
	packageName := path.Join(registry.Owner, registry.Name)

	log.WithFields(ctx, map[string]any{
		"package": packageName,
		"owner":   commitInfo.Owner,
		"repo":    commitInfo.Repo,
		"version": commitInfo.Hash,
		"aliases": commitInfo.Aliases,
		"j5Org":   registry.Owner,
		"j5Name":  registry.Name,
	}).Info("uploading jsonapi")

	root := s.fs.Join("repo", commitInfo.Owner, commitInfo.Repo, "commit", commitInfo.Hash, "j5", packageName)

	imageBytes, err := protojson.Marshal(img) // nolint:forbidigo
	if err != nil {
		return fmt.Errorf("marshal image: %w", err)
	}

	storageKey := s.fs.Join(root, "image.json")

	if err := s.fs.Put(ctx, storageKey, bytes.NewReader(imageBytes), map[string]string{
		"Content-Type": "application/json",
	}); err != nil {
		return err
	}

	versionDests := make([]string, 0, len(commitInfo.Aliases)+1)
	versionDests = append(versionDests, commitInfo.Hash)

	for _, alias := range commitInfo.Aliases {
		versionDests = append(versionDests, strings.TrimPrefix(alias, "refs/heads/"))
	}

	pkg := &registry_pb.J5Package{
		Owner:      registry.Owner,
		Name:       registry.Name,
		Version:    commitInfo.Hash,
		StorageKey: storageKey,
		Aliases:    versionDests,
	}

	pkgJSON, err := j5codec.Global.ProtoToJSON(pkg.ProtoReflect())
	if err != nil {
		return err
	}

	log.WithFields(ctx, map[string]any{
		"versions": versionDests,
	}).Info("Storing J5 Version")

	if err := s.db.Transact(ctx, &sqrlx.TxOptions{
		Isolation: sql.LevelReadCommitted,
		Retryable: true,
	}, func(ctx context.Context, tx sqrlx.Transaction) error {
		for _, version := range versionDests {
			_, err := tx.Insert(ctx, sqrlx.Upsert("j5_version").
				Key("owner", registry.Owner).
				Key("repo", registry.Name).
				Key("version", version).
				Set("data", pkgJSON))
			if err != nil {
				return err
			}

		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
