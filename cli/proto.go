package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pentops/jsonapi/builder/builder"
	"github.com/pentops/jsonapi/builder/docker"
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func protoSet() *commander.CommandSet {
	protoGroup := commander.NewCommandSet()
	protoGroup.Add("build", commander.NewCommand(runProtoBuild))
	protoGroup.Add("request", commander.NewCommand(runProtoRequest))
	protoGroup.Add("test", commander.NewCommand(runTestBuild))
	return protoGroup
}

func runTestBuild(ctx context.Context, cfg struct {
	SourceConfig
	Pull     bool     `flag:"pull" default:"false" description:"Pull images from registry, even if they already exist"`
	Output   []string `flag:"output" default:"" description:"Not a dry run - actually output the built files (e.g. for go mod replace). "`
	Builders []string `flag:",remaining" description:"Builders to run - 'j5', 'proto/$label' 'proto/$label/$plugin'"`
}) error {

	remote := builder.NewRawUploader()
	if len(cfg.Output) > 0 {
		for _, output := range cfg.Output {
			parts := strings.SplitN(output, "=", 2)
			if len(parts) != 2 {
				if len(cfg.Output) != 1 {
					return fmt.Errorf("invalid output: %s, specify either a single dir, or key=val pairs", output)
				}
				remote.J5Output = output
			}
			if strings.HasPrefix(parts[0], "j5") {
				remote.J5Output = parts[1]
			} else if strings.HasPrefix(parts[0], "proto/") {
				key := strings.TrimPrefix(parts[0], "proto/")
				remote.ProtoGenOutputs[key] = parts[1]
			}
		}
	}

	source, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	if !cfg.Pull {
		for _, builder := range source.J5Config().ProtoBuilds {
			for _, plugin := range builder.Plugins {
				plugin.Docker.Pull = false
			}
		}
	}

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	bb := builder.NewBuilder(dockerWrapper, remote)

	err = bb.BuildAll(ctx, source, cfg.Builders...)
	if err != nil {
		return err
	}

	fmt.Println("All plugins built successfully")
	return nil
}

func runProtoRequest(ctx context.Context, cfg struct {
	SourceConfig
	PackagePrefix string `flag:"package-prefix" env:"PACKAGE_PREFIX" default:""`
	Command       string `flag:"command" default:"" description:"Pipe the output to a builder command and print files"`
}) error {

	src, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	protoBuildRequest, err := src.ProtoCodeGeneratorRequest(ctx)
	if err != nil {
		return err
	}

	protoBuildRequestBytes, err := proto.Marshal(protoBuildRequest)
	if err != nil {
		return err
	}

	if cfg.Command == "" {
		_, err = os.Stdout.Write(protoBuildRequestBytes)
		return err
	}

	cmd := exec.CommandContext(ctx, cfg.Command)

	inPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer inPipe.Close()

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer outPipe.Close()

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	outErr := make(chan error)
	outBuf := &bytes.Buffer{}
	go func() {
		_, err := io.Copy(outBuf, outPipe)
		outErr <- err
	}()

	if _, err := inPipe.Write(protoBuildRequestBytes); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	outPipe.Close()

	if err := <-outErr; err != nil {
		return err
	}

	res := pluginpb.CodeGeneratorResponse{}
	if err := proto.Unmarshal(outBuf.Bytes(), &res); err != nil {
		return err
	}

	for _, file := range res.File {
		fmt.Println(file.GetName())
	}

	return nil
}

func runProtoBuild(ctx context.Context, cfg struct {
	SourceConfig
	Dest          string `flag:"dest" default:"" description:"Destination directory for generated files"`
	PackagePrefix string `flag:"package-prefix" env:"PACKAGE_PREFIX" default:""`
	Pull          bool   `flag:"pull" default:"false" description:"Pull images from registry, even if they already exist"`
}) error {

	src, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	dest, err := NewLocalFS(cfg.Dest)
	if err != nil {
		return err
	}
	remote := builder.NewFSUploader(dest)

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	bb := builder.NewBuilder(dockerWrapper, remote)

	if !cfg.Pull {
		for _, builder := range src.J5Config().ProtoBuilds {
			for _, plugin := range builder.Plugins {
				plugin.Docker.Pull = false
			}
		}
	}

	fmt.Println("All plugins built successfully")

	return bb.BuildAll(ctx, src)
}
