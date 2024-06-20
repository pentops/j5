package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gogen"
	"github.com/pentops/j5/schema/j5reflect"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
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

	descriptors := &descriptorpb.FileDescriptorSet{
		File: image.File,
	}

	if len(descriptors.File) < 1 {
		panic("Expected at least one descriptor file, found none")
	}

	config := &config_j5pb.Config{
		Packages: image.Packages,
		Options:  image.Codec,
	}

	if config.Packages == nil || len(config.Packages) < 1 {
		return fmt.Errorf("no packages to generate")
	}

	if config.Options == nil {
		config.Options = &config_j5pb.CodecOptions{}
	}

	descFiles, err := protodesc.NewFiles(descriptors)
	if err != nil {
		return fmt.Errorf("descriptor files: %w", err)
	}

	rootPackages, err := structure.BuildPackages(config, descFiles, nil)
	if err != nil {
		return err
	}

	options := gogen.Options{
		TrimPackagePrefix: cfg.TrimPackagePrefix,
		PackagePrefix:     cfg.PackagePrefix,
		AddGoPrefix:       cfg.AddGoPrefix,
	}

	output := gogen.DirFileWriter(cfg.OutputDir)

	schemaSet := j5reflect.NewSchemaResolver(descFiles)

	for _, j5Package := range rootPackages { // Only generate packages within the prefix.
		if !strings.HasPrefix(j5Package.Name, options.PackagePrefix) {
			continue
		}

		if err := gogen.WriteGoCode(j5Package, schemaSet, output, options); err != nil {
			return err
		}
	}
	return nil
}
