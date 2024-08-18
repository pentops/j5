package ast

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/internal/j5lang/lexer"
)

func tokenErrf(tok lexer.Token, format string, args ...interface{}) error {
	return &lexer.PositionError{
		Position: tok.Start,
		Msg:      fmt.Sprintf(format, args...),
	}
}

func unexpectedToken(tok lexer.Token, expected ...lexer.TokenType) error {
	return &unexpectedTokenError{
		tok:      tok,
		expected: expected,
	}
}

type unexpectedTokenError struct {
	context  string
	tok      lexer.Token
	expected []lexer.TokenType
}

func (e *unexpectedTokenError) Error() string {
	return fmt.Sprintf("%s %s", e.tok.Start, e.msg())
}

func (e *unexpectedTokenError) msg() string {
	if len(e.expected) == 1 {
		return fmt.Sprintf("unexpected %s in %s, expected %s", e.tok, e.context, e.expected[0])
	}
	expectSet := make([]string, len(e.expected))
	for i, e := range e.expected {
		expectSet[i] = e.String()
	}
	return fmt.Sprintf("unexpected %s in %s, expected one of %s", e.tok, e.context, strings.Join(expectSet, ", "))
}

func (e *unexpectedTokenError) ErrorPosition() (lexer.Position, string) {
	return e.tok.Start, e.msg()
}
