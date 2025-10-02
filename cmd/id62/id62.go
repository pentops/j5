package main

import (
	"fmt"
	"os"

	"github.com/pentops/j5/lib/id62"
)

func main() {
	var id string

	if len(os.Args) > 1 {
		id = id62.NewHash("", os.Args[1:]...).String()
	} else {
		id = id62.NewString()
	}

	fmt.Println(id)
}
