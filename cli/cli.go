package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"runtime/debug"

	"github.com/bufbuild/protoyaml-go"
	"github.com/pentops/jsonapi/builder/builder"
	"github.com/pentops/jsonapi/builder/git"
	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/jsonapi/schema/source"
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
	cmdGroup.Add("generate", generateSet())
	cmdGroup.Add("proto", protoSet())

	cmdGroup.Add("version", commander.NewCommand(runVersion))

	return cmdGroup
}

func runVersion(ctx context.Context, cfg struct{}) error {
	fmt.Printf("jsonapi version %v\n", Commit)
	return nil
}

type SourceConfig struct {
	Source        string   `flag:"src" default:"." description:"Source directory containing j5.yaml and buf.lock.yaml"`
	CommitHash    string   `flag:"commit-hash" env:"COMMIT_HASH" default:""`
	CommitTime    string   `flag:"commit-time" env:"COMMIT_TIME" default:""`
	CommitAliases []string `flag:"commit-alias" env:"COMMIT_ALIAS" default:""`

	GitAuto bool `flag:"git-auto" env:"COMMIT_INFO_GIT_AUTO" default:"false" description:"Automatically pull commit info from git"`
}

func (cfg SourceConfig) GetSource(ctx context.Context) (builder.Source, error) {

	sourceDir := cfg.Source
	japiConfig, err := loadConfig(sourceDir)
	if err != nil {
		return nil, err
	}

	var commitInfo *builder_j5pb.CommitInfo
	if cfg.GitAuto {
		commitInfo, err = git.ExtractGitMetadata(ctx, japiConfig.Git, cfg.Source)
		if err != nil {
			return nil, err
		}
	} else if cfg.CommitHash == "" || cfg.CommitTime == "" {
		return nil, fmt.Errorf("commit hash and time are required, or set --git-auto")
	} else {
		commitInfo.Hash = cfg.CommitHash
		commitTime, err := time.Parse(time.RFC3339, cfg.CommitTime)
		if err != nil {
			return nil, fmt.Errorf("parsing commit time: %w", err)
		}
		commitInfo.Time = timestamppb.New(commitTime)
		commitInfo.Aliases = cfg.CommitAliases
	}

	return source.NewLocalDirSource(ctx, commitInfo, japiConfig, sourceDir)
}

func loadConfig(src string) (*source_j5pb.Config, error) {
	var configData []byte
	var err error
	for _, filename := range source.ConfigPaths {
		configData, err = os.ReadFile(filepath.Join(src, filename))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		break
	}

	if configData == nil {
		return nil, fmt.Errorf("no config found")
	}

	config := &source_j5pb.Config{}
	if err := protoyaml.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	return config, nil
}

type LocalFS struct {
	root string
}

func NewLocalFS(root string) (*LocalFS, error) {
	return &LocalFS{
		root: root,
	}, nil
}

func (local *LocalFS) Put(ctx context.Context, subPath string, body io.Reader, metadata map[string]string) error {
	key := filepath.Join(local.root, subPath)
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
