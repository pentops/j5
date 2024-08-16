package lexer

import (
	"fmt"
	"unicode"
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Lexer struct {
	pos    Position
	ch     rune
	offset int
	data   []rune
}

func NewLexer(data string) *Lexer {
	return &Lexer{
		pos:  Position{Line: 1, Column: 0},
		data: []rune(data),
	}
}

const eof = -1

func (l *Lexer) next() {
	if l.offset >= len(l.data) {
		l.ch = eof
		return
	}
	r := rune(l.data[l.offset])
	l.offset++

	if r == '\n' {
		l.pos.Line++
		l.pos.Column = 0
	}

	l.ch = r
	l.pos.Column++
}

func (l *Lexer) peek() rune {
	if l.offset >= len(l.data) {
		return eof
	}
	return rune(l.data[l.offset])
}

func (l *Lexer) peekAhead(n int) rune {
	if l.offset+n >= len(l.data) {
		return eof
	}
	return rune(l.data[l.offset+n])
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.peek()) {
		l.next()
	}
}

func (l *Lexer) peekPastWhitespace() rune {
	n := 0
	for unicode.IsSpace(l.peekAhead(n)) {
		n++
	}
	return l.peekAhead(n)
}

func (l *Lexer) tokenOf(ty TokenType) Token {
	return Token{
		Type:  ty,
		Start: l.pos,
		End:   l.pos,
	}
}

func (l *Lexer) AllTokens() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, fmt.Errorf("scanning file: %w", err)
		}
		if tok.Type == EOF {
			return tokens, nil
		}
		tokens = append(tokens, tok)
	}
}

// NextToken scans the input for the next token. It returns the position of the token,
// the token's type, and the literal value.
func (l *Lexer) NextToken() (Token, error) {
	// keep looping until we return a token
	for {
		next := l.peek()
		if next == eof {
			return l.tokenOf(EOF), nil
		}

		if op, ok := operators[next]; ok {
			l.next()
			return l.tokenOf(op), nil
		}

		switch next {
		case '/':
			l.next()
			l.next()
			if l.ch != '/' && l.ch != '*' {
				return Token{}, l.errf("unexpected character %c", l.ch)
			}
			commentStart := l.pos
			lit, err := l.lexComment()
			if err != nil {
				return Token{}, err
			}
			return Token{
				Type:  COMMENT,
				Start: commentStart,
				End:   l.pos,
				Lit:   lit,
			}, nil

		case '"':
			startPos := l.pos
			lit, err := l.lexString()
			if err != nil {
				return Token{}, err
			}
			return Token{
				Type:  STRING,
				Start: startPos,
				End:   l.pos,
				Lit:   lit,
			}, nil

		case '|':
			startPos := l.pos
			lit, err := l.lexDescription()
			if err != nil {
				return Token{}, err
			}
			return Token{
				Type:  DESCRIPTION,
				Start: startPos,
				End:   l.pos,
				Lit:   lit,
			}, nil

		case '\n':
			l.next()
			return l.tokenOf(EOL), nil

		default:
			if unicode.IsSpace(next) {
				l.next()
				continue // nothing to do here, just move on
			} else if unicode.IsDigit(next) {
				// backup and let lexInt rescan the beginning of the int
				return l.lexNumber()
			} else if unicode.IsLetter(next) {
				startPos := l.pos
				lit := l.lexIdent()
				if keyword, ok := asKeyword(lit); ok {
					return Token{
						Type:  keyword,
						Start: startPos,
						End:   l.pos,
					}, nil
				}
				if lit == "true" || lit == "false" {
					return Token{
						Type:  BOOL,
						Start: startPos,
						End:   l.pos,
						Lit:   lit,
					}, nil
				}

				return Token{
					Type:  IDENT,
					Start: startPos,
					End:   l.pos,
					Lit:   lit,
				}, nil
			} else {
				l.next()
				return Token{}, l.errf("unexpected character: %c", next)
			}
		}
	}
}

