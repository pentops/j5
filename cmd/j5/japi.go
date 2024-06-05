package main

import "github.com/pentops/j5/cli"

var Version = "dev"

func main() {
	cmdGroup := cli.CommandSet()
	cmdGroup.RunMain("j5", Version)
}
