package source

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/log.go/log"
	"gopkg.in/yaml.v2"
)

type BufLockFile struct {
	Version string                   `yaml:"version"`
	Deps    []*BufLockFileDependency `yaml:"deps"`
}

type BufLockFileDependency struct {
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Commit     string `yaml:"commit"`
	Remote     string `yaml:"remote"`
	Digest     string `yaml:"digest"`
	Name       string `yaml:"name"`
}

func ConvertBufDeps(ctx context.Context, root fs.FS, subDir string) ([]*config_j5pb.Input, error) {

	var lockFileData []byte

	searchPath := subDir
	for {
		lockFile, err := fs.ReadFile(root, path.Join(searchPath, "buf.lock"))
		if err == nil {
			lockFileData = lockFile

			log.WithFields(ctx, map[string]interface{}{
				"lockFile": path.Join(searchPath, "buf.lock"),
			}).Debug("found lock file")
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if searchPath == "." {
			break
		}
		searchPath = filepath.Dir(searchPath)
	}

	if lockFileData == nil {
		return nil, fmt.Errorf("buf.lock not found")
	}

	bufLockFile := &BufLockFile{}
	if err := yaml.Unmarshal(lockFileData, bufLockFile); err != nil {
		return nil, err
	}

	switch bufLockFile.Version {
	case "", "v1":

	case "v2":
		for _, dep := range bufLockFile.Deps {
			parts := strings.Split(dep.Name, "/")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid remote %s", dep.Remote)
			}

			if parts[0] != "buf.build" {
				return nil, fmt.Errorf("unsupported remote %s", parts[0])
			}
			dep.Owner = parts[1]
			dep.Repository = parts[2]
		}

	default:
		return nil, fmt.Errorf("unsupported buf.lock version %s", bufLockFile.Version)

	}

	allDeps := make([]*config_j5pb.Input, 0, len(bufLockFile.Deps))
	for _, dep := range bufLockFile.Deps {
		allDeps = append(allDeps, &config_j5pb.Input{
			Type: &config_j5pb.Input_BufRegistry_{
				BufRegistry: &config_j5pb.Input_BufRegistry{
					Owner: dep.Owner,
					Name:  dep.Repository,
				},
			},
		})
	}

	return allDeps, nil

}
