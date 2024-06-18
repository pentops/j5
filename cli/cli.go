package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"runtime/debug"

	"github.com/pentops/j5/builder/builder"
	"github.com/pentops/j5/builder/docker"
	"github.com/pentops/j5/builder/git"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/source"
	"github.com/pentops/log.go/log"
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "dev"
}()

func CommandSet() *commander.CommandSet {

	cmdGroup := commander.NewCommandSet()

	cmdGroup.Add("registry", registrySet())
	cmdGroup.Add("schema", schemaSet())
	cmdGroup.Add("codegen", generateSet())
	cmdGroup.Add("proto", protoSet())

	cmdGroup.Add("version", commander.NewCommand(runVersion))
	cmdGroup.Add("generate", commander.NewCommand(runGenerate))

	return cmdGroup
}

func runVersion(ctx context.Context, cfg struct{}) error {
	fmt.Printf("jsonapi version %v\n", Commit)
	return nil
}

func runGenerate(ctx context.Context, cfg struct {
	SourceConfig
}) error {

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	src, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	j5Config := src.J5Config()
	commitInfo, err := src.CommitInfo(ctx)
	if err != nil {
		return err
	}

	for _, generator := range j5Config.Generate {

		dest, err := NewLocalFS(generator.Out)
		if err != nil {
			return err
		}

		bb := builder.NewBuilder(dockerWrapper)
		resolvedPlugins := make([]*config_j5pb.BuildPlugin, 0, len(generator.Plugins))
		for _, plugin := range generator.Plugins {
			plugin, err = src.ResolvePlugin(plugin)
			if err != nil {
				return err
			}
			for k, v := range generator.Opts {
				plugin.Opts[k] = v
			}
			resolvedPlugins = append(resolvedPlugins, plugin)
		}

		for _, genSrc := range generator.Src {
			protoBuildRequest, err := src.ProtoCodeGeneratorRequest(ctx, genSrc)
			if err != nil {
				return err
			}

			for _, plugin := range resolvedPlugins {
				if plugin.Docker == nil {
					return fmt.Errorf("plugin %q has no docker spec", plugin.Name)
				}
				envVars, err := builder.MapEnvVars(plugin.Docker.Env, commitInfo)
				if err != nil {
					return err
				}
				if err := bb.RunProtocPlugin(ctx, dest, plugin, protoBuildRequest, os.Stderr, envVars); err != nil {
					return err
				}
			}
		}

	}

	return nil
}

type SourceConfig struct {
	Source string `flag:"src" default:"." description:"Source directory containing j5.yaml and buf.lock.yaml"`
	Bundle string `flag:"bundle" default:"" description:"When the bundle j5.yaml is in a subdirectory"`

	CommitHash    string   `flag:"commit-hash" env:"COMMIT_HASH" default:""`
	CommitTime    string   `flag:"commit-time" env:"COMMIT_TIME" default:""`
	CommitAliases []string `flag:"commit-alias" env:"COMMIT_ALIAS" default:""`

	GitAuto bool `flag:"git-auto" env:"COMMIT_INFO_GIT_AUTO" default:"false" description:"Automatically pull commit info from git"`
}

func (cfg SourceConfig) GetSource(ctx context.Context) (builder.Source, error) {

	sourceDir := cfg.Source

	var commitInfo *source_j5pb.CommitInfo
	var err error

	if cfg.CommitHash != "" && cfg.CommitTime != "" {
		commitInfo = &source_j5pb.CommitInfo{}
		commitInfo.Hash = cfg.CommitHash
		commitTime, err := time.Parse(time.RFC3339, cfg.CommitTime)
		if err != nil {
			return nil, fmt.Errorf("parsing commit time: %w", err)
		}
		commitInfo.Time = timestamppb.New(commitTime)
		commitInfo.Aliases = cfg.CommitAliases
	} else {
		commitInfo, err = git.ExtractGitMetadata(ctx, cfg.Source)
		if err != nil {
			return nil, fmt.Errorf("extracting git metadata from dir: %w", err)
		}
	}

	return source.ReadLocalSource(ctx, commitInfo, os.DirFS(sourceDir), cfg.Bundle)
}

type DiscardFS struct{}

func NewDiscardFS() *DiscardFS {
	return &DiscardFS{}
}

func (d *DiscardFS) Put(ctx context.Context, subPath string, body io.Reader) error {
	return nil
}

type LocalFS struct {
	root string
}

func NewLocalFS(root string) (*LocalFS, error) {
	return &LocalFS{
		root: root,
	}, nil
}

func (local *LocalFS) Put(ctx context.Context, subPath string, body io.Reader) error {
	key := filepath.Join(local.root, subPath)
	log.WithField(ctx, "filename", key).Debug("writing file")
	err := os.MkdirAll(filepath.Dir(key), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(key)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err
	}

	return nil
}
