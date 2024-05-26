package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pentops/jsonapi/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/log.go/log"
	"go.opentelemetry.io/otel/trace/noop"

	glob "github.com/ryanuber/go-glob"
)

var DefaultRegistryAuths = []*source_j5pb.DockerRegistryAuth{{
	Registry: "ghcr.io/*",
	Auth: &source_j5pb.DockerRegistryAuth_Github_{
		Github: &source_j5pb.DockerRegistryAuth_Github{},
	},
}, {
	Registry: "*.dkr.ecr.*.amazonaws.com/*",
	Auth: &source_j5pb.DockerRegistryAuth_AwsEcs{
		AwsEcs: &source_j5pb.DockerRegistryAuth_AWSECS{},
	},
}}

type DockerWrapper struct {
	pulledImages map[string]bool
	client       *client.Client
	auth         []*source_j5pb.DockerRegistryAuth
}

func NewDockerWrapper(registryAuth []*source_j5pb.DockerRegistryAuth) (*DockerWrapper, error) {

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithTraceProvider(noop.NewTracerProvider()),
	)
	if err != nil {
		return nil, err
	}

	return &DockerWrapper{
		pulledImages: make(map[string]bool),
		client:       cli,
		auth:         registryAuth,
	}, nil
}

func (dw *DockerWrapper) Close() error {
	return dw.client.Close()
}

func (dw *DockerWrapper) Run(ctx context.Context, spec *source_j5pb.DockerSpec, input io.Reader, output io.Writer, errOutput io.Writer, envVars []string) error {
	if spec.Pull {
		// skip if pulled...
		if !dw.pulledImages[spec.Image] {
			// only pull once for all plugins
			dw.pulledImages[spec.Image] = true

			log.Info(ctx, "pulling image")
			if err := dw.Pull(ctx, spec); err != nil {
				log.WithError(ctx, err).Error("failed to pull image")
				return err
			}
			log.Info(ctx, "image pulled")
		}
	}

	resp, err := dw.client.ContainerCreate(ctx, &container.Config{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    true,
		OpenStdin:    true,

		Tty: false,

		Image:      spec.Image,
		Env:        envVars,
		Entrypoint: spec.Entrypoint,
		Cmd:        spec.Command,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}
	defer func() {
		if err := dw.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
			log.WithError(ctx, err).Error("failed to remove container")
		}
	}()

	hj, err := dw.client.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
		Logs:   true,
	})
	if err != nil {
		return err
	}

	defer hj.Close()

	log.Debug(ctx, "ContainerStart")

	if err := dw.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	chOut := make(chan error)
	go func() {
		_, err = stdcopy.StdCopy(output, errOutput, hj.Reader)
		chOut <- err
	}()

	if _, err := io.Copy(hj.Conn, input); err != nil {
		return err
	}
	if err := hj.CloseWrite(); err != nil {
		return err
	}

	if err := <-chOut; err != nil {
		return fmt.Errorf("output copy error: %w", err)
	}

	log.Debug(ctx, "ContainerWait")

	statusCh, errCh := dw.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case st := <-statusCh:
		log.WithField(ctx, "exitStatus", st).Info("container exited")
		if st.StatusCode != 0 {
			return fmt.Errorf("non-zero exit code: %d", st.StatusCode)
		}

	}

	log.Debug(ctx, "Builder Done")

	return nil
}

func (dw *DockerWrapper) Pull(ctx context.Context, spec *source_j5pb.DockerSpec) error {

	type basicAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	pullOptions := image.PullOptions{}

	var registryAuth *source_j5pb.DockerRegistryAuth
	for _, auth := range dw.auth {
		// If auth's registry pattern with * wildcards matches the spec's image, use it.
		if glob.Glob(auth.Registry, spec.Image) {
			registryAuth = auth
			log.WithField(ctx, "registry", auth.Registry).Debug("using auth")
			break
		}
	}
	if registryAuth == nil {
		log.WithField(ctx, "image", spec.Image).Debug("no registry auth matched")
	}

	if registryAuth != nil {
		pullOptions.PrivilegeFunc = func() (string, error) {
			var authConfig *basicAuth

			switch authType := registryAuth.Auth.(type) {
			case *source_j5pb.DockerRegistryAuth_Basic_:
				val := os.Getenv(authType.Basic.PasswordEnvVar)
				if val == "" {
					return "", fmt.Errorf("basic auth password (%s) not set", authType.Basic.PasswordEnvVar)
				}

				authConfig = &basicAuth{
					Username: authType.Basic.Username,
					Password: val,
				}

			case *source_j5pb.DockerRegistryAuth_Github_:
				envVar := authType.Github.TokenEnvVar
				if envVar == "" {
					envVar = "GITHUB_TOKEN"
				}
				val := os.Getenv(envVar)
				if val == "" {
					return "", fmt.Errorf("github token (%s) not set", envVar)
				}

				authConfig = &basicAuth{
					Username: "GITHUB",
					Password: val,
				}

			case *source_j5pb.DockerRegistryAuth_AwsEcs:

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

				authConfig = &basicAuth{
					Username: parts[0],
					Password: parts[1],
				}

			default:
				return "", fmt.Errorf("unknown auth type: %T", authType)
			}
			cred, _ := json.Marshal(authConfig)
			return base64.StdEncoding.EncodeToString(cred), nil
		}
	}

	reader, err := dw.client.ImagePull(ctx, spec.Image, pullOptions)
	if err != nil {
		// The ECS registry seems to return the 'wrong' status code for PrivilegeFunc errors.
		// This is a workaround.
		if strings.Contains(err.Error(), "no basic auth credentials") {
			token, err := pullOptions.PrivilegeFunc()
			if err != nil {
				return fmt.Errorf("image pull: %w", err)
			}
			pullOptions.PrivilegeFunc = nil
			pullOptions.RegistryAuth = token
			reader, err = dw.client.ImagePull(ctx, spec.Image, pullOptions)
			if err != nil {
				return fmt.Errorf("image pull: %w", err)
			}
		} else {
			return fmt.Errorf("image pull: %w", err)
		}
	}

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
