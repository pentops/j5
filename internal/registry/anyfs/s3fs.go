package anyfs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pentops/log.go/log"
)

var ErrNotFound = fmt.Errorf("not found")

type S3API interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, options ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, input *s3.PutObjectInput, options ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(ctx context.Context, input *s3.HeadObjectInput, options ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2(ctx context.Context, input *s3.ListObjectsV2Input, options ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3FS struct {
	bucket string
	prefix string
	client S3API
}

func NewS3EnvFS(ctx context.Context, location string) (*S3FS, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	return NewS3FS(s3Client, location)
}

func NewS3FS(client S3API, location string) (*S3FS, error) {
	bucketURL, err := url.Parse(location)
	if err != nil {
		return nil, err
	}

	if bucketURL.Scheme != "s3" {
		return nil, fmt.Errorf("bucket must be an s3:// url")
	}

	bucketName := bucketURL.Host
	if bucketName == "" {
		return nil, fmt.Errorf("bucket must be an s3:// url")
	}

	keyPrefix := strings.TrimPrefix(bucketURL.Path, "/")

	return &S3FS{
		bucket: bucketName,
		prefix: keyPrefix,
		client: client,
	}, nil
}

func (s3fs *S3FS) Join(paths ...string) string {
	return path.Join(paths...)
}

func (s3fs *S3FS) Put(ctx context.Context, subPath string, body io.Reader, metadata map[string]string) error {
	key := path.Join(s3fs.prefix, subPath)
	ctx = log.WithFields(ctx, map[string]any{
		"s3Bucket": s3fs.bucket,
		"s3Key":    key,
	})
	log.Debug(ctx, "uploading to s3")

	_, err := s3fs.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   &s3fs.bucket,
		Key:      &key,
		Body:     body,
		Metadata: metadata,
	})
	if err != nil {
		log.WithError(ctx, err).Error("failed to upload to s3")
		return fmt.Errorf("failed to upload to s3: 's3://%s/%s' : %w", s3fs.bucket, key, err)
	}
	log.Info(ctx, "uploaded to s3")
	return nil
}

func (s3fs *S3FS) Head(ctx context.Context, subPath string) (*FileInfo, error) {
	key := path.Join(s3fs.prefix, subPath)
	ctx = log.WithFields(ctx, map[string]any{
		"s3Bucket": s3fs.bucket,
		"s3Key":    key,
	})
	log.Debug(ctx, "s3 Head")
	head, err := s3fs.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s3fs.bucket,
		Key:    &key,
	})
	if err != nil {
		log.WithError(ctx, err).Error("s3 Head Error")
		if strings.Contains(err.Error(), "NotFound") {
			return nil, ErrNotFound
		}
		return nil, err
	}

	log.Info(ctx, "s3 Head Success")

	return &FileInfo{
		Size:     *head.ContentLength,
		Metadata: head.Metadata,
	}, nil

}

func (s3fs *S3FS) GetReader(ctx context.Context, subPath string) (io.ReadCloser, *FileInfo, error) {
	key := path.Join(s3fs.prefix, subPath)
	ctx = log.WithFields(ctx, map[string]any{
		"s3Bucket": s3fs.bucket,
		"s3Key":    key,
	})
	log.Debug(ctx, "s3 Get")

	obj, err := s3fs.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s3fs.bucket,
		Key:    &key,
	})
	if err != nil {
		log.WithError(ctx, err).Error("s3 Get Error")
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	log.Info(ctx, "s3 Get Success")

	fileInfo := &FileInfo{
		Size:     *obj.ContentLength,
		Metadata: obj.Metadata,
	}

	return obj.Body, fileInfo, nil
}

func (s3fs *S3FS) GetBytes(ctx context.Context, subPath string) ([]byte, error) {
	body, _, err := s3fs.GetReader(ctx, subPath)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

type ListInfo struct {
	Name string
	Size int64
}

func (s3fs *S3FS) List(ctx context.Context, subPath string) ([]ListInfo, error) {
	key := path.Join(s3fs.prefix, subPath)
	ctx = log.WithFields(ctx, map[string]any{
		"s3Bucket": s3fs.bucket,
		"s3Key":    key,
	})
	log.Debug(ctx, "s3 List")

	var results []ListInfo

	var continuationToken *string
	for {
		page, err := s3fs.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &s3fs.bucket,
			Prefix:            &key,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			log.WithError(ctx, err).Error("s3 List Error")
			return nil, fmt.Errorf("list s3://%s/%s: %w", s3fs.bucket, key, err)
		}

		for _, obj := range page.Contents {
			results = append(results, ListInfo{
				Size: *obj.Size,
				Name: strings.TrimPrefix(*obj.Key, s3fs.prefix+"/"),
			})
		}
		if page.NextContinuationToken == nil {
			break
		}
		continuationToken = page.NextContinuationToken
	}

	log.Info(ctx, "s3 List Success")

	return results, nil
}
