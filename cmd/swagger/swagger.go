package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pentops/custom-proto-api/jsonapi"
	"github.com/pentops/custom-proto-api/swagger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func main() {
	src := flag.String("proto-src", "-", "Protobuf binary input file (- for stdin)")
	flag.Parse()
	descriptors := &descriptorpb.FileDescriptorSet{}

	if *src == "-" {
		protoData, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err.Error())
		}
		if err := proto.Unmarshal(protoData, descriptors); err != nil {
			log.Fatal(err.Error())
		}
	} else {
		protoData, err := os.ReadFile(*src)
		if err != nil {
			log.Fatal(err.Error())
		}
		if err := proto.Unmarshal(protoData, descriptors); err != nil {
			log.Fatal(err.Error())
		}
	}

	options := jsonapi.Options{}
	document, err := swagger.BuildFromDescriptors(options, descriptors)
	if err != nil {
		log.Fatal(err.Error())
	}

	json, err := json.Marshal(document)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(string(json))

}
