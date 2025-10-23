package buildwrap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	"github.com/pentops/j5/gen/j5/registry/v1/registry_tpb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/registry/github"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-messaging/o5msg"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Storage interface {
	UploadGoModule(ctx context.Context, commitInfo *source_j5pb.CommitInfo, fs fs.FS) error
	UploadJ5Image(ctx context.Context, commitInfo *source_j5pb.CommitInfo, img *source_j5pb.SourceImage, reg *config_j5pb.RegistryConfig) error
	GetJ5Image(ctx context.Context, orgName, imageName, version string) (*source_j5pb.SourceImage, error)
}

type Publisher interface {
	Publish(ctx context.Context, msg o5msg.Message) error
}

type BuildWorker struct {
	registry_tpb.UnimplementedBuilderRequestTopicServer

	builder J5Builder
	github  IGithub

	store     Storage
	publisher Publisher
}

type J5Builder interface {
	RunPublishBuild(ctx context.Context, pc PluginContext, input *source_j5pb.SourceImage, build *config_j5pb.PublishConfig) error
	MutateImageWithMods(img *source_j5pb.SourceImage, mods []*config_j5pb.ProtoMod) error
	SourceImage(ctx context.Context, fs fs.FS, bundleName string) (*source_j5pb.SourceImage, *config_j5pb.BundleConfigFile, error)
}

type IGithub interface {
	GetContent(ctx context.Context, ref *github.Commit, intoDir string) error
	GetCommit(ctx context.Context, ref *github.Commit) (*source_j5pb.CommitInfo, error)
}

func NewBuildWorker(builder J5Builder, github IGithub, store Storage, publisher o5msg.Publisher) *BuildWorker {
	return &BuildWorker{
		builder:   builder,
		github:    github,
		store:     store,
		publisher: publisher,
	}
}

func (bw *BuildWorker) RegisterGRPC(s *grpc.Server) {
	registry_tpb.RegisterBuilderRequestTopicServer(s, bw)
}

func (bw *BuildWorker) replyStatus(ctx context.Context, request *messaging_j5pb.RequestMetadata, status registry_tpb.BuildStatus, output *registry_tpb.BuildOutput) error {
	if request == nil {
		return nil
	}
	return bw.publisher.Publish(ctx, &registry_tpb.J5BuildStatusMessage{
		Request: request,
		Status:  status,
		Output:  output,
	})
}

func (bw *BuildWorker) Publish(ctx context.Context, req *registry_tpb.PublishMessage) (*emptypb.Empty, error) {
	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, registry_tpb.BuildStatus_IN_PROGRESS, nil); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
	}

	logBuffer := &bytes.Buffer{}
	err := bw.runPublish(ctx, req, logBuffer)
	if err != nil {
		if req.Request == nil {
			return nil, fmt.Errorf("build: %w", err)
		}

		errorMessage := err.Error()
		fullText := fmt.Sprintf("%s\n\n```%s```", errorMessage, logBuffer.String())
		if err := bw.replyStatus(ctx, req.Request, registry_tpb.BuildStatus_FAILURE, &registry_tpb.BuildOutput{
			Title:   "proto build error",
			Summary: errorMessage,
			Text:    some(fullText),
		}); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
		return &emptypb.Empty{}, nil
	}

	logStr := logBuffer.String()
	if len(logStr) >= 65535 {
		trunc := "... (truncated see logs for full error)"
		logStr = logStr[:65535-len(trunc)] + trunc
	}

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, registry_tpb.BuildStatus_SUCCESS, &registry_tpb.BuildOutput{
			Title: "proto build success",
			Text:  some(logStr),
		}); err != nil {
			log.Error(ctx, logStr)
			return nil, fmt.Errorf("update checkrun: completed: %w", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func (bw *BuildWorker) runPublish(ctx context.Context, req *registry_tpb.PublishMessage, logBuffer io.Writer) error {
	img, cfg, err := bw.BundleImageFromCommit(ctx, req.Commit, req.Bundle)
	if err != nil {
		return fmt.Errorf("new fs input: %w", err)
	}

	dest, err := newTmpDest()
	if err != nil {
		return fmt.Errorf("make tmp dest: %w", err)
	}
	defer dest.Close()

	pc := PluginContext{
		Variables: map[string]string{}, // TODO: Commit / Source Info
		ErrOut:    logBuffer,
		Dest:      dest,
	}

	var publishConfig *config_j5pb.PublishConfig
	for _, publish := range cfg.Publish {
		if publish.Name == req.Name {
			publishConfig = publish
			break
		}
	}
	if publishConfig == nil {
		return fmt.Errorf("publish config %q not found", req.Name)
	}
	if err := bw.builder.MutateImageWithMods(img, publishConfig.Mods); err != nil {
		return fmt.Errorf("mutate image: %w", err)
	}

	// Build
	if err := bw.builder.RunPublishBuild(ctx, pc, img, publishConfig); err != nil {
		return err
	}

	// Package And Upload
	if publishConfig.OutputFormat == nil || publishConfig.OutputFormat.Type == nil {
		return fmt.Errorf("output format not set")
	}
	switch publishConfig.OutputFormat.Type.(type) {
	case *config_j5pb.OutputType_GoProxy_:

		err = bw.store.UploadGoModule(ctx, req.Commit, dest)
		if err != nil {
			return fmt.Errorf("upload go module: %w", err)
		}
	default:
		return fmt.Errorf("unsupported package type")
	}

	return nil
}

func (bw *BuildWorker) BuildAPI(ctx context.Context, req *registry_tpb.BuildAPIMessage) (*emptypb.Empty, error) {
	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, registry_tpb.BuildStatus_IN_PROGRESS, nil); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
	}

	log.WithField(ctx, "commit", req.Commit).Info("Build API")

	err := bw.buildAPI(ctx, req.Commit, req)
	if err != nil {
		if req.Request == nil {
			return nil, err
		}
		errorMessage := err.Error()
		if err := bw.replyStatus(ctx, req.Request, registry_tpb.BuildStatus_FAILURE, &registry_tpb.BuildOutput{
			Title:   "proto build error",
			Summary: errorMessage,
		}); err != nil {
			return nil, fmt.Errorf("reply status: %w", err)
		}
		return &emptypb.Empty{}, nil
	}

	if req.Request != nil {
		if err := bw.replyStatus(ctx, req.Request, registry_tpb.BuildStatus_SUCCESS, &registry_tpb.BuildOutput{
			Title: "API Build Success",
		}); err != nil {
			return nil, fmt.Errorf("update checkrun: completed: %w", err)
		}
	}

	return &emptypb.Empty{}, nil

}

func (bw *BuildWorker) buildAPI(ctx context.Context, commit *source_j5pb.CommitInfo, req *registry_tpb.BuildAPIMessage) error {
	img, bundleConfig, err := bw.BundleImageFromCommit(ctx, commit, req.Bundle)
	if err != nil {
		return fmt.Errorf("new fs input: %w", err)
	}

	return bw.store.UploadJ5Image(ctx, req.Commit, img, bundleConfig.Registry)
}

func some[T any](s T) *T {
	return &s
}
