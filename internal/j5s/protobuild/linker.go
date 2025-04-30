package protobuild

import (
	"context"
	"fmt"

	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/options"
	"github.com/bufbuild/protocompile/parser"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type SourceType int

const (
	LocalJ5Source SourceType = iota
	LocalProtoSource
	BuiltInProtoSource
	ExternalProtoSource
)

var sourceTypeNames = map[SourceType]string{
	LocalJ5Source:       "Local J5",
	LocalProtoSource:    "Local Proto",
	BuiltInProtoSource:  "Built-in Proto",
	ExternalProtoSource: "External Proto",
}

func (st SourceType) String() string {
	if name, ok := sourceTypeNames[st]; ok {
		return name
	}
	return fmt.Sprintf("Unknown SourceType %d", st)
}

type SearchResult struct {
	Filename string
	Summary  *j5convert.FileSummary

	SourceType SourceType

	// Results are checked in lexical order
	Linked      linker.File
	Refl        protoreflect.FileDescriptor
	Desc        *descriptorpb.FileDescriptorProto
	ParseResult *parser.Result
}

type searchLinker struct {
	symbols  *linker.Symbols
	resolver fileResolver // usually *PackageSet
	errs     *ErrCollector
}

func newLinker(src fileResolver, symbols *linker.Symbols) *searchLinker {
	return &searchLinker{
		symbols:  symbols,
		errs:     &ErrCollector{},
		resolver: src,
	}
}

func (ll *searchLinker) resolveAll(ctx context.Context, filenames []string) (linker.Files, error) {
	files := make(linker.Files, 0, len(filenames))
	for _, filename := range filenames {
		file, err := ll._resolveFile(ctx, filename)
		if err != nil {
			return nil, fmt.Errorf("linker, resolve file %s: %w", filename, err)
		}

		files = append(files, file)
	}

	return files, nil
}

func (ll *searchLinker) _resolveFile(ctx context.Context, filename string) (linker.File, error) {
	ctx = log.WithField(ctx, "askFilename", filename)
	result, err := ll.resolver.findFileByPath(filename)
	if err != nil {
		return nil, fmt.Errorf("findFileByPath: %w", err)
	}

	return ll.linkResult(ctx, result)
}

func (ll *searchLinker) linkResult(ctx context.Context, result *SearchResult) (linker.File, error) {
	if result.Linked != nil {
		return result.Linked, nil
	}
	log.WithField(ctx,
		"sourceFilename", result.Summary.SourceFilename,
		"sourceType", result.SourceType.String(),
	).Debug("link-new")

	linked, err := ll._linkNewResult(ctx, result)
	if err != nil {
		return nil, err
	}

	result.Linked = linked

	err = ll.symbols.Import(linked, ll.errs.Handler())
	if err != nil {
		return nil, fmt.Errorf("importing new file into symbols: %w", err)
	}
	return linked, nil

}

func (ll *searchLinker) _linkNewResult(ctx context.Context, result *SearchResult) (linker.File, error) {
	if result.Refl != nil {
		file, err := ll._linkReflection(ctx, result.Refl)
		if err != nil {
			return nil, fmt.Errorf("linking Refl: %w", err)
		}
		return file, nil
	}

	if result.Desc != nil {
		file, err := ll._linkDescriptorProto(ctx, result.Desc)
		if err != nil {
			return nil, fmt.Errorf("linking Desc: %w", err)
		}
		return file, nil
	}

	if result.ParseResult != nil {
		file, err := ll._linkParserResult(ctx, *result.ParseResult)
		if err != nil {
			return nil, fmt.Errorf("linking ParseResult: %w", err)
		}
		return file, nil
	}

	return nil, fmt.Errorf("search result type not unknown")
}

func (ll *searchLinker) _linkParserResult(ctx context.Context, result parser.Result) (linker.File, error) {
	desc := result.FileDescriptorProto()
	deps, err := ll.resolveAll(ctx, desc.Dependency)
	if err != nil {
		return nil, fmt.Errorf("loading dependencies for %s: %w", desc.GetName(), err)
	}

	linked, err := linker.Link(result, deps, ll.symbols, ll.errs.Handler())
	if err != nil {
		return nil, fmt.Errorf("linking using protocompile linker: %w", err)
	}

	_, err = options.InterpretOptions(linked, ll.errs.Handler())
	if err != nil {
		return nil, err
	}

	linked.CheckForUnusedImports(ll.errs.Handler())
	return linked, nil
}

func (ll *searchLinker) _linkReflection(ctx context.Context, refl protoreflect.FileDescriptor) (linker.File, error) {
	imports := refl.Imports()
	importFilenames := make([]string, 0, imports.Len())
	for i := range imports.Len() {
		imports := imports.Get(i)
		importFilenames = append(importFilenames, imports.Path())
	}
	log.WithField(ctx, "importFilenames", importFilenames).Debug("linking reflection file")
	deps, err := ll.resolveAll(ctx, importFilenames)
	if err != nil {
		return nil, fmt.Errorf("loading dependencies: %w", err)
	}
	file, err := linker.NewFile(refl, deps)
	if err != nil {
		return nil, fmt.Errorf("creating file from refl: %w", err)
	}

	return file, nil
}

func (ll *searchLinker) _linkDescriptorProto(ctx context.Context, desc *descriptorpb.FileDescriptorProto) (linker.File, error) {
	deps, err := ll.resolveAll(ctx, desc.Dependency)
	if err != nil {
		return nil, fmt.Errorf("loading dependencies: %w", err)
	}

	result := parser.ResultWithoutAST(desc)
	log.WithField(ctx, "descName", desc.GetName()).Debug("descriptorToFile")

	linked, err := linker.Link(result, deps, ll.symbols, ll.errs.Handler())
	if err != nil {
		return nil, fmt.Errorf("descriptorToFile, link: %w", err)
	}

	_, err = options.InterpretOptions(linked, ll.errs.Handler())
	if err != nil {
		return nil, err
	}

	err = markExtensionImportsUsed(linked)
	if err != nil {
		return nil, err
	}

	linked.PopulateSourceCodeInfo()

	linked.CheckForUnusedImports(ll.errs.Handler())
	return linked, nil
}

// hacks the underlying linker to mark the imports which are used in extensions
// as 'used' to prevent a compiler warning.
func markExtensionImportsUsed(file linker.File) error {
	resolver := linker.ResolverFromFile(file)
	messages := file.Messages()
	for i := range messages.Len() {
		message := messages.Get(i)
		err := markMessageExtensionImportsUsed(resolver, message)
		if err != nil {
			return err
		}
	}

	services := file.Services()
	for i := range services.Len() {
		service := services.Get(i)
		err := markOptionImportsUsed(resolver, service.Options())
		if err != nil {
			return err
		}

		methods := service.Methods()
		for j := range methods.Len() {
			method := methods.Get(j)
			err = markOptionImportsUsed(resolver, method.Options())
			if err != nil {
				return err
			}
		}
	}

	enums := file.Enums()
	for i := range enums.Len() {
		enum := enums.Get(i)
		err := markOptionImportsUsed(resolver, enum.Options())
		if err != nil {
			return err
		}

		values := enum.Values()
		for j := range values.Len() {
			value := values.Get(j)
			err = markOptionImportsUsed(resolver, value.Options())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func markMessageExtensionImportsUsed(resolver linker.Resolver, message protoreflect.MessageDescriptor) error {
	err := markOptionImportsUsed(resolver, message.Options())
	if err != nil {
		return err
	}

	fields := message.Fields()
	for j := range fields.Len() {
		field := fields.Get(j)
		err = markOptionImportsUsed(resolver, field.Options())
		if err != nil {
			return err
		}

	}

	oneofs := message.Oneofs()
	for j := range oneofs.Len() {
		oneof := oneofs.Get(j)
		err = markOptionImportsUsed(resolver, oneof.Options())
		if err != nil {
			return err
		}
	}

	nested := message.Messages()
	for j := range nested.Len() {
		nestedMessage := nested.Get(j)
		err = markMessageExtensionImportsUsed(resolver, nestedMessage)
		if err != nil {
			return err
		}

	}
	return nil
}

func markOptionImportsUsed(resolver linker.Resolver, opts proto.Message) error {
	var outerErr error

	proto.RangeExtensions(opts, func(ext protoreflect.ExtensionType, value any) bool {
		td := ext.TypeDescriptor()
		name := td.FullName()
		_, err := resolver.FindExtensionByName(name)
		if err != nil {
			outerErr = err
			return false
		}
		return true
	})

	return outerErr
}
