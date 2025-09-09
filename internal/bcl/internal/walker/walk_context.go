package walker

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/bcl/internal/parser"
	"github.com/pentops/j5/internal/bcl/internal/walker/schema"
)

type Context interface {
	WithFreshScope(path ScopePath, fn SpanCallback) error
	WithMergedScope(path ScopePath, fn SpanCallback) error

	SetDescription(desc parser.ASTValue) error
	SetAttribute(path ScopePath, value parser.ASTValue) error
	AppendAttribute(path ScopePath, value parser.ASTValue) error

	setContainerFromScalar(bs schema.BlockSpec, vals parser.ASTValue) error

	Logf(format string, args ...any)
	WrapErr(err error, pos HasPosition) error
}

type SpanCallback func(Context, schema.BlockSpec) error

type walkContext struct {
	scope *schema.Scope

	// path is the full path from the root to this context, as field names
	path []string

	// depth is the nested level of walk context. It may not equal len(name)
	// as depth skips blocks
	depth         int
	blockLocation schema.SourceLocation

	verbose bool
}

func newSchemaError(err error) error {
	return fmt.Errorf("schema error: %w", err)
}

type pathElement struct {
	name     string
	position *schema.SourceLocation
}

type ScopePath struct {
	Schema schema.PathSpec
	User   []parser.Ident
}

func (sp ScopePath) GoString() string {
	nn := make([]string, 0, len(sp.Schema)+len(sp.User))
	for _, ident := range sp.Schema {
		nn = append(nn, ident)
	}
	uu := make([]string, 0, len(sp.User))
	for _, ident := range sp.User {
		uu = append(uu, ident.String())
	}
	return fmt.Sprintf("idents{%s/%s}", strings.Join(nn, "."), strings.Join(uu, "."))
}

func (sp ScopePath) combined() []pathElement {
	return combinePath(sp.Schema, sp.User)
}

func combinePath(path schema.PathSpec, ref []parser.Ident) []pathElement {
	pathToBlock := make([]pathElement, len(path)+len(ref))
	for i, ident := range path {
		pathToBlock[i] = pathElement{
			name: ident,
		}
	}

	for i, ident := range ref {
		pathToBlock[i+len(path)] = pathElement{
			name: ident.String(),
			position: &schema.SourceLocation{
				Start: ident.Start,
				End:   ident.End,
			},
		}
	}
	return pathToBlock
}

func (sc *walkContext) SetLocation(loc schema.SourceLocation) {
	sc.blockLocation = loc
}

func (sc *walkContext) walkScopePath(path []pathElement) (*schema.Scope, error) {
	scope := sc.scope
	loc := sc.blockLocation
	for _, ident := range path {
		if ident.position != nil {
			loc = *ident.position
		}

		next, werr := scope.ChildBlock(ident.name, loc)
		if werr == nil { // INVERSION
			sc.Logf("walkScopePath %q found %s", ident.name, next.CurrentBlock().SchemaName())
			scope = next.Orphan()
			continue
		}

		if ident.position == nil {
			return nil, newSchemaError(werr)
		}

		var err error
		switch werr.Type {
		case schema.RootNotFound:
			blocks := scope.SchemaNames()
			if len(blocks) == 1 {
				err = fmt.Errorf("root type %q has no field %s - expecting %q",
					blocks[0],
					werr.Field,
					werr.Available) // ", "))
			} else if len(blocks) > 1 {
				err = fmt.Errorf("no field %q in any of %q - expecting %q",
					werr.Field,
					blocks,
					werr.Available)
			}
		case schema.NodeNotFound:
			err = fmt.Errorf("type %q has no field %q - expecting %q",
				werr.Schema,
				werr.Field,
				werr.Available) //strings.Join(werr.Available, ", "))

		default:
			err = fmt.Errorf("%s", werr.LongMessage())
		}

		err = errpos.AddPosition(err, *ident.position)
		return nil, err
	}
	return scope, nil
}

type BadTypeError struct {
	WantType string
	GotType  string
}

