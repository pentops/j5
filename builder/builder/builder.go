package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/plugin/v1/plugin_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/export"
	"github.com/pentops/j5/schema/source"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type Source interface {
	GetInput(context.Context, *config_j5pb.Input) (source.Input, error)
	NamedInput(string) (source.Input, error)
	SourceFile(ctx context.Context, filename string) ([]byte, error)
}

type IDockerWrapper interface {
	Run(ctx context.Context, spec *config_j5pb.DockerSpec, input io.Reader, output, errOutput io.Writer, envVars []string) error
}

type Builder struct {
	Docker IDockerWrapper
}

func NewBuilder(docker IDockerWrapper) *Builder {
	return &Builder{
		Docker: docker,
	}
}

func (b *Builder) BuildJsonAPI(ctx context.Context, img *source_j5pb.SourceImage) (*J5Upload, error) {
	log.Info(ctx, "build json API")

	apiReflection, err := structure.ReflectFromSource(img)
	if err != nil {
		return nil, fmt.Errorf("reflection from image: %w", err)
	}

	apiDescriptor, err := apiReflection.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("descriptor from reflection: %w", err)
	}

	swaggerDoc, err := export.BuildSwagger(apiDescriptor)
	if err != nil {
		return nil, fmt.Errorf("build swagger: %w", err)
	}

	return &J5Upload{
		Image:   img,
		J5API:   apiDescriptor,
		Swagger: swaggerDoc,
	}, nil
}

func (b *Builder) RunGenerate(ctx context.Context, src Source, dst FS, build *config_j5pb.GenerateConfig, errOut io.Writer) error {
	inputs := make([]source.Input, 0, len(build.Inputs))

	for _, inputDef := range build.Inputs {
		input, err := src.GetInput(ctx, inputDef)
		if err != nil {
			return fmt.Errorf("get input: %w", err)
		}
		inputs = append(inputs, input)
	}

	for _, input := range inputs {
		if err := b.runPlugins(ctx, input, dst, build.Plugins, errOut); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) RunPublishBuild(ctx context.Context, input source.Input, dst FS, build *config_j5pb.PublishConfig, errOut io.Writer) error {
	if err := b.runPlugins(ctx, input, dst, build.Plugins, errOut); err != nil {
		return err
	}
	return nil
}

func (b *Builder) RunPublish(ctx context.Context, src Source, bundle string, dst FS, build *config_j5pb.PublishConfig, errOut io.Writer) error {

	input, err := src.NamedInput(bundle)
	if err != nil {
		return fmt.Errorf("named input: %w", err)
	}

	if err := b.RunPublishBuild(ctx, input, dst, build, errOut); err != nil {
		return err
	}

	if build.OutputFormat != nil {
		switch pkg := build.OutputFormat.Type.(type) {
		case *config_j5pb.OutputType_GoProxy_:

			gomodFile, err := src.SourceFile(ctx, pkg.GoProxy.GoModFile)
			if err != nil {
				return err
			}

			err = dst.Put(ctx, "go.mod", bytes.NewReader(gomodFile))
			if err != nil {
				return err
			}
			return nil

		}
		// Fallthrough default, is OK to not specify
	}
	return nil
}

func (b *Builder) runPlugins(ctx context.Context, input source.Input, dst FS, plugins []*config_j5pb.BuildPlugin, errOut io.Writer) error {
	variables := map[string]string{}
	for _, plugin := range plugins {

		switch plugin.Type {
		case config_j5pb.Plugin_PLUGIN_PROTO:
			protoBuildRequest, err := input.ProtoCodeGeneratorRequest(ctx)
			if err != nil {
				return fmt.Errorf("ProtoCodeGeneratorRequest: %w", err)
			}

			if err := b.RunProtocPlugin(ctx, variables, dst, plugin, protoBuildRequest, errOut); err != nil {
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
				return fmt.Errorf("no packages found for input %s", input.Name())
			}
			for _, pkg := range descriptorAPI.Packages {
				log.WithField(ctx, "package", pkg.Name).Debug("Package")
			}

			if err := b.RunJ5ClientPlugin(ctx, variables, dst, plugin, descriptorAPI, errOut); err != nil {
				return fmt.Errorf("plugin %s for input %s: %w", plugin.Name, input.Name(), err)
			}

		default:
			return fmt.Errorf("unsupported plugin type: %s", plugin.Type)
		}
	}

	return nil
}

func mapEnvVars(spec []string, vars map[string]string) ([]string, error) {
	env := make([]string, len(spec))
	for idx, src := range spec {

		parts := strings.Split(src, "=")
		if len(parts) == 1 {
			env[idx] = fmt.Sprintf("%s=%s", src, os.Getenv(src))
			continue
		}
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env var: %s", src)
		}
		val := os.Expand(parts[1], func(key string) string {
			if v, ok := vars[key]; ok {
				return v
			}
			return ""
		})

		env[idx] = fmt.Sprintf("%s=%s", parts[0], val)
	}
	return env, nil
}

