package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pentops/j5/builder/git"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/j5/schema/swagger"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Source interface {
	J5Config() *config_j5pb.Config
	CommitInfo(ctx context.Context) (*source_j5pb.CommitInfo, error)
	ProtoCodeGeneratorRequest(ctx context.Context, root string) (*pluginpb.CodeGeneratorRequest, error)
	SourceImage(ctx context.Context) (*source_j5pb.SourceImage, error)
	SourceDescriptors(ctx context.Context) ([]*descriptorpb.FileDescriptorProto, error)
	SourceFile(ctx context.Context, filename string) ([]byte, error)
	ResolvePlugin(plugin *config_j5pb.BuildPlugin) (*config_j5pb.BuildPlugin, error)
	PackageBuildConfig(name string) (*config_j5pb.ProtoBuildConfig, error)
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

func (b *Builder) BuildAll(ctx context.Context, src Source, dst FS) error {

	spec := src.J5Config()

	commitInfo, err := src.CommitInfo(ctx)
	if err != nil {
		return err
	}

	img, err := src.SourceImage(ctx)
	if err != nil {
		return fmt.Errorf("read image: %w", err)
	}

	if spec.Git != nil {
		git.ExpandGitAliases(spec.Git, commitInfo)
	}

	if _, err := b.BuildJsonAPI(ctx, img); err != nil {
		return fmt.Errorf("building JSON API: %w", err)
	}

	for _, dockerBuild := range spec.ProtoBuilds {
		ctx = log.WithField(ctx, "dockerBuild", dockerBuild.Name)
		lineWriter := &lineWriter{
			writeLine: func(line string) {
				log.WithField(ctx, "line", line).Info("docker build")
			},
		}
		if err := b.BuildProto(ctx, src, dst, dockerBuild.Name, lineWriter); err != nil {
			lineWriter.flush()
			return fmt.Errorf("running proto build %s: %w", dockerBuild.Name, err)
		}
		lineWriter.flush()
	}

	return nil
}

type lineWriter struct {
	buf       []byte
	writeLine func(string)
}

func (w *lineWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if b == '\n' {
			w.writeLine(string(w.buf))
			w.buf = w.buf[:0]
		} else {
			w.buf = append(w.buf, b)
		}
	}
	return len(p), nil
}

func (w *lineWriter) flush() {
	if len(w.buf) > 0 {
		w.writeLine(string(w.buf))
	}
	w.buf = []byte{}
}

func (b *Builder) BuildJsonAPI(ctx context.Context, img *source_j5pb.SourceImage) (*J5Upload, error) {
	log.Info(ctx, "build json API")

	jdefDoc, err := structure.BuildFromImage(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("build from image: %w", err)
	}

	swaggerDoc, err := swagger.BuildSwagger(jdefDoc)
	if err != nil {
		return nil, fmt.Errorf("build swagger: %w", err)
	}

	return &J5Upload{
		Image:   img,
		JDef:    jdefDoc,
		Swagger: swaggerDoc,
	}, nil
}

func (b *Builder) BuildProto(ctx context.Context, src Source, dst FS, buildName string, errOut io.Writer) error {

	var dockerBuild *config_j5pb.ProtoBuildConfig
	for _, b := range src.J5Config().ProtoBuilds {
		if b.Name == buildName {
			dockerBuild = b
			break
		}
	}

	if dockerBuild == nil {
		return fmt.Errorf("build not found: %s", buildName)
	}

	commitInfo, err := src.CommitInfo(ctx)
	if err != nil {
		return err
	}

	protoBuildRequest, err := src.ProtoCodeGeneratorRequest(ctx, ".")
	if err != nil {
		return fmt.Errorf("ProtoCodeGeneratorRequest: %w", err)
	}

	switch pkg := dockerBuild.PackageType.(type) {
	case *config_j5pb.ProtoBuildConfig_GoProxy_:

		for _, plugin := range dockerBuild.Plugins {

			envVars, err := MapEnvVars(plugin.Docker.Env, commitInfo)
			if err != nil {
				return err
			}
			if err := b.RunProtocPlugin(ctx, dst, plugin, protoBuildRequest, errOut, envVars); err != nil {
				return err
			}
		}

		gomodFile, err := src.SourceFile(ctx, pkg.GoProxy.GoModFile)
		if err != nil {
			return err
		}

		err = dst.Put(ctx, "go.mod", bytes.NewReader(gomodFile))
		if err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unsupported package type: %T", pkg)
	}
}

func MapEnvVars(spec []string, commitInfo *source_j5pb.CommitInfo) ([]string, error) {
	env := make([]string, len(spec))
	for idx, src := range spec {
		if strings.Contains(src, "PROTOC_GEN_GO_MESSAGING_EXTRA_HEADERS") && strings.Contains(src, "$GIT_HASH") {
			env[idx] = fmt.Sprintf("PROTOC_GEN_GO_MESSAGING_EXTRA_HEADERS=api-version:%v", commitInfo.Hash)
			continue
		}
		parts := strings.Split(src, "=")
		if len(parts) == 1 {
			env[idx] = fmt.Sprintf("%s=%s", src, os.Getenv(src))
			continue
		}
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env var: %s", src)
		}
		val := os.ExpandEnv(src)

		env[idx] = fmt.Sprintf("%s=%s", parts[0], val)
	}
	return env, nil
}

func (b *Builder) RunProtocPlugin(ctx context.Context, dest FS, plugin *config_j5pb.BuildPlugin, sourceProto *pluginpb.CodeGeneratorRequest, errOut io.Writer, env []string) error {

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

	resp := pluginpb.CodeGeneratorResponse{}

	outBuffer := &bytes.Buffer{}
	inBuffer := bytes.NewReader(reqBytes)

	err = b.Docker.Run(ctx, plugin.Docker, inBuffer, outBuffer, errOut, env)
	if err != nil {
		return fmt.Errorf("running docker %s: %w", plugin.GetName(), err)
	}

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
