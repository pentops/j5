package source

import (
	"context"
	"fmt"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
)

func ResolveIncludes(ctx context.Context, rr RemoteResolver, img *source_j5pb.SourceImage, locks *config_j5pb.LockFile) (*source_j5pb.SourceImage, error) {

	ib := newImageBuilderFromImage(img)
	for _, include := range img.Includes {
		inputSpec := &config_j5pb.Input{
			Type: &config_j5pb.Input_Registry_{
				Registry: &config_j5pb.Input_Registry{
					Owner:     include.Owner,
					Name:      include.Name,
					Version:   include.Version,
					Reference: include.Reference,
				},
			},
		}

		includedImage, err := rr.GetRemoteDependency(ctx, inputSpec, locks)
		if err != nil {
			return nil, fmt.Errorf("resolving included dependency %s/%s: %w", include.Owner, include.Name, err)
		}
		ctx := log.WithFields(ctx, "including", includedImage.SourceName)
		log.Debug(ctx, "Resolver: resolved include")

		resolvedIncluded, err := ResolveIncludes(ctx, rr, includedImage, locks)
		if err != nil {
			return nil, fmt.Errorf("resolving includes for %s/%s: %w", include.Owner, include.Name, err)
		}

		if err := ib.include(ctx, resolvedIncluded); err != nil {
			return nil, fmt.Errorf("including dependency %s/%s: %w", include.Owner, include.Name, err)
		}
	}

	ib.img.Includes = nil // clear includes to avoid duplication in the final image

	return ib.img, nil
}
