package psrc

import (
	"fmt"
	"strings"

	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5s/j5convert"
	"github.com/pentops/j5/internal/j5s/protobuild/errset"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func buildSummaryFromReflect(res protoreflect.FileDescriptor, errs *errset.ErrCollector) (*j5convert.FileSummary, error) {
	return SummaryFromDescriptor(protodesc.ToFileDescriptorProto(res), errs)
}

func SummaryFromDescriptor(res *descriptorpb.FileDescriptorProto, errs *errset.ErrCollector) (*j5convert.FileSummary, error) {
	filename := res.GetName()
	exports := map[string]*j5convert.TypeRef{}

	for _, msg := range res.MessageType {
		typeRef := &j5convert.TypeRef{
			Name:    msg.GetName(),
			File:    filename,
			Package: res.GetPackage(),
		}
		exports[msg.GetName()] = typeRef
		options := proto.GetExtension(msg.Options, ext_j5pb.E_Message).(*ext_j5pb.MessageOptions)
		if options != nil {
			switch et := options.Type.(type) {
			case *ext_j5pb.MessageOptions_Oneof:
				typeRef.Oneof = &j5convert.OneofRef{}
			case *ext_j5pb.MessageOptions_Polymorph:
				typeRef.Polymorph = &j5convert.PolymorphRef{
					Members: et.Polymorph.Members,
				}
			case *ext_j5pb.MessageOptions_Object:
				typeRef.Object = &j5convert.ObjectRef{}
			}
		} else {
			typeRef.Object = &j5convert.ObjectRef{}
		}
	}

	for idx, en := range res.EnumType {
		built, err := buildEnumRef(res, int32(idx), en, errs)
		if err != nil {
			return nil, err
		}
		exports[en.GetName()] = &j5convert.TypeRef{
			Name:    en.GetName(),
			Package: res.GetPackage(),
			File:    filename,
			Enum:    built,
		}
	}

	packageExt := proto.GetExtension(res.Options, ext_j5pb.E_Package).(*ext_j5pb.PackageOptions)
	if packageExt != nil {
		for _, stringFormat := range packageExt.StringFormats {
			exports[stringFormat.Name] = &j5convert.TypeRef{
				Name:    stringFormat.Name,
				Package: res.GetPackage(),
				File:    filename,

				StringFormat: &schema_j5pb.StringFormat{
					Regex:       stringFormat.Regex,
					Name:        stringFormat.Name,
					Description: stringFormat.Description,
				},
			}
		}
	}

	return &j5convert.FileSummary{
		SourceFilename:   filename,
		Exports:          exports,
		FileDependencies: res.Dependency,
		ProducesFiles:    []string{filename},
		Package:          res.GetPackage(),

		// No type dependencies for proto files, all deps come from the files.
		TypeDependencies: nil,
	}, nil
}

func buildEnumRef(file *descriptorpb.FileDescriptorProto, idx int32, enumDescriptor *descriptorpb.EnumDescriptorProto, errs *errset.ErrCollector) (*j5convert.EnumRef, error) {
	for idx, value := range enumDescriptor.Value {
		if value.Number == nil {
			return nil, fmt.Errorf("enum value[%d] does not have a number", idx)
		}
		if value.Name == nil {
			return nil, fmt.Errorf("enum value[%d] does not have a name", idx)
		}
	}

	suffix := "UNSPECIFIED"
	var trimPrefix string

	if len(enumDescriptor.Value) < 1 {
		return nil, fmt.Errorf("enum has no values")
	} else {
		if *enumDescriptor.Value[0].Number != 0 {
			return nil, fmt.Errorf("enum does not have a value 0")
		}
		unspecifiedVal := *enumDescriptor.Value[0].Name

		if strings.HasSuffix(unspecifiedVal, suffix) {
			trimPrefix = strings.TrimSuffix(unspecifiedVal, suffix)
		} else {
			errs.WarnProtoDesc(file, []int32{5, idx}, fmt.Errorf("enum value 0 should have suffix %s", suffix))
			// proceed without prefix.
		}
	}

	ref := &j5convert.EnumRef{
		Prefix: trimPrefix,
		ValMap: map[string]int32{},
	}

	for _, value := range enumDescriptor.Value {
		ref.ValMap[value.GetName()] = value.GetNumber()
	}
	return ref, nil
}
