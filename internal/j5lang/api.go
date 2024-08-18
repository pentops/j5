package j5lang

import (
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5lang/ast"
	"github.com/pentops/j5/internal/j5lang/lexer"
)

func ConvertFileToSource(input string) (*source_j5pb.SourceFile, error) {
	tree, err := ParseFile(input)
	if err != nil {
		return nil, err
	}

	file, err := ConvertTreeToSource(tree)
	if err != nil {

		return nil, lexer.AddPositionErrorSource(err, input)
	}

	return file, nil
}

func ParseFile(input string) (*ast.File, error) {
	l := lexer.NewLexer(input)

	tokens, err := l.AllTokens()
	if err != nil {
		return nil, lexer.AddPositionErrorSource(err, input)
	}

	tree, err := ast.Walk(tokens)
	if err != nil {
		return nil, lexer.AddPositionErrorSource(err, input)
	}

	return tree, nil
}

type SyntaxError = lexer.PositionErrorWithSource

type SyntaxErrors []SyntaxError

func (e SyntaxErrors) Error() string {
	if len(e) > 0 {
		return e[0].Error()
	}
	return "syntax errors"
}

func AsSyntaxErrors(err error) (SyntaxErrors, bool) {
	errors, ok := lexer.AsPositionErrorsWithSource(err)
	return SyntaxErrors(errors), ok
}

func (se SyntaxErrors) Print(printer func(string, ...interface{})) {
	lexer.PositionErrorsWithSource(se).Print(printer)
}
