package main

import (
	"fmt"

	"github.com/pentops/j5/lib/id62"
)

func main() {
	id := id62.NewString()
	fmt.Println(id)
}
