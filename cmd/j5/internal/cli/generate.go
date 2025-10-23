package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/builder"
	"github.com/pentops/j5/internal/j5s/protobuild/protomod"
	"github.com/pentops/j5/internal/source"
	"github.com/pentops/log.go/log"
)

func runGenerate(ctx context.Context, cfg struct {
	SourceConfig
	Name    string `flag:"name" required:"false" description:"Name of the generate to build"`
	NoClean bool   `flag:"no-clean" description:"Do not remove the directories in config as 'managedPaths' before generating"`
}) error {

	src, err := cfg.GetSource(ctx)
	if err != nil {
		if ep, ok := errpos.AsErrorsWithSource(err); ok {
			fmt.Fprintln(os.Stderr, ep.ShortString())
		}
		return err
	}

	repoConfig := src.RepoConfig()
	if repoConfig.GenerateJ5SProto {
		if err := runJ5sGenProto(ctx, j5sGenProtoConfig{
			SourceConfig: cfg.SourceConfig,
		}); err != nil {
			return err
		}
	}

	dockerWrapper, err := builder.NewRunner(builder.DefaultRegistryAuths)
	if err != nil {
		return err
	}
	bb := builder.NewBuilder(dockerWrapper)

	outRoot, err := NewLocalFS(cfg.Source)
	if err != nil {
		return err
	}

	if !cfg.NoClean {
		repoConfig := src.RepoConfig()
		if err := outRoot.Clean(repoConfig.ManagedPaths); err != nil {
			return err
		}
	}

	j5Config := src.RepoConfig()
	for _, generator := range j5Config.Generate {
		if cfg.Name != "" && generator.Name != cfg.Name {
			continue
		}
		if err := runGeneratePlugin(ctx, bb, src, generator, outRoot); err != nil {
			return err

		}
	}
	return nil
}

func runGeneratePlugin(ctx context.Context, bb *builder.Builder, src *source.RepoRoot, generator *config_j5pb.GenerateConfig, out Dest) error {
	img, err := src.CombinedSourceImage(ctx, generator.Inputs)
	if err != nil {
		if ep, ok := errpos.AsErrorsWithSource(err); ok {
			fmt.Fprintln(os.Stderr, ep.ShortString())
		}
		return err
	}

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

	if err := protomod.MutateImageWithMods(img, generator.Mods); err != nil {
		return fmt.Errorf("MutateImageWithMods: %w", err)
	}

	err = bb.RunGenerateBuild(ctx, pc, img, generator)
	errOut.flush()
	if err != nil {
		return err
	}

	return nil
}
