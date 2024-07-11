package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/pentops/j5/builder"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
	"github.com/pentops/j5/internal/source"
	"github.com/pentops/j5/internal/structure"
	"github.com/pentops/log.go/log"
)

func runVerify(ctx context.Context, cfg struct {
	Dir string `flag:"dir" default:"." description:"Directory with j5.yaml root"`
}) error {

	src, err := source.ReadLocalSource(ctx, os.DirFS(cfg.Dir))
	if err != nil {
		return err
	}

	dockerWrapper, err := builder.NewRunner(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}
	bb := builder.NewBuilder(dockerWrapper)

	for _, bundle := range src.AllBundles() {

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
		for _, pkg := range descriptorAPI.Packages {
			fmt.Printf("Package %s OK\n", pkg.Name)
		}

		bundleConfig, err := bundle.J5Config()
		if err != nil {
			return err
		}

		for _, publish := range bundleConfig.Publish {
			if err := bb.Publish(ctx, builder.PluginContext{
				Variables: map[string]string{},
				ErrOut:    os.Stderr,
				Dest:      NewDiscardFS(),
			}, bundle, publish); err != nil {
				return err
			}
		}
	}

	if err := runGeneratePlugins(ctx, bb, src, NewDiscardFS()); err != nil {
		return err
	}

	return nil
}

func runGenerate(ctx context.Context, cfg struct {
	Dir string `flag:"dir" default:"." description:"Directory with j5.yaml generate configured"`
}) error {

	src, err := source.ReadLocalSource(ctx, os.DirFS(cfg.Dir))
	if err != nil {
		return err
	}

	dockerWrapper, err := builder.NewRunner(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}
	bb := builder.NewBuilder(dockerWrapper)

	outRoot, err := NewLocalFS(cfg.Dir)
	if err != nil {
		return err
	}
	return runGeneratePlugins(ctx, bb, src, outRoot)
}

func runGeneratePlugins(ctx context.Context, bb *builder.Builder, src *source.Source, out Dest) error {
	j5Config := src.J5Config()
	for _, generator := range j5Config.Generate {
		errOut := &lineWriter{
			writeLine: func(line string) {
				log.WithField(ctx, "generator", generator.Name).Info(line)
			},
		}

		dest := out.Sub(generator.Output)

		pc := builder.PluginContext{
			Variables: map[string]string{},
			Dest:      dest,
			ErrOut:    errOut,
		}

		for _, inputDef := range generator.Inputs {
			input, err := src.GetInput(ctx, inputDef)
			if err != nil {
				return err
			}
			err = bb.Generate(ctx, pc, input, generator)
			if err != nil {
				return err
			}
		}
		errOut.flush()

	}

	return nil
}

func runPublish(ctx context.Context, cfg struct {
	SourceConfig
	Dest    string `flag:"dest" description:"Destination directory for published files"`
	Publish string `flag:"publish" optional:"true" description:"Name of the 'publish' to run (required when more than one exists)"`
}) error {

	input, err := cfg.GetInput(ctx)
	if err != nil {
		return err
	}

	inputConfig, err := input.J5Config()
	if err != nil {
		return err
	}

	var publish *config_j5pb.PublishConfig
	if cfg.Publish == "" {
		if len(inputConfig.Publish) != 1 {
			return fmt.Errorf("no publish specified and %d publishes found", len(inputConfig.Publish))
		}
		publish = inputConfig.Publish[0]
	} else {
		for _, p := range inputConfig.Publish {
			if p.Name == cfg.Publish {
				publish = p
				break
			}
		}
		if publish == nil {
			return fmt.Errorf("no publish found with name %q", cfg.Publish)
		}
	}

	dockerWrapper, err := builder.NewRunner(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}
	bb := builder.NewBuilder(dockerWrapper)

	outRoot, err := NewLocalFS(cfg.Dest)
	if err != nil {
		return err
	}

	pc := builder.PluginContext{
		Variables: map[string]string{},
		Dest:      outRoot,
		ErrOut:    os.Stderr,
	}
	return bb.Publish(ctx, pc, input, publish)
}
