package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"runtime/debug"

	"github.com/pentops/j5/builder/builder"
	"github.com/pentops/j5/builder/docker"
	"github.com/pentops/j5/schema/j5reflect"
	"github.com/pentops/j5/schema/source"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/log.go/log"
	"github.com/pentops/runner/commander"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "dev"
}()

func CommandSet() *commander.CommandSet {

	cmdGroup := commander.NewCommandSet()

	cmdGroup.Add("registry", registrySet())
	cmdGroup.Add("schema", schemaSet())
	cmdGroup.Add("proto", protoSet())

	cmdGroup.Add("version", commander.NewCommand(runVersion))
	cmdGroup.Add("generate", commander.NewCommand(runGenerate))
	cmdGroup.Add("verify", commander.NewCommand(runVerify))

	return cmdGroup
}

func runVersion(ctx context.Context, cfg struct{}) error {
	fmt.Printf("jsonapi version %v\n", Commit)
	return nil
}

func runVerify(ctx context.Context, cfg struct {
	Dir string `flag:"dir" default:"." description:"Directory with j5.yaml root"`
}) error {

	src, err := source.ReadLocalSource(ctx, os.DirFS(cfg.Dir))
	if err != nil {
		return err
	}

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	j5Config := src.J5Config()

	for _, bundle := range src.AllBundles() {

		image, err := bundle.SourceImage(ctx)
		if err != nil {
			return err
		}

		reflectionAPI, err := structure.ReflectFromSource(image)
		if err != nil {
			return fmt.Errorf("ReflectFromSource: %w", err)
		}

		descriptorAPI, err := structure.DescriptorFromReflection(reflectionAPI)
		if err != nil {
			return fmt.Errorf("DescriptorFromReflection: %w", err)
		}

		if err := structure.ResolveProse(image, descriptorAPI); err != nil {
			return fmt.Errorf("ResolveProse: %w", err)
		}

		_, err = j5reflect.APIFromDesc(descriptorAPI)
		if err != nil {
			return fmt.Errorf("building reflection from descriptor: %w", err)
		}
		for _, pkg := range descriptorAPI.Packages {
			fmt.Printf("Package %s OK\n", pkg.Name)
		}

	}

	dest := NewDiscardFS()
	for _, generator := range j5Config.Generate {
		out := &lineWriter{
			writeLine: func(line string) {
				log.WithField(ctx, "generator", generator.Name).Info(line)
			},
		}

		bb := builder.NewBuilder(dockerWrapper)
		err = bb.RunGenerate(ctx, src, dest, generator, out)
		out.flush()
		if err != nil {
			return err
		}

	}

	return nil

}

func runGenerate(ctx context.Context, cfg struct {
	Dir string `flag:"dir" default:"." description:"Directory with j5.yaml generate configured"`
}) error {

	src, err := source.ReadLocalSource(ctx, os.DirFS(cfg.Dir))
	if err != nil {
		return err
	}

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	j5Config := src.J5Config()

	for _, generator := range j5Config.Generate {
		out := &lineWriter{
			writeLine: func(line string) {
				log.WithField(ctx, "generator", generator.Name).Info(line)
			},
		}

		outDir := filepath.Join(cfg.Dir, generator.Output)
		dest, err := NewLocalFS(outDir)
		if err != nil {
			return err
		}

		bb := builder.NewBuilder(dockerWrapper)
		err = bb.RunGenerate(ctx, src, dest, generator, out)
		out.flush()
		if err != nil {
			return err
		}

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

type SourceConfig struct {
	Source string `flag:"src" default:"." description:"Source directory containing j5.yaml and buf.lock.yaml"`
	Bundle string `flag:"bundle" default:"" description:"When the bundle j5.yaml is in a subdirectory"`
}

func (cfg SourceConfig) GetSource(ctx context.Context) (*source.Source, error) {

	sourceDir := cfg.Source

	return source.ReadLocalSource(ctx, os.DirFS(sourceDir))
}

func (cfg SourceConfig) GetInput(ctx context.Context) (source.Input, error) {
	source, err := cfg.GetSource(ctx)
	if err != nil {
		return nil, err
	}
	return source.NamedInput(cfg.Bundle)
}

type DiscardFS struct{}

func NewDiscardFS() *DiscardFS {
	return &DiscardFS{}
}

func (d *DiscardFS) Put(ctx context.Context, subPath string, body io.Reader) error {
	return nil
}

type LocalFS struct {
	root string
}

func NewLocalFS(root string) (*LocalFS, error) {
	return &LocalFS{
		root: root,
	}, nil
}

func (local *LocalFS) Put(ctx context.Context, subPath string, body io.Reader) error {
	key := filepath.Join(local.root, subPath)
	log.WithField(ctx, "filename", key).Debug("writing file")
	err := os.MkdirAll(filepath.Dir(key), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(key)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err
	}

	return nil
}
