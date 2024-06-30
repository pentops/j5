package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/schema/export"
	"github.com/pentops/j5/schema/structure"
	"github.com/pentops/log.go/log"
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/encoding/protojson"
)

func schemaSet() *commander.CommandSet {
	genGroup := commander.NewCommandSet()
	genGroup.Add("image", commander.NewCommand(RunImage))
	genGroup.Add("jdef", commander.NewCommand(RunJDef))
	genGroup.Add("descriptor", commander.NewCommand(RunDescriptor))
	genGroup.Add("swagger", commander.NewCommand(RunSwagger))
	return genGroup
}

type BuildConfig struct {
	SourceConfig
	Output string `flag:"output" default:"-" description:"Destination to push image to. - for stdout, s3://bucket/prefix, otherwise a file"`
}

func (cfg BuildConfig) sourceImage(ctx context.Context) (*source_j5pb.SourceImage, error) {
	source, err := cfg.GetInput(ctx)
	if err != nil {
		return nil, fmt.Errorf("getSource: %w", err)
	}

	image, err := source.SourceImage(ctx)
	if err != nil {
		return nil, fmt.Errorf("SourceImage: %w", err)
	}

	return image, nil
}

func (cfg BuildConfig) descriptorAPI(ctx context.Context) (*schema_j5pb.API, error) {
	image, err := cfg.sourceImage(ctx)
	if err != nil {
		return nil, err
	}

	reflectionAPI, err := structure.ReflectFromSource(image)
	if err != nil {
		return nil, fmt.Errorf("ReflectFromSource: %w", err)
	}

	descriptorAPI, err := reflectionAPI.ToJ5Proto()
	if err != nil {
		return nil, fmt.Errorf("DescriptorFromReflection: %w", err)
	}

	if err := structure.ResolveProse(image, descriptorAPI); err != nil {
		return nil, fmt.Errorf("ResolveProse: %w", err)
	}

	return descriptorAPI, nil
}

func RunImage(ctx context.Context, cfg BuildConfig) error {
	image, err := cfg.sourceImage(ctx)
	if err != nil {
		return err
	}

	bb, err := protojson.Marshal(image)
	if err != nil {
		return err
	}
	return writeBytes(ctx, cfg.Output, bb)
}

func RunDescriptor(ctx context.Context, cfg BuildConfig) error {

	descriptorAPI, err := cfg.descriptorAPI(ctx)
	if err != nil {
		return err
	}

	bb, err := protojson.Marshal(descriptorAPI)
	if err != nil {
		return err
	}

	return writeBytes(ctx, cfg.Output, bb)
}

func RunSwagger(ctx context.Context, cfg BuildConfig) error {
	descriptorAPI, err := cfg.descriptorAPI(ctx)
	if err != nil {
		return err
	}

	swaggerDoc, err := export.BuildSwagger(descriptorAPI)
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
	descriptorAPI, err := cfg.descriptorAPI(ctx)
	if err != nil {
		return err
	}

	jDefJSON, err := export.FromProto(descriptorAPI)
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

		log.WithField(ctx, "dest", dest).Debug("Uploading to S3")

		// url.Parse will take s3://foobucket/keyname and turn keyname into "/keyname" which we want to be "keyname"
		k := strings.Replace(s3URL.Path, "/", "", 1)

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &s3URL.Host,
			Key:    &k,
			Body:   strings.NewReader(string(bb)),
		})

		if err != nil {
			return fmt.Errorf("failed to upload to s3://%s/%s: %w", s3URL.Host, k, err)
		}
	}

	return nil
}
