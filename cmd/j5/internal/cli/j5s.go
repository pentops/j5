package cli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pentops/j5/internal/bcl"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/j5s/j5parse"
	"github.com/pentops/j5/internal/j5s/protoprint"
	"github.com/pentops/j5/internal/source"
	"github.com/pentops/j5/internal/source/resolver"
	"github.com/pentops/log.go/log"
	"github.com/pentops/runner/commander"
	"github.com/pmezard/go-difflib/difflib"
	"google.golang.org/protobuf/reflect/protodesc"
)

func j5sSet() *commander.CommandSet {
	genGroup := commander.NewCommandSet()
	genGroup.Add("fmt", commander.NewCommand(runJ5sFmt))
	genGroup.Add("lint", commander.NewCommand(runJ5sLint))
	genGroup.Add("genproto", commander.NewCommand(runJ5sGenProto))
	return genGroup
}

func runJ5sLint(ctx context.Context, cfg struct {
	Dir  string `flag:"dir" required:"false" description:"Source / working directory containing j5.yaml"`
	File string `flag:"file" required:"false" description:"Single file to format"`
}) error {

	imageResolver, err := resolver.NewEnvResolver()
	if err != nil {
		return err
	}

	if cfg.Dir == "" {
		cfg.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	fsRoot := os.DirFS(cfg.Dir)
	srcRoot, err := source.NewFSRepoRoot(ctx, fsRoot, imageResolver)
	if err != nil {
		return err
	}

	if cfg.File != "" {
		fullDir, err := filepath.Abs(cfg.Dir)
		if err != nil {
			return err
		}
		fileRel, err := filepath.Rel(fullDir, cfg.File)
		if err != nil {
			return err
		}
		return runJ5sLintFile(ctx, srcRoot, fileRel)
	}

	return runJ5sLintAll(ctx, srcRoot)
}

func runJ5sLintFile(ctx context.Context, srcRoot *source.RepoRoot, fileRel string) error {
	bundle, relToBundle, err := srcRoot.BundleForFile(fileRel)
	if err != nil {
		return err
	}

	compiler, err := bundle.Compiler(ctx, srcRoot)
	if err != nil {
		return err
	}

	data, err := fs.ReadFile(bundle.FS(), relToBundle)
	if err != nil {
		return err
	}

	sourceFile, err := j5parse.ParseFile(relToBundle, string(data))
	if err != nil {
		return err
	}

	lintErr, err := compiler.LintFile(ctx, relToBundle, sourceFile)
	if err != nil {
		return fmt.Errorf("root err: %w", err)
	}
	if lintErr == nil {
		fmt.Fprintln(os.Stderr, "No linting errors")
		return nil
	}
	withSource := lintErr.AsErrorsWithSource(relToBundle, string(data))
	fmt.Fprintln(os.Stderr, withSource.HumanString(2))
	return fmt.Errorf("linting failed")
}

func runJ5sLintAll(ctx context.Context, srcRoot *source.RepoRoot) error {
	bundles, externalDeps, err := srcRoot.LocalBundlesSorted(ctx)
	if err != nil {
		return err
	}

	for _, bundle := range bundles {

		ps, err := bundle.Compiler(ctx, srcRoot)
		if err != nil {
			return fmt.Errorf("compiler: %w", err)
		}

		allPackages := ps.ListLocalPackages()

		built, err := ps.BuildPackages(ctx, allPackages)
		if err != nil {
			if ep, ok := errpos.AsErrorsWithSource(err); ok {
				fmt.Printf("Linting errors in bundle %s\n", bundle.DebugName())
				fmt.Fprintln(os.Stderr, ep.ShortString())
			}
			return fmt.Errorf("unhandled: %w", err)
		}
		for _, pkg := range built {
			for _, file := range pkg.Proto {
				protoFile := protodesc.ToFileDescriptorProto(file.Linked)
				externalDeps[file.Linked.Path()] = protoFile
			}
		}
	}

	fmt.Fprintln(os.Stderr, "No linting errors")
	return nil
}

func runJ5sFmt(ctx context.Context, cfg struct {
	Dir   string `flag:"dir" required:"false" description:"Source / working directory containing j5.yaml and buf.lock.yaml"`
	File  string `flag:"file" required:"false" description:"Single file to format"`
	Write bool   `flag:"write" default:"false" desc:"Write fixes to files"`
	Check bool   `flag:"check" default:"false" desc:"Return a non-zero exit code if files need formatting"`
}) error {
	var outWriter *fileWriter

	if cfg.Check && cfg.Write {
		return fmt.Errorf("cannot specify both check and write")
	}

	checkFailed := false
	doFile := func(ctx context.Context, pathname string, data []byte) error {
		fixed, err := bcl.Fmt(pathname, string(data))
		if err != nil {
			return err
		}

		if cfg.Write {
			return outWriter.PutFile(ctx, pathname, []byte(fixed))
		}

		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			FromFile: pathname,
			B:        difflib.SplitLines(fixed),
			ToFile:   pathname,
			Context:  3,
		})

		if err != nil {
			return err
		}

		if diff != "" {
			fmt.Println(diff)

			if cfg.Check {
				checkFailed = true
			}
		}

		return nil
	}

	if cfg.File != "" {
		if cfg.Dir != "" {
			return fmt.Errorf("cannot specify both dir and file")
		}

		dir, pathname := path.Split(cfg.File)
		outWriter = &fileWriter{dir: dir}

		data, err := os.ReadFile(cfg.File)
		if err != nil {
			return err
		}

		err = doFile(ctx, pathname, data)
		if err != nil {
			return err
		}

		return nil
	}

	var err error
	if cfg.Dir == "" {
		cfg.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	outWriter = &fileWriter{dir: cfg.Dir}
	fsRoot := os.DirFS(cfg.Dir)

	err = runForJ5Files(ctx, fsRoot, doFile)
	if err != nil {
		return err
	}

	if checkFailed {
		return fmt.Errorf("one or more files need formatting")
	}

	return nil
}

