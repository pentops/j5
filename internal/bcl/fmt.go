package bcl

import "github.com/pentops/j5/internal/bcl/internal/parser"

func Fmt(filename string, data string) (string, error) {
	fixed, err := parser.Fmt(filename, data)
	if err != nil {
		return "", err
	}
	return fixed, nil
}
