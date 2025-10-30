package schema

import (
	"fmt"
	"sort"

	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/lib/j5reflect"
)

type ScalarField interface {
	SetASTValue(j5reflect.ASTValue) error
	FullTypeName() string
}

type ArrayOfScalarField interface {
	AppendASTValue(j5reflect.ASTValue) (int, error)
	FullTypeName() string
}

type Field interface {
	j5reflect.Field
}

type field struct {
	j5reflect.Field
	location *bcl_j5pb.SourceLocation
}

type SourceLocation = errpos.Position

type Scope struct {
	parent    *Scope // parent scope, if any
	leafBlock *containerField
	schemaSet *SchemaSet
}

func (sw *Scope) CurrentBlock() Container {
	return sw.leafBlock
}

func (sw *Scope) RootBlock() Container {
	if sw.parent != nil {
		return sw.parent.RootBlock()
	}
	return sw.leafBlock
}

func NewRootSchemaWalker(ss *SchemaSet, root j5reflect.Object, sourceLoc *bcl_j5pb.SourceLocation) (*Scope, error) {
	if sourceLoc == nil {
		return nil, fmt.Errorf("source location required")
	}

	rootWrapped, err := ss.wrapContainer(root, []string{}, sourceLoc)
	if err != nil {
		return nil, err
	}

	rootWrapped.isRoot = true
	return &Scope{
		schemaSet: ss,
		leafBlock: rootWrapped,
	}, nil
}

func (sw *Scope) newChild(container *containerField, newScope bool) *Scope {
	var parent *Scope
	if newScope {
		parent = sw
	}
	return &Scope{
		parent:    parent,
		leafBlock: container,
		schemaSet: sw.schemaSet,
	}
}

func (sw *Scope) SchemaNames() []string {
	var parent []string
	if sw.parent != nil {
		parent = sw.parent.SchemaNames()
	}
	return append(parent, sw.leafBlock.schemaName)
}

func (sw *Scope) allChildFields() map[string]*schema_j5pb.Field {
	var children map[string]*schema_j5pb.Field
	if sw.parent != nil {
		children = sw.parent.allChildFields()
	} else {
		children = map[string]*schema_j5pb.Field{}
	}

	for name, schema := range sw.leafBlock.allFields() {
		if _, ok := children[name]; !ok {
			children[name] = schema
		}
	}
	return children
}

func (sw *Scope) listBlocks() []string {
	fields := sw.allChildFields()
	fieldNames := []string{}

	for name, field := range fields {
		if schemaCan(field.GetType()).canBlock {
			fieldNames = append(fieldNames, name)
		}
	}
	sort.Strings(fieldNames)
	return fieldNames
}

func (sw *Scope) ChildBlock(name string, source SourceLocation) (*Scope, *WalkPathError) {
	root, spec, ok := sw.findBlock(name)
	if !ok {
		return nil, &WalkPathError{
			Field:     name,
			Type:      RootNotFound,
			Available: sw.listBlocks(),
		}
	}

	container, err := sw.walkToChild(root, spec.Path, source)
	if err != nil {
		switch err.Type {
		case NodeNotContainer:
			err.Path = []string{name}
		}
		return nil, err
	}

	newWalker := sw.newChild(container, true)
	return newWalker, nil
}

func (sw *Scope) ScalarField(name string, source SourceLocation) (ScalarField, *WalkPathError) {
	finalField, _, err := sw.field(name, source, false)
	if err != nil {
		return nil, err
	}

	asScalar, ok := finalField.AsScalar()
	if ok {
		return asScalar, nil
	}

	return nil, &WalkPathError{
		Path:   []string{name},
		Type:   NodeNotScalar,
		Schema: finalField.FullTypeName(),
	}
}

func (sw *Scope) Field(name string, source SourceLocation, existingIsOk bool) (Field, *WalkPathError) {
	finalField, _, err := sw.field(name, source, existingIsOk)
	if err != nil {
		return nil, err
	}

	return finalField, nil
}

