package j5lang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pentops/j5/internal/j5lang/lexer"
)

type BlockHeader struct {
	Name        Ident
	Description string
}

type Body struct {
	Decls []Decl
}

type File struct {
	Body
}

type Node interface {
	Start() lexer.Position
	End() lexer.Position
}

type basicNode struct {
	start lexer.Position
	end   lexer.Position
}

func (n basicNode) Start() lexer.Position {
	return n.start
}

func (n basicNode) End() lexer.Position {
	return n.end
}

type tokenNode struct {
	lexer.Token
}

func (n tokenNode) Start() lexer.Position {
	return n.Token.Start
}

func (n tokenNode) End() lexer.Position {
	return n.Token.End
}

// Decl is any declaration of a type - the types which implement Decl all end
// with 'Decl'.
// EnumDecl
// ObjectDecl
// FieldDecl
type Decl interface {
	Node
}

// Ident is a simple name used when *declaring* a type (field, object, etc). It
// allows a-z, A-Z, 0-9, and _
type Ident string

func (i Ident) String() string {
	return fmt.Sprintf("ident(%s)", string(i))
}

// Reference is a list of identifiers separated by dots. It is used to reference
// a type or field.
type Reference []Ident

func (r Reference) ToString() string {
	parts := make([]string, len(r))
	for i, p := range r {
		parts[i] = string(p)
	}
	return strings.Join(parts, ".")
}

func (r Reference) String() string {
	return fmt.Sprintf("ref(%s)", r.ToString())
}

// Value is a literal being used as a value in the source, such as a string, number, etc.
type Value struct {
	tokenNode
}

func (v Value) AsString() string {
	return v.Lit
}

func (v Value) AsBoolean() (bool, error) {
	if v.Type != lexer.BOOL {
		return false, fmt.Errorf("expected bool, got %s", v.Type)
	}
	return v.Lit == "true", nil
}

func (v Value) AsUint(size int) (uint64, error) {
	if v.Type != lexer.INT {
		return 0, fmt.Errorf("expected int, got %s", v.Type)
	}
	return strconv.ParseUint(v.Lit, 10, size)

}

// ValueAssign is a key-value pair with a Value directly from the source code.
type ValueAssign struct {
	basicNode
	Key   Reference
	Value Value
}

type SpecialDecl struct {
	basicNode
	Key   lexer.TokenType
	Value Reference
}

// ObjectDecl is a declaration of a an object
type ObjectDecl struct {
	basicNode
	Body
	BlockHeader
}

// FieldDecl is a declaration of a field in an object
type FieldDecl struct {
	basicNode
	IsKey bool
	BlockHeader
	Body
	Type Ident
}

// EnumDecl is a declaration of an enum
type EnumDecl struct {
	basicNode
	BlockHeader
	Options []EnumOptionDecl
}

type EnumOptionDecl struct {
	Name        Ident
	Description string
}

type EntityDecl struct {
	basicNode
	BlockHeader
	Body
}
