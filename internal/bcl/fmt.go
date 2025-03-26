package bcl

import "github.com/pentops/j5/internal/bcl/internal/parser"

func Fmt(data string) (string, error) {
	fixed, err := parser.Fmt(string(data))
	if err != nil {
		return "", err
	}
	return fixed, nil
}
