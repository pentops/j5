package cli

import (
	"context"
	"fmt"
	"path"

	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/proto"
)

func registrySet() *commander.CommandSet {
	registryGroup := commander.NewCommandSet()
	registryGroup.Add("push", commander.NewCommand(runPush))
	return registryGroup
}

func runPush(ctx context.Context, cfg struct {
	SourceConfig
	Version string `flag:"version" default:"" description:"Version to push"`
	Latest  bool   `flag:"latest" description:"Push as latest"`
	Bucket  string `flag:"bucket" description:"S3 bucket to push to"`
	Prefix  string `flag:"prefix" description:"S3 prefix to push to"`
}) error {

	if (!cfg.Latest) && cfg.Version == "" {
		return fmt.Errorf("version, latest or both are required")
	}

	source, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	image, err := source.SourceImage(ctx)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	versions := []string{}

	if cfg.Latest {
		versions = append(versions, "latest")
	}

	if cfg.Version != "" {
		versions = append(versions, cfg.Version)
	}

	destinations := make([]string, len(versions))
	for i, version := range versions {
		p := path.Join(cfg.Prefix, image.Registry.Organization, image.Registry.Name, version, "image.bin")
		destinations[i] = fmt.Sprintf("s3://%s/%s", cfg.Bucket, p)
	}

	return pushS3(ctx, bb, destinations...)

}
