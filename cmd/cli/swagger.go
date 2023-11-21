package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pentops/jsonapi/gogen"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/jsonapi/swagger"
	"google.golang.org/protobuf/proto"
)

func main() {
	ctx := context.Background()

	args := os.Args[1:]
	if len(args) == 0 {
		usage()
	}

	switch args[0] {

	case "image":
		if err := runImage(ctx, os.Args[2:]); err != nil {
			log.Fatal(err.Error())
		}

	case "gocode":
		if err := runGocode(ctx, os.Args[2:]); err != nil {
			log.Fatal(err.Error())
		}

	case "jdef":
		if err := runJdef(ctx, os.Args[2:]); err != nil {
			log.Fatal(err.Error())
		}

	case "swagger":
		if err := runSwagger(ctx, os.Args[2:]); err != nil {
			log.Fatal(err.Error())
		}

	default:
		usage()
	}

}

func usage() {
	name := os.Args[0]
	fmt.Printf(`Usage: %s <command> [options]`, name)
	os.Exit(1)

}

func runGocode(ctx context.Context, osArgs []string) error {
	args := flag.NewFlagSet("swagger", flag.ExitOnError)
	src := args.String("src", ".", "Source directory containing jsonapi.yaml and buf.lock.yaml")
	outputDir := args.String("output-dir", "", "Directory to write go source")
	trimPackagePrefix := args.String("trim-package-prefix", "", "Prefix to trim from go package names")
	addGoPrefix := args.String("add-go-prefix", "", "Prefix to add to go package names")
	if err := args.Parse(osArgs); err != nil {
		return err
	}

	if *outputDir == "" {
		log.Fatal("output-dir is required")
	}

	image, err := structure.ReadImageFromSourceDir(ctx, *src)
	if err != nil {
		return err
	}

	jdefDoc, err := structure.BuildFromImage(image)
	if err != nil {
		return err
	}

	options := gogen.Options{
		TrimPackagePrefix: *trimPackagePrefix,
		AddGoPrefix:       *addGoPrefix,
	}

	if err := gogen.WriteGoCode(jdefDoc, *outputDir, options); err != nil {
		return err
	}

	return nil
}

func runImage(ctx context.Context, osArgs []string) error {
	args := flag.NewFlagSet("image", flag.ExitOnError)
	src := args.String("src", ".", "Source directory containing jsonapi.yaml and buf.lock.yaml")
	if err := args.Parse(osArgs); err != nil {
		return err
	}

	image, err := structure.ReadImageFromSourceDir(ctx, *src)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	fmt.Println(string(bb))
	return nil
}

func runSwagger(ctx context.Context, osArgs []string) error {
	args := flag.NewFlagSet("swagger", flag.ExitOnError)
	src := args.String("src", ".", "Source directory containing jsonapi.yaml and buf.lock.yaml")
	if err := args.Parse(osArgs); err != nil {
		return err
	}

	image, err := structure.ReadImageFromSourceDir(ctx, *src)
	if err != nil {
		return err
	}

	jdefDoc, err := structure.BuildFromImage(image)
	if err != nil {
		return err
	}

	swaggerDoc, err := swagger.BuildSwagger(jdefDoc)
	if err != nil {
		return err
	}

	asJson, err := json.Marshal(swaggerDoc)
	if err != nil {
		return err
	}

	fmt.Println(string(asJson))
	return nil
}

func runJdef(ctx context.Context, osArgs []string) error {
	args := flag.NewFlagSet("jdef", flag.ExitOnError)
	src := args.String("src", ".", "Source directory containing jsonapi.yaml and buf.lock.yaml")
	if err := args.Parse(osArgs); err != nil {
		return err
	}

	image, err := structure.ReadImageFromSourceDir(ctx, *src)
	if err != nil {
		log.Fatal(err.Error())
	}

	document, err := structure.BuildFromImage(image)
	if err != nil {
		log.Fatal(err.Error())
	}

	json, err := json.Marshal(document)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(string(json))
	return nil
}
