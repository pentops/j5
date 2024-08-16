package lexer

import (
	"strings"
	"testing"
)

func tIdent(lit string) Token {
	return Token{
		Type: IDENT,
		Lit:  lit,
	}
}

func tInt(lit string) Token {
	return Token{
		Type: INT,
		Lit:  lit,
	}
}

func tString(lit string) Token {
	return Token{
		Type: STRING,
		Lit:  lit,
	}
}

func tComment(lit string) Token {
	return Token{
		Type: COMMENT,
		Lit:  lit,
	}
}

func tDescription(lit string) Token {
	return Token{
		Type: DESCRIPTION,
		Lit:  lit,
	}
}

func tDecimal(lit string) Token {
	return Token{
		Type: DECIMAL,
		Lit:  lit,
	}
}

func tBool(lit string) Token {
	return Token{
		Type: BOOL,
		Lit:  lit,
	}
}

var (
	tAssign = Token{
		Type: ASSIGN,
	}
	tEOF = Token{
		Type: EOF,
	}
	tObject = Token{
		Type: OBJECT,
	}
	tField = Token{
		Type: FIELD,
	}

	tLBrace = Token{
		Type: LBRACE,
	}

	tRBrace = Token{
		Type: RBRACE,
	}
	tDot = Token{
		Type: DOT,
	}
	tEOL = Token{
		Type: EOL,
	}
	tPackage = Token{
		Type: PACKAGE,
	}
)

