package builder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	dockerConfig "github.com/docker/cli/cli/config"
	"github.com/pentops/j5/gen/j5/config/v1/config_j5pb"
	"github.com/pentops/log.go/log"

	glob "github.com/ryanuber/go-glob"
)

var DefaultRegistryAuths = []*config_j5pb.DockerRegistryAuth{{
	Registry: "ghcr.io/*",
	Auth: &config_j5pb.DockerRegistryAuth_Github_{
		Github: &config_j5pb.DockerRegistryAuth_Github{},
	},
}, {
	Registry: "*.dkr.ecr.*.amazonaws.com/*",
	Auth: &config_j5pb.DockerRegistryAuth_AwsEcs{
		AwsEcs: &config_j5pb.DockerRegistryAuth_AWSECS{},
	},
}}

type dockerRun struct {
	lock        sync.Mutex
	containerID string
	client      *client.Client
	started     bool
}

func (dr *dockerRun) close(ctx context.Context) {
	dr.lock.Lock()
	defer dr.lock.Unlock()
	ctx = context.WithoutCancel(ctx)

	if dr.started {
		if err := dr.client.ContainerStop(ctx, dr.containerID, container.StopOptions{}); err != nil {
			log.WithError(ctx, err).Warn("failed to stop container")
			return
		}
		log.WithField(ctx, "container-id", dr.containerID).Debug("Container Stopped")
	}

	if remErr := dr.client.ContainerRemove(ctx, dr.containerID, container.RemoveOptions{}); remErr != nil {
		log.WithError(ctx, remErr).Warn("failed to remove container")
	}
}

func (dr *dockerRun) attach(ctx context.Context) (*types.HijackedResponse, error) {
	dr.lock.Lock()
	defer dr.lock.Unlock()
	hj, err := dr.client.ContainerAttach(ctx, dr.containerID, container.AttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
		Logs:   true,
	})
	if err != nil {
		log.WithError(ctx, err).Error("failed to attach to container")
		return nil, fmt.Errorf("container attach: %w", err)
	}
	return &hj, nil
}

func (dr *dockerRun) start(ctx context.Context) error {
	dr.lock.Lock()
	defer dr.lock.Unlock()

	if dr.started {
		return nil
	}

	if err := dr.client.ContainerStart(ctx, dr.containerID, container.StartOptions{}); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.WithError(ctx, err).Error("failed to start container")
		}
		return fmt.Errorf("container start: %w", err)
	}

	dr.started = true
	log.WithField(ctx, "container-id", dr.containerID).Debug("Container Started")
	return nil
}

func (dw *Runner) containerCreate(ctx context.Context, rc RunContext) (*dockerRun, error) {
	rcDocker := rc.Command.RunType.GetDocker()
	if rcDocker == nil {
		return nil, fmt.Errorf("run type is not docker")
	}
	resp, err := dw.client.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    true,
		OpenStdin:    true,

		Tty: false,

		Env:        rcDocker.Env,
		Image:      rcDocker.Image,
		Entrypoint: rcDocker.Entrypoint,
		Cmd:        rcDocker.Cmd,
	}, nil, nil, nil, "")
	if err != nil {
		log.WithError(ctx, err).Error("failed to start container")
		return nil, err
	}

	log.WithField(ctx, "container-id", resp.ID).Debug("Container Created")
	return &dockerRun{
		client:      dw.client,
		containerID: resp.ID,
	}, nil
}

func (dr *dockerRun) wait(ctx context.Context) error {
	dr.lock.Lock()
	statusCh, errCh := dr.client.ContainerWait(ctx, dr.containerID, container.WaitConditionNotRunning)
	dr.lock.Unlock()

	select {
	case err := <-errCh:
		if err != nil {
			log.WithError(ctx, err).Debug("Container Wait Error")
			return fmt.Errorf("container-wait error: %w", err)
		}
		return nil
	case st := <-statusCh:
		log.WithField(ctx, "status-code", st.StatusCode).Debug("Container Exit")
		if st.StatusCode != 0 {
			return fmt.Errorf("non-zero exit code: %d", st.StatusCode)
		}
		return nil
	}
}

