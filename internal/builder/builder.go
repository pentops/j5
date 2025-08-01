package builder

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"maps"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/plugin/v1/plugin_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/builder/protogen/j5go"
	"github.com/pentops/j5/internal/j5client"
	"github.com/pentops/j5/internal/structure"
	"github.com/pentops/log.go/log"
	"github.com/pentops/runner/parallel"
	"golang.org/x/mod/modfile"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func NewBuilder(runner PipeRunner) *Builder {
	return &Builder{
		runner: runner,
	}
}

type PluginContext struct {
	Variables map[string]string
	ErrOut    io.Writer
	Dest      Dest
}

type PipeRunner interface {
	Run(ctx context.Context, rc RunContext) error
}
type Builder struct {
	runner PipeRunner
}

type Dest interface {
	PutFile(ctx context.Context, path string, body io.Reader) error
}

func (b *Builder) RunGenerateBuild(ctx context.Context, pc PluginContext, input *source_j5pb.SourceImage, build *config_j5pb.GenerateConfig) error {
	return b.runPlugins(ctx, pc, input, build.Plugins)
}

func (b *Builder) RunPublishBuild(ctx context.Context, pc PluginContext, input *source_j5pb.SourceImage, build *config_j5pb.PublishConfig) error {
	err := b.runPlugins(ctx, pc, input, build.Plugins)
	if err != nil {
		return err
	}

	if build.OutputFormat != nil {
		switch pkg := build.OutputFormat.Type.(type) {
		case *config_j5pb.OutputType_GoProxy_:

			gomodFile, err := buildGomodFile(pkg.GoProxy)
			if err != nil {
				return err
			}

			err = pc.Dest.PutFile(ctx, "go.mod", bytes.NewReader(gomodFile))
			if err != nil {
				return err
			}
			return nil

		}
		// Fallthrough default, is OK to not specify
	}
	return nil
}

func buildGomodFile(pkg *config_j5pb.OutputType_GoProxy) ([]byte, error) {
	mm := modfile.File{}
	if err := mm.AddModuleStmt(pkg.Path); err != nil {
		return nil, err
	}

	if pkg.GoVersion == "" {
		pkg.GoVersion = "1.22.3"
	}

	if err := mm.AddGoStmt(pkg.GoVersion); err != nil {
		return nil, err
	}

	for _, dep := range pkg.Deps {
		if err := mm.AddRequire(dep.Path, dep.Version); err != nil {
			return nil, err
		}
	}

	return mm.Format()
}

func (b *Builder) runPlugins(ctx context.Context, pc PluginContext, input *source_j5pb.SourceImage, plugins []*config_j5pb.BuildPlugin) error {

	if len(plugins) == 0 {
		return fmt.Errorf("no plugins")
	}

	runGroup := parallel.NewGroup(ctx)

	for _, plugin := range plugins {

		switch plugin.Type {
		case config_j5pb.Plugin_PLUGIN_PROTO:
			protoBuildRequest, err := structure.CodeGeneratorRequestFromImage(input)
			if err != nil {
				return fmt.Errorf("CodeGeneratorRequestFromImage for image %s: %w", input.SourceName, err)
			}

			plugin := plugin
			runGroup.Go(func(ctx context.Context) error {
				ctx = log.WithField(ctx, "plugin", plugin.Name)
				log.Info(ctx, "Running Plugin")
				if plugin.Type != config_j5pb.Plugin_PLUGIN_PROTO {
					return fmt.Errorf("plugin type mismatch: %s", plugin.Type)
				}
				if err := b.runProtocPlugin(ctx, pc, plugin, protoBuildRequest); err != nil {
					return fmt.Errorf("proto plugin %s: %w", plugin.Name, err)
				}
				return nil
			})

		case config_j5pb.Plugin_J5_CLIENT:

			sourceAPI, err := structure.APIFromImage(input)
			if err != nil {
				return err
			}

			clientAPI, err := j5client.APIFromSource(sourceAPI)
			if err != nil {
				return err
			}

			if len(clientAPI.Packages) == 0 {
				return fmt.Errorf("no packages found")
			}

			runGroup.Go(func(ctx context.Context) error {
				ctx = log.WithField(ctx, "plugin", plugin.Name)
				log.Info(ctx, "Running Plugin")
				if plugin.Type != config_j5pb.Plugin_PLUGIN_J5_CLIENT {
					return fmt.Errorf("plugin type mismatch: %s", plugin.Type)
				}
				if err := b.runJ5ClientPlugin(ctx, pc, plugin, clientAPI); err != nil {
					return fmt.Errorf("j5 client plugin %s: %w", plugin.Name, err)
				}
				return nil
			})

		default:
			return fmt.Errorf("unsupported plugin type: %s", plugin.Type)
		}
	}

	return runGroup.Wait()
}

