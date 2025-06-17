package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pentops/log.go/log"

	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/bcl/genlsp"
	"github.com/pentops/j5/internal/j5s/j5parse"
	"github.com/pentops/j5/internal/j5s/protobuild"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/j5/internal/source"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func runLSP(ctx context.Context, cfg struct {
	Dir string `flag:"project-root" default:"" desc:"Root schema directory"`
}) error {

	if cfg.Dir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		cfg.Dir = pwd
	}

	fileOut, err := makeLogFile()
	if err != nil {
		return fmt.Errorf("makeLogFile: %w", err)
	}
	defer fileOut.Close()
	logger := log.NewCallbackLogger(log.JSONLog(fileOut))
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		logger.SetLevel(slog.LevelDebug)
	case "info":
		logger.SetLevel(slog.LevelInfo)
	case "warn":
		logger.SetLevel(slog.LevelWarn)
	case "error":
		logger.SetLevel(slog.LevelError)
	default:
		logger.SetLevel(slog.LevelInfo)
	}
	log.DefaultLogger = logger

	fileTypes := []genlsp.FileTypeConfig{}
	cc, err := newLspCompiler(ctx, cfg.Dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		log.WithError(ctx, err).Warn("No j5 config found, running in basic mode")
	} else {
		j5s := genlsp.FileTypeConfig{
			FileFactory: j5parse.FileStub,
			OnChange:    cc.updateFile,
			Match: func(filename string) bool {
				return strings.HasSuffix(filename, ".j5")
			},
		}
		fileTypes = append(fileTypes, j5s)
	}

	// generic fallback for basic bcl syntax and format
	fileTypes = append(fileTypes, genlsp.FileTypeConfig{
		Match: func(filename string) bool {
			return true
		},
	})

	return genlsp.RunLSP(ctx, genlsp.Config{
		ProjectRoot: cfg.Dir,
		FileTypes:   fileTypes,
	})
}

type lspCompiler struct {
	srcRoot *source.RepoRoot
	rootDir string
}

func newLspCompiler(ctx context.Context, dir string) (*lspCompiler, error) {
	resolver, err := source.NewEnvResolver()
	if err != nil {
		return nil, err
	}
	fullDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	fsRoot := os.DirFS(fullDir)
	srcRoot, err := source.NewFSRepoRoot(ctx, fsRoot, resolver)
	if err != nil {
		return nil, err
	}

	cc := &lspCompiler{
		srcRoot: srcRoot,
		rootDir: fullDir,
	}
	return cc, nil
}

func (cc *lspCompiler) updateFile(ctx context.Context, filename string, locs *bcl_j5pb.SourceLocation, msg protoreflect.Message) error {

	fileRel, err := filepath.Rel(cc.rootDir, filename)
	if err != nil {
		return err
	}

	bundle, relToBundle, err := cc.srcRoot.BundleForFile(fileRel)
	if err != nil {
		return err
	}

	deps, err := bundle.GetDependencies(ctx, cc.srcRoot)
	if err != nil {
		return err
	}
	fileSource, err := bundle.FileSource()
	if err != nil {
		return err
	}

	compiler, err := protobuild.NewPackageSet(psrc.DescriptorFiles(deps), fileSource)
	if err != nil {
		return err
	}

	parsed, ok := msg.Interface().(*sourcedef_j5pb.SourceFile)
	if !ok {
		return fmt.Errorf("file %s is not a SourceFile", filename)
	}
	parsed.SourceLocations = locs

	lintErr, err := compiler.LintFile(ctx, relToBundle, parsed)
	if err != nil {
		log.WithError(ctx, err).Error("linting file")
		return err
	}
	if lintErr == nil {
		fmt.Fprintln(os.Stderr, "No linting errors")
		return nil
	}

	for _, err := range lintErr {
		if parsed.SourceLocations == nil {
			log.WithError(ctx, errors.New("no source locations in parsed file")).Error("Preview Diagnostic")
		}
		printSourceTree(ctx, parsed.SourceLocations)
		posStr := err.Pos.String()
		log.WithFields(ctx, "pos", posStr, "error", err.Err.Error()).Debug("Preview Diagnostic")
	}

	return lintErr

}

func printSourceTree(ctx context.Context, loc *bcl_j5pb.SourceLocation) {
	log.WithFields(ctx, "children", len(loc.Children)).Debug("Source Location")
	for name, child := range loc.Children {
		line := fmt.Sprintf("%s - %d:%d\n", name, child.StartLine, child.StartColumn)
		log.WithFields(ctx, "source", line).Debug("Source Location")
		printSourceTree(ctx, child)
	}

}

func makeLogFile() (io.WriteCloser, error) {
	stateDir, err := getStateDir()
	if err != nil {
		return nil, fmt.Errorf("getStateDir: %w", err)
	}
	logDir := filepath.Join(stateDir, "j5")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", logDir, err)
	}
	logFile := filepath.Join(logDir, "j5-lsp.log")
	fileOut, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("create log file %s: %w", logFile, err)
	}
	return fileOut, nil
}

func getStateDir() (string, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("neither $XDG_STATE_HOME nor $HOME are defined")
		}
		dir += "/.local/state"
	} else if !filepath.IsAbs(dir) {
		return "", errors.New("path in $XDG_STATE_HOME is relative")
	}

	return dir, nil
}
