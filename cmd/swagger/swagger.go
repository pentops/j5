package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bufbuild/protoyaml-go"
	"github.com/pentops/custom-proto-api/gen/v1/jsonapi_pb"
	"github.com/pentops/custom-proto-api/structure"
)

func main() {
	src := flag.String("proto-src", "-", "Protobuf binary input file (- for stdin)")
	configFile := flag.String("config", "", "Config file to use")
	flag.Parse()

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

	json, err := json.Marshal(document)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(string(json))

}
