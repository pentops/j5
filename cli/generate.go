package cli

import (
	"context"
	"fmt"

	"github.com/pentops/jsonapi/gogen"
	"github.com/pentops/jsonapi/source"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/runner/commander"
)

func GenerateSet() *commander.CommandSet {
	genGroup := commander.NewCommandSet()
	genGroup.Add("gocode", commander.NewCommand(runGocode))
	return genGroup
}

func runGocode(ctx context.Context, cfg struct {
	Source            string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	OutputDir         string `flag:"output-dir" description:"Directory to write go source"`
	TrimPackagePrefix string `flag:"trim-package-prefix" default:"" description:"Prefix to trim from go package names"`
	AddGoPrefix       string `flag:"add-go-prefix" default:"" description:"Prefix to add to go package names"`
}) error {
	image, err := source.ReadImageFromSourceDir(ctx, cfg.Source, true)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}

	jdefDoc, err := structure.BuildFromImage(image)
	if err != nil {
		return fmt.Errorf("build structure: %w", err)
	}

	options := gogen.Options{
		TrimPackagePrefix: cfg.TrimPackagePrefix,
		AddGoPrefix:       cfg.AddGoPrefix,
	}

	output := gogen.DirFileWriter(cfg.OutputDir)

	if err := gogen.WriteGoCode(jdefDoc, output, options); err != nil {
		return err
	}

	return nil
}
