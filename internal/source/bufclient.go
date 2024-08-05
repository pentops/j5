package source

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"

	registry_spb "buf.build/gen/go/bufbuild/buf/grpc/go/buf/alpha/registry/v1alpha1/registryv1alpha1grpc"
	registry_pb "buf.build/gen/go/bufbuild/buf/protocolbuffers/go/buf/alpha/registry/v1alpha1"
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type bufClient struct {
	downloadClient registry_spb.DownloadServiceClient
	versionClient  registry_spb.RepositoryCommitServiceClient
}

func NewBufClient() (*bufClient, error) {

	grpcClient, err := grpc.NewClient("buf.build:443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return nil, err
	}
	registryClient := registry_spb.NewDownloadServiceClient(grpcClient)
	versionClient := registry_spb.NewRepositoryCommitServiceClient(grpcClient)
	return &bufClient{
		downloadClient: registryClient,
		versionClient:  versionClient,
	}, nil
}

func (bc *bufClient) LatestImage(ctx context.Context, owner, repoName string, reference *string) (*source_j5pb.SourceImage, error) {
	refLock := ""
	if reference != nil {
		refLock = *reference
	}
	res, err := bc.versionClient.GetRepositoryCommitByReference(ctx, &registry_pb.GetRepositoryCommitByReferenceRequest{
		RepositoryOwner: owner,
		RepositoryName:  repoName,
		Reference:       refLock,
	})
	if err != nil {
		return nil, err
	}

	version := res.RepositoryCommit.Name
	if version == "" || version == refLock {
		return nil, fmt.Errorf("no version found")
	}

	img, err := bc.GetImage(ctx, owner, repoName, version)
	if err != nil {
		return nil, err
	}

	img.Version = &version
	return img, nil
}

func (bc *bufClient) GetImage(ctx context.Context, owner, name, version string) (*source_j5pb.SourceImage, error) {

	downloadRes, err := bc.downloadClient.Download(ctx, &registry_pb.DownloadRequest{
		Owner:      owner,
		Repository: name,
		Reference:  version,
	})
	if err != nil {
		return nil, err
	}

	filenames := []string{}
	fileMap := map[string][]byte{}
	for _, file := range downloadRes.Module.Files {
		if _, ok := fileMap[file.Path]; ok {
			return nil, fmt.Errorf("duplicate file %s", file.Path)
		}
		fileMap[file.Path] = file.Content
		filenames = append(filenames, file.Path)
	}

	parser := protoparse.Parser{
		ImportPaths:                     []string{""},
		IncludeSourceCodeInfo:           true,
		InterpretOptionsInUnlinkedFiles: true,

		WarningReporter: func(err reporter.ErrorWithPos) {
			log.WithFields(ctx, map[string]interface{}{
				"error": err.Error(),
			}).Warn("protoparse warning")
		},

		Accessor: func(filename string) (io.ReadCloser, error) {
			content, ok := fileMap[filename]
			if !ok {
				return nil, fs.ErrNotExist
			}
			return io.NopCloser(bytes.NewReader(content)), nil
		},
	}

	customDesc, err := parser.ParseFiles(filenames...)
	if err != nil {
		panicErr := protocompile.PanicError{}
		if errors.As(err, &panicErr) {
			fmt.Printf("PANIC: %s\n", panicErr.Stack)
		}

		return nil, fmt.Errorf("parsing buf input %q/%q: %w", owner, name, err)
	}

	realDesc := desc.ToFileDescriptorSet(customDesc...)

	img := &source_j5pb.SourceImage{
		File:            realDesc.File,
		SourceFilenames: filenames,
	}

	return img, nil
}
