package integration

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/pentops/flowtest"
	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGomodStore(t *testing.T) {

	flow, uu := NewUniverse(t)
	defer flow.RunSteps(t)

	commitHash := "we5rbvfcb94vjn69nv6j95dn"
	flow.Step("Upload Go Module", func(ctx context.Context, t flowtest.Asserter) {

		files := fstest.MapFS{
			"go.mod": &fstest.MapFile{
				Mode: 0644,

				Data: []byte(strings.Join([]string{
					"module example.com/org/repo",
					"",
					"go 1.22.3",
				}, "\n")),
			},
			"dir1/file1.go": &fstest.MapFile{
				Mode: 0644,
				Data: []byte("package main"),
			},
			"dir2/file1.go": &fstest.MapFile{
				Mode: 0644,
				Data: []byte("package main"),
			},
		}

		if err := uu.PackageStore.UploadGoModule(ctx, &source_j5pb.CommitInfo{
			Owner: "commitorg",
			Repo:  "commitrepo",
			Hash:  commitHash,
			Time:  timestamppb.Now(),
			Aliases: []string{
				"refs/heads/main",
			},
		}, files); err != nil {
			t.Fatalf("failed to upload go module: %v", err)
		}
	})

	flow.Step("Go Mod Info", func(ctx context.Context, t flowtest.Asserter) {
		res := uu.HTTPGet(ctx, "/gopkg/example.com/org/repo/@v/main.info")
		t.Equal(200, res.StatusCode)
		t.Log(string(res.Body))

		aa := jsontest.NewTestAsserter(t, res.Body)

		val, ok := aa.Get("Version")
		if !ok {
			t.Fatalf("missing version")
		}
		parts := strings.Split(val.(string), "-")
		if len(parts) != 3 {
			t.Fatalf("unexpected version: %v", val)
		}
		t.Equal("v0.0.0", parts[0])

		if !strings.HasPrefix(val.(string), "v0.0.0-") {
			t.Fatalf("unexpected version: %v", val)
		}

		if !strings.HasPrefix(commitHash, parts[2]) {
			t.Fatalf("unexpected version: %v", val)
		}
	})

	flow.Step("Go Mod File", func(ctx context.Context, t flowtest.Asserter) {
		res := uu.HTTPGet(ctx, "/gopkg/example.com/org/repo/@v/main.mod")
		t.Equal(200, res.StatusCode)
		t.Log(string(res.Body))

		lines := strings.Split(string(res.Body), "\n")
		line := lines[0]
		t.Equal("module example.com/org/repo", line)
	})

	flow.Step("Go Mod Zip", func(ctx context.Context, t flowtest.Asserter) {
		res := uu.HTTPGet(ctx, "/gopkg/example.com/org/repo/@v/main.zip")
		t.Equal(200, res.StatusCode)
		//	t.Log(string(res.Body))
	})

	flow.Step("List", func(ctx context.Context, t flowtest.Asserter) {
		res := uu.HTTPGet(ctx, "/gopkg/example.com/org/repo/@v/list")
		t.Equal(200, res.StatusCode)
		str := string(res.Body)
		if str != "" {
			t.Fatal("body should be empty for list until Tags are supported")
		}
	})

}
