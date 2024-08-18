package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/pentops/j5/codec"
	"github.com/pentops/j5/internal/j5lang"
	"github.com/pentops/j5/internal/protobuild"
	"github.com/pentops/j5/internal/protobuild/protoprint"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/descriptorpb"
)

func runGenProto(ctx context.Context, cfg struct {
	Src string `flag:"src" description:"Source directory"`
	Dst string `flag:"dst" description:"Destination directory"`
}) error {
	outRoot, err := NewLocalFS(cfg.Dst)
	if err != nil {
		return err
	}

	desc := &descriptorpb.FileDescriptorSet{
		File: make([]*descriptorpb.FileDescriptorProto, 0),
	}

	fsRoot := os.DirFS(cfg.Src)

	if err := fs.WalkDir(fsRoot, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".j5s") {
			return nil
		}

		protoFile, err := fsRoot.Open(path)
		if err != nil {
			return err
		}
		defer protoFile.Close()

		srcBytes, err := io.ReadAll(protoFile)
		if err != nil {
			return err
		}

		parsed, err := j5lang.ParseFile(string(srcBytes))
		if err != nil {
			return fmt.Errorf("file %s: %w", path, err)
		}

		j5Desc, err := j5lang.ConvertFile(parsed)
		if err != nil {
			return err
		}

		jj, err := codec.NewCodec().ProtoToJSON(j5Desc.ProtoReflect())
		if err != nil {
			return err
		}
		buf := new(bytes.Buffer)
		if err := json.Indent(buf, jj, "", "  "); err != nil {
			return err
		}
		fmt.Println(buf.String())

		protoFilename := strings.TrimSuffix(path, ".j5") + ".proto"
		j5Desc.Path = protoFilename
		protoDesc, err := protobuild.BuildFile(j5Desc)
		if err != nil {
			return err
		}

		fmt.Println(protojson.Format(protoDesc))

		desc.File = append(desc.File, protoDesc)

		return nil
	}); err != nil {
		return err
	}

	err = protoprint.PrintProtoFiles(ctx, outRoot, desc, protoprint.Options{})
	if err != nil {
		return fmt.Errorf("printing: %w", err)
	}

	return nil
}
