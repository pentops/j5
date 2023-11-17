package main

import (
	"flag"
	"log"

	"github.com/pentops/custom-proto-api/gogen"
	"github.com/pentops/custom-proto-api/jsonapi"
	"github.com/pentops/custom-proto-api/structure"
)

func main() {
	src := flag.String("proto-src", "-", "Protobuf binary input file (- for stdin)")
	outputDir := flag.String("output-dir", "", "Directory to write go source")
	trimPackagePrefix := flag.String("trim-package-prefix", "", "Prefix to trim from go package names")
	addGoPrefix := flag.String("add-go-prefix", "", "Prefix to add to go package names")
	flag.Parse()

	if *outputDir == "" {
		log.Fatal("output-dir is required")
	}

	codecOptions := jsonapi.Options{
		ShortEnums: &jsonapi.ShortEnumsOption{
			UnspecifiedSuffix: "UNSPECIFIED",
			StrictUnmarshal:   true,
		},
		WrapOneof: true,
	}

	descriptors, err := structure.ReadFileDescriptorSet(*src)
	if err != nil {
		log.Fatal(err.Error())
	}

	document, err := structure.BuildFromDescriptors(codecOptions, descriptors)
	if err != nil {
		log.Fatal(err.Error())
	}

	options := gogen.Options{
		TrimPackagePrefix: *trimPackagePrefix,
		AddGoPrefix:       *addGoPrefix,
	}

	if err := gogen.WriteGoCode(document, *outputDir, options); err != nil {
		log.Fatal(err.Error())
	}

}