func (b *Builder) runProtocPlugin(ctx context.Context, pc PluginContext, plugin *config_j5pb.BuildPlugin, sourceProto *pluginpb.CodeGeneratorRequest) error {

	start := time.Now()

	parameters := make([]string, 0, len(plugin.Opts))
	for k, v := range plugin.Opts {
		parameters = append(parameters, fmt.Sprintf("%s=%s", k, v))
	}
	parameter := strings.Join(parameters, ",")
	sourceProto.Parameter = &parameter

	switch pt := plugin.RunType.Type.(type) {
	case *config_j5pb.PluginRunType_Local:
		ctx = log.WithField(ctx, "local-cmd", pt.Local.Cmd)
	case *config_j5pb.PluginRunType_Docker:
		ctx = log.WithField(ctx, "docker-runner", pt.Docker.Image)
	case *config_j5pb.PluginRunType_Builtin:
		ctx = log.WithField(ctx, "builtin", pt.Builtin)
		return b.runBuiltinProtogen(ctx, pc, plugin, sourceProto)
	}

	log.Debug(ctx, "Running Protoc Plugin")

	reqBytes, err := proto.Marshal(sourceProto)
	if err != nil {
		return err
	}

	outBuffer := &bytes.Buffer{}
	inBuffer := bytes.NewReader(reqBytes)

	if err := b.runner.Run(ctx, RunContext{
		Vars:    pc.Variables,
		StdIn:   inBuffer,
		StdOut:  outBuffer,
		StdErr:  pc.ErrOut,
		Command: plugin,
	}); err != nil {
		return err
	}

	resp := &pluginpb.CodeGeneratorResponse{}
	if err := proto.Unmarshal(outBuffer.Bytes(), resp); err != nil {
		return fmt.Errorf("parsing CodeGeneratorResponse: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("plugin error: %s", *resp.Error)
	}

	if err := handleResponseFiles(ctx, resp, pc.Dest); err != nil {
		return fmt.Errorf("handling response files: %w", err)
	}

	log.WithFields(ctx, map[string]any{
		"files":           len(resp.File),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil
}

func handleResponseFiles(ctx context.Context, resp *pluginpb.CodeGeneratorResponse, dest Dest) error {
	for _, f := range resp.File {
		name := f.GetName()
		reader := bytes.NewReader([]byte(f.GetContent()))
		log.WithField(ctx, "file", name).Debug("Writing File")
		if err := dest.PutFile(ctx, name, reader); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) runJ5ClientPlugin(ctx context.Context, pc PluginContext, plugin *config_j5pb.BuildPlugin, descriptorAPI *client_j5pb.API) error {

	start := time.Now()

	buildRequest := &plugin_j5pb.CodeGenerationRequest{
		Packages: descriptorAPI.Packages,
		Options:  map[string]string{},
	}
	maps.Copy(buildRequest.Options, plugin.Opts)

	reqBytes, err := proto.Marshal(buildRequest)
	if err != nil {
		return err
	}

	outBuffer := &bytes.Buffer{}
	inBuffer := bytes.NewReader(reqBytes)

	if err := b.runner.Run(ctx, RunContext{
		Vars:    pc.Variables,
		StdIn:   inBuffer,
		StdOut:  outBuffer,
		StdErr:  pc.ErrOut,
		Command: plugin,
	}); err != nil {
		return err
	}

	resp := &plugin_j5pb.CodeGenerationResponse{}
	if err := proto.Unmarshal(outBuffer.Bytes(), resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("plugin error: %s", *resp.Error)
	}

	for _, f := range resp.Files {
		name := f.GetName()
		reader := bytes.NewReader([]byte(f.GetContent()))
		if err := pc.Dest.PutFile(ctx, name, reader); err != nil {
			return err
		}
	}

	log.WithFields(ctx, map[string]any{
		"files":           len(resp.Files),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil

}

func (b *Builder) runBuiltinProtogen(ctx context.Context, pc PluginContext, plugin *config_j5pb.BuildPlugin, req *pluginpb.CodeGeneratorRequest) error {
	name := plugin.RunType.GetBuiltin()
	var f func(*protogen.Plugin) error
	switch name {
	case "j5-go":
		f = j5go.ProtocPlugin()
	default:
		return fmt.Errorf("unknown builtin plugin: %s", name)
	}

	var flags flag.FlagSet
	opts := protogen.Options{
		ParamFunc: flags.Set,
	}

	gen, err := opts.New(req)
	if err != nil {
		return err
	}
	if err := f(gen); err != nil {
		// Errors from the plugin function are reported by setting the
		// error field in the CodeGeneratorResponse.
		//
		// In contrast, errors that indicate a problem in protoc
		// itself (unparsable input, I/O errors, etc.) are reported
		// to stderr.
		gen.Error(err)
	}
	resp := gen.Response()
	return handleResponseFiles(ctx, resp, pc.Dest)
}
