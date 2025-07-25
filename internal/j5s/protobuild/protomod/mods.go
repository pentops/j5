package protomod

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/structure"
	"google.golang.org/protobuf/types/descriptorpb"
)

// MutateImageWithMods applies a set of image mods to a source image.
// This is a minor foot gun, hence the name.
// The caller is responsible for ensuring that the image is fresh and not reused
// if that isn't desired. Adding a clone protection is too expensive for the
// likely image sizes, and all existing use cases only use the image once.
func MutateImageWithMods(img *source_j5pb.SourceImage, mods []*config_j5pb.ProtoMod) error {
	ms, err := BuildMutators(mods)
	if err != nil {
		return err
	}
	return ms.MutateImage(img)
}

type ModSet []mutator

func (ms ModSet) MutateImage(img *source_j5pb.SourceImage) error {
	return runMods(img, ms)
}

func (ms ModSet) MutateFile(file *descriptorpb.FileDescriptorProto) error {
	for _, m := range ms {
		if err := m.MutateFile(file); err != nil {
			return fmt.Errorf("mutating file %q: %w", *file.Name, err)
		}
	}
	return nil
}

func BuildMutators(mods []*config_j5pb.ProtoMod) (ModSet, error) {
	var mutators []mutator

	for _, mod := range mods {
		switch mod := mod.Type.(type) {
		case *config_j5pb.ProtoMod_GoPackageNames_:
			mutators = append(mutators, newGoPackageNames(mod.GoPackageNames))
		default:
			return nil, fmt.Errorf("unknown image mod type: %T", mod)
		}
	}

	return mutators, nil
}

func runMods(img *source_j5pb.SourceImage, mods []mutator) error {
	isSource := make(map[string]bool)
	for _, file := range img.SourceFilenames {
		isSource[file] = true
	}

	for _, file := range img.File {
		if !isSource[*file.Name] {
			continue
		}
		for _, m := range mods {
			if err := m.MutateFile(file); err != nil {
				return fmt.Errorf("mutating file %q: %w", *file.Name, err)
			}
		}
	}
	return nil
}

func firstNonEmptyString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

type mutator interface {
	MutateFile(file *descriptorpb.FileDescriptorProto) error
}

type goPackageNames struct {
	suffixes     map[string]string
	trimPrefixes []string
	prefix       string
}

func newGoPackageNames(mod *config_j5pb.ProtoMod_GoPackageNames) *goPackageNames {
	if mod.Suffixes == nil {
		mod.Suffixes = map[string]string{}
	}

	suffixes := map[string]string{
		"":        firstNonEmptyString(mod.BaseSuffix, mod.Suffixes[""], "_pb"),
		"service": firstNonEmptyString(mod.ServiceSuffix, mod.Suffixes["service"], "_spb"),
		"topic":   firstNonEmptyString(mod.TopicSuffix, mod.Suffixes["topic"], "_tpb"),
	}

	return &goPackageNames{
		suffixes:     suffixes,
		trimPrefixes: mod.TrimPrefixes,
		prefix:       mod.Prefix,
	}
}

func (mod goPackageNames) MutateFile(file *descriptorpb.FileDescriptorProto) error {

	if file.Options == nil {
		file.Options = &descriptorpb.FileOptions{}
	} else if file.Options.GoPackage != nil {
		// already set, nothing to do.
		return nil
	}

	pkg := *file.Package

	for _, prefix := range mod.trimPrefixes {
		if strings.HasPrefix(pkg, prefix) {
			pkg = pkg[len(prefix):]
			pkg = strings.TrimPrefix(pkg, ".")
			break
		}
	}
	basePkg, subPkg, err := structure.SplitPackageParts(pkg)
	if err != nil {
		return fmt.Errorf("splitting package %q: %w", pkg, err)
	}

	suffix := "_pb"
	if subPkg != nil {
		sub := *subPkg
		suffix = fmt.Sprintf("_%spb", sub[0:1])
	}
	subPkgName := ""
	if subPkg != nil {
		subPkgName = *subPkg
	}

	if modSuffix, ok := mod.suffixes[subPkgName]; ok {
		suffix = modSuffix
	}
	pkgParts := strings.Split(basePkg, ".")
	baseName := strings.Join(pkgParts, "/")
	namePart := pkgParts[len(pkgParts)-2]
	lastPart := fmt.Sprintf("%s%s", namePart, suffix)
	fullName := strings.Join([]string{mod.prefix, baseName, lastPart}, "/")
	file.Options.GoPackage = &fullName
	return nil
}
