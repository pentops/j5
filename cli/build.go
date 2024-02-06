package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pentops/jsonapi/source"
	"github.com/pentops/jsonapi/structure"
	"github.com/pentops/jsonapi/structure/jdef"
	"github.com/pentops/jsonapi/swagger"
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/encoding/protojson"
)

func BuildSet() *commander.CommandSet {
	genGroup := commander.NewCommandSet()
	genGroup.Add("image", commander.NewCommand(RunImage))
	genGroup.Add("jdef", commander.NewCommand(RunJDef))
	genGroup.Add("swagger", commander.NewCommand(RunSwagger))
	return genGroup
}

type BuildConfig struct {
	Source string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
	Output string `flag:"output" default:"-" description:"Destination to push image to. - for stdout, s3://bucket/prefix, otherwise a file"`
}

func RunImage(ctx context.Context, cfg BuildConfig) error {
	image, err := source.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	bb, err := protojson.Marshal(image)
	if err != nil {
		return err
	}
	return writeBytes(ctx, cfg.Output, bb)

}

func RunSwagger(ctx context.Context, cfg BuildConfig) error {
	image, err := source.ReadImageFromSourceDir(ctx, cfg.Source)
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

func RunJDef(ctx context.Context, cfg BuildConfig) error {
	image, err := source.ReadImageFromSourceDir(ctx, cfg.Source)
	if err != nil {
		return err
	}

	document, err := structure.BuildFromImage(image)
	if err != nil {
		return err
	}

	jDefJSON, err := jdef.FromProto(document)
	if err != nil {
		return err
	}

	jDefJSONBytes, err := json.Marshal(jDefJSON)
	if err != nil {
		return err
	}

	return writeBytes(ctx, cfg.Output, jDefJSONBytes)
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
