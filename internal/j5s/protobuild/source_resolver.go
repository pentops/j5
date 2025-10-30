package protobuild

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/protocompile/parser"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/j5/internal/j5s/j5parse"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/log.go/log"
)

type SourceFile struct {
	Summary *j5convert.FileSummary

	Warnings []*errpos.Err

	// Oneof:
	J5File    *sourcedef_j5pb.SourceFile
	ProtoFile *parser.Result
}

// toSearchResults converts the source file to the protocompile searchResults it
// produces. Proto files are 1:1 with Search Results, but one J5S file can
// produce multiple Search Results
func (sf *SourceFile) toSearchResults(typeResolver j5convert.TypeResolver) ([]*psrc.File, error) {
	if sf.ProtoFile != nil {
		return []*psrc.File{{
			SourceType:  psrc.LocalProtoSource,
			Filename:    sf.Summary.SourceFilename,
			Summary:     sf.Summary,
			ParseResult: sf.ProtoFile,
		}}, nil
	}

	if sf.J5File != nil {
		descs, err := j5convert.ConvertJ5File(typeResolver, sf.J5File)
		if err != nil {
			return nil, fmt.Errorf("convertJ5File %s: %w", sf.Summary.SourceFilename, err)
		}

		files := make([]*psrc.File, 0, len(descs))
		for _, desc := range descs {
			files = append(files, &psrc.File{
				Filename:   desc.GetName(),
				Summary:    sf.Summary,
				Desc:       desc,
				SourceType: psrc.LocalJ5Source,
			})
		}
		return files, nil
	}
	return nil, fmt.Errorf("source file %s has no result and is not j5s", sf.Summary.SourceFilename)
}

type LocalFileSource interface {
	GetLocalFile(context.Context, string) ([]byte, error)
	ProseFiles(pkgName string) ([]*source_j5pb.ProseFile, error)
	ListPackages() []string
	ListSourceFiles(ctx context.Context, pkgName string) ([]string, error)
}

type sourceResolver struct {
	bundleFiles       LocalFileSource
	j5Parser          *j5parse.Parser
	localPrefixes     []string
	localPackageNames map[string]struct{}
}

func NewSourceResolver(localFiles LocalFileSource) (*sourceResolver, error) {
	packages := localFiles.ListPackages()

	localPackageNames := map[string]struct{}{}
	localPrefixes := make([]string, len(packages))
	for i, p := range packages {
		s := strings.ReplaceAll(p, ".", "/")
		localPrefixes[i] = s + "/"
		localPackageNames[p] = struct{}{}
	}

	j5Parser, err := j5parse.NewParser()
	if err != nil {
		return nil, err
	}

	sr := &sourceResolver{
		j5Parser: j5Parser,

		bundleFiles:       localFiles,
		localPackageNames: localPackageNames,
		localPrefixes:     localPrefixes,
	}

	return sr, nil
}

func (sr *sourceResolver) ListPackages() []string {
	return sr.bundleFiles.ListPackages()
}

func (sr *sourceResolver) PackageProseFiles(pkgName string) ([]*source_j5pb.ProseFile, error) {
	return sr.bundleFiles.ProseFiles(pkgName)
}

func hasAPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func (sr *sourceResolver) PackageForFile(filename string) (string, bool, error) {
	if !hasAPrefix(filename, sr.localPrefixes) {
		// not a local file, not in scope.
		return "", false, nil
	}

	pkg, _, err := j5convert.SplitPackageFromFilename(filename)
	if err != nil {
		return "", false, err
	}
	return pkg, true, nil
}

func (sr *sourceResolver) IsLocalPackage(pkgName string) bool {
	_, ok := sr.localPackageNames[pkgName]
	return ok
}

func (sr *sourceResolver) listPackageFiles(ctx context.Context, pkgName string) ([]string, error) {
	root := strings.ReplaceAll(pkgName, ".", "/")

	files, err := sr.bundleFiles.ListSourceFiles(ctx, root)
	if err != nil {
		return nil, err
	}
	filtered := make([]string, 0)
	for _, f := range files {
		if strings.HasSuffix(f, ".j5s.proto") {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered, nil
}

func (sr *sourceResolver) getFile(ctx context.Context, sourceFilename string) (*SourceFile, error) {
	log.WithField(ctx, "sourceFilename", sourceFilename).Debug("read local source file")

	data, err := sr.bundleFiles.GetLocalFile(ctx, sourceFilename)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(sourceFilename, ".j5s") {
		return sr.parseJ5s(sourceFilename, data)
	}

	if strings.HasSuffix(sourceFilename, ".proto") {
		return sr.parseProto(sourceFilename, data)
	}

	return nil, fmt.Errorf("unsupported file type: %s", sourceFilename)
}

func (sr *sourceResolver) PackageSourceFiles(ctx context.Context, packageName string) ([]*SourceFile, error) {
	fileNames, err := sr.listPackageFiles(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("package files for (local) %s: %w", packageName, err)
	}

	files := make([]*SourceFile, 0, len(fileNames))
	for _, filename := range fileNames {
		file, err := sr.getFile(ctx, filename)
		if err != nil {
			return nil, fmt.Errorf("GetLocalFile %s: %w", filename, err)
		}
		files = append(files, file)
	}
	return files, nil

}

func (sr *sourceResolver) parseJ5s(sourceFilename string, data []byte) (*SourceFile, error) {
	sourceFile, err := sr.j5Parser.ParseFile(sourceFilename, string(data))
	if err != nil {
		if ep, ok := errpos.AsErrors(err); ok {
			return nil, ep.AsErrorsWithSource(sourceFilename, string(data))
		}
		return nil, err
	}

	errs := errset.NewCollector()
	summary, err := j5convert.SourceSummary(sourceFile, errs)
	if err != nil {
		if ep, ok := errpos.AsErrors(err); ok {
			return nil, ep.AsErrorsWithSource(sourceFilename, string(data))
		}
		return nil, err
	}

	return &SourceFile{
		Summary:  summary,
		Warnings: errs.Warnings,
		J5File:   sourceFile,
	}, nil
}

func (sr *sourceResolver) parseProto(filename string, data []byte) (*SourceFile, error) {
	errs := errset.NewCollector()
	fileNode, err := parser.Parse(filename, bytes.NewReader(data), errs.Handler())
	if err != nil {
		return nil, err
	}

	result, err := parser.ResultFromAST(fileNode, true, errs.Handler())
	if err != nil {
		return nil, err
	}

	summary, err := psrc.SummaryFromDescriptor(result.FileDescriptorProto(), errs)
	if err != nil {
		return nil, err
	}

	return &SourceFile{
		Summary:   summary,
		Warnings:  errs.Warnings,
		ProtoFile: &result,
	}, nil
}
