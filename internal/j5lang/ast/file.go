package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pentops/j5/internal/j5lang/lexer"
)

type File struct {
	Package string
	Imports []Import
	Body    Body
}

type Comment struct {
	Text string
}

type SourceNode struct {
	Start   lexer.Position
	End     lexer.Position
	Comment *Comment
}

func (sn SourceNode) Source() SourceNode {
	return sn
}

type Import struct {
	Path  string
	Alias string
}

type Body struct {
	Includes   []Reference
	Statements []Statement
}

// Ident is a simple name used when declaring a type, or as parts of a
// reference.
type Ident struct {
	Value string
	SourceNode
}

func (i Ident) String() string {
	return i.Value
}

func (i Ident) GoString() string {
	return fmt.Sprintf("ident(%s)", i.Value)
}

// Reference is a dot separates set of Idents
type Reference []Ident

func (r Reference) String() string {
	parts := make([]string, len(r))
	for i, part := range r {
		parts[i] = part.Value
	}
	return strings.Join(parts, ".")
}

func ReferencesToStrings(refs []Reference) []string {
	out := make([]string, len(refs))
	for i, ref := range refs {
		out[i] = ref.String()
	}
	return out
}

// AsValue converts the reference to a Value type, which is used when it is on
// the RHS of an assignment or directive.
func (r Reference) AsValue() Value {
	return Value{
		token: lexer.Token{
			Type: lexer.IDENT,
			Lit:  r.String(),
		},
	}
}

func (r Reference) GoString() string {
	return fmt.Sprintf("reference(%s)", r)
}

type BlockHeader struct {
	Name        []Reference
	Qualifier   Reference
	Description string
	Export      bool
	SourceNode
}

// ScanTags scans the tags after the first 'type' tag, all elements must be
// single Ident, not joined references. (for other cases parse it directly)
func (bs BlockHeader) ScanTags(into ...*string) error {
	wantLen := 1 + len(into)
	// idx 0 is the type
	if len(bs.Name) != wantLen {
		return fmt.Errorf("expected %d tags, got %v", wantLen, bs.Name) //ast.ReferencesToStrings(tags))
	}

	for idx, dest := range into {
		tag := bs.Name[1+idx]
		if len(tag) != 1 {
			return fmt.Errorf("expected single tag, got %v", tag)
		}
		str := tag[0].String()
		*dest = str
	}

	return nil
}

func (bs BlockHeader) GoString() string {
	return fmt.Sprintf("block(%s)", bs.Name)
}

func (bs BlockHeader) RootName() string {
	if len(bs.Name) == 0 {
		return ""
	}
	return bs.Name[0].String()
}

func (bs BlockHeader) guessName() string {
	if len(bs.Name) == 0 {
		return "?"
	}
	first := bs.Name[0].String()
	if len(bs.Name) == 1 {
		return first
	}
	second := bs.Name[1].String()
	return fmt.Sprintf("%s(%s)", first, second)
}

func (bs BlockHeader) NamePart(idx int) (string, bool) {
	if idx >= len(bs.Name) {
		return "", false
	}
	return bs.Name[idx].String(), true
}

type BlockStatement struct {
	BlockHeader
	Body Body
	statement
}

type Statement interface {
	fmt.GoStringer
	Source() SourceNode
	isStatement()
}

type statement struct{}

func (s statement) isStatement() {}

type Assignment struct {
	Key   Reference
	Value Value
	SourceNode
	statement
}

func (a Assignment) GoString() string {
	return fmt.Sprintf("assign(%s = %#v)", a.Key, a.Value)
}

type Directive struct {
	Key   Reference
	Value *Value
	SourceNode
	statement
}

func (d Directive) GoString() string {
	return fmt.Sprintf("directive(%s %#v)", d.Key, d.Value)
}

type TypeError struct {
	Expected string
	Got      string
}

func (te *TypeError) Error() string {
	return fmt.Sprintf("expected a %s, got %s", te.Expected, te.Got)
}

type Value struct {
	token lexer.Token
	SourceNode
}

func (v Value) GoString() string {
	return fmt.Sprintf("value(%s:%s)", v.token.Type, v.token.Lit)
}

func (v Value) AsString() (string, error) {
	if v.token.Type != lexer.STRING {

		return "", &TypeError{
			Expected: "string",
			Got:      v.token.String(),
		}
	}
	return v.token.Lit, nil
}

func (v Value) AsBoolean() (bool, error) {
	if v.token.Type != lexer.BOOL {
		return false, &TypeError{
			Expected: "bool",
			Got:      v.token.String(),
		}
	}
	return v.token.Lit == "true", nil
}

func (v Value) AsUint(size int) (uint64, error) {
	if v.token.Type != lexer.INT {
		return 0, &TypeError{
			Expected: fmt.Sprintf("uint%d", size),
			Got:      v.token.String(),
		}
	}
	parsed, err := strconv.ParseUint(v.token.Lit, 10, size)
	return parsed, err

}
