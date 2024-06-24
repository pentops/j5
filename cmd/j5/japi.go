package main

import "github.com/pentops/j5/internal/cli"

var Version = "dev"

func main() {
	cmdGroup := cli.CommandSet()
	cmdGroup.RunMain("j5", Version)
}
