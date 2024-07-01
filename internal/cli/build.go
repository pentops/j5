package cli

import (
	"context"
	"fmt"

	"github.com/pentops/j5/builder/builder"
	"github.com/pentops/j5/builder/docker"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/schema/j5reflect"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/log.go/log"
)

func runBuild(ctx context.Context, cfg struct {
	SourceConfig
	Name   string `flag:"name" default:"" description:"Name of the publisher, required when more than one specified"`
	Output string `flag:"output" description:"Output directory"`
}) error {

	bundle, err := cfg.GetInput(ctx)
	if err != nil {
		return err
	}

	j5Config, err := bundle.J5Config()
	if err != nil {
		return err
	}

	var publishCfg *config_j5pb.PublishConfig

	if len(j5Config.Publish) == 1 {
		publishCfg = j5Config.Publish[0]
	} else if len(j5Config.Publish) < 1 {
		return fmt.Errorf("no publishers found in j5.yaml")
	} else if cfg.Name == "" {
		return fmt.Errorf("more than one publisher found in j5.yaml, please specify one with --name")
	} else {
		for _, pub := range j5Config.Publish {
			if pub.Name == cfg.Name {
				publishCfg = pub
				break
			}
		}
		if publishCfg == nil {
			return fmt.Errorf("no publisher found with name %s", cfg.Name)
		}
	}

	image, err := bundle.SourceImage(ctx)
	if err != nil {
		return err
	}

	reflectionAPI, err := structure.ReflectFromSource(image)
	if err != nil {
		return fmt.Errorf("ReflectFromSource: %w", err)
	}

	descriptorAPI, err := reflectionAPI.ToJ5Proto()
	if err != nil {
		return fmt.Errorf("DescriptorFromReflection: %w", err)
	}

	if err := structure.ResolveProse(image, descriptorAPI); err != nil {
		return fmt.Errorf("ResolveProse: %w", err)
	}

	_, err = j5reflect.APIFromDesc(descriptorAPI)
	if err != nil {
		return fmt.Errorf("building reflection from descriptor: %w", err)
	}

	out := &lineWriter{
		writeLine: func(line string) {
			log.WithField(ctx, "publisher", publishCfg.Name).Info(line)
		},
	}

	dest, err := NewLocalFS(cfg.Output)
	if err != nil {
		return err
	}

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}
	bb := builder.NewBuilder(dockerWrapper)
	err = bb.RunPublishBuild(ctx, bundle, dest, publishCfg, out)
	out.flush()
	if err != nil {
		return err
	}

	return nil
}
