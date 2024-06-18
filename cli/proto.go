package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pentops/j5/builder/builder"
	"github.com/pentops/j5/builder/docker"
	"github.com/pentops/runner/commander"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func protoSet() *commander.CommandSet {
	protoGroup := commander.NewCommandSet()
	protoGroup.Add("build", commander.NewCommand(runProtoBuild))
	protoGroup.Add("request", commander.NewCommand(runProtoRequest))
	return protoGroup
}

func runProtoBuild(ctx context.Context, cfg struct {
	SourceConfig
	Out  string `flag:"out" default:"" description:"Output directory for generated files. Default discards output, use to test that the proto will build"`
	Pull bool   `flag:"pull" default:"false" description:"Pull images from registry, even if they already exist"`
}) error {

	src, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	var dest builder.FS

	if cfg.Out == "" {
		dest = NewDiscardFS()
	} else {
		dest, err = NewLocalFS(cfg.Out)
		if err != nil {
			return err
		}
	}

	dockerWrapper, err := docker.NewDockerWrapper(docker.DefaultRegistryAuths)
	if err != nil {
		return err
	}

	bb := builder.NewBuilder(dockerWrapper)

	if !cfg.Pull {
		for _, builder := range src.J5Config().ProtoBuilds {
			for _, plugin := range builder.Plugins {
				plugin.Docker.Pull = false
			}
		}
	}

	err = bb.BuildAll(ctx, src, dest)
	if err != nil {
		return err
	}

	fmt.Println("All plugins built successfully")

	return nil
}

func runProtoRequest(ctx context.Context, cfg struct {
	SourceConfig
	Command string `flag:"command" default:"" description:"Pipe the output to a builder command and print files"`
}) error {

	src, err := cfg.GetSource(ctx)
	if err != nil {
		return err
	}

	protoBuildRequest, err := src.ProtoCodeGeneratorRequest(ctx, ".")
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
