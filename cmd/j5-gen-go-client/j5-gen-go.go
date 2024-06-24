package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pentops/j5/gen/j5/plugin/v1/plugin_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/gogen"
	"github.com/pentops/j5/schema/j5reflect"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := do(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func do() error {

	req := &plugin_j5pb.CodeGenerationRequest{}

	inBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	if err := proto.Unmarshal(inBytes, req); err != nil {
		return err
	}

	if req.Packages == nil || len(req.Packages) < 1 {
		return fmt.Errorf("no packages to generate")
	}

	if req.Config == nil {
		req.Config = &plugin_j5pb.Config{}
	}

	options := gogen.Options{
		TrimPackagePrefix:   req.Config.TrimPackagePrefix,
		FilterPackagePrefix: req.Config.FilterPackagePrefix,
		GoPackagePrefix:     req.Options["go_package_prefix"],
	}

	output := &protoFileWriter{
		resp: &plugin_j5pb.CodeGenerationResponse{},
	}

	reflect, err := j5reflect.APIFromDesc(&schema_j5pb.API{
		Packages: req.Packages,
	})
	if err != nil {
		return err
	}

	for _, j5Package := range reflect.Packages { // Only generate packages within the prefix.
		if options.FilterPackagePrefix != "" {
			if !strings.HasPrefix(j5Package.Name, options.FilterPackagePrefix) {
				continue
			}
		}

		if err := gogen.WriteGoCode(j5Package, output, options); err != nil {
			return err
		}
	}

	outBytes, err := proto.Marshal(output.resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	if _, err := os.Stdout.Write(outBytes); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

type protoFileWriter struct {
	resp *plugin_j5pb.CodeGenerationResponse
}

func (w *protoFileWriter) WriteFile(name string, data []byte) error {
	dataStr := string(data)
	w.resp.Files = append(w.resp.Files, &plugin_j5pb.File{
		Name:    name,
		Content: dataStr,
	})
	return nil
}
