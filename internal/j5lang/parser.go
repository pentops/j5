package j5lang

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/internal/j5lang/lexer"
)

func ParseFile(input string) (*File, error) {
	l := lexer.NewLexer(input)

	tokens, err := l.AllTokens()
	if err != nil {
		return nil, err
	}

	ww := &Walker{
		tokens: tokens,
	}

	return ww.file()
}

type Walker struct {
	tokens []lexer.Token
	offset int
}

func (ww *Walker) file() (*File, error) {
	file := &File{}

	for {
		stmt, err := ww.popRootStatement()
		if err != nil {
			return nil, err
		}

		if stmt == nil {
			break
		}

		file.Decls = append(file.Decls, stmt)
	}

	return file, nil
}

func (w *Walker) currentPos() lexer.Position {
	if w.offset == 0 {
		return lexer.Position{}
	}
	return w.tokens[w.offset-1].End
}

func (w *Walker) popToken() (lexer.Token, error) {
	if w.offset >= len(w.tokens) {
		return lexer.Token{
			Type: lexer.EOF,
		}, nil
	}
	tok := w.tokens[w.offset]
	w.offset++
	return tok, nil
}

func (w *Walker) Peek() (lexer.Token, error) {
	if w.offset >= len(w.tokens) {
		return lexer.Token{
			Type: lexer.EOF,
		}, nil
	}
	return w.tokens[w.offset], nil
}

func (l *Walker) popRootStatement() (Decl, error) {
	tok, err := l.popToken()
	if err != nil {
		return nil, err
	}

	return l.tokenToDecl(tok)
}

func (l *Walker) tokenToDecl(tok lexer.Token) (Decl, error) {

	switch tok.Type {
	case lexer.EOF:
		return nil, nil

	case lexer.EOL:
		// Empty line
		next, err := l.popToken()
		if err != nil {
			return nil, err
		}
		return l.tokenToDecl(next)

	case lexer.COMMENT:
		// skip comments for now
		next, err := l.popToken()
		if err != nil {
			return nil, err
		}
		return l.tokenToDecl(next)

	case lexer.IDENT:
		return l.walkAsign(tok)

	case lexer.OBJECT:
		return l.walkObject()

	case lexer.FIELD:
		return l.walkField()

	case lexer.ENUM:
		return l.walkEnum()

	case lexer.ENTITY:
		return l.walkEntity()

	case lexer.KEY:
		field, err := l.walkField()
		if err != nil {
			return nil, err
		}
		field.IsKey = true
		return field, nil

	case lexer.PACKAGE, lexer.REF:
		return l.walkSpecialDecl(tok)

	default:
		return nil, tokenErrf(tok, "unexpected token %s, looking for statement", tok)
	}

}

func (ww *Walker) popEOL() error {
	tok, err := ww.popToken()
	if err != nil {
		return err
	}

	if tok.Type == lexer.COMMENT {
		return ww.popEOL()
	}

	if tok.Type != lexer.EOL && tok.Type != lexer.EOF {
		return unexpectedTokenErr(tok, lexer.EOL)
	}
	return nil
}

func (ww *Walker) popType(tt lexer.TokenType) (lexer.Token, error) {
	tok, err := ww.popToken()
	if err != nil {
		return lexer.Token{}, err
	}
	if tok.Type != tt {
		return lexer.Token{}, unexpectedTokenErr(tok, tt)
	}
	return tok, nil
}

func (ww *Walker) popValue() (Value, error) {
	token, err := ww.popToken()
	if err != nil {
		return Value{}, err
	}

	if token.Type.IsLiteral() {
		return Value{
			tokenNode: tokenNode{
				Token: token,
			},
		}, nil
	}

	return Value{}, tokenErrf(token, "unexpected token %s, expected a literal or value", token)

}

func (ww *Walker) popIdent() (Ident, error) {
	tok, err := ww.popType(lexer.IDENT)
	if err != nil {
		return Ident(""), err
	}

	if tok.Type != lexer.IDENT {
		return Ident(""), unexpectedTokenErr(tok, lexer.IDENT)
	}

	return Ident(tok.Lit), nil
}

func (ww *Walker) popReference() (Reference, error) {
	ident, err := ww.popToken()
	if err != nil {
		return nil, err
	}
	return ww.continueReference(ident)
}

