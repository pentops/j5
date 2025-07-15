package integration

import (
	"context"
	"testing"

	"github.com/pentops/flowtest"
	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/gen/j5/registry/v1/registry_spb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestJ5Store(t *testing.T) {

	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	commitHash := "we5rbvfcb"

	regConfig := &config_j5pb.RegistryConfig{
		Owner: "cfgorg",
		Name:  "cfgrepo",
	}
	sourceImage := &source_j5pb.SourceImage{
		File: []*descriptorpb.FileDescriptorProto{{
			Name:    proto.String("test/v1/test.proto"),
			Package: proto.String("test.v1"),
		}},
		Packages: []*source_j5pb.Package{{
			Name:  "test.v1",
			Label: "Test Package",
		}},
	}
	flow.Step("Upload J5 Config", func(ctx context.Context, t flowtest.Asserter) {
		if err := uu.PackageStore.UploadJ5Image(ctx, &source_j5pb.CommitInfo{
			Owner: "commitorg",
			Repo:  "commitrepo",
			Hash:  commitHash,
			Time:  timestamppb.Now(),
			Aliases: []string{
				"refs/heads/main",
			},
		}, sourceImage, regConfig); err != nil {
			t.Fatalf("failed to upload j5 image: %v", err)
		}
	})

	flow.Step("Download By Branch", func(ctx context.Context, t flowtest.Asserter) {
		res, err := uu.RegistryDownload.DownloadClientAPI(ctx, &registry_spb.DownloadClientAPIRequest{
			Owner:   "cfgorg",
			Name:    "cfgrepo",
			Version: "main",
		})
		t.NoError(err)

		aa := jsontest.NewTestAsserter(t, res.Api)

		aa.AssertEqual("packages.0.name", "test.v1")
	})

	flow.Step("Download By Commit", func(ctx context.Context, t flowtest.Asserter) {
		res, err := uu.RegistryDownload.DownloadClientAPI(ctx, &registry_spb.DownloadClientAPIRequest{
			Owner:   "cfgorg",
			Name:    "cfgrepo",
			Version: commitHash,
		})
		t.NoError(err)

		aa := jsontest.NewTestAsserter(t, res.Api)

		aa.AssertEqual("packages.0.name", "test.v1")
	})

	flow.Step("Next Commit", func(ctx context.Context, t flowtest.Asserter) {
		if err := uu.PackageStore.UploadJ5Image(ctx, &source_j5pb.CommitInfo{
			Owner: "commitorg",
			Repo:  "commitrepo",
			Hash:  "commit2",
			Time:  timestamppb.Now(),
			Aliases: []string{
				"refs/heads/main",
			},
		}, sourceImage, regConfig); err != nil {
			t.Fatalf("failed to upload j5 image: %v", err)
		}
	})

}