func (bte BadTypeError) Error() string {
	return fmt.Sprintf("bad type: want %s, got %s", bte.WantType, bte.GotType)
}

func (sc *walkContext) SetDescription(description parser.ASTValue) error {
	root := sc.scope.RootBlock()
	descSpec := root.Spec().Description
	if descSpec == nil {
		return newSchemaError(fmt.Errorf("no description field"))
	}

	return sc.SetAttribute(ScopePath{Schema: schema.PathSpec{*descSpec}}, description)
}

func (sc *walkContext) AppendAttribute(path ScopePath, val parser.ASTValue) error {
	sc.Logf("AppendAttribute(%#v, %#v, %s)", path, val, val.Position())
	return sc.setAttribute(path, val, true)
}

func (sc *walkContext) SetAttribute(path ScopePath, val parser.ASTValue) error {
	sc.Logf("SetAttribute(%#v,  %#v, %s)", path, val, val.Position())
	return sc.setAttribute(path, val, false)
}

func (sc *walkContext) setAttribute(path ScopePath, val parser.ASTValue, appendValue bool) error {

	fullPath := path.combined()
	if len(fullPath) == 0 {
		return newSchemaError(fmt.Errorf("empty path for SetAttribute"))
	}

	last := fullPath[len(fullPath)-1]
	pathToBlock := fullPath[:len(fullPath)-1]
	parentScope := sc.scope
	if len(pathToBlock) > 0 {
		foundScope, err := sc.walkScopePath(pathToBlock)
		if err != nil {
			return err
		}
		parentScope = foundScope.Orphan()
	}

	sc.Logf(">> SetAttribute %q in %q", last.name, parentScope.CurrentBlock().SchemaName())
	sc.logScope(parentScope)

	field, walkPathErr := parentScope.Field(last.name, val.Position(), appendValue)
	if walkPathErr != nil {
		sc.Logf("parentScope.Field(%q) failed: %s", last.name, walkPathErr)
		if last.position != nil {
			return sc.WrapErr(walkPathErr, *last.position)
		} else {
			return newSchemaError(walkPathErr)
		}
	}

	_, ok := field.AsContainer()
	if ok {
		if appendValue {
			return sc.WrapErr(fmt.Errorf("cannot append to container"), val.Position())
		}
		containerScope, err := parentScope.ChildBlock(last.name, val.Position())
		if err != nil {
			return sc.WrapErr(err, val.Position())
		}

		sc.Logf("|>>> Entering Clean Container for Block-Attribute >>>")
		return sc.withScope(containerScope.Orphan(), func(sc Context, bs schema.BlockSpec) error {
			return sc.setContainerFromScalar(bs, val)
		})
	}

	isArray := val.IsArray()
	vals, _ := val.AsArray()
	if !isArray && appendValue {
		vals = []parser.ASTValue{val}
		isArray = true
	}

	if isArray {
		fieldArray, ok := field.AsArrayOfScalar()

		if ok { // Field and Value are both arrays.
			if !appendValue && fieldArray.Length() > 0 {
				return sc.WrapErr(fmt.Errorf("value already set"), val.Position())
			}
			for _, val := range vals {
				_, err := fieldArray.AppendASTValue(val)
				if err != nil {
					err = fmt.Errorf("SetAttribute %s, Append value: %w", field.FullTypeName(), err)
					return sc.WrapErr(err, val.Position())
				}
			}
			sc.Logf("SetAttribute Array Done")
			return nil
		}

		return sc.WrapErr(BadTypeError{
			WantType: "ArrayOfScalar",
			GotType:  field.FullTypeName(),
		}, val.Position())
	}

	scalarField, ok := field.AsScalar()
	if !ok {
		return sc.WrapErr(BadTypeError{
			WantType: "Scalar",
			GotType:  field.FullTypeName(),
		}, val.Position())
	}

	err := scalarField.SetASTValue(val)
	if err != nil {
		err = fmt.Errorf("SetAttribute %s: %w", field.FullTypeName(), err)
		return sc.WrapErr(err, val.Position())
	}

	sc.Logf("SetAttribute Non-Array Done")
	return nil
}