func (sw *Scope) field(name string, source SourceLocation, existingIsOk bool) (Field, *ChildSpec, *WalkPathError) {
	// Root, Parent and Field.
	// The 'Root' is the container within the current scope which is identified
	// by the block name.

	// Parent is the second last element in the path, the object/oneof etc which
	// holds the field we are looking for.

	// The 'Field' is the leaf at the end of the path.

	// A Path from 'Root' to 'Parent' gives us the place we can get the field,
	// but we can't walk all the way to the field because it is a scalar, so we
	// need it in context.

	root, spec, ok := sw.findBlock(name)
	if !ok {
		return nil, nil, &WalkPathError{
			Field:     name,
			Type:      RootNotFound,
			Schema:    sw.leafBlock.schemaName,
			Available: sw.listBlocks(),
		}
	}
	if len(spec.Path) == 0 {
		return nil, nil, &WalkPathError{
			Field:  name,
			Type:   UnknownPathError,
			Schema: root.schemaName,
			Err:    fmt.Errorf("empty path, spec issue"),
		}
	}

	final, pathToParent := popLast(spec.Path)
	parentScope, err := sw.walkToChild(root, pathToParent, source)
	if err != nil {
		return nil, nil, err
	}

	if !parentScope.container.HasAvailableProperty(final) {
		return nil, nil, &WalkPathError{
			Field:     final,
			Schema:    parentScope.schemaName,
			Type:      NodeNotFound,
			Available: sw.listBlocks(),
		}
	}

	if existingIsOk {
		field, err := parentScope.getOrSetValue(final, source)
		if err != nil {
			return nil, nil, &WalkPathError{
				Type: UnknownPathError,
				Err:  err,
			}
		}
		return field, nil, nil
	}

	finalField, newValErr := parentScope.newValue(final, source)
	if newValErr != nil {
		return nil, nil, &WalkPathError{
			Type: UnknownPathError,
			Err:  newValErr,
		}
	}

	return finalField, spec, nil
}

func (sw *Scope) walkToChild(blockSchema *containerField, path []string, sourceLocation SourceLocation) (*containerField, *WalkPathError) {
	if len(path) == 0 {
		return blockSchema, nil
	}

	// walk the block to the path specified in the config.
	visitedFields, pathErr := blockSchema.walkPath(path, sourceLocation)
	if pathErr != nil {
		return nil, pathErr
	}

	for _, field := range visitedFields {
		spec, err := sw.schemaSet.blockSpec(field.container)
		if err != nil {
			return nil, unexpectedPathError(field.name, err)
		}
		field.spec = *spec
	}

	mainField := visitedFields[0]
	mainField.transparentPath = visitedFields[1:]
	return mainField, nil
}

func (sw *Scope) findBlock(name string) (*containerField, *ChildSpec, bool) {
	// Search parents first, then search the current scope.
	if sw.parent != nil {
		f, s, ok := sw.parent.findBlock(name)
		if ok {
			return f, s, true
		}
	}
	return sw.findLeafBlock(name)
}

func (sw *Scope) findLeafBlock(name string) (*containerField, *ChildSpec, bool) {
	pathToChild, ok := sw.leafBlock.spec.Aliases[name]
	if ok {
		return sw.leafBlock, &ChildSpec{
			Path: pathToChild,
		}, true
	}

	if sw.leafBlock.container.HasAvailableProperty(name) {
		return sw.leafBlock, &ChildSpec{
			Path: []string{name},
		}, true
	}

	return nil, nil, false
}

func popLast[T any](list []T) (T, []T) {
	return list[len(list)-1], list[:len(list)-1]
}

func (sw *Scope) TailScope() *Scope {
	return &Scope{
		leafBlock: sw.leafBlock,
		schemaSet: sw.schemaSet,
	}
}

func (sw *Scope) Parent() *Scope {
	if sw.parent == nil {
		return nil
	}
	return sw.parent
}

func (sw *Scope) MergeScope(other *Scope) *Scope {
	return &Scope{
		parent:    sw,
		leafBlock: other.leafBlock,
		schemaSet: sw.schemaSet,
	}
}

func (sw *Scope) Orphan() *Scope {
	return &Scope{
		leafBlock: sw.leafBlock,
		schemaSet: sw.schemaSet,
	}
}

func (sw *Scope) PrintScope(logf func(string, ...any)) {
	logf("available blocks:")
	sw.printScope(logf)
	if sw.leafBlock == nil {
		logf("no leaf spec")
	} else {

		spec := sw.leafBlock.spec
		logf("leaf spec Tags: %s", spec.ErrName())
		if spec.Name != nil {
			logf(" - tag[name]: %#v", spec.Name)
		}
		if spec.TypeSelect != nil {
			logf(" - tag[type]: %#v", spec.TypeSelect)
		}
	}
	logf("-------")
}

func (sw *Scope) printScope(logf func(string, ...any)) {
	if sw.parent != nil {
		sw.parent.printScope(logf)
	}

	if sw.leafBlock.spec.DebugName != "" {
		logf("from %s : %s %q", sw.leafBlock.schemaName, sw.leafBlock.spec.source, sw.leafBlock.spec.DebugName)
	} else {
		logf("from %s : %s", sw.leafBlock.schemaName, sw.leafBlock.spec.source)
	}
	for name, subBlock := range sw.leafBlock.allFields() {
		logf(" - [%s] %#v", name, schemaCan(subBlock.GetType()))
	}

	logf("-------")
}
