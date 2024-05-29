package source

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pentops/jsonapi/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Source struct {
	// input
	commitInfo *source_j5pb.CommitInfo
	config     *config_j5pb.Config
	sourceDir  string

	// cache
	codegenReqs map[string]*pluginpb.CodeGeneratorRequest
	sourceImg   *source_j5pb.SourceImage
}

func NewLocalDirSource(ctx context.Context, commitInfo *source_j5pb.CommitInfo, config *config_j5pb.Config, sourceDir string) (*Source, error) {
	return &Source{
		config:      config,
		commitInfo:  commitInfo,
		sourceDir:   sourceDir,
		codegenReqs: map[string]*pluginpb.CodeGeneratorRequest{},
	}, nil
}

func ReadLocalSource(ctx context.Context, commitInfo *source_j5pb.CommitInfo, dir string) (*Source, error) {
	config, err := ReadDirConfigs(dir)
	if err != nil {
		return nil, err
	}

	return NewLocalDirSource(ctx, commitInfo, config, dir)
}

func (src Source) J5Config() *config_j5pb.Config {
	return src.config
}

func (src Source) CommitInfo(context.Context) (*source_j5pb.CommitInfo, error) {
	return src.commitInfo, nil
}

func (src *Source) ProtoCodeGeneratorRequest(ctx context.Context, root string) (*pluginpb.CodeGeneratorRequest, error) {
	if src.codegenReqs[root] == nil {
		rr, err := CodeGeneratorRequestFromSource(ctx, filepath.Join(src.sourceDir, root))
		if err != nil {
			return nil, err
		}
		src.codegenReqs[root] = rr
	}
	return src.codegenReqs[root], nil
}

func (src *Source) SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	if src.sourceImg == nil {
		img, err := ReadImageFromSourceDir(ctx, src.sourceDir)
		if err != nil {
			return nil, err
		}
		src.sourceImg = img
	}

	return src.sourceImg, nil
}

func (src *Source) SourceDescriptors(ctx context.Context) ([]*descriptorpb.FileDescriptorProto, error) {
	img, err := src.SourceImage(ctx)
	if err != nil {
		return nil, err
	}
	includeMap := map[string]struct{}{}
	for _, include := range img.SourceFilenames {
		includeMap[include] = struct{}{}
	}

	out := []*descriptorpb.FileDescriptorProto{}
	for _, file := range img.File {
		if _, ok := includeMap[*file.Name]; !ok {
			continue
		}
		out = append(out, file)
	}

	return out, nil
}

func (src *Source) SourceFile(ctx context.Context, filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(src.sourceDir, filename))
}

func (src *Source) PackageBuildConfig(name string) (*config_j5pb.ProtoBuildConfig, error) {
	for _, plugin := range src.config.ProtoBuilds {
		if plugin.Name == name {
			return plugin, nil
		}
	}
	return nil, fmt.Errorf("package build %q not found", name)
}

func (src *Source) ResolvePlugin(plugin *config_j5pb.BuildPlugin) (*config_j5pb.BuildPlugin, error) {
	return src.resolvePlugin(map[string]struct{}{}, plugin)
}

var ErrPluginCycle = errors.New("plugin cycle detected")

func (src *Source) resolvePlugin(visited map[string]struct{}, plugin *config_j5pb.BuildPlugin) (*config_j5pb.BuildPlugin, error) {
	if plugin.Base == nil {
		if plugin.Opts == nil {
			plugin.Opts = map[string]string{}
		}
		return plugin, nil
	}
	if _, ok := visited[*plugin.Base]; ok {
		return nil, ErrPluginCycle
	}
	visited[*plugin.Base] = struct{}{}
	for _, search := range src.config.Plugins {
		if search.Name == *plugin.Base {
			resolvedBase, err := src.resolvePlugin(visited, search)
			if err != nil {
				return nil, err
			}
			if plugin.Type != config_j5pb.Plugin_PLUGIN_UNSPECIFIED {
				if plugin.Type != resolvedBase.Type {
					return nil, fmt.Errorf("base plugin %q has type %v, but extension has type %v", *plugin.Base, resolvedBase.Type, plugin.Type)
				}
			}
			return extendPlugin(resolvedBase, plugin), nil
		}
	}
	return nil, fmt.Errorf("base plugin %q not found", *plugin.Base)
}

func extendPlugin(base, ext *config_j5pb.BuildPlugin) *config_j5pb.BuildPlugin {
	out := proto.Clone(base).(*config_j5pb.BuildPlugin)
	if out.Opts == nil {
		out.Opts = map[string]string{}
	}
	if ext.Name != "" {
		out.Name = ext.Name
	}
	if ext.Docker != nil {
		out.Docker = ext.Docker
	}
	if ext.Command != nil {
		out.Command = ext.Command
	}
	for k, v := range ext.Opts {
		out.Opts[k] = v
	}
	return out
}
