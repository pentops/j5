package cli

import (
	"context"

	"github.com/pentops/j5/gogen"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/runner/commander"
)

func generateSet() *commander.CommandSet {
	genGroup := commander.NewCommandSet()
	genGroup.Add("gocode", commander.NewCommand(runGocode))
	return genGroup
}

func runGocode(ctx context.Context, cfg struct {
	SourceConfig
	OutputDir         string `flag:"output-dir" description:"Directory to write go source"`
	PackagePrefix     string `flag:"package-prefix" default:"" description:"Only generate files matching this prefix"`
	TrimPackagePrefix string `flag:"trim-package-prefix" default:"" description:"Proto package name to remove from go package names"`
	AddGoPrefix       string `flag:"add-go-prefix" default:"" description:"Prefix to add to go package names"`
}) error {

	source, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	image, err := source.SourceImage(ctx)
	if err != nil {
		return err
	}

	jdefDoc, err := structure.BuildFromImage(ctx, image)
	if err != nil {
		return err
	}

	options := gogen.Options{
		TrimPackagePrefix: cfg.TrimPackagePrefix,
		PackagePrefix:     cfg.PackagePrefix,
		AddGoPrefix:       cfg.AddGoPrefix,
	}

	output := gogen.DirFileWriter(cfg.OutputDir)

	if err := gogen.WriteGoCode(jdefDoc, output, options); err != nil {
		return err
	}

	return nil
}
