package structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
)

type ProseResolver interface {
	ResolveProse(filename string) (string, error)
}

type DirResolver string

func (dr DirResolver) ResolveProse(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(string(dr), filename))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type mapResolver map[string]string

func (mr mapResolver) ResolveProse(filename string) (string, error) {
	data, ok := mr[filename]
	if !ok {
		return "", fmt.Errorf("prose file %q not found", filename)
	}
	return data, nil
}

func removeMarkdownHeader(data string) string {
	// only look at the first 5 lines, that should be well enough to deal with
	// both title formats (# or \n===), and a few trailing empty lines

	lines := strings.SplitN(data, "\n", 5)
	if len(lines) == 0 {
		return ""
	}
	if strings.HasPrefix(lines[0], "# ") {
		lines = lines[1:]
	} else if strings.HasPrefix(lines[1], "==") {
		lines = lines[2:]
	}

	// Remove any leading empty lines
	for len(lines) > 1 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	return strings.Join(lines, "\n")
}

func imageResolver(proseFiles []*source_j5pb.ProseFile) ProseResolver {
	mr := make(mapResolver)
	for _, proseFile := range proseFiles {
		mr[proseFile.Path] = string(proseFile.Content)
	}
	return mr
}
