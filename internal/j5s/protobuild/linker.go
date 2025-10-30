package protobuild

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/options"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/bufbuild/protocompile/sourceinfo"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"github.com/pentops/j5/internal/j5s/protobuild/psrc"
	"github.com/pentops/log.go/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type searchLinker struct {
	symbols  *linker.Symbols
	resolver psrc.Resolver
	errs     *errset.ErrCollector
}

func newLinker(src psrc.Resolver, symbols *linker.Symbols) *searchLinker {
	return &searchLinker{
		symbols:  symbols,
		errs:     &errset.ErrCollector{},
		resolver: src,
	}
}

func (ll *searchLinker) resolveAll(ctx context.Context, filenames []string) ([]*psrc.File, error) {
	files := make([]*psrc.File, 0, len(filenames))
	for _, filename := range filenames {
		file, err := ll._resolveFile(ctx, filename)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

func (ll *searchLinker) _resolveFile(ctx context.Context, filename string) (*psrc.File, error) {
	ctx = log.WithField(ctx, "askFilename", filename)
	result, err := ll.resolver.FindFileByPath(filename)
	if err != nil {
		return nil, err
	}

	err = ll.linkResult(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("linkResult: %w", err)
	}
	return result, nil
}

func (ll *searchLinker) linkResult(ctx context.Context, result *psrc.File) error {
	if result.Linked != nil {
		// result already linked
		return nil
	}

	log.WithField(ctx,
		"sourceFilename", result.Summary.SourceFilename,
		"sourceType", result.SourceType.String(),
	).Debug("link-new")

	dependencyFilenames, err := result.ListDependencies()
	if err != nil {
		return fmt.Errorf("listing dependencies: %w", err)
	}

	dependencies, err := ll.resolveAll(ctx, dependencyFilenames)
	if err != nil {
		return fmt.Errorf("loading dependencies for %s: %w", result.Filename, err)
	}

	result.Dependencies = dependencies

	info := linkInfo{
		deps:    dependencies,
		symbols: ll.symbols,
		errs:    ll.errs.Handler(),
	}

	linked, err := linkResult(result, info)
	if err != nil {
		info.debugState(os.Stderr)
		return err
	}

	result.Linked = linked

	err = ll.symbols.Import(linked, ll.errs.Handler())
	if err != nil {
		info.debugState(os.Stderr)
		return fmt.Errorf("importing new file into symbols: %w", err)
	}
	return nil
}

type linkInfo struct {
	deps    []*psrc.File //linker.Files
	symbols *linker.Symbols
	errs    *reporter.Handler
}

func (info *linkInfo) linkerDeps() linker.Files {
	return resultsToLinkerFiles(info.deps)
}

func resultsToLinkerFiles(results []*psrc.File) linker.Files {
	deps := make(linker.Files, 0, len(results))
	for _, dep := range results {
		if dep.Linked != nil {
			deps = append(deps, dep.Linked)
		}
	}
	return deps
}

func (info *linkInfo) debugState(ww io.Writer) {
	fmt.Fprintln(ww, "Linker State:")
	fmt.Fprintln(ww, "  Dependencies:")
	for _, dep := range info.deps {
		fmt.Fprintf(ww, "   - %s (%s)\n", dep.Filename, dep.SourceType.String())
	}
}

func linkResult(result *psrc.File, ll linkInfo) (linker.File, error) {
	// REFACTOR: The types should be an interface implementation
	if result.Refl != nil {
		file, err := _linkReflection(ll, result.Refl)
		if err != nil {
			return nil, fmt.Errorf("linking Refl: %w", err)
		}
		return file, nil
	}

	if result.Desc != nil {
		file, err := _linkDescriptorProto(ll, result.Desc)
		if err != nil {
			return nil, fmt.Errorf("linking Desc: %w", err)
		}
		return file, nil
	}

	if result.ParseResult != nil {
		file, err := _linkParserResult(ll, *result.ParseResult)
		if err != nil {
			return nil, fmt.Errorf("linking ParseResult: %w", err)
		}
		return file, nil
	}

	return nil, fmt.Errorf("search result type not unknown")
}

func _linkParserResult(ll linkInfo, result parser.Result) (linker.File, error) {
	linked, err := linker.Link(result, ll.linkerDeps(), ll.symbols, ll.errs)
	if err != nil {
		return nil, fmt.Errorf("linking using protocompile linker: %w", err)
	}

	optsIndex, err := options.InterpretOptions(linked, ll.errs)
	if err != nil {
		return nil, err
	}

	astNode := result.AST()
	sourceInfo := sourceinfo.GenerateSourceInfo(astNode, optsIndex, sourceinfo.WithExtraComments())
	linked.FileDescriptorProto().SourceCodeInfo = sourceInfo

	linked.PopulateSourceCodeInfo()
	linked.CheckForUnusedImports(ll.errs)

	return linked, nil
}

func _linkReflection(ll linkInfo, refl protoreflect.FileDescriptor) (linker.File, error) {
	file, err := linker.NewFile(refl, ll.linkerDeps())
	if err != nil {
		return nil, fmt.Errorf("creating file from refl: %w", err)
	}

	return file, nil
}

func _linkDescriptorProto(ll linkInfo, desc *descriptorpb.FileDescriptorProto) (linker.File, error) {
	result := parser.ResultWithoutAST(desc)

	linked, err := linker.Link(result, ll.linkerDeps(), ll.symbols, ll.errs)
	if err != nil {
		return nil, fmt.Errorf("link: %w", err)
	}

	_, err = options.InterpretOptions(linked, ll.errs)
	if err != nil {
		return nil, fmt.Errorf("options: %w", err)
	}

	linked.FileDescriptorProto().SourceCodeInfo = desc.SourceCodeInfo

	err = markExtensionImportsUsed(linked)
	if err != nil {
		return nil, fmt.Errorf("ext-used: %w", err)
	}

	linked.PopulateSourceCodeInfo()
	linked.CheckForUnusedImports(ll.errs)

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
			outerErr = fmt.Errorf("resolve extension %s: %w", name, err)
			return false
		}
		return true
	})

	return outerErr
}
