package source

import (
	"errors"
	"fmt"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"google.golang.org/protobuf/proto"
)

var ErrPluginCycle = errors.New("plugin cycle detected")

func resolveConfigReferences(config *config_j5pb.RepoConfigFile) error {

	rootPlugins := map[string]*config_j5pb.BuildPlugin{}

	for _, plugin := range config.Plugins {
		if plugin.Base == nil {
			rootPlugins[plugin.Name] = plugin
			continue
		}
		base, ok := rootPlugins[*plugin.Base]
		if !ok {
			didMatch := false
			for _, search := range config.Plugins {
				if search.Name == *plugin.Base {
					didMatch = true
					break
				}
			}
			if !didMatch {
				return fmt.Errorf("plugin %q extends base plugin %q which is not defined", plugin.Name, *plugin.Base)
			} else {
				return fmt.Errorf("plugin %q extends %q which is defined later in the source", plugin.Name, *plugin.Base)
			}
		}

		extended := extendPlugin(base, plugin)
		rootPlugins[plugin.Name] = extended
	}

	config.Plugins = nil

	for _, gen := range config.Generate {
		localBases := map[string]*config_j5pb.BuildPlugin{}
		genOpts := gen.Opts
		for idx, plugin := range gen.Plugins {
			if plugin.Opts == nil {
				plugin.Opts = map[string]string{}
			}
			for k, v := range genOpts {
				if _, ok := plugin.Opts[k]; !ok {
					plugin.Opts[k] = v
				}
			}

			if plugin.Base != nil {
				found, ok := rootPlugins[*plugin.Base]
				if !ok {
					found, ok = localBases[*plugin.Base]
					if !ok {
						return fmt.Errorf("plugin %q extends base plugin %q which is not defined", plugin.Name, *plugin.Base)
					}
				}
				plugin = extendPlugin(found, plugin)
			}

			gen.Plugins[idx] = plugin
			localBases[plugin.Name] = plugin
		}
	}

	return nil
}

func extendPlugin(base, ext *config_j5pb.BuildPlugin) *config_j5pb.BuildPlugin {
	out := proto.Clone(base).(*config_j5pb.BuildPlugin)
	if out.Opts == nil {
		out.Opts = map[string]string{}
	}
	if ext.Name != "" {
		out.Name = ext.Name
	}
	if ext.Runner != nil {
		out.Runner = ext.Runner
	}
	if ext.Type == config_j5pb.Plugin_UNSPECIFIED {
		ext.Type = out.Type
	}

	for k, v := range ext.Opts {
		out.Opts[k] = v
	}
	return out
}
