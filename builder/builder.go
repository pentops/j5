package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/plugin/v1/plugin_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/source"
	"github.com/pentops/j5/internal/structure"
	"github.com/pentops/log.go/log"
	"golang.org/x/mod/modfile"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func NewFSInput(ctx context.Context, fs fs.FS, bundleName string) (Input, error) {
	src, err := source.ReadLocalSource(ctx, fs)
	if err != nil {
		return nil, err
	}
	return src.NamedInput(bundleName)
}

type PipeRunner interface {
	Run(ctx context.Context, rc RunContext) error
}

type Builder struct {
	runner PipeRunner
}

type Input interface {
	Name() string
	J5Config() (*config_j5pb.BundleConfigFile, error)
	ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error)
	SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error)
}

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

type Dest interface {
	PutFile(ctx context.Context, path string, body io.Reader) error
}

func (b *Builder) Generate(ctx context.Context, pc PluginContext, input Input, build *config_j5pb.GenerateConfig) error {
	for _, plugin := range build.Plugins {
		err := b.runPlugin(ctx, pc, input, plugin)
		if err != nil {
			return fmt.Errorf("input %s, plugin %s: %w", input.Name(), plugin.Name, err)
		}
	}
	return nil
}

func (b *Builder) Publish(ctx context.Context, pc PluginContext, input Input, build *config_j5pb.PublishConfig) error {

	for _, plugin := range build.Plugins {
		err := b.runPlugin(ctx, pc, input, plugin)
		if err != nil {
			return fmt.Errorf("plugin %s: %w", plugin.Name, err)
		}
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

	for _, dep := range pkg.Deps {
		if err := mm.AddRequire(dep.Path, dep.Version); err != nil {
			return nil, err
		}
	}

	return mm.Format()
}

func (b *Builder) runPlugin(ctx context.Context, pc PluginContext, input Input, plugin *config_j5pb.BuildPlugin) error {
	ctx = log.WithField(ctx, "plugin", plugin.Name)
	log.Info(ctx, "Running Plugin")

	switch plugin.Type {
	case config_j5pb.Plugin_PLUGIN_PROTO:
		protoBuildRequest, err := input.ProtoCodeGeneratorRequest(ctx)
		if err != nil {
			return fmt.Errorf("ProtoCodeGeneratorRequest: %w", err)
		}

		if err := b.runProtocPlugin(ctx, pc, plugin, protoBuildRequest); err != nil {
			return fmt.Errorf("plugin %s for input %s: %w", plugin.Name, input.Name(), err)
		}

	case config_j5pb.Plugin_J5_CLIENT:
		sourceImage, err := input.SourceImage(ctx)
		if err != nil {
			return fmt.Errorf("source image: %w", err)
		}
		reflectionAPI, err := structure.ReflectFromSource(sourceImage)
		if err != nil {
			return fmt.Errorf("ReflectFromSource: %w", err)
		}

		descriptorAPI, err := reflectionAPI.ToJ5Proto()
		if err != nil {
			return fmt.Errorf("DescriptorFromReflection: %w", err)
		}

		if len(descriptorAPI.Packages) == 0 {
			return fmt.Errorf("no packages found")
		}

		if err := b.runJ5ClientPlugin(ctx, pc, plugin, descriptorAPI); err != nil {
			return fmt.Errorf("plugin %s for input %s: %w", plugin.Name, input.Name(), err)
		}

	default:
		return fmt.Errorf("unsupported plugin type: %s", plugin.Type)
	}

	return nil
}

func (b *Builder) runProtocPlugin(ctx context.Context, pc PluginContext, plugin *config_j5pb.BuildPlugin, sourceProto *pluginpb.CodeGeneratorRequest) error {

	start := time.Now()

	ctx = log.WithField(ctx, "builder", plugin.GetName())
	log.Debug(ctx, "Running Protoc Plugin")

	parameters := make([]string, 0, len(plugin.Opts))
	for k, v := range plugin.Opts {
		parameters = append(parameters, fmt.Sprintf("%s=%s", k, v))
	}
	parameter := strings.Join(parameters, ",")
	sourceProto.Parameter = &parameter

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
		Command: plugin.Command,
	}); err != nil {
		return err
	}

	resp := pluginpb.CodeGeneratorResponse{}
	if err := proto.Unmarshal(outBuffer.Bytes(), &resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("plugin error: %s", *resp.Error)
	}

	for _, f := range resp.File {
		name := f.GetName()
		reader := bytes.NewReader([]byte(f.GetContent()))
		if err := pc.Dest.PutFile(ctx, name, reader); err != nil {
			return err
		}
	}

	log.WithFields(ctx, map[string]interface{}{
		"files":           len(resp.File),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil
}

func (b *Builder) runJ5ClientPlugin(ctx context.Context, pc PluginContext, plugin *config_j5pb.BuildPlugin, descriptorAPI *schema_j5pb.API) error {

	start := time.Now()

	buildRequest := &plugin_j5pb.CodeGenerationRequest{
		Packages: descriptorAPI.Packages,
		Options:  map[string]string{},
	}
	for key, opt := range plugin.Opts {
		buildRequest.Options[key] = opt
	}

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
		Command: plugin.Command,
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

	log.WithFields(ctx, map[string]interface{}{
		"files":           len(resp.Files),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil

}