func (dw *Runner) runDocker(ctx context.Context, rc RunContext) error {

	rcDocker := rc.Command.RunType.GetDocker()
	if rcDocker == nil {
		return fmt.Errorf("run type is not docker")
	}

	ctx = log.WithField(ctx, "image", rcDocker.Image)
	t0 := time.Now()
	log.WithField(ctx, "t0", time.Since(t0).String()).Debug("Pull If Needed")
	if err := dw.pullIfNeeded(ctx, rcDocker.Image); err != nil {
		log.WithError(ctx, err).Error("failed to pull image")
		return err
	}

	dr, err := dw.containerCreate(ctx, rc)
	if err != nil {
		return fmt.Errorf("container create: %w", err)
	}

	defer dr.close(ctx)

	hj, err := dr.attach(ctx)
	if err != nil {
		return err
	}

	defer hj.Close()

	log.WithField(ctx, "t0", time.Since(t0).String()).Debug("Container Start")

	if err := dr.start(ctx); err != nil {
		log.WithError(ctx, err).Debug("Container Start Error")
	}

	log.WithField(ctx, "t0", time.Since(t0).String()).Debug("Container Started")

	chOut := make(chan error)
	go func() {
		_, err = stdcopy.StdCopy(rc.StdOut, rc.StdErr, hj.Reader)
		chOut <- err
	}()

	if _, err := io.Copy(hj.Conn, rc.StdIn); err != nil {
		return err
	}
	if err := hj.CloseWrite(); err != nil {
		return err
	}

	if err := <-chOut; err != nil {
		return fmt.Errorf("output copy error: %w", err)
	}

	log.WithField(ctx, "t0", time.Since(t0).String()).Debug("Container Wait")

	if err := dr.wait(ctx); err != nil {
		return err
	}

	log.WithField(ctx, "t0", time.Since(t0).String()).Debug("ContainerDone")

	return nil
}

func (dw *Runner) markPull(img string) bool {
	dw.pullLock.Lock()
	defer dw.pullLock.Unlock()

	// skip if pulled...
	if dw.pulledImages[img] {
		return true
	}

	// only pull once for all plugins
	dw.pulledImages[img] = true
	return false
}

// dockerCliAuth uses the docker CLI config to get auth for a registry
func dockerCliAuth(img string) (string, error) {
	cfg := dockerConfig.LoadDefaultConfigFile(os.Stderr)
	registryEnd := strings.Index(img, "/")
	if registryEnd == -1 {
		return "", fmt.Errorf("invalid image: %s", img)
	}

	registry := img[:registryEnd]
	auth, err := cfg.GetAuthConfig(registry)
	if err != nil || auth.Password == "" {
		return "", fmt.Errorf("get auth config: %w", err)
	}

	return basicAuth(auth.Username, auth.Password)
}

// ghCliAuth uses the GitHub CLI to get auth for ghcr.io
func ghCliAuth(img string) (string, error) {
	output, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", fmt.Errorf("gh auth token: %w", err)
	}

	return basicAuth("GITHUB", string(output))
}

// basicAuth returns base64 encoded credentials for a docker registry
func basicAuth(username, password string) (string, error) {
	cred, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("marshal auth: %w", err)
	}

	return base64.StdEncoding.EncodeToString(cred), nil
}