// lexInt scans the input until the end of an integer and then returns the
// literal.
func (l *Lexer) lexNumber() (Token, error) {
	tt := Token{
		Type:  INT,
		Start: l.pos,
		End:   l.pos,
		Lit:   "",
	}
	var seenDot bool
	for {
		next := l.peek()
		if unicode.IsDigit(next) {
			l.next()
			tt.Lit = tt.Lit + string(l.ch)
		} else if next == '.' {
			if seenDot {
				return tt, l.errf("unexpected second dot in number literal")
			}
			l.next()
			seenDot = true
			tt.Type = DECIMAL
			tt.Lit = tt.Lit + string('.')

		} else {
			// scanned something not in the integer
			tt.End = l.pos
			return tt, nil
		}
	}
}

// lexIdent scans the input until the end of an identifier and then returns the
// literal.
func (l *Lexer) lexIdent() string {
	var lit string
	for {
		next := l.peek()
		if unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_' {
			l.next()
			lit = lit + string(l.ch)
		} else {
			// scanned something not in the identifier
			return lit
		}
	}
}

// lexString scans the input until the end of a string and then returns the
// literal.
func (l *Lexer) lexString() (string, error) {
	var lit string
	l.next()
	quote := l.ch
	for {
		next := l.peek()

		if next == eof {
			return "", l.unexpectedEOF()
		}
		if next == quote {
			l.next()
			// at the end of the string
			return lit, nil
		}
		if next == '\n' {
			return "", l.errf("unesacped newline in string")
		}

		if next == '\\' {
			if err := l.lexEscape(quote); err != nil {
				return "", err
			}
			// continue, having consumed the escape sequence, the next character
			// is just 'normal'
		}
		l.next()
		lit = lit + string(l.ch)
	}
}

// lexEscape scans the input for an escape sequence and returns an error if the
// escape sequence is invalid.
func (l *Lexer) lexEscape(quote rune) error {
	l.next()
	switch l.peek() {
	case '\\', '\n', quote:
		return nil
	}
	return l.errf("invalid escape sequence: \\%c", l.ch)
}

// lexDescription scans the input lines the next line is not a description
func (l *Lexer) lexDescription() (string, error) {
	var lit string
	for {
		line, err := l.lexDescriptionLine()
		if err != nil {
			return "", err
		}
		if l.peekPastWhitespace() != '|' {
			lit = lit + line
			return lit, nil
		}
		l.skipWhitespace() // leading whitespace on newline
		l.next()           // consume the |
		lit = lit + line + "\n"
	}
}

func (l *Lexer) lexDescriptionLine() (string, error) {

	var lit string
	l.next()
	l.skipWhitespace()
	for {
		next := l.peek()
		if next == eof {
			return lit, nil
		}
		if next == '\n' {
			return lit, nil
		}
		l.next()
		lit = lit + string(l.ch)
	}
}

// lineComment scans the input until the end of a line and then returns the
// literal.
func (l *Lexer) lexComment() (string, error) {
	if l.ch == '/' {
		return l.lexLineComment()
	}

	commentText := ""
	for {
		if l.peek() == '*' && l.peekAhead(1) == '/' {
			l.next()
			l.next()
			return commentText, nil
		}
		l.next()
		if l.ch == eof {
			return commentText, nil
		}
		commentText = commentText + string(l.ch)
	}
}

func (l *Lexer) lexLineComment() (string, error) {
	var lit string
	for {
		next := l.peek()
		if next == eof || next == '\n' {
			return lit, nil
		}
		l.next()
		lit = lit + string(l.ch)
	}
}

type PositionError struct {
	Position
	Msg string
}

func (e PositionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Position, e.Msg)
}

func (l *Lexer) unexpectedEOF() error {
	return l.errf("unexpected EOF")
}

func (l *Lexer) errf(format string, args ...interface{}) error {
	return &PositionError{
		Position: l.pos,
		Msg:      fmt.Sprintf(format, args...),
	}
}
