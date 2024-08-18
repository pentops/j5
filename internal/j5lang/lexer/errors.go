package lexer

import (
	"errors"
	"fmt"
	"strings"
)

type errorWithPosition interface {
	error
	ErrorPosition() (Position, string)
}

func positionOfError(err error) (*Position, string) {
	var posErr errorWithPosition
	ok := errors.As(err, &posErr)
	if !ok {
		return nil, ""
	}
	pos, msg := posErr.ErrorPosition()
	return &pos, msg
}

type PositionError struct {
	Position
	Msg string
}

func (e PositionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Position, e.Msg)
}

func (e PositionError) ErrorPosition() (Position, string) {
	return e.Position, e.Msg
}

type PositionErrors []PositionError

func (e PositionErrors) Error() string {
	if len(e) > 0 {
		return e[0].Error()
	}
	return "syntax errors"
}

type PositionErrorWithSource struct {
	Position   Position
	Msg        string
	SourceLine string
}

func (e *PositionErrorWithSource) Error() string {
	return fmt.Sprintf("%s: %s", e.Position, e.Msg)
}

func (e *PositionErrorWithSource) ErrorPosition() (Position, string) {
	return e.Position, e.Msg
}

type PositionErrorsWithSource []PositionErrorWithSource

func (e PositionErrorsWithSource) Print(printf func(string, ...any)) {
	for idx, err := range e {
		if idx > 0 {
			printf("-----\n")
		}
		err.Print(printf)
	}
}

func (e PositionErrorsWithSource) Error() string {
	if len(e) > 0 {
		return e[0].Error()
	}
	return "syntax errors"
}

func AddPositionErrorSource(err error, input string) error {
	var positionErrors PositionErrors
	if !errors.As(err, &positionErrors) {

		var positionErr *PositionError
		if errors.As(err, &positionErr) {
			positionErrors = PositionErrors{*positionErr}
		}

		position, msg := positionOfError(err)
		if position == nil {
			return err
		}

		positionErrors = PositionErrors{{
			Position: *position,
			Msg:      msg,
		}}

	}

	lines := strings.Split(input, "\n")

	errors := make(PositionErrorsWithSource, 0, len(positionErrors))

	for _, srcErr := range positionErrors {

		pe := &PositionErrorWithSource{
			Position: srcErr.Position,
			Msg:      srcErr.Msg,
		}

		if srcErr.Line > len(lines) {
			pe.SourceLine = "<out of range>"
			return pe
		}
		pe.SourceLine = lines[srcErr.Position.Line-1]
		errors = append(errors, *pe)
	}

	return errors

}

func (e *PositionErrorWithSource) Print(printf func(string, ...any)) {
	printf("%v", e.Msg)

	line := replaceRunes(e.SourceLine, func(r string) string {
		// Replace tabs with double space for console consistency
		if r == "\t" {
			return "  "
		}
		return r
	})
	lineNum := fmt.Sprintf("line %d: ", e.Position.Line)
	printf("%s%s", lineNum, line)
	if e.Position.Column < 1 {
		// negative columns should not occur but let's not crash.
		printf("<column out of range>")
		return
	}

	prefix := replaceRunes(lineNum+e.SourceLine[:e.Position.Column-1], func(r string) string {
		if r == "\t" {
			return "  "
		}
		return " "
	})

	printf("%s^", prefix)
}

func replaceRunes(s string, cb func(string) string) string {
	runes := []rune(s)
	out := make([]string, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		out = append(out, cb(string(runes[i])))
	}
	return strings.Join(out, "")
}

func AsPositionErrorsWithSource(err error) (PositionErrorsWithSource, bool) {
	var posErr PositionErrorsWithSource
	ok := errors.As(err, &posErr)
	if !ok {
		return nil, false
	}
	return posErr, ok
}