func (sc *walkContext) setContainerFromScalar(bs schema.BlockSpec, val parser.ASTValue) error {
	ss := bs.ScalarSplit
	if ss == nil {
		return fmt.Errorf("container %s has no method to set from array", bs.ErrName())
	}

	var setVals []parser.ASTValue

	if ss.Delimiter != nil {
		strVal, err := val.AsString()
		if err != nil {
			return sc.WrapErr(err, val.Position())
		}
		sc.Logf("Splitting scalar %#v -> %q", val, strVal)
		valStrings := strings.Split(strVal, *bs.ScalarSplit.Delimiter)
		vals := make([]parser.ASTValue, len(valStrings))
		for idx, str := range valStrings {
			vals[idx] = parser.NewStringValue(str, parser.SourceNode{
				Start: val.Position().Start,
				End:   val.Position().End,
			})
		}
		setVals = vals

	} else {

		vals, isArray := val.AsArray()
		if !isArray {
			return fmt.Errorf("container %s requires an array when setting from value, got a scalar", bs.ErrName())
		}
		setVals = vals
	}
	sc.Logf("setContainerFromArray(%#v)", setVals)

	if ss.RightToLeft {
		slices.Reverse(setVals)
	}

	if len(setVals) < len(ss.Required) {
		return fmt.Errorf("container %s requires %d values, got %d", bs.ErrName(), len(ss.Required), len(setVals))
	}
	intoRequired, remaining := setVals[:len(ss.Required)], setVals[len(ss.Required):]
	for idx, val := range intoRequired {
		rr := ss.Required[idx]
		if err := sc.SetAttribute(ScopePath{Schema: rr}, val); err != nil {
			return err
		}
	}

	if len(remaining) == 0 {
		return nil
	}

	var optional []parser.ASTValue
	if len(remaining) > len(ss.Optional) {
		optional, remaining = remaining[:len(ss.Optional)], remaining[len(ss.Optional):]
	} else {
		optional, remaining = remaining, nil
	}

	for idx, val := range optional {
		ro := ss.Optional[idx]
		if err := sc.SetAttribute(ScopePath{Schema: ro}, val); err != nil {
			return err
		}
	}

	if len(remaining) == 0 {
		return nil
	}

	if ss.Remainder == nil {
		return fmt.Errorf("container %s has more array fields than we know what to do with", bs.ErrName())
	}

	// We reverse at the start to pop values from the end of the array, but when
	// placing back into remainder it should be in the specified order.
	// a.b.c, with RTL, pop `c` as a required element, then a.b is remainder,
	// not b.a
	if ss.RightToLeft {
		slices.Reverse(remaining)
	}

	remainingStr := make([]string, len(remaining))
	for idx, val := range remaining {
		var err error
		remainingStr[idx], err = val.AsString()
		if err != nil {
			return sc.WrapErr(err, val.Position())
		}
	}

	delim := "."
	if ss.Delimiter != nil {
		delim = *ss.Delimiter
	}
	singleString := strings.Join(remainingStr, delim)

	return sc.SetAttribute(ScopePath{Schema: *ss.Remainder}, parser.NewStringValue(singleString, parser.SourceNode{
		Start: remaining[0].Position().Start,
		End:   remaining[len(remaining)-1].Position().End,
	}))

}

func (wc *walkContext) WithFreshScope(path ScopePath, fn SpanCallback) error {
	fullPath := path.combined()
	if len(fullPath) == 0 {
		newScope := wc.scope.TailScope()
		if wc.verbose {
			wc.Logf("|>>> Entering FRESH Reset %q >>>", newScope.CurrentBlock().Name())
		}
		return wc.withScope(newScope, fn)
	}

	container, err := wc.walkScopePath(fullPath)
	if err != nil {
		return err
	}

	if wc.verbose {
		wc.Logf("|>>> Entering FRESH %q >>>", container.CurrentBlock().Name())
	}

	return wc.withScope(container.Orphan(), fn)
}

