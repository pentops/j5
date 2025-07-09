package github

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"buf.build/go/protoyaml"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/pentops/envconf.go/envconf"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/go-github/v58/github"
)

type Client struct {
	repositories RepositoriesService
}

type RepositoriesService interface {
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	ListByOrg(context.Context, string, *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	GetArchiveLink(ctx context.Context, owner string, repo string, archiveFormat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error)
	GetCommit(ctx context.Context, owner string, repo string, ref string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
	ListBranchesHeadCommit(ctx context.Context, owner string, repo string, sha string) ([]*github.BranchCommit, *github.Response, error)
}

type Commit struct {
	Owner string
	Repo  string
	Sha   string
}

func NewEnvClient(ctx context.Context) (*Client, error) {

	config := struct {
		// Method 1
		GithubToken string `env:"GITHUB_TOKEN" default:""`

		// Method 2
		GithubPrivateKey     string `env:"GH_PRIVATE_KEY" default:""`
		GithubAppID          int64  `env:"GH_APP_ID" default:"0"`
		GithubInstallationID int64  `env:"GH_INSTALLATION_ID" default:"0"`
	}{}

	if err := envconf.Parse(&config); err != nil {
		return nil, err
	}

	var err error
	var client *Client

	if config.GithubPrivateKey != "" {

		if config.GithubAppID == 0 || config.GithubInstallationID == 0 {
			return nil, fmt.Errorf("no github app id or installation id")
		}
		tr := http.DefaultTransport
		privateKey, err := base64.StdEncoding.DecodeString(config.GithubPrivateKey)
		if err != nil {
			return nil, err
		}

		itr, err := ghinstallation.New(tr, config.GithubAppID, int64(config.GithubInstallationID), privateKey)
		if err != nil {
			return nil, err
		}

		client, err = NewClient(&http.Client{Transport: itr})
		if err != nil {
			return nil, err
		}

	} else if config.GithubToken != "" {

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GithubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client, err = NewClient(tc)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no valid github config in environment")
	}

	return client, nil
}

func NewClient(tc *http.Client) (*Client, error) {
	ghcl := github.NewClient(tc)
	cl := &Client{
		repositories: ghcl.Repositories,
	}
	return cl, nil
}
func (cl Client) PullConfig(ctx context.Context, ref *Commit, into proto.Message, tryPaths []string) error {

	opts := &github.RepositoryContentGetOptions{
		Ref: ref.Sha,
	}
	for _, path := range tryPaths {

		file, _, err := cl.repositories.DownloadContents(ctx, ref.Owner, ref.Repo, path, opts)
		if err != nil {
			errStr := err.Error()
			if strings.HasPrefix(errStr, "no file named") {
				continue
			}

			return err
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return fmt.Errorf("reading bytes: %s", err)
		}

		if err := protoyaml.Unmarshal(data, into); err != nil {
			return fmt.Errorf("unmarshalling yaml: %s", err)
		}

		return nil
	}

	return fmt.Errorf("no config found")
}

func (cl Client) GetCommit(ctx context.Context, ref *Commit) (*source_j5pb.CommitInfo, error) {

	commit, _, err := cl.repositories.GetCommit(ctx, ref.Owner, ref.Repo, ref.Sha, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	ts := commit.GetCommit().GetCommitter().GetDate()
	info := &source_j5pb.CommitInfo{
		Hash:  commit.GetSHA(),
		Time:  timestamppb.New(ts.Time),
		Owner: ref.Owner,
		Repo:  ref.Repo,
	}

	heads, _, err := cl.repositories.ListBranchesHeadCommit(ctx, ref.Owner, ref.Repo, info.Hash)

	if err != nil {
		return nil, err
	}

	for _, head := range heads {
		info.Aliases = append(info.Aliases, fmt.Sprintf("refs/heads/%s", *head.Name))
	}

	return info, nil
}

func (cl Client) GetContent(ctx context.Context, ref *Commit, destDir string) error {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref.Sha,
	}

	linkURL, _, err := cl.repositories.GetArchiveLink(ctx, ref.Owner, ref.Repo, github.Zipball, opts, 5)
	if err != nil {
		return err
	}

	log.WithField(ctx, "url", linkURL.String()).Debug("downloading")

	getRes, err := http.DefaultClient.Get(linkURL.String())
	if err != nil {
		return err
	}

	if getRes.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d", getRes.StatusCode)
	}

	defer getRes.Body.Close()

	// TODO: Use a disk.
	zipBody, err := io.ReadAll(getRes.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBody), int64(len(zipBody)))
	if err != nil {
		return err
	}

	prefix := ""

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		if prefix == "" {
			parts := strings.Split(file.Name, "/")
			prefix = parts[0]
		}

		if !strings.HasPrefix(file.Name, prefix) {
			return fmt.Errorf("invalid file name %q", file.Name)
		}
		destFile := filepath.Join(destDir, file.Name[len(prefix):])
		destDir := filepath.Dir(destFile)

		if err := func() error {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return err
			}

			dest, err := os.Create(destFile)
			if err != nil {
				return err
			}
			defer dest.Close()

			archiveFile, err := file.Open()
			if err != nil {
				return err
			}

			defer archiveFile.Close()

			if _, err := io.Copy(dest, archiveFile); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}

	}

	return nil
}