func (dw *Runner) pullIfNeeded(ctx context.Context, img string) error {
	alreadyPulled := dw.markPull(img)
	if alreadyPulled {
		log.Debug(ctx, "image already pulled")
		return nil
	}

	images, err := dw.client.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", img)),
	})
	if err != nil {
		return fmt.Errorf("image list: %w", err)
	}
	if len(images) > 0 {
		log.Debug(ctx, "found images")
		return nil
	}

	pullOptions := image.PullOptions{}

	var registryAuth *config_j5pb.DockerRegistryAuth
	for _, auth := range dw.auth {
		// If auth's registry pattern with * wildcards matches the spec's image, use it.
		if glob.Glob(auth.Registry, img) {
			registryAuth = auth
			log.WithField(ctx, "registry", auth.Registry).Debug("using auth")
			break
		}
	}
	if registryAuth == nil {
		log.WithField(ctx, "image", img).Debug("no registry auth matched")
	}

	if registryAuth != nil {
		pullOptions.PrivilegeFunc = func(ctx context.Context) (string, error) {
			switch authType := registryAuth.Auth.(type) {
			case *config_j5pb.DockerRegistryAuth_Basic_:
				val := os.Getenv(authType.Basic.PasswordEnvVar)
				if val == "" {
					return "", fmt.Errorf("basic auth password (%s) not set", authType.Basic.PasswordEnvVar)
				}

				return basicAuth(authType.Basic.Username, val)

			case *config_j5pb.DockerRegistryAuth_Github_:
				envVar := authType.Github.TokenEnvVar
				if envVar == "" {
					envVar = "GITHUB_TOKEN"
				}

				// Prefer env var over docker config
				val := os.Getenv(envVar)
				if val != "" {
					log.Debug(ctx, "using token from env")
					return basicAuth("GITHUB", val)
				}

				val, _ = dockerCliAuth(img)
				if val != "" {
					log.Debug(ctx, "using token from docker cli")
					return val, nil
				}

				val, _ = ghCliAuth(img)
				if val != "" {
					log.Debug(ctx, "using token from gh cli")
					return val, nil
				}

				return "", fmt.Errorf("github token (%s) not set, or docker/gh not logged in to github", envVar)

			case *config_j5pb.DockerRegistryAuth_AwsEcs:
				// TODO: This is a little too magic.
				awsConfig, err := config.LoadDefaultConfig(ctx)
				if err != nil {
					return "", fmt.Errorf("failed to load configuration: %w", err)
				}

				ecrClient := ecr.NewFromConfig(awsConfig)
				resp, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
				if err != nil {
					return "", fmt.Errorf("failed to get authorization token: %w", err)
				}

				if len(resp.AuthorizationData) == 0 {
					return "", fmt.Errorf("no authorization data returned")
				}

				authData, err := base64.StdEncoding.DecodeString(*resp.AuthorizationData[0].AuthorizationToken)
				if err != nil {
					return "", fmt.Errorf("failed to decode authorization token: %w", err)
				}

				parts := strings.SplitN(string(authData), ":", 2)
				if len(parts) != 2 {
					return "", fmt.Errorf("invalid authorization token")
				}

				return basicAuth(parts[0], parts[1])

			default:
				return "", fmt.Errorf("unknown auth type: %T", authType)
			}
		}
	}

	reader, err := dw.client.ImagePull(ctx, img, pullOptions)
	if err != nil {
		// The ECS registry & ghcr seems to return the 'wrong' status code for PrivilegeFunc errors.
		// This is a workaround.
		if strings.Contains(err.Error(), "no basic auth credentials") || strings.Contains(err.Error(), "unauthorized") {
			token, err := pullOptions.PrivilegeFunc(ctx)
			if err != nil {
				return fmt.Errorf("image pull: %w", err)
			}
			pullOptions.PrivilegeFunc = nil
			pullOptions.RegistryAuth = token
			reader, err = dw.client.ImagePull(ctx, img, pullOptions)
			if err != nil {
				return fmt.Errorf("image pull: %w", err)
			}
		} else {
			return fmt.Errorf("image pull: %w", err)
		}
	}

	log.Debug(ctx, "wait for imagePull")

	// cli.ImagePull is asynchronous.
	// The reader needs to be read completely for the pull operation to complete.
	// If stdout is not required, consider using io.Discard instead of os.Stdout.
	_, err = io.Copy(os.Stdout, reader)
	reader.Close()
	if err != nil {
		return err
	}
	return nil
}
