package genlsp

import (
	"context"
	"os"

	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/internal/bcl"
	"github.com/pentops/j5/internal/bcl/internal/linter"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Config struct {
	ProjectRoot string

	FileTypes []FileTypeConfig
}

type FileTypeConfig struct {
	Match       func(filename string) bool
	FileFactory func(filename string) protoreflect.Message
	OnChange    func(ctx context.Context, filename string, sourceLocs *bcl_j5pb.SourceLocation, parsed protoreflect.Message) error
}

type fileType struct {
	match func(filename string) bool
	FileHandler
}

func (ft fileType) MatchFilename(filename string) bool {
	return ft.match(filename)
}

func BuildLSPHandler(config Config) (*lspConfig, error) {
	lspc := lspConfig{
		ProjectRoot: config.ProjectRoot,
	}

	if config.ProjectRoot == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		config.ProjectRoot = pwd
	}

	for _, ft := range config.FileTypes {
		built := fileType{
			match: ft.Match,
		}

		if ft.FileFactory != nil {
			parser, err := bcl.NewParser()
			if err != nil {
				return nil, err
			}
			built.FileHandler = linter.New(parser, ft.FileFactory, ft.OnChange, config.ProjectRoot)
		} else {
			built.FileHandler = linter.NewGeneric(config.ProjectRoot)
		}
		lspc.Handlers = append(lspc.Handlers, built)
	}

	lspc.Formatter = astFormatter{}

	return &lspc, nil

}

func RunLSP(ctx context.Context, config Config) error {
	lspc, err := BuildLSPHandler(config)
	if err != nil {
		return err
	}

	ctx = log.WithField(ctx, "ProjectRoot", config.ProjectRoot)
	log.Info(ctx, "Starting LSP server")

	ss, err := newServerStream(*lspc)
	if err != nil {
		return err
	}

	return ss.Run(ctx, stdIO())
}