func (ww *Walker) popLooseIdent() (Ident, error) {
	tok, err := ww.popToken()
	if err != nil {
		return Ident(""), err
	}

	return looseIdent(tok)
}

func looseIdent(tok lexer.Token) (Ident, error) {
	switch tok.Type {
	case lexer.IDENT: // good, normal
		return Ident(tok.Lit), nil
	case lexer.OBJECT, lexer.ENUM, lexer.KEY:
		return Ident(tok.Type.String()), nil
	}
	return Ident(""), unexpectedTokenErr(tok, lexer.IDENT)
}

func (ww *Walker) continueReference(keyTok lexer.Token) (Reference, error) {

	ref := make(Reference, 0)

	ident, err := looseIdent(keyTok)
	if err != nil {
		return nil, err
	}

	ref = append(ref, ident)

	for {

		next, err := ww.Peek()
		if err != nil {
			return nil, err
		}

		if next.Type != lexer.DOT {
			return ref, nil
		}

		_, err = ww.popToken()
		if err != nil {
			return nil, err
		}

		ident, err := ww.popLooseIdent()
		if err != nil {
			return nil, err
		}
		ref = append(ref, ident)
	}

}

func (ww *Walker) walkSpecialDecl(keyword lexer.Token) (Decl, error) {
	start := ww.currentPos()

	decl := SpecialDecl{
		Key: keyword.Type,
		basicNode: basicNode{
			start: start,
			end:   ww.currentPos(),
		},
	}

	value, err := ww.popReference()
	if err != nil {
		return ValueAssign{}, err
	}

	decl.Value = value
	decl.end = ww.currentPos()

	err = ww.popEOL()
	if err != nil {
		return nil, err
	}

	return decl, nil
}

func (ww *Walker) walkAsign(keyTok lexer.Token) (ValueAssign, error) {

	ref, err := ww.continueReference(keyTok)
	if err != nil {
		return ValueAssign{}, err
	}

	assign := ValueAssign{
		Key: ref,
		basicNode: basicNode{
			start: keyTok.Start,
			end:   keyTok.End,
		},
	}

	_, err = ww.popType(lexer.ASSIGN)
	if err != nil {
		return assign, err
	}

	value, err := ww.popValue()
	if err != nil {
		return ValueAssign{}, err
	}

	assign.Value = value
	assign.end = value.End()

	return assign, ww.popEOL()
}

func (ww *Walker) walkObject() (Decl, error) {
	name, err := ww.popIdent()
	if err != nil {
		return nil, err
	}

	hdr, isOpen, err := ww.blockHeader(name)
	if err != nil {
		return nil, err
	}

	decl := ObjectDecl{
		BlockHeader: hdr,
	}

	if isOpen {
		body, err := ww.walkBody()
		if err != nil {
			return nil, err
		}
		decl.Body = body
	}

	return decl, nil
}

func (ww *Walker) walkEntity() (Decl, error) {
	start := ww.currentPos()

	name, err := ww.popIdent()
	if err != nil {
		return nil, err
	}

	hdr, isOpen, err := ww.blockHeader(name)
	if err != nil {
		return nil, err
	}

	decl := EntityDecl{
		BlockHeader: hdr,
		basicNode: basicNode{
			start: start,
			end:   ww.currentPos(),
		},
	}

	if isOpen {
		body, err := ww.walkBody()
		if err != nil {
			return nil, err
		}
		decl.Body = body
	}

	return decl, nil
}

type TypeName struct {
	Package string
	Schema  string
}

func (ww *Walker) walkField() (FieldDecl, error) {
	start := ww.currentPos()

	name, err := ww.popIdent()
	if err != nil {
		return FieldDecl{}, err
	}

	fieldType, err := ww.popLooseIdent()
	if err != nil {
		return FieldDecl{}, err
	}

	hdr, isOpen, err := ww.blockHeader(name)
	if err != nil {
		return FieldDecl{}, err
	}

	decl := FieldDecl{
		BlockHeader: hdr,
		Type:        fieldType,
		basicNode: basicNode{
			start: start,
			end:   ww.currentPos(),
		},
	}

	if !isOpen {
		return decl, nil
	}

	decl.Body = Body{}

	return decl, ww.walkBlock(func(tok lexer.Token) error {
		stmt, err := ww.tokenToDecl(tok)
		if err != nil {
			return err
		}

		if stmt == nil {
			return errStopWalk
		}

		decl.Body.Decls = append(decl.Body.Decls, stmt)
		return nil
	})

}

