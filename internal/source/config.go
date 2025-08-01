package source

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/lib/j5codec"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"
)

var ErrPluginCycle = errors.New("plugin cycle detected")

var configPaths = []string{
	"j5.repo.yaml",
	"j5.yaml",
	"ext/j5/j5.yaml",
}

var bundleConfigPaths = []string{
	"j5.bundle.yaml",
	"j5.yaml",
}

func readBytesFromAny(root fs.FS, filenames []string) ([]byte, string, error) {
	for _, filename := range filenames {
		data, err := fs.ReadFile(root, filename)
		if err == nil {
			return data, filename, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, "", fmt.Errorf("reading file %s: %w", filename, err)
		}
	}
	return nil, "", fmt.Errorf("searching %s: %w", strings.Join(filenames, ", "), fs.ErrNotExist)
}

func unmarshalFile(filename string, data []byte, out proto.Message) error {

	switch filepath.Ext(filename) {
	case ".yaml", ".yml":
		jsonData, err := yaml.YAMLToJSON(data)
		if err != nil {
			return fmt.Errorf("unmarshal %s: %w", filename, err)
		}
		return j5codec.Global.JSONToProto(jsonData, out.ProtoReflect())

	case ".json":
		return j5codec.Global.JSONToProto(data, out.ProtoReflect())

	default:
		return fmt.Errorf("unmarshal %s: unknown file extension %q", filename, filepath.Ext(filename))
	}
}

