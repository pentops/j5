package main

import (
	"context"
	"encoding/json"
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
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var Version = "dev"

func main() {

	cmdGroup := commander.NewCommandSet()

	cmdGroup.Add("push", commander.NewCommand(runPush))

	genGroup := commander.NewCommandSet()
	genGroup.Add("image", commander.NewCommand(runImage))
	genGroup.Add("gocode", commander.NewCommand(runGocode))
	genGroup.Add("jdef", commander.NewCommand(runJdef))
	genGroup.Add("swagger", commander.NewCommand(runSwagger))
	genGroup.Add("code-generator-request", commander.NewCommand(runCodeGeneratorRequest))

	cmdGroup.Add("generate", genGroup)

	cmdGroup.RunMain("japi", Version)
}

func runGocode(ctx context.Context, cfg struct {
	Source            string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	OutputDir         string `flag:"output-dir" description:"Directory to write go source"`
	TrimPackagePrefix string `flag:"trim-package-prefix" default:"" description:"Prefix to trim from go package names"`
	AddGoPrefix       string `flag:"add-go-prefix" default:"" description:"Prefix to add to go package names"`
}) error {
	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	jdefDoc, err := structure.BuildFromImage(image)
	if err != nil {
		return err
	}

	options := gogen.Options{
		TrimPackagePrefix: cfg.TrimPackagePrefix,
		AddGoPrefix:       cfg.AddGoPrefix,
	}

	output := gogen.DirFileWriter(cfg.OutputDir)

	if err := gogen.WriteGoCode(jdefDoc, output, options); err != nil {
		return err
	}

	return nil
}

func runPush(ctx context.Context, cfg struct {
	Source  string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Version string `flag:"version" default:"" description:"Version to push"`
	Latest  bool   `flag:"latest" description:"Push as latest"`
	Bucket  string `flag:"bucket" description:"S3 bucket to push to"`
	Prefix  string `flag:"prefix" description:"S3 prefix to push to"`
}) error {

	if (!cfg.Latest) && cfg.Version == "" {
		return fmt.Errorf("version, latest or both are required")
	}

	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}

	versions := []string{}

	if cfg.Latest {
		versions = append(versions, "latest")
	}

	if cfg.Version != "" {
		versions = append(versions, cfg.Version)
	}

	destinations := make([]string, len(versions))
	for i, version := range versions {
		p := path.Join(cfg.Prefix, image.Registry.Organization, image.Registry.Name, version, "image.bin")
		destinations[i] = fmt.Sprintf("s3://%s/%s", cfg.Bucket, p)
	}

	return pushS3(ctx, bb, destinations...)

}

func runImage(ctx context.Context, cfg struct {
	Source string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Output string `flag:"output" default:"-" description:"Destination to push image to. - for stdout, s3://bucket/prefix, otherwise a file"`
}) error {

	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	bb, err := proto.Marshal(image)
	if err != nil {
		return err
	}
	return writeBytes(ctx, cfg.Output, bb)

}

func writeBytes(ctx context.Context, to string, data []byte) error {
	if to == "-" {
		os.Stdout.Write(data)
		return nil
	}

	if strings.HasPrefix(to, "s3://") {
		return pushS3(ctx, data, to)
	}

	return os.WriteFile(to, data, 0644)
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

		// url.Parse will take s3://foobucket/keyname and turn keyname into "/keyname" which we want to be "keyname"
		k := strings.Replace(s3URL.Path, "/", "", 1)

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &s3URL.Host,
			Key:    &k,
			Body:   strings.NewReader(string(bb)),
		})

		if err != nil {
			return fmt.Errorf("failed to upload object: %w", err)
		}
	}

	return nil
}

func runSwagger(ctx context.Context, cfg struct {
	Source string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Output string `flag:"output" default:"-" description:"Destination to push image to. - for stdout, s3://bucket/key, otherwise a file"`
}) error {

	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
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

	return writeBytes(ctx, cfg.Output, asJson)
}

func runJdef(ctx context.Context, cfg struct {
	Source string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Output string `flag:"output" default:"-" description:"Destination to push json image to. - for stdout, s3://bucket/key, otherwise a local file"`
}) error {

	image, err := structure.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		log.Fatal(err.Error())
	}

	document, err := structure.BuildFromImage(image)
	if err != nil {
		log.Fatal(err.Error())
	}

	asJson, err := json.Marshal(document)
	if err != nil {
		log.Fatal(err.Error())
	}

	return writeBytes(ctx, cfg.Output, asJson)
}

func runCodeGeneratorRequest(ctx context.Context, cfg struct {
	Source        string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Output        string `flag:"output" default:"-" description:"Destination to push proto binary image to. - for stdout, s3://bucket/key, otherwise a local file. Format determined by suffix: .txt = textproto .json = protojson, else proto binary"`
	PackagePrefix string `flag:"package-prefix" default:"" description:"proto package prefix to generate"`
}) error {

	descriptors, err := structure.ReadFileDescriptorSet(ctx, cfg.Source)
	if err != nil {
		return err
	}

	options := structure.CodeGenOptions{
		PackagePrefix: cfg.PackagePrefix,
	}
	genRequest, err := structure.CodeGeneratorRequestFromDescriptors(options, descriptors)
	if err != nil {
		return err
	}

	if strings.HasSuffix(cfg.Output, ".txt") {
		return writeBytes(ctx, cfg.Output, []byte(prototext.Format(genRequest)))
	}

	if strings.HasSuffix(cfg.Output, ".json") {
		return writeBytes(ctx, cfg.Output, []byte(protojson.Format(genRequest)))
	}

	bb, err := proto.Marshal(genRequest)
	if err != nil {
		return err
	}

	return writeBytes(ctx, cfg.Output, bb)
}
