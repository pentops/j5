package protobuild

import (
	"context"
	"fmt"

	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/log.go/log"
)

func (ps *PackageSet) LintFile(ctx context.Context, filename string, fileData string) (*errpos.ErrorsWithSource, error) {
	pkgName, isLocal, err := ps.PackageForLocalFile(filename)
	if err != nil {
		return nil, fmt.Errorf("packageForFile %s: %w", filename, err)
	}
	if !isLocal {
		return nil, fmt.Errorf("file %s is not a local bundle file", filename)
	}

	// LoadLocalPackage parses both BCL and Proto files, but does not link them
	pkg, err := ps.LoadLocalPackage(ctx, pkgName)
	if err != nil {
		if ep, ok := errpos.AsErrorsWithSource(err); ok {
			return ep, nil
		}
		return nil, fmt.Errorf("loadLocalPackage %s: %w", pkgName, err)
	}

	var srcFile *SourceFile
	for _, file := range pkg.SourceFiles {
		if file.Summary.SourceFilename == filename {
			srcFile = file
			break
		}
	}
	if srcFile == nil {
		return nil, fmt.Errorf("source file %s not found in package %s", filename, pkgName)
	}

	linker := newLinker(ps, ps.symbols)

	results, err := srcFile.toSearchResults(pkg)
	if err != nil {
		return nil, fmt.Errorf("searchResults %s: %w", srcFile.Summary.SourceFilename, err)
	}

	for _, result := range results {
		pkg.Files[result.Filename] = result
		ctx := log.WithField(ctx, "linking", result.Filename)
		log.Info(ctx, "linking for lint")
		err := linker.linkResult(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("linking j5 file %s: %w", filename, err)
		}
		if linker.errs.HasAny() {
			return convertLintErrors(filename, "", linker.errs)
		}
	}

	return convertLintErrors(filename, fileData, &ErrCollector{
		Warnings: srcFile.Warnings,
	})
}

func (ps *PackageSet) LintAll(ctx context.Context) (*errpos.ErrorsWithSource, error) {
	allPackages := ps.ListLocalPackages()

	allErrs := &ErrCollector{}

	for _, pkgName := range allPackages {
		// LoadLocalPackage parses both BCL and Proto files, but does not fully link.
		pkg, err := ps.LoadLocalPackage(ctx, pkgName)
		if err != nil {
			if ep, ok := errpos.AsErrorsWithSource(err); ok {
				return ep, nil
			}
			return nil, fmt.Errorf("loadLocalPackage %s: %w", pkgName, err)
		}

		for _, src := range pkg.SourceFiles {
			allErrs.Warnings = append(allErrs.Warnings, src.Warnings...)
		}

		for _, file := range pkg.Files {
			linker := newLinker(ps, ps.symbols)
			err := linker.linkResult(ctx, file)
			if err != nil {
				return nil, fmt.Errorf("linking file %s: %w", file.Summary.SourceFilename, err)
			}
			if linker.errs.HasAny() {
				data, err := ps.GetLocalFileContent(ctx, file.Summary.SourceFilename)
				if err != nil {
					return nil, fmt.Errorf("getRawFile %s: %w", file.Summary.SourceFilename, err)
				}
				return convertLintErrors(file.Summary.SourceFilename, data, linker.errs)
			}
		}

	}
	return convertLintErrors("", "", allErrs)

}

func convertLintErrors(filename string, fileData string, errs *ErrCollector) (*errpos.ErrorsWithSource, error) {

	errors := errpos.Errors{}
	for _, err := range errs.Errors {
		errors = append(errors, err)
	}
	for _, err := range errs.Warnings {
		errors = append(errors, err)
	}

	if len(errors) == 0 {
		return nil, nil
	}

	ws := errpos.AddSourceFile(errors, filename, fileData)
	as, ok := errpos.AsErrorsWithSource(ws)
	if !ok {
		return nil, fmt.Errorf("error not valid for source: (%T) %w", ws, ws)
	}
	return as, nil
}