func (ww *Walker) walkEnum() (Decl, error) {
	start := ww.currentPos()

	name, err := ww.popIdent()
	if err != nil {
		return nil, fmt.Errorf("enum name: %w", err)
	}

	hdr, isOpen, err := ww.blockHeader(name)
	if err != nil {
		return nil, fmt.Errorf("enum header: %w", err)
	}

	decl := EnumDecl{
		BlockHeader: hdr,
		basicNode: basicNode{
			start: start,
			end:   ww.currentPos(),
		},
	}

	if !isOpen {
		return decl, nil
	}

	return decl, ww.walkBlock(func(tok lexer.Token) error {

		switch tok.Type {

		case lexer.IDENT:
			val := Ident(tok.Lit)

			hdr, isOpen, err := ww.blockHeader(val)
			if err != nil {
				return err
			}

			if isOpen {
				// not expecting anything other than a comment, so should
				// immediately close
				_, err = ww.popType(lexer.RBRACE)
				if err != nil {
					return err
				}
			}

			enumVal := EnumOptionDecl{
				Name:        val,
				Description: hdr.Description,
			}

			decl.Options = append(decl.Options, enumVal)

			err = ww.popEOL()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("enum body: %w", unexpectedTokenErr(tok, lexer.IDENT))
		}

		return nil
	})

}

func (ww *Walker) blockHeader(name Ident) (BlockHeader, bool, error) {

	next, err := ww.Peek()
	if err != nil {
		return BlockHeader{}, false, err
	}

	bb := BlockHeader{
		Name: name,
	}

	if next.Type == lexer.DESCRIPTION {
		// Short Form Description without a block
		// thing | description
		desc, err := ww.popToken()
		if err != nil {
			return BlockHeader{}, false, err
		}
		bb.Description = desc.Lit
		return bb, false, nil
	}

	if next.Type != lexer.LBRACE {
		// no body.
		return bb, false, nil
	}

	// consume the brace.
	_, err = ww.popToken()
	if err != nil {
		return BlockHeader{}, false, err
	}

	err = ww.popEOL()
	if err != nil {
		return BlockHeader{}, true, err
	}

	// What is next?
	next, err = ww.Peek()
	if err != nil {
		return BlockHeader{}, true, err
	}

	if next.Type == lexer.DESCRIPTION {
		desc, err := ww.popToken()
		if err != nil {
			return BlockHeader{}, true, err
		}
		bb.Description = desc.Lit
		err = ww.popEOL()
		if err != nil {
			return BlockHeader{}, true, err
		}
	}

	return bb, true, nil
}

var errStopWalk = fmt.Errorf("stop walking")

func (ww *Walker) walkBlock(cb func(lexer.Token) error) error {
	for {
		tok, err := ww.popToken()
		if err != nil {
			return err
		}

		switch tok.Type {
		case lexer.EOF:
			return unexpectedTokenErr(tok, lexer.RBRACE)

		case lexer.EOL:
			// Empty line
			continue

		case lexer.COMMENT:
			// skip comments for now
			continue

		case lexer.RBRACE:
			return nil // End Body

		default:
			err := cb(tok)
			if err == errStopWalk {
				return nil
			} else if err != nil {
				return err
			}
		}
	}
}

func (ww *Walker) walkBody() (Body, error) {
	bb := Body{}
	return bb, ww.walkBlock(func(tok lexer.Token) error {
		stmt, err := ww.tokenToDecl(tok)
		if err != nil {
			return err
		}

		if stmt == nil {
			return errStopWalk
		}

		bb.Decls = append(bb.Decls, stmt)
		return nil
	})

}

func tokenErrf(tok lexer.Token, format string, args ...interface{}) error {
	return &lexer.PositionError{
		Position: tok.Start,
		Msg:      fmt.Sprintf(format, args...),
	}
}

func unexpectedTokenErr(tok lexer.Token, expected ...lexer.TokenType) error {
	if len(expected) == 1 {
		return tokenErrf(tok, "unexpected token %s, expected %s", tok, expected)
	}
	expectSet := make([]string, len(expected))
	for i, e := range expected {
		expectSet[i] = e.String()
	}
	return tokenErrf(tok, "unexpected token %s, expected one of %s", tok, strings.Join(expectSet, ", "))
}
