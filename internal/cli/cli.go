package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"runtime/debug"

	"github.com/pentops/j5/builder"
	"github.com/pentops/j5/internal/source"
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

	cmdGroup.Add("schema", schemaSet())
	cmdGroup.Add("protoc", protoSet())

	cmdGroup.Add("version", commander.NewCommand(runVersion))
	cmdGroup.Add("generate", commander.NewCommand(runGenerate))
	cmdGroup.Add("publish", commander.NewCommand(runPublish))
	cmdGroup.Add("verify", commander.NewCommand(runVerify))

	return cmdGroup
}

func runVersion(ctx context.Context, cfg struct{}) error {
	fmt.Printf("jsonapi version %v\n", Commit)
	return nil
}

type SourceConfig struct {
	Source string `flag:"src" default:"." description:"Source directory containing j5.yaml and buf.lock.yaml"`
	Bundle string `flag:"bundle" default:"" description:"When the bundle j5.yaml is in a subdirectory"`
}

func (cfg SourceConfig) GetSource(ctx context.Context) (*source.Source, error) {
	return source.ReadLocalSource(ctx, os.DirFS(cfg.Source))
}

func (cfg SourceConfig) GetInput(ctx context.Context) (source.Input, error) {
	source, err := cfg.GetSource(ctx)
	if err != nil {
		return nil, err
	}
	return source.NamedInput(cfg.Bundle)
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

type Dest interface {
	builder.Dest
	Sub(subPath string) Dest
}

type DiscardFS struct{}

func NewDiscardFS() *DiscardFS {
	return &DiscardFS{}
}

func (d *DiscardFS) Sub(subPath string) Dest {
	return d
}

func (d *DiscardFS) PutFile(ctx context.Context, subPath string, body io.Reader) error {
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

func (local *LocalFS) Sub(subPath string) Dest {
	return &LocalFS{
		root: filepath.Join(local.root, subPath),
	}
}

func (local *LocalFS) PutFile(ctx context.Context, subPath string, body io.Reader) error {
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
