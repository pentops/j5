package main

import "github.com/pentops/jsonapi/cli"

var Version = "dev"

func main() {
	cmdGroup := cli.CommandSet()
	cmdGroup.RunMain("japi", Version)
}
