package cli

import (
	"context"
	"fmt"

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
}) error {

	if (!cfg.Latest) && cfg.Version == "" {
		return fmt.Errorf("version, latest or both are required")
	}

	source, err := cfg.GetInput(ctx)
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

	_ = bb

	return fmt.Errorf("NOT IMPLEMENTED")

}