func readDirConfigs(root fs.FS) (*config_j5pb.RepoConfigFile, error) {
	data, filename, err := readBytesFromAny(root, configPaths)
	if err != nil {
		return nil, fmt.Errorf("repo config: %w", err)
	}
	config := &config_j5pb.RepoConfigFile{}
	if err := unmarshalFile(filename, data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func readBundleConfigFile(root fs.FS) (*config_j5pb.BundleConfigFile, error) {
	data, filename, err := readBytesFromAny(root, bundleConfigPaths)
	if err != nil {
		return nil, fmt.Errorf("bundle config: %w", err)
	}
	config := &config_j5pb.BundleConfigFile{}
	if err := unmarshalFile(filename, data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func readLockFile(root fs.FS, filename string) (*config_j5pb.LockFile, error) {
	data, err := fs.ReadFile(root, filename)
	if err != nil {
		return nil, err
	}
	lockFile := &config_j5pb.LockFile{}
	if err := unmarshalFile(filename, data, lockFile); err != nil {
		return nil, err
	}
	return lockFile, nil
}

func resolveRepoPluginReferences(base *pluginBase, config *config_j5pb.RepoConfigFile) error {
	for _, gen := range config.Generate {
		if err := resolvePlugins(base, gen.Plugins, gen.Opts); err != nil {
			return err
		}
	}
	for _, pub := range config.Publish {
		if err := resolvePlugins(base, pub.Plugins, pub.Opts); err != nil {
			return err
		}
	}

	return nil
}

func resolveBundlePluginReferences(base *pluginBase, config *config_j5pb.BundleConfigFile) error {

	bundleRoot, err := buildRootPlugins(config.Plugins, base.rootPlugins)
	if err != nil {
		return err
	}
	bundleBase := &pluginBase{
		overrides:   base.overrides,
		rootPlugins: maps.Clone(base.rootPlugins),
	}

	maps.Copy(bundleBase.rootPlugins, bundleRoot)

	for _, pub := range config.Publish {
		if err := resolvePlugins(bundleBase, pub.Plugins, pub.Opts); err != nil {
			return err
		}
	}
	return nil
}

func repoPluginBase(config *config_j5pb.RepoConfigFile) (*pluginBase, error) {
	overrides := map[string]*config_j5pb.PluginOverride{}
	for _, override := range config.PluginOverrides {
		overrides[override.Name] = override
	}

	rootPlugins, err := buildRootPlugins(config.Plugins, nil)
	if err != nil {
		return nil, err
	}
	config.Plugins = nil

	return &pluginBase{
		rootPlugins: rootPlugins,
		overrides:   overrides,
	}, nil
}

func upscalePlugin(plugin *config_j5pb.BuildPlugin) error {
	if plugin.RunType != nil {
		if plugin.Docker != nil {
			return fmt.Errorf("plugin %q has both runType (new format) and docker (legacy), please use only one", plugin.Name)
		}
		if plugin.Local != nil {
			return fmt.Errorf("plugin %q has both runType (new format) and local (legacy), please use only one", plugin.Name)
		}
		return nil // uses latest
	}

	plugin.RunType = &config_j5pb.PluginRunType{}
	if plugin.Docker != nil {
		plugin.RunType.Type = &config_j5pb.PluginRunType_Docker{
			Docker: plugin.Docker,
		}
		plugin.Docker = nil // remove legacy field
		return nil
	}

	if plugin.Local != nil {
		plugin.RunType.Type = &config_j5pb.PluginRunType_Local{
			Local: plugin.Local,
		}
		plugin.Local = nil // remove legacy field
		return nil
	}

	// some don't specify any, if they extend.
	return nil
}
func buildRootPlugins(specified []*config_j5pb.BuildPlugin, parentPlugins map[string]*config_j5pb.BuildPlugin) (map[string]*config_j5pb.BuildPlugin, error) {
	rootPlugins := map[string]*config_j5pb.BuildPlugin{}

	for _, plugin := range specified {
		if err := upscalePlugin(plugin); err != nil {
			return nil, fmt.Errorf("plugin %q: %w", plugin.Name, err)
		}
		if plugin.Base == nil {
			rootPlugins[plugin.Name] = plugin
			continue
		}
		base, ok := rootPlugins[*plugin.Base]
		if !ok {
			base, ok = parentPlugins[*plugin.Base]
			if !ok {
				// this logic is only building a better error message.
				didMatch := false
				for _, search := range specified {
					if search.Name == *plugin.Base {
						didMatch = true
						break
					}
				}
				if !didMatch {
					return nil, fmt.Errorf("plugin %q extends base plugin %q which is not defined", plugin.Name, *plugin.Base)
				} else {
					return nil, fmt.Errorf("plugin %q extends %q which is defined later in the source (plugins are resolved in lexical order)", plugin.Name, *plugin.Base)
				}
			}
		}

		extended := extendPlugin(base, plugin)
		rootPlugins[plugin.Name] = extended
	}
	return rootPlugins, nil
}

func extendPlugin(base, ext *config_j5pb.BuildPlugin) *config_j5pb.BuildPlugin {
	ext = proto.Clone(ext).(*config_j5pb.BuildPlugin)
	if ext.Name == "" {
		ext.Name = base.Name
	}

	if ext.RunType == nil {
		ext.RunType = base.RunType

	}

	if ext.Type == config_j5pb.Plugin_UNSPECIFIED {
		ext.Type = base.Type
	}

	// MERGE options.
	if ext.Opts == nil {
		ext.Opts = map[string]string{}
	}
	for k, v := range base.Opts {
		if _, ok := ext.Opts[k]; !ok {
			ext.Opts[k] = v
		}
	}
	return ext
}

type pluginBase struct {
	rootPlugins map[string]*config_j5pb.BuildPlugin
	overrides   map[string]*config_j5pb.PluginOverride
}

func resolvePlugins(base *pluginBase, plugins []*config_j5pb.BuildPlugin, baseOpts map[string]string) error {
	localBases := map[string]*config_j5pb.BuildPlugin{}
	for idx, plugin := range plugins {
		if plugin.Opts == nil {
			plugin.Opts = map[string]string{}
		}
		for k, v := range baseOpts {
			if _, ok := plugin.Opts[k]; !ok {
				plugin.Opts[k] = v
			}
		}

		if plugin.Base != nil {
			found, ok := base.rootPlugins[*plugin.Base]
			if !ok {
				found, ok = localBases[*plugin.Base]
				if !ok {
					return fmt.Errorf("plugin %q extends base plugin %q which is not defined", plugin.Name, *plugin.Base)
				}
			}
			plugin = extendPlugin(found, plugin)
		}

		if plugin.Type == config_j5pb.Plugin_UNSPECIFIED {
			if plugin.Base == nil {
				return fmt.Errorf("plugin %q has no type, did you mean to set 'base'?", plugin.Name)
			}
		}

		if plugin.Name != "" {
			localBases[plugin.Name] = plugin
		}

		// override AFTER using as a base.
		if override, ok := base.overrides[plugin.Name]; ok {
			plugin.RunType = override.RunType
		}
		plugins[idx] = plugin
	}
	return nil
}
