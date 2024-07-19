package j5client

import (
	"context"
	"os"
	"testing"

	"github.com/pentops/j5/internal/source"
	"github.com/pentops/j5/internal/structure"
)

func TestFooSchema(t *testing.T) {

	ctx := context.Background()
	rootFS := os.DirFS("../../")
	thisRoot, err := source.ReadLocalSource(ctx, rootFS)
	if err != nil {
		t.Fatalf("ReadLocalSource: %v", err)
	}

	input, err := thisRoot.NamedInput("test")
	if err != nil {
		t.Fatalf("NamedInput: %v", err)
	}

	srcImg, err := input.SourceImage(ctx)
	if err != nil {
		t.Fatalf("SourceImage: %v", err)
	}

	sourceAPI, err := structure.APIFromImage(srcImg)
	if err != nil {
		t.Fatalf("APIFromImage: %v", err)
	}

	for _, pkg := range sourceAPI.Packages {
		t.Logf("Package: %s", pkg.Name)
		for name := range pkg.Schemas {
			t.Logf("Schema: %s", name)
		}
	}

	clientAPI, err := APIFromSource(sourceAPI)
	if err != nil {
		t.Fatalf("APIFromSource: %v", err)
	}

	t.Logf("ClientAPI: %v", clientAPI)

}