func (wc *walkContext) WithMergedScope(path ScopePath, fn SpanCallback) error {
	fullPath := path.combined()
	if len(fullPath) == 0 {
		if wc.verbose {
			wc.Logf("|>>> Entering MERGED Re-Entry %q >>>", wc.scope.CurrentBlock().Name())
		}
		return wc.withScope(wc.scope, fn)
	}
	container, err := wc.walkScopePath(fullPath)
	if err != nil {
		return err
	}
	newScope := wc.scope.MergeScope(container)

	if wc.verbose {
		wc.Logf("|>>> Entering MERGED %q >>>", container.CurrentBlock().Name())
	}
	return wc.withScope(newScope, fn)
}

func (wc *walkContext) logScope(scope *schema.Scope) {
	if !wc.verbose {
		return
	}
	wc.Logf("|>>> SCOPE %q >>>", scope.CurrentBlock().SchemaName())
	prefix := strings.Repeat("| ", wc.depth) + "|> "
	entry := prefixer(log.Printf, prefix)
	scope.PrintScope(entry)
}

func (wc *walkContext) withScope(newScope *schema.Scope, fn SpanCallback) error {
	lastBlock := newScope.CurrentBlock()

	newPath := append(wc.path, lastBlock.Name())

	if wc.verbose {
		wc.Logf("|>>> Entering %q >>>", lastBlock.Name())
		prefix := strings.Repeat("| ", wc.depth) + "|> "
		entry := prefixer(log.Printf, prefix)
		entry("Src = %q", strings.Join(lastBlock.Path(), "."))
		entry("Path = %q", strings.Join(newPath, "."))

		newScope.PrintScope(entry)
	}

	childContext := &walkContext{
		scope:         newScope,
		path:          newPath,
		depth:         wc.depth + 1,
		verbose:       wc.verbose,
		blockLocation: wc.blockLocation,
	}

	err := fn(childContext, lastBlock.Spec())
	if err == nil {
		wc.Logf("|<<< Exiting OK %q <<<", lastBlock.Name())
		return nil
	}

	// already scoped, pass it up the tree.
	scoped := &scopedError{}
	if errors.As(err, &scoped) {
		return err
	}

	wc.Logf("New Error %s", err)

	posErr, ok := errpos.AsError(err)
	if !ok {
		wc.Logf("Not errpos")
		posErr = &errpos.Err{
			Err: err,
		}
	}

	return &scopedError{
		err:    posErr,
		schema: newScope,
	}
}

type HasPosition interface {
	Position() errpos.Position
}

func (wc *walkContext) WrapErr(err error, pos HasPosition) error {
	if err == nil {
		panic("WrapErr called with nil error")
	}

	wc.Logf("Wrapping Error %s with %s", err, pos.Position())
	err = errpos.AddContext(err, strings.Join(wc.path, "."))
	err = errpos.AddPosition(err, pos.Position())
	return err
}

type logger func(format string, args ...any)

func prefixer(parent logger, prefix string) logger {
	return func(format string, args ...any) {
		parent(prefix+format+"\n", args...)
	}
}

func (wc *walkContext) Logf(format string, args ...any) {
	if !wc.verbose {
		return
	}
	prefix := strings.Repeat("| ", wc.depth)
	prefixer(log.Printf, prefix)(format, args...)
}

type scopedError struct {
	err    *errpos.Err
	schema *schema.Scope
}

func (se *scopedError) Error() string {
	return se.err.Error()
}

func (se *scopedError) Unwrap() error {
	return se.err
}

func logError(err error) {
	scoped := &scopedError{}
	pf := prefixer(log.Printf, "ERR | ")
	if !errors.As(err, &scoped) {
		pf("Error (unscoped): %s\n", err)
		return
	}
	msg := scoped.err.Err.Error()
	pf("Error: %s", msg)
	pf("Location: %s", scoped.err.Pos)
	pf("Scope:")
	scoped.schema.PrintScope(pf)
	pf("Got Error %s\n", msg)
}
