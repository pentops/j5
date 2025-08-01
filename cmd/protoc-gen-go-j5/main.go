package main

import (
	"flag"
	"fmt"

	"github.com/pentops/j5/internal/builder/protogen/j5go"
	"google.golang.org/protobuf/compiler/protogen"
)

var Version = "1.0"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-psm %v\n", Version)
		return
	}

	var flags flag.FlagSet
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(j5go.ProtocPlugin())
}
