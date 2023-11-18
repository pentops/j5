package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/pentops/custom-proto-api/gen/v1/jsonapi_pb"
	"github.com/pentops/custom-proto-api/gogen"
	"github.com/pentops/custom-proto-api/jsonapi"
	"github.com/pentops/custom-proto-api/structure"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"

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

	codecOptions := jsonapi.Options{
		ShortEnums: &jsonapi.ShortEnumsOption{
			UnspecifiedSuffix: "UNSPECIFIED",
			StrictUnmarshal:   true,
		},
		WrapOneof: config.Options.WrapOneof,
	}

	if config.Options.ShortEnums != nil {
		codecOptions.ShortEnums = &jsonapi.ShortEnumsOption{
			UnspecifiedSuffix: config.Options.ShortEnums.UnspecifiedSuffix,
			StrictUnmarshal:   config.Options.ShortEnums.StrictUnmarshal,
		}
	}

	descriptors, err := structure.ReadFileDescriptorSet(*src)
	if err != nil {
		log.Fatal(err.Error())
	}

	services := make([]protoreflect.ServiceDescriptor, 0)
	descFiles, err := protodesc.NewFiles(descriptors)
	if err != nil {
		log.Fatal(err.Error())
	}

	descFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		fileServices := file.Services()
		for ii := 0; ii < fileServices.Len(); ii++ {
			service := fileServices.Get(ii)
			services = append(services, service)
		}
		return true
	})

	filteredServices := make([]protoreflect.ServiceDescriptor, 0)
	for _, service := range services {
		name := service.FullName()
		if !strings.HasSuffix(string(name), "Service") {
			continue
		}

		filteredServices = append(filteredServices, service)
	}

	document, err := structure.Build(codecOptions, filteredServices)
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
