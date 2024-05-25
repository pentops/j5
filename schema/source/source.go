package source

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Source struct {
	// input
	commitInfo *builder_j5pb.CommitInfo
	config     *source_j5pb.Config
	sourceDir  string

	// cache
	codegenReq *pluginpb.CodeGeneratorRequest
	sourceImg  *source_j5pb.SourceImage
}

func NewLocalDirSource(ctx context.Context, commitInfo *builder_j5pb.CommitInfo, config *source_j5pb.Config, sourceDir string) (*Source, error) {
	return &Source{
		config:     config,
		commitInfo: commitInfo,
		sourceDir:  sourceDir,
	}, nil
}

func (src Source) J5Config() *source_j5pb.Config {
	return src.config
}

func (src Source) CommitInfo(context.Context) (*builder_j5pb.CommitInfo, error) {
	return src.commitInfo, nil
}

func (src *Source) ProtoCodeGeneratorRequest(ctx context.Context) (*pluginpb.CodeGeneratorRequest, error) {
	if src.codegenReq == nil {
		rr, err := CodeGeneratorRequestFromSource(ctx, src.sourceDir)
		if err != nil {
			return nil, err
		}
		src.codegenReq = rr
	}
	return src.codegenReq, nil
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

func (src *Source) SourceFile(ctx context.Context, filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(src.sourceDir, filename))
}