func runForJ5Files(ctx context.Context, root fs.FS, doFile func(ctx context.Context, pathname string, data []byte) error) error {
	err := fs.WalkDir(root, ".", func(pathname string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if path.Ext(pathname) != ".j5s" {
			return nil
		}

		data, err := fs.ReadFile(root, pathname)
		if err != nil {
			return err
		}

		return doFile(ctx, pathname, data)
	})
	if err != nil {
		return err
	}

	return nil
}

type j5sGenProtoConfig struct {
	SourceConfig
	Verbose bool `flag:"verbose" env:"BCL_VERBOSE" default:"false" desc:"Verbose output"`
}

func runJ5sGenProto(ctx context.Context, cfg j5sGenProtoConfig) error {
	src, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	genComment := fmt.Sprintf("Generated by j5build %s. DO NOT EDIT", Version)

	err = cfg.EachBundle(ctx, func(bundle source.Bundle) error {

		ctx = log.WithField(ctx, "bundle", bundle.DebugName())
		log.Debug(ctx, "GenProto for Bundle")

		compiler, err := bundle.Compiler(ctx, src)
		if err != nil {
			return err
		}

		outWriter, err := cfg.FileWriterAt(ctx, bundle.DirInRepo())
		if err != nil {
			return fmt.Errorf("fw: %w", err)
		}

		err = outWriter.DeleteFilesMatching(ctx, func(name string) bool {
			// not using path.Ext because it returns .proto
			return strings.HasSuffix(name, ".j5s.proto")
		})
		if err != nil {
			return fmt.Errorf("clean: %w", err)
		}

		for _, pkg := range compiler.ListLocalPackages() {

			out, err := compiler.CompilePackage(ctx, pkg)
			if err != nil {
				return fmt.Errorf("compile package %q: %w", pkg, err)
			}

			for _, file := range out.Proto {
				filename := file.Linked.Path()
				if !strings.HasSuffix(filename, ".j5s.proto") {
					continue
				}

				out, err := protoprint.PrintFile(ctx, file.Linked, genComment)
				if err != nil {
					log.WithFields(ctx, map[string]any{
						"error":    err.Error(),
						"filename": file.Filename,
					}).Error("Error printing file")
					return err
				}

				err = outWriter.PutFile(ctx, filename, []byte(out))
				if err != nil {
					return err
				}

			}

		}

		return nil
	})

	if err == nil {
		return nil
	}

	e, ok := errpos.AsErrorsWithSource(err)
	if !ok {
		return err
	}
	fmt.Fprintln(os.Stderr, e.HumanString(3))

	return err
}
