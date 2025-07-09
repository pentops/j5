package gomodproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pentops/log.go/log"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

func RunUploadCommand(ctx context.Context, cfg struct {
	SrcDir     string   `env:"SRC_DIR" flag:"src" default:"."`
	DstDir     string   `env:"GOMOD_REMOTE" flag:"remote"`
	CommitHash string   `env:"COMMIT_HASH" flag:"commit-hash" default:""`
	CommitTime string   `env:"COMMIT_TIME" flag:"commit-time" default:""`
	Alias      []string `env:"ALIAS" flag:"alias" default:""`
}) error {

	modfileSrc, err := os.ReadFile(filepath.Join(cfg.SrcDir, "go.mod"))
	if err != nil {
		return err
	}

	parsed, err := modfile.Parse("go.mod", modfileSrc, nil)
	if err != nil {
		return err
	}

	packageName := parsed.Module.Mod.Path

	commitTime, err := time.Parse(time.RFC3339, cfg.CommitTime)
	if err != nil {
		return fmt.Errorf("parsing commit time: %w", err)
	}

	commitHashPrefix := cfg.CommitHash
	if len(commitHashPrefix) > 12 {
		commitHashPrefix = commitHashPrefix[:12]
	}

	canonicalVersion := module.PseudoVersion("", "", commitTime, commitHashPrefix)

	var s3Dest string
	var fullDest string

	if strings.HasPrefix(cfg.DstDir, "s3://") {
		s3Dest = cfg.DstDir
		tmpDir, err := os.MkdirTemp("", "mod-proxy")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		fullDest = tmpDir

	} else {
		fullDest = filepath.Join(cfg.DstDir, packageName)
		if err := os.MkdirAll(fullDest, 0755); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(fullDest, canonicalVersion+".mod"), modfileSrc, 0644); err != nil {
		return err
	}

	info := Info{
		Version: canonicalVersion,
		Time:    commitTime,
	}

	outWriter, err := os.Create(filepath.Join(fullDest, canonicalVersion+".zip"))
	if err != nil {
		return err
	}
	defer outWriter.Close()

	if err := zip.CreateFromDir(outWriter, module.Version{
		Path:    parsed.Module.Mod.Path,
		Version: canonicalVersion,
	}, cfg.SrcDir); err != nil {
		return err
	}

	if s3Dest != "" {
		if err := s3Copy(ctx, packageName, info, s3Dest, os.DirFS(fullDest), cfg.Alias); err != nil {
			return err
		}
	} else {
		infoJSON, err := json.Marshal(info)
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(fullDest, canonicalVersion+".info"), infoJSON, 0644); err != nil {
			return err
		}

		for _, alias := range cfg.Alias {
			if err := os.WriteFile(filepath.Join(fullDest, alias+".alias"), []byte(canonicalVersion), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func s3Copy(ctx context.Context, packageName string, info Info, s3Dest string, copyFrom fs.FS, aliases []string) error {
	bucketURL, err := url.Parse(s3Dest)
	if err != nil {
		return err
	}

	if bucketURL.Scheme != "s3" {
		return fmt.Errorf("bucket must be an s3:// url")
	}

	bucketName := bucketURL.Host
	if bucketName == "" {
		return fmt.Errorf("bucket must be an s3:// url")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	keyPrefix := strings.TrimPrefix(bucketURL.Path, "/")
	keyPrefix = path.Join(keyPrefix, packageName)

	err = fs.WalkDir(copyFrom, ".", func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		file, err := copyFrom.Open(srcPath)
		if err != nil {
			return err
		}

		defer file.Close()

		keyDest := path.Join(keyPrefix, srcPath)
		log.WithField(ctx, "key", keyDest).Info("uploading file")
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &keyDest,
			Body:   file,
			Metadata: map[string]string{
				S3MetadataCommitTime: info.Time.Format(time.RFC3339),
			},
		})

		if err != nil {
			return fmt.Errorf("failed to put object to bucket: '%s' key: '%s': %w", bucketName, keyDest, err)
		}

		return nil
	})

	for _, alias := range aliases {
		key := path.Join(keyPrefix, alias+".zip")
		_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Body:   strings.NewReader(info.Version),
			Bucket: &bucketName,
			Key:    &key,
			Metadata: map[string]string{
				S3MetadataAlias:      alias,
				S3MetadataCommitTime: info.Time.Format(time.RFC3339),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to put object to bucket: '%s' key: '%s': %w", bucketName, key, err)
		}
	}

	return err

}
