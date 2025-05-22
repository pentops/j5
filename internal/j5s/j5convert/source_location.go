package j5convert

import (
	"strings"

	"github.com/pentops/golib/gl"
)

type comment struct {
	path        []int32
	description *string
}

type commentSet []*comment

func (cs *commentSet) comment(path []int32, description string) {
	cc := &comment{
		path: path,
	}

	if description != "" {
		lines := strings.Split(description, "\n")
		joined := " " + strings.Join(lines, "\n ") + "\n"
		cc.description = gl.Ptr(joined)
	}
	*cs = append(*cs, cc)
}

// mergeAt adds the comments in the nested set to this set rooted at 'path'
func (cs *commentSet) mergeAt(path []int32, nested commentSet) {
	for _, input := range nested {

		thisPath := make([]int32, len(path)+len(input.path))
		copy(thisPath, path)
		copy(thisPath[len(path):], input.path)

		newComment := &comment{
			path:        thisPath,
			description: input.description,
		}
		*cs = append(*cs, newComment)
	}
}