func TestSimple(t *testing.T) {

	for _, tc := range []struct {
		name        string
		input       []string
		expected    []Token
		expectError *Position
	}{{
		name:  "assign",
		input: []string{`vv=123`},
		expected: []Token{
			tIdent("vv"),
			tAssign,
			tInt("123"),
			tEOF,
		},
	}, {
		name:  "assign with spaces",
		input: []string{`vv = 123`},
		expected: []Token{
			tIdent("vv"),
			tAssign,
			tInt("123"),
			tEOF,
		},
	}, {
		name: "identifier with dots",
		input: []string{
			`vv.with.dots = 123`,
		},
		expected: []Token{
			tIdent("vv"),
			tDot,
			tIdent("with"),
			tDot,
			tIdent("dots"),
			tAssign,
			tInt("123"),
			tEOF,
		},
	}, {
		name: "literal types",
		input: []string{
			`vv = 123`,
			`vv = "value"`,
			`vv = 123.456`,
			`vv = true`,
			`vv = false`,
		},
		expected: []Token{
			tIdent("vv"), tAssign, tInt("123"), tEOL,
			tIdent("vv"), tAssign, tString("value"), tEOL,
			tIdent("vv"), tAssign, tDecimal("123.456"), tEOL,
			tIdent("vv"), tAssign, tBool("true"), tEOL,
			tIdent("vv"), tAssign, tBool("false"), tEOF,
		},
	}, {
		name: "type declaration",
		input: []string{
			`object Foo {}`,
		},
		expected: []Token{
			tObject,
			tIdent("Foo"),
			tLBrace,
			tRBrace,
			tEOF,
		},
	}, {
		name: "string quotes",
		input: []string{
			`vv = "value"`,
		},
		expected: []Token{
			tIdent("vv"),
			tAssign,
			tString("value"),
			tEOF,
		},
	}, {
		name: "string escaped quotes",
		input: []string{
			`vv = "value \"with\" quotes"`,
		},
		expected: []Token{
			tIdent("vv"),
			tAssign,
			tString("value \"with\" quotes"),
			tEOF,
		},
	}, {
		name: "string with useless escapes",
		input: []string{
			`vv = "value \\ with \\ useless \\ escapes"`,
		},
		expected: []Token{
			tIdent("vv"),
			tAssign,
			tString("value \\ with \\ useless \\ escapes"),
			tEOF,
		},
	}, {
		name: "string with invalid escape",
		input: []string{
			`vv = "value \ with invalid escape"`,
		},
		expectError: &Position{Line: 1, Column: 13},
	}, {
		name: "Newline in string is bad",
		input: []string{
			`vv = "value`,
			`with newline"`,
		},
		expectError: &Position{Line: 1, Column: 11},
	}, {
		name: "Escaped is fine",
		input: []string{
			`vv = "value\`,
			`with newline"`,
		},
		// note no EOL token, strings and comments and descriptions include the
		// newline
		expected: []Token{
			tIdent("vv"),
			tAssign,
			tString("value\nwith newline"),
			tEOF,
		},
	}, {
		name: "extend identifier",
		input: []string{
			`key123_ü = 123`,
		},
		expected: []Token{
			tIdent("key123_ü"),
			tAssign,
			tInt("123"),
			tEOF,
		},
	}, {
		name: "comment line",
		input: []string{
			"vv = 123 // c1",
			"vv = 123",
			"// c2",
			" //c3",
		},
		expected: []Token{
			tIdent("vv"), tAssign, tInt("123"),
			tComment(" c1"), tEOL,
			tIdent("vv"), tAssign, tInt("123"), tEOL,
			tComment(" c2"), tEOL,
			tComment("c3"), tEOF,
		},
	}, {
		name: "block comment empty",
		input: []string{
			"/**/ vv",
		},
		expected: []Token{
			tComment(""),
			tIdent("vv"),
			tEOF,
		},
	}, {
		name: "block comment",
		input: []string{
			"/* line1",
			"line2 */",
			"vv",
		},
		expected: []Token{
			tComment(" line1\nline2 "),
			tEOL,
			tIdent("vv"),
			tEOF,
		},
	}, {
		name: "description",
		input: []string{
			`  | line1 of description`,
			`  | line2 of description`,
			"vv = 123",
		},
		expected: []Token{
			tDescription("line1 of description\nline2 of description"),
			tEOL,
			tIdent("vv"), tAssign, tInt("123"),
			tEOF,
		},
	}} {

		t.Run(tc.name, func(t *testing.T) {

			tokens, err := scanAll(strings.Join(tc.input, "\n"))
			if tc.expectError != nil {
				if err == nil {
					t.Fatalf("expected error at %s", tc.expectError)
				}
				posErr, ok := err.(*PositionError)
				if !ok {
					t.Fatalf("expected position error, got %T", err)
				}
				if posErr.Position.String() != tc.expectError.String() {
					t.Fatalf("expected error at %s, got %s", tc.expectError, posErr.Position)
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertTokensEqual(t, tokens, tc.expected)

		})
	}
}

func assertTokensEqual(t *testing.T, tokens, expected []Token) {

	for idx, tok := range tokens {
		if len(expected) <= idx {
			t.Errorf("BAD %d %s (extra)", idx, tok)
			continue
		}
		if tok.Type != expected[idx].Type {
			t.Errorf("BAD %d %s want %s", idx, tok, expected[idx])
			continue
		} else if tok.Lit != expected[idx].Lit {
			t.Errorf("BAD %d %s want %s", idx, tok, expected[idx])
			continue
		}
		t.Logf("OK  %d: %s at %s to %s", idx, tok, tok.Start, tok.End)
	}

	if len(expected) > len(tokens) {
		for _, tok := range expected[len(tokens):] {
			t.Errorf("missing %s", tok)
		}
	}
}

func scanAll(input string) ([]Token, error) {
	lexer := NewLexer(input)
	tokens := []Token{}
	for {
		tok, err := lexer.NextToken()
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens, nil
}

func TestFullExample(t *testing.T) {
	input := `
package pentops.j5lang.example
version = "v1"

// Comment Line
object Foo {
	| Foo is an example object
	| from ... Python I guess?
	| Unsure.

	field id uuid {}

	field name string {
		min_len = 10
	}
}

/* Comment Block

With Lines
*/`

	tokens, err := scanAll(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTokensEqual(t, tokens, []Token{
		tEOL,
		tPackage, tIdent("pentops"), tDot, tIdent("j5lang"), tDot, tIdent("example"), tEOL,
		tIdent("version"), tAssign, tString("v1"), tEOL,
		tEOL,
		tComment(" Comment Line"), tEOL,
		tObject, tIdent("Foo"), tLBrace, tEOL,
		tDescription("Foo is an example object\nfrom ... Python I guess?\nUnsure."), tEOL,
		tEOL,
		tField, tIdent("id"), tIdent("uuid"), tLBrace, tRBrace, tEOL,
		tEOL,
		tField, tIdent("name"), tIdent("string"), tLBrace, tEOL,
		tIdent("min_len"), tAssign, tInt("10"), tEOL,
		tRBrace, tEOL,
		tRBrace, tEOL,
		tEOL,
		tComment(" Comment Block\n\nWith Lines\n"),
		tEOF,
	})

}