type runnerContext struct {
	vars   map[string]string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (b *Builder) runRunner(ctx context.Context, run runnerContext, plugin *config_j5pb.BuildPlugin) error {

	switch runner := plugin.Runner.(type) {
	case *config_j5pb.BuildPlugin_Docker:
		envVars, err := mapEnvVars(runner.Docker.Env, run.vars)
		if err != nil {
			return err
		}
		err = b.Docker.Run(ctx, runner.Docker, run.stdin, run.stdout, run.stderr, envVars)
		if err != nil {
			return fmt.Errorf("running docker %s: %w", plugin.GetName(), err)
		}

	case *config_j5pb.BuildPlugin_Command:

		envVars, err := mapEnvVars(runner.Command.Env, run.vars)
		if err != nil {
			return err
		}

		err = b.RunCommand(ctx, runner.Command, run.stdin, run.stdout, run.stderr, envVars)
		if err != nil {
			return fmt.Errorf("running command %s: %w", plugin.GetName(), err)
		}

	default:
		return fmt.Errorf("unsupported runner: %T", runner)
	}

	return nil

}

func (b *Builder) RunProtocPlugin(ctx context.Context, variables map[string]string, dest FS, plugin *config_j5pb.BuildPlugin, sourceProto *pluginpb.CodeGeneratorRequest, errOut io.Writer) error {

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

	if err := b.runRunner(ctx, runnerContext{
		vars:   variables,
		stdin:  inBuffer,
		stdout: outBuffer,
		stderr: errOut,
	}, plugin); err != nil {
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
		if err := dest.Put(ctx, name, reader); err != nil {
			return err
		}
	}

	log.WithFields(ctx, map[string]interface{}{
		"files":           len(resp.File),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil
}

func (b *Builder) RunJ5ClientPlugin(ctx context.Context, variables map[string]string, dest FS, plugin *config_j5pb.BuildPlugin, descriptorAPI *schema_j5pb.API, errOut io.Writer) error {

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

	if err := b.runRunner(ctx, runnerContext{
		vars:   variables,
		stdin:  inBuffer,
		stdout: outBuffer,
		stderr: errOut,
	}, plugin); err != nil {
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
		if err := dest.Put(ctx, name, reader); err != nil {
			return err
		}
	}

	log.WithFields(ctx, map[string]interface{}{
		"files":           len(resp.Files),
		"durationSeconds": time.Since(start).Seconds(),
	}).Info("Protoc Plugin Complete")

	return nil

}

func (b *Builder) RunCommand(ctx context.Context, spec *config_j5pb.CommandSpec, input io.Reader, output, errOutput io.Writer, envVars []string) error {
	cmd := exec.CommandContext(ctx, spec.Command, spec.Args...)
	cmd.Stdin = input
	cmd.Stdout = output
	cmd.Stderr = errOutput
	cmd.Env = envVars
	return cmd.Run()
}
