package j5convert

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
)

type PackageSummary struct {
	Exports map[string]*TypeRef
	Files   []*FileSummary
}

type FileSummary struct {
	SourceFilename   string
	Package          string
	Exports          map[string]*TypeRef
	FileDependencies []string
	TypeDependencies []*schema_j5pb.Ref

	ProducesFiles []string
}

// TypeRef is the summary of an exported type
type TypeRef struct {
	Package  string
	Name     string
	File     string
	Position *errpos.Position

	// Oneof
	Enum      *EnumRef
	Object    *ObjectRef
	Oneof     *OneofRef
	Polymorph *PolymorphRef
}

func (typeRef TypeRef) protoTypeName() *string {
	if typeRef.Package == "" {
		return gl.Ptr(typeRef.Name)
	}
	return gl.Ptr(fmt.Sprintf(".%s.%s", typeRef.Package, typeRef.Name))
}

func (typeRef TypeRef) typeName() string {
	if typeRef.Enum != nil {
		return "enum"
	} else if typeRef.Object != nil {
		return "object"
	} else if typeRef.Oneof != nil {
		return "oneof"
	} else if typeRef.Polymorph != nil {
		return "polymorph"
	}
	return "unknown"
}

func (typeRef TypeRef) debugName() string {
	return fmt.Sprintf("%s.%s[%s]", typeRef.Package, typeRef.Name, typeRef.typeName())
}

type ObjectRef struct {
}

type OneofRef struct {
}

type PolymorphRef struct {
	Members  []string
	Includes []string
}

// EnumRef is the summary of an enum definition
type EnumRef struct {
	Prefix string
	ValMap map[string]int32
}

func (er *EnumRef) mapValues(vals []string) ([]int32, error) {
	out := make([]int32, len(vals))
	for idx, in := range vals {
		if !strings.HasPrefix(in, er.Prefix) {
			in = er.Prefix + in
		}
		val, ok := er.ValMap[in]
		if !ok {
			return nil, fmt.Errorf("enum value %q not found", in)
		}
		out[idx] = val
	}
	return out, nil
}

type TypeResolver interface {
	ResolveType(pkg string, name string) (*TypeRef, error)
}
type ErrCollector interface {
	WarnPos(pos *errpos.Position, err error)
}

// SourceSummary collects the exports and imports for a j5 source file
func SourceSummary(sourceFile *sourcedef_j5pb.SourceFile, ec ErrCollector) (*FileSummary, error) {

	cc := &summaryWalker{}
	err := cc.collectFileRefs(sourceFile)
	if err != nil {
		return nil, err
	}

	importPath := sourceFile.Path + ".proto"

	allFilenames := []string{importPath}
	for _, subPackage := range cc.subPackageFiles {
		allFilenames = append(allFilenames, subPackageFileName(sourceFile.Path, subPackage))
	}

	fs := &FileSummary{
		SourceFilename: sourceFile.Path,
		Package:        sourceFile.Package.Name,
		Exports:        make(map[string]*TypeRef),
		ProducesFiles:  allFilenames,
	}

	importMap, err := j5Imports(sourceFile)
	if err != nil {
		return nil, err
	}

	for _, refSrc := range cc.refs {
		expanded := importMap.expand(refSrc.Ref)
		if expanded == nil {
			err := fmt.Errorf("package %q not imported (for schema %s)", refSrc.Package, refSrc.Schema)
			err = errpos.AddContext(err, strings.Join(refSrc.Source.Path, "."))
			err = errpos.AddPosition(err, refSrc.Source.GetPos())
			return nil, err
		}

		fs.TypeDependencies = append(fs.TypeDependencies, expanded.ref)
	}

	for _, export := range cc.exports {
		export.Package = sourceFile.Package.Name
		export.File = importPath
		fs.Exports[export.Name] = export
	}

	for _, ref := range importMap.vals {
		if ref.used {
			continue
		}
		err := fmt.Errorf("import %q not used", ref.fullPath)
		var pos *errpos.Position
		if ref.source != nil {
			pos = &errpos.Position{
				Start: errpos.Point{
					Line:   int(ref.source.StartLine),
					Column: int(ref.source.StartColumn),
				},
				End: errpos.Point{
					Line:   int(ref.source.EndLine),
					Column: int(ref.source.EndColumn),
				},
			}
		}
		ec.WarnPos(pos, err)
	}

	return fs, nil

}

