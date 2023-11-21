package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/pentops/custom-proto-api/gen/v1/jsonapi_pb"
	"github.com/pentops/custom-proto-api/gogen"
	"github.com/pentops/custom-proto-api/structure"

	protoyaml "github.com/bufbuild/protoyaml-go"
)

func main() {
	src := flag.String("proto-src", "-", "Protobuf binary input file (- for stdin)")
	configFile := flag.String("config", "", "Config file to use")
	outputDir := flag.String("output-dir", "", "Directory to write go source")
	trimPackagePrefix := flag.String("trim-package-prefix", "", "Prefix to trim from go package names")
	addGoPrefix := flag.String("add-go-prefix", "", "Prefix to add to go package names")
	flag.Parse()

	if *outputDir == "" {
		log.Fatal("output-dir is required")
	}
	if *configFile == "" {
		log.Fatal("config is required")
	}

	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	config := &jsonapi_pb.Config{}
	if err := protoyaml.Unmarshal(configData, config); err != nil {
		log.Fatal(err.Error())
	}

	descriptors, err := structure.ReadFileDescriptorSet(*src)
	if err != nil {
		log.Fatal(err.Error())
	}

	document, err := structure.BuildFromDescriptors(config, descriptors, structure.DirResolver(filepath.Dir(*configFile)))
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
