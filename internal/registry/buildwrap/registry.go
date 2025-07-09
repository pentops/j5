package buildwrap

import (
	"context"
	"fmt"

	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
)

type RegStore interface {
	GetJ5Image(ctx context.Context, orgName, imageName, version string) (*source_j5pb.SourceImage, error)
}

type registryClient struct {
	store RegStore
}

func NewRegistryClient(store RegStore) *registryClient {
	return &registryClient{
		store: store,
	}
}

func (rc *registryClient) LatestImage(ctx context.Context, owner, repoName string, reference *string) (*source_j5pb.SourceImage, error) {
	if rc == nil {
		return nil, fmt.Errorf("registry client not set")
	}

	branch := "main"
	if reference != nil {
		branch = *reference
	}

	// registry returns the canonical version in the image
	return rc.GetImage(ctx, owner, repoName, branch)
}

func (rc *registryClient) GetImage(ctx context.Context, owner, repoName, version string) (*source_j5pb.SourceImage, error) {

	fullName := fmt.Sprintf("registry/v1/%s/%s", owner, repoName)
	ctx = log.WithField(ctx, "bundle", fullName)

	return rc.store.GetJ5Image(ctx, owner, repoName, version)
}
