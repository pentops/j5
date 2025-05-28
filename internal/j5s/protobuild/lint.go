package protobuild

import (
	"context"
	"fmt"

	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
)

func (ps *PackageSet) LintFile(ctx context.Context, filename string, parsed *sourcedef_j5pb.SourceFile) (errpos.Errors, error) {
	pkgName, isLocal, err := ps.PackageForLocalFile(filename)
	if err != nil {
		return nil, fmt.Errorf("packageForFile %s: %w", filename, err)
	}
	if !isLocal {
		return nil, fmt.Errorf("file %s is not a local bundle file", filename)
	}

	pkg, err := ps.localPackageIO(ctx, pkgName)
	if err != nil {
		if ep, ok := errpos.AsErrors(err); ok {
			return ep, nil
		}
		return nil, fmt.Errorf("loadLocalPackage %s: %w", pkgName, err)
	}

	rb := newResolveBaton()
	err = ps.resolveDependencies(ctx, rb, pkg)
	if err != nil {
		return nil, fmt.Errorf("resolveDependencies for %s: %w", pkgName, err)
	}

	errs := errset.NewCollector()
	_, err = j5convert.SourceSummary(parsed, errs)
	if err != nil {
		if ep, ok := errpos.AsErrors(err); ok {
			return ep, nil
		}
		return nil, err
	}

	_, err = j5convert.ConvertJ5File(pkg, parsed)
	if err != nil {
		if ep, ok := errpos.AsErrors(err); ok {
			return ep, nil
		}
		return nil, fmt.Errorf("convertJ5File: %w", err)
	}

	return nil, nil
}
