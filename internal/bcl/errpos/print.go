package errpos

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorsWithSource struct {
	lines  map[string][]string
	Errors Errors
}

func (e ErrorsWithSource) HumanString(contextLines int) string {
	if len(e.Errors) == 0 {
		// should not happen, this is not an error.
		return "<ErrorsWithSource - no errors>"
	}

	lines := make([]string, 0)

	for idx, err := range e.Errors {
		if idx > 0 {
			lines = append(lines, "-----")
		}

		var srcLines []string
		if err.Pos != nil && err.Pos.Filename != nil {
			if fileLines, ok := e.lines[*err.Pos.Filename]; ok {
				srcLines = fileLines
			} else {
				lines = append(lines, fmt.Sprintf("<no source lines for %s>", *err.Pos.Filename))
			}
		}

		str := humanString(err, srcLines, contextLines)
		lines = append(lines, str)
	}

	return strings.Join(lines, "\n")
}

func (e ErrorsWithSource) ShortString() string {
	lines := make([]string, 0)
	for _, err := range e.Errors {
		lines = append(lines, shortString(err))
	}
	return strings.Join(lines, "\n")

}

func (e *ErrorsWithSource) Append(err *ErrorsWithSource) {
	if e.lines == nil {
		e.lines = make(map[string][]string)
	}

	for filename, lines := range err.lines {
		if _, ok := e.lines[filename]; !ok {
			e.lines[filename] = lines
		}
	}

	e.Errors = append(e.Errors, err.Errors...)
}

func (e ErrorsWithSource) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	if len(e.Errors) == 0 {
		// should not happen, this is not an error.
		return "<ErrorsWithWource - no errors>"
	}

	return fmt.Sprintf("<ErrorsWithSource - %d errors>", len(e.Errors))
}

func AsErrorsWithSource(err error) (*ErrorsWithSource, bool) {
	var posErr *ErrorsWithSource
	ok := errors.As(err, &posErr)
	if ok {
		return posErr, true
	}

	return nil, false
}

func shortString(err *Err) string {

	if err.Pos == nil || err.Pos.isEmpty() {
		return fmt.Sprintf("? %s", err.Err.Error())
	} else {
		return fmt.Sprintf("%s %s", err.Pos.String(), err.Err.Error())

	}
}

func humanString(err *Err, lines []string, context int) string {
	out := &strings.Builder{}

	func() {
		if err.Pos == nil || err.Pos.isEmpty() {
			out.WriteString("<no position information>")
			out.WriteString("\n")
			return
		}

		fmt.Fprintf(out, "Position: %s\n", err.Pos.String())

		if err.Pos.Start.isEmpty() {
			return
		}

		pos := *err.Pos

		startLine := pos.Start.Line + 1
		startCol := pos.Start.Column + 1
		if startLine > len(lines) {
			fmt.Fprintf(out, "<line %d out of range (len %d) - a>", startLine, len(lines))
			out.WriteString("\n")
			return
		}

		for lineNum := startLine - context; lineNum < startLine; lineNum++ {
			if lineNum < 1 {
				continue
			}
			line := lines[lineNum-1]
			fmt.Fprintf(out, "  > %03d: ", lineNum)
			out.WriteString(tabsToSpaces(line))
			out.WriteString("\n")
			context--
		}

		if startLine > len(lines) || startLine < 1 {
			fmt.Fprintf(out, "<line %d out of range (len %d) - b>", startLine, len(lines))
			out.WriteString("\n")
			return
		}

		errLine := lines[startLine-1]

		prefix := fmt.Sprintf("  > %03d", startLine)
		out.WriteString(prefix)
		out.WriteString(": ")
		out.WriteString(tabsToSpaces(errLine))
		out.WriteString("\n")

		if startCol == len(errLine)+1 {
			// allows for the column to reference the EOF or EOL
			errLine += " "
		}

		if startCol < 1 || startCol > len(errLine) {
			// negative columns should not occur but let's not crash.
			out.WriteString(strings.Repeat(">", len(prefix)))
			out.WriteString(": ")
			fmt.Fprintf(out, "<column %d out of range>\n", startCol)
			out.WriteString("\n")
			return
		}

		errCol := replaceRunes(errLine[:startCol-1], func(r string) string {
			if r == "\t" {
				return "  "
			}
			return " "
		})

		out.WriteString(strings.Repeat(">", len(prefix)))
		out.WriteString(": ")
		out.WriteString(errCol)
		out.WriteString("^\n")

	}()
	if err.Ctx != nil {
		out.WriteString("Context: ")
		out.WriteString(err.Ctx.String())
		out.WriteString("\n")
	}
	if err.Err != nil {
		out.WriteString("Message: ")
		out.WriteString(err.Err.Error())
		out.WriteString("\n")
	}
	return out.String()
}

func tabsToSpaces(s string) string {
	return replaceRunes(s, func(r string) string {
		if r == "\t" {
			return "  "
		}
		return r
	})
}

func replaceRunes(s string, cb func(string) string) string {
	runes := []rune(s)
	out := make([]string, 0, len(runes))
	for i := range runes {
		out = append(out, cb(string(runes[i])))
	}
	return strings.Join(out, "")
}

func MustAddSource(err error, filename string, fileSource []byte) (*ErrorsWithSource, error) {
	input, ok := AsErrors(err)
	if !ok {
		return nil, fmt.Errorf("error not valid for source: (%T) %w", err, err)
	}

	return &ErrorsWithSource{
		lines: map[string][]string{
			filename: strings.Split(string(fileSource), "\n"),
		},
		Errors: input,
	}, nil
}
