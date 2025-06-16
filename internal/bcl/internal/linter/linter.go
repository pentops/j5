package linter

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/internal/bcl"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/bcl/internal/parser"
	"github.com/pentops/log.go/log"
	"go.lsp.dev/protocol"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type FileFactory func(filename string) protoreflect.Message
type OnChange func(ctx context.Context, filename string, sourceLocs *bcl_j5pb.SourceLocation, msg protoreflect.Message) error

type Linter struct {
	parser      *bcl.Parser
	rootDir     string
	fileFactory FileFactory
	onChange    OnChange
}

func New(parser *bcl.Parser, fileFactory FileFactory, validate OnChange, rootDir string) *Linter {
	return &Linter{
		parser:      parser,
		fileFactory: fileFactory,
		onChange:    validate,
		rootDir:     rootDir,
	}
}

func NewGeneric(rootDir string) *Linter {
	return &Linter{
		rootDir: rootDir,
	}
}

func errorToDiagnostics(ctx context.Context, relFilename string, mainError error) ([]protocol.Diagnostic, error) {
	if mainError == nil {
		log.Debug(ctx, "No errors")
		return []protocol.Diagnostic{}, nil
	}
	locErr, ok := errpos.AsErrors(mainError)
	if !ok {
		withSourceErr, ok := errpos.AsErrorsWithSource(mainError) // try to convert to ErrorsWithSource
		if !ok {
			log.WithError(ctx, mainError).Error("Error not errpos.Errors")
			return nil, mainError
		}
		locErr = withSourceErr.Errors
	}

	diagnostics := make([]protocol.Diagnostic, 0, len(locErr))

	for _, err := range locErr {
		log.WithFields(ctx,
			"pos", err.Pos.String(),
			"error", err.Err.Error(),
			"relFilename", relFilename,
		).Debug("Lint Diagnostic")
		if err.Pos == nil {
			continue
		}
		if err.Pos.Filename == nil {
			continue
		}
		if *err.Pos.Filename != relFilename {
			continue
		}

		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(err.Pos.Start.Line),
					Character: uint32(err.Pos.Start.Column),
				},
				End: protocol.Position{
					Line:      uint32(err.Pos.End.Line),
					Character: uint32(err.Pos.End.Column),
				},
			},
			Code:     ptr("LINT"),
			Message:  err.Err.Error(),
			Severity: protocol.DiagnosticSeverityError,
			Source:   "bcl",
		})
	}

	return diagnostics, nil

}
func (l *Linter) FileChanged(ctx context.Context, req *protocol.TextDocumentItem) ([]protocol.Diagnostic, error) {
	relFilename, err := filepath.Rel(l.rootDir, req.URI.Filename())
	if err != nil {
		return nil, err
	}

	// Step 1: Parse BCL
	tree, err := parser.ParseFile(relFilename, req.Text, false)
	if err != nil {
		log.WithError(ctx, err).Error("parser.ParseFile error")
		if ews, ok := errpos.AsErrorsWithSource(err); ok {
			return errorToDiagnostics(ctx, relFilename, ews)
		} else {
			return nil, fmt.Errorf("parse file not HadErrors - : %w", err)
		}
	}

	if l.fileFactory == nil || l.parser == nil {
		return nil, nil
	}

	// Step 2: Parse AST
	msg := l.fileFactory(req.URI.Filename())
	sourceLocs, err := l.parser.ParseAST(tree, msg)
	if err != nil {
		if ep, ok := errpos.AsErrors(err); ok {
			return errorToDiagnostics(ctx, relFilename, ep.AsErrorsWithSource(relFilename, req.Text))
		}
		return errorToDiagnostics(ctx, relFilename, err)
	}

	// Step 3: Validate
	if l.onChange == nil {
		return nil, nil
	}

	err = l.onChange(ctx, req.URI.Filename(), sourceLocs, msg)
	if err != nil {

		//err = errpos.AddSourceFile(err, req.URI.Filename(), req.Text)
		return errorToDiagnostics(ctx, relFilename, err)
	}

	log.Debug(ctx, "No errors")
	return []protocol.Diagnostic{}, nil

}

func ptr[T any](v T) *T {
	return &v
}
