package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	case "push":
		if err := runPush(ctx, os.Args[2:]); err != nil {
			log.Fatal(err.Error())
		}

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

	output := gogen.DirFileWriter(*outputDir)

	if err := gogen.WriteGoCode(jdefDoc, output, options); err != nil {
		return err
	}

	return nil
}

func runPush(ctx context.Context, osArgs []string) error {
	args := flag.NewFlagSet("image", flag.ExitOnError)
	src := args.String("src", ".", "Source directory containing jsonapi.yaml and buf.lock.yaml")
	version := args.String("version", "", "Version to push")
	latest := args.Bool("latest", false, "Push as latest")
	bucket := args.String("bucket", "", "S3 bucket to push to")
	prefix := args.String("prefix", "", "S3 prefix to push to")
	if err := args.Parse(osArgs); err != nil {
		return err
	}

	if *bucket == "" {
		return fmt.Errorf("bucket is required")
	}

	if (!*latest) && *version == "" {
		return fmt.Errorf("version, latest or both are required")
	}

	image, err := structure.ReadImageFromSourceDir(ctx, *src)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	versions := []string{}

	if *latest {
		versions = append(versions, "latest")
	}

	if *version != "" {
		versions = append(versions, *version)
	}

	destinations := make([]string, len(versions))
	for i, version := range versions {
		p := path.Join(*prefix, image.Registry.Organization, image.Registry.Name, version, "image.bin")
		destinations[i] = fmt.Sprintf("s3://%s/%s", *bucket, p)
	}

	return pushS3(ctx, bb, destinations...)

}

func runImage(ctx context.Context, osArgs []string) error {
	args := flag.NewFlagSet("image", flag.ExitOnError)
	src := args.String("src", ".", "Source directory containing jsonapi.yaml and buf.lock.yaml")
	pushDest := args.String("output", "-", "Destination to push image to. - for stdout, s3://bucket/prefix, otherwise a file")
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

	if *pushDest == "-" {
		os.Stdout.Write(bb)
		return nil
	}

	if strings.HasPrefix(*pushDest, "s3://") {
		return pushS3(ctx, bb, *pushDest)
	}

	return os.WriteFile(*pushDest, bb, 0644)
}

func pushS3(ctx context.Context, bb []byte, destinations ...string) error {

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)
	for _, dest := range destinations {
		s3URL, err := url.Parse(dest)
		if err != nil {
			return err
		}
		if s3URL.Scheme != "s3" || s3URL.Host == "" {
			return fmt.Errorf("invalid s3 url: %s", dest)
		}

		log.Printf("Uploading to %s", dest)

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &s3URL.Host,
			Key:    &s3URL.Path,
			Body:   strings.NewReader(string(bb)),
		})

		if err != nil {
			return fmt.Errorf("failed to upload object: %w", err)
		}
	}

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