type summaryWalker struct {
	exports         []*TypeRef
	refs            []*sourcewalk.RefNode
	subPackageFiles []string
}

func (c *summaryWalker) includeSubFile(subPackage string) {
	if slices.Contains(c.subPackageFiles, subPackage) {
		return
	}
	c.subPackageFiles = append(c.subPackageFiles, subPackage)
}

func (c *summaryWalker) addExport(ref *TypeRef) {
	c.exports = append(c.exports, ref)
}

func (c *summaryWalker) addRef(ref *sourcewalk.RefNode) {
	c.refs = append(c.refs, ref)
}

func (cc *summaryWalker) collectFileRefs(sourceFile *sourcedef_j5pb.SourceFile) error {
	file := sourcewalk.NewRoot(sourceFile)

	visitor := &sourcewalk.DefaultVisitor{
		Property: func(node *sourcewalk.PropertyNode) error {
			if node.Field.Ref != nil {
				cc.addRef(node.Field.Ref)
			} else if node.Field.Items != nil && node.Field.Items.Ref != nil {
				cc.addRef(node.Field.Items.Ref)
			}
			return nil
		},
		Object: func(node *sourcewalk.ObjectNode) error {
			cc.addExport(objectTypeRef(node))
			return nil
		},
		Oneof: func(node *sourcewalk.OneofNode) error {
			cc.addExport(oneofTypeRef(node))
			return nil
		},
		Enum: func(node *sourcewalk.EnumNode) error {
			valMap := make(map[string]int32)
			for _, value := range node.Schema.Options {
				valMap[node.Schema.Prefix+value.Name] = value.Number
			}
			cc.addExport(enumTypeRef(node))
			return nil
		},
		Polymorph: func(node *sourcewalk.PolymorphNode) error {
			cc.addExport(polymorphTypeRef(node))
			return nil
		},
		Service: func(node *sourcewalk.ServiceNode) error {
			cc.includeSubFile("service")
			return nil
		},
		Topic: func(node *sourcewalk.TopicNode) error {
			cc.includeSubFile("topic")
			return nil
		},
	}

	return file.RangeRootElements(visitor)

}

func objectTypeRef(node *sourcewalk.ObjectNode) *TypeRef {
	return &TypeRef{
		Name:     node.NameInPackage(),
		Position: gl.Ptr(node.Source.GetPos()),
		Object:   &ObjectRef{},
	}
}

func oneofTypeRef(node *sourcewalk.OneofNode) *TypeRef {
	return &TypeRef{
		Name:     node.NameInPackage(),
		Position: gl.Ptr(node.Source.GetPos()),
		Oneof:    &OneofRef{},
	}
}

func enumTypeRef(node *sourcewalk.EnumNode) *TypeRef {
	valMap := make(map[string]int32)
	for _, value := range node.Options {
		valMap[value.Name] = value.Number
	}
	return &TypeRef{
		Name:     node.NameInPackage(),
		Position: gl.Ptr(node.Source.GetPos()),

		Enum: &EnumRef{
			Prefix: node.Schema.Prefix,
			ValMap: valMap,
		},
	}
}

func polymorphTypeRef(node *sourcewalk.PolymorphNode) *TypeRef {
	return &TypeRef{
		Name:     node.NameInPackage(),
		Position: gl.Ptr(node.Source.GetPos()),
		Polymorph: &PolymorphRef{
			Members:  node.Members,
			Includes: node.Includes,
		},
	}
}
